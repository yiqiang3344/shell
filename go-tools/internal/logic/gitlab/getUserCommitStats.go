package gitlab

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gtype"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/schollz/progressbar/v3"
	"github.com/xanzy/go-gitlab"
	"github.com/xuri/excelize/v2"
	"go-tools/internal/utility"
	"math"
	"strings"
	"sync"
	"time"
)

type userInfo struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at"`
}

type projectInfo struct {
	ID                int        `json:"id"`
	Description       string     `json:"description"`
	Name              string     `json:"name"`
	Path              string     `json:"path"`
	PathWithNamespace string     `json:"path_with_namespace"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
}

type UserCommitStats struct {
	UserInfo           userInfo
	Projects           *gmap.IntAnyMap
	TotalCommitStats   *gitlab.CommitStats
	ProjectCommitStats *gmap.IntAnyMap
	ProjectCommit      *gmap.IntAnyMap
}

// GetUserCommitStats 获取用户提交统计信息
func (s *sGitlab) GetUserCommitStats(ctx context.Context, parse *gcmd.Parser) {
	var (
		err       error
		userInput []string
		userMap   = gmap.NewStrAnyMap(true)
		startTime *gtime.Time
		endTime   *gtime.Time
		sTime     = gtime.Now()
		bar       *progressbar.ProgressBar
	)
	//1.检查用户列表参数
	usernamesStr := utility.GetArgString(ctx, parse, "gitlab.commitStats.usernames", "usernames")
	for {
		userInput = strings.Split(usernamesStr, ",")
		if strings.Trim(usernamesStr, "") == "" || len(userInput) == 0 {
			usernamesStr = utility.Scanf("请输入要统计的用户名列表(逗号分割):")
			continue
		}
		break
	}

	//2.检查开始时间参数
	startTimeStr := utility.GetArgString(ctx, parse, "gitlab.commitStats.startTime", "startTime")
	for {
		startTime = gtime.NewFromStr(startTimeStr)
		if strings.Trim(startTimeStr, "") == "" || startTime == nil {
			startTimeStr = utility.Scanf("请输入要统计开始时间(yyyy-mm-dd HH:ii:ss):")
			continue
		}
		break
	}

	//3.检查截止时间参数
	endTimeStr := utility.GetArgString(ctx, parse, "gitlab.commitStats.endTime", "endTime")
	for {
		endTime = gtime.NewFromStr(endTimeStr)
		if strings.Trim(endTimeStr, "") == "" || endTime == nil {
			endTimeStr = utility.Scanf("请输入要统计截止时间(yyyy-mm-dd HH:ii:ss):")
			continue
		}
		break
	}

	err = s.initClient(ctx, parse)
	if err != nil {
		utility.Errorf("客户端初始化失败:%+v\n", err.Error())
		return
	}

	//获取用户信息
	allUserMap := s.getAllUserMap(ctx, parse, &gitlab.ListUsersOptions{})
	utility.Debugf(ctx, parse, "用户总数:%d\n", len(allUserMap))

	//遍历用户，构建用户map
	for _, v := range userInput {
		if _, ok := allUserMap[v]; !ok {
			utility.Debugf(ctx, parse, "用户[%s]不存在\n", v)
			continue
		}
		uInfo := userInfo{}
		gconv.ConvertWithRefer(allUserMap[v], &uInfo)
		userMap.Set(v, &UserCommitStats{
			UserInfo:           uInfo,
			Projects:           gmap.NewIntAnyMap(true),
			ProjectCommitStats: gmap.NewIntAnyMap(true),
			TotalCommitStats:   &gitlab.CommitStats{},
			ProjectCommit:      gmap.NewIntAnyMap(true),
		})
	}

	// 遍历所有项目，查询用户加入的项目列表
	//search := "hutta-web"
	allProjectMap := s.getAllProjectMap(ctx, parse, &gitlab.ListProjectsOptions{
		LastActivityAfter:  &startTime.Time,
		LastActivityBefore: &endTime.Time,
		//Search:             &search,
	})
	utility.Debugf(ctx, parse, "仓库总数:%d\n", len(allProjectMap))

	// 并发遍历项目，获取用户和项目的关联关系
	wg := sync.WaitGroup{}
	wg.Add(len(allProjectMap))
	if !utility.IsDebug(ctx, parse) {
		bar = progressbar.Default(int64(len(allProjectMap)), "查询用户项目信息")
	}
	for k, v := range allProjectMap {
		go func(pid int, project *gitlab.Project) {
			defer func() {
				if !utility.IsDebug(ctx, parse) {
					bar.Add(1)
				}
				wg.Done()
			}()
			//控制并发数，按pid的百位数来，最大不超过4秒
			waitT := gconv.Int64(math.Min(gconv.Float64(pid/100), 4))
			time.Sleep(time.Duration(waitT) * time.Second)

			projectUserMap := s.getProjectUserMap(ctx, parse, project.ID)
			for _, u := range userMap.Map() {
				if _, ok := projectUserMap[u.(*UserCommitStats).UserInfo.Username]; ok {
					pInfo := projectInfo{}
					gconv.ConvertWithRefer(project, &pInfo)
					u.(*UserCommitStats).Projects.Set(project.ID, pInfo)
				}
			}
		}(k, v)
	}
	wg.Wait()
	if !utility.IsDebug(ctx, parse) {
		bar.Finish()
	}

	withStats, withAllCmt := true, true
	barMax := gtype.NewInt(0)
	if !utility.IsDebug(ctx, parse) {
		//先预设一个总数
		bar = progressbar.Default(1000, "统计提交信息")
	}
	wgU := sync.WaitGroup{}
	wgU.Add(userMap.Size())
	for _, uc := range userMap.Map() {
		//并发处理用户
		go func(uc *UserCommitStats) {
			defer func() {
				wgU.Done()
			}()
			//并发遍历项目，获取提交信息
			wg := sync.WaitGroup{}
			wg.Add(uc.Projects.Size())
			if !utility.IsDebug(ctx, parse) {
				barMax.Add(uc.Projects.Size())
				bar.ChangeMax(barMax.Val())
			}
			for _, p := range uc.Projects.Map() {
				go func(uc1 *UserCommitStats, project projectInfo) {
					defer func() {
						if !utility.IsDebug(ctx, parse) {
							bar.Add(1)
						}
						wg.Done()
					}()
					commits := s.getProjectCommits(ctx, parse, project.ID, &gitlab.ListCommitsOptions{
						Since:     &startTime.Time,
						Until:     &endTime.Time,
						WithStats: &withStats,
						All:       &withAllCmt,
					})
					//遍历提交信息，获取代码行数
					for _, cmt := range commits {
						// 过滤，按提交者的邮箱、名称和gitlab用户的邮箱、用户名、全名来比较
						if cmt.CommitterEmail != uc1.UserInfo.Email && cmt.CommitterName != uc1.UserInfo.Username && cmt.CommitterName != uc1.UserInfo.Name {
							continue
						}
						//记录项目提交明细
						uc1.ProjectCommit.LockFunc(func(m map[int]interface{}) {
							if c, ok := m[project.ID]; ok {
								c = append(c.([]*gitlab.Commit), cmt)
								m[project.ID] = c
							} else {
								m[project.ID] = []*gitlab.Commit{cmt}
							}
						})
						//统计用户每项目的代码行数
						uc1.ProjectCommitStats.LockFunc(func(m map[int]interface{}) {
							if c, ok := m[project.ID]; ok {
								c.(*gitlab.CommitStats).Total += cmt.Stats.Total
								c.(*gitlab.CommitStats).Deletions += cmt.Stats.Deletions
								c.(*gitlab.CommitStats).Additions += cmt.Stats.Additions
							} else {
								m[project.ID] = &gitlab.CommitStats{
									Total:     cmt.Stats.Total,
									Deletions: cmt.Stats.Deletions,
									Additions: cmt.Stats.Additions,
								}
							}
						})

						//统计用户总代码行数
						uc1.TotalCommitStats.Total += cmt.Stats.Total
						uc1.TotalCommitStats.Deletions += cmt.Stats.Deletions
						uc1.TotalCommitStats.Additions += cmt.Stats.Additions
					}
				}(uc, p.(projectInfo))
			}
			wg.Wait()
		}(uc.(*UserCommitStats))
	}
	wgU.Wait()
	if !utility.IsDebug(ctx, parse) {
		bar.Finish()
	}

	//输出excel
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			utility.Errorf("excel关闭异常:%v", err)
		}
	}()
	// 创建一个工作表
	excelFilepath := fmt.Sprintf("%s/用户提交统计%s-%s.xlsx", gfile.Pwd(), startTime.Format("YmdHis"), endTime.Format("YmdHis"))
	sheet1Name := "用户维度"
	sheet1Head := []string{"姓名", "增加数", "删除数", "总数"}
	sheet2Name := "用户项目维度"
	sheet2Head := []string{"姓名", "项目", "增加数", "删除数", "总数"}
	sheet3Name := "用户项目提交明细"
	sheet3Head := []string{"姓名", "项目", "提交时间", "提交标识", "提交备注", "增加数", "删除数", "总数"}
	sheet1Index, err := f.NewSheet(sheet1Name)
	if err != nil {
		utility.Errorf("sheet[%s]创建异常:%v", sheet1Name, err)
		return
	}
	_, err = f.NewSheet(sheet2Name)
	if err != nil {
		utility.Errorf("sheet[%s]创建异常:%v", sheet2Name, err)
		return
	}
	_, err = f.NewSheet(sheet3Name)
	if err != nil {
		utility.Errorf("sheet[%s]创建异常:%v", sheet3Name, err)
		return
	}
	// 设置头部
	for k, v := range sheet1Head {
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(k+1)+gconv.String(1), v)
	}
	for k, v := range sheet2Head {
		f.SetCellValue(sheet2Name, utility.ConvertNumToChar(k+1)+gconv.String(1), v)
	}
	for k, v := range sheet3Head {
		f.SetCellValue(sheet3Name, utility.ConvertNumToChar(k+1)+gconv.String(1), v)
	}
	// 设置值
	n, m, x := 2, 2, 2
	sortUsers := garray.NewStrArrayFrom(userMap.Keys()).Sort()
	for _, u := range sortUsers.Slice() {
		v := userMap.Get(u).(*UserCommitStats)
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(1)+gconv.String(n), v.UserInfo.Name)
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(2)+gconv.String(n), v.TotalCommitStats.Additions)
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(3)+gconv.String(n), v.TotalCommitStats.Deletions)
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(4)+gconv.String(n), v.TotalCommitStats.Total)
		sortPids := garray.NewIntArrayFrom(v.ProjectCommitStats.Keys()).Sort()
		for _, pid := range sortPids.Slice() {
			v1 := v.ProjectCommitStats.Get(pid).(*gitlab.CommitStats)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(1)+gconv.String(m), v.UserInfo.Name)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(2)+gconv.String(m), v.Projects.Get(pid).(projectInfo).PathWithNamespace)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(3)+gconv.String(m), v1.Additions)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(4)+gconv.String(m), v1.Deletions)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(5)+gconv.String(m), v1.Total)

			sortCmts := v.ProjectCommit.Get(pid).([]*gitlab.Commit)
			for _, c := range sortCmts {
				f.SetCellValue(sheet3Name, utility.ConvertNumToChar(1)+gconv.String(x), v.UserInfo.Name)
				f.SetCellValue(sheet3Name, utility.ConvertNumToChar(2)+gconv.String(x), v.Projects.Get(pid).(projectInfo).PathWithNamespace)
				f.SetCellValue(sheet3Name, utility.ConvertNumToChar(3)+gconv.String(x), c.CommittedDate.Format("2006-01-02 15:04:05"))
				f.SetCellValue(sheet3Name, utility.ConvertNumToChar(4)+gconv.String(x), c.ShortID)
				f.SetCellValue(sheet3Name, utility.ConvertNumToChar(5)+gconv.String(x), c.Title)
				f.SetCellValue(sheet3Name, utility.ConvertNumToChar(6)+gconv.String(x), c.Stats.Additions)
				f.SetCellValue(sheet3Name, utility.ConvertNumToChar(7)+gconv.String(x), c.Stats.Deletions)
				f.SetCellValue(sheet3Name, utility.ConvertNumToChar(8)+gconv.String(x), c.Stats.Total)
				x++
			}
			m++
		}
		n++
	}
	// 设置工作簿的默认工作表
	f.SetActiveSheet(sheet1Index)
	f.DeleteSheet("Sheet1")
	// 根据指定路径保存文件
	if err := f.SaveAs(excelFilepath); err != nil {
		utility.Errorf("excel保存异常:%v", err)
	}
	fmt.Printf("excel文件生成完毕，地址:%s\n", excelFilepath)
	fmt.Printf("总耗时:%s\n", utility.FormatDuration(gtime.Now().Sub(sTime)))
	return
}

func (s *sGitlab) getAllProjectMap(ctx context.Context, parse *gcmd.Parser, options *gitlab.ListProjectsOptions) (projectMap map[int]*gitlab.Project) {
	var bar *progressbar.ProgressBar
	projectMap = make(map[int]*gitlab.Project)
	gmap := gmap.New(true)
	defer func() {
		for k, v := range gmap.Map() {
			projectMap[k.(int)] = v.(*gitlab.Project)
		}
	}()
	page := 1
	perPage := 100
	searchSimple := true
	utility.Debugf(ctx, parse, "开始查询所有仓库项目信息\n")
	//先查询一次，获取总页数，如果大于2页，则并发读取
	options.ListOptions.Page = page
	options.ListOptions.PerPage = perPage
	options.Simple = &searchSimple
	firstRet, firstRes, err := s.gitClient.Projects.ListProjects(options)
	if err != nil {
		utility.Errorf("查询项目仓库信息失败:%+v\n", err.Error())
		return
	}
	for _, v := range firstRet {
		gmap.Set(v.ID, v)
	}
	utility.Debugf(ctx, parse, "查询项目仓库信息进度:%d/%d\n", page, firstRes.TotalPages)
	if firstRes.TotalPages <= 1 {
		return
	}
	page++

	if !utility.IsDebug(ctx, parse) {
		bar = progressbar.Default(int64(firstRes.TotalItems), "查询项目仓库信息")
		bar.Add(gmap.Size())
	}
	wg := sync.WaitGroup{}
	for ; page <= firstRes.TotalPages; page++ {
		wg.Add(1)
		optionsCopy := *options
		go func(page int, options *gitlab.ListProjectsOptions) {
			defer func() {
				wg.Done()
			}()
			options.ListOptions.Page = page
			ret, res, err1 := s.gitClient.Projects.ListProjects(options)
			if err1 != nil {
				utility.Errorf("查询项目仓库信息失败:%+v\n", err1.Error())
				return
			}
			if !utility.IsDebug(ctx, parse) {
				bar.Add(len(ret))
			}
			for _, v := range ret {
				gmap.Set(v.ID, v)
			}
			utility.Debugf(ctx, parse, "查询项目仓库信息进度:%d/%d\n", page, res.TotalPages)
		}(page, &optionsCopy)
	}
	wg.Wait()
	if !utility.IsDebug(ctx, parse) {
		bar.Finish()
	}
	return
}

func (s *sGitlab) getAllUserMap(ctx context.Context, parse *gcmd.Parser, options *gitlab.ListUsersOptions) (userMap map[string]*gitlab.User) {
	var bar *progressbar.ProgressBar
	userMap = make(map[string]*gitlab.User)
	gmap := gmap.New(true)
	defer func() {
		for k, v := range gmap.Map() {
			userMap[k.(string)] = v.(*gitlab.User)
		}
	}()
	page := 1
	perPage := 100
	searchActive := true
	utility.Debugf(ctx, parse, "开始查询所有用户信息\n")
	//先查询一次，获取总页数，如果大于2页，则并发读取
	options.ListOptions.Page = page
	options.ListOptions.PerPage = perPage
	options.Active = &searchActive
	firstRet, firstRes, err := s.gitClient.Users.ListUsers(options)
	if err != nil {
		utility.Errorf("查询用户信息失败:%+v\n", err.Error())
		return
	}
	for _, v := range firstRet {
		gmap.Set(v.Username, v)
	}
	utility.Debugf(ctx, parse, "查询用户信息进度:%d/%d\n", page, firstRes.TotalPages)
	if firstRes.TotalPages <= 1 {
		return
	}
	page++

	if !utility.IsDebug(ctx, parse) {
		bar = progressbar.Default(int64(firstRes.TotalItems), "查询用户信息")
		bar.Add(gmap.Size())
	}
	wg := sync.WaitGroup{}
	for ; page <= firstRes.TotalPages; page++ {
		wg.Add(1)
		optionsCopy := *options
		go func(page int, options *gitlab.ListUsersOptions) {
			defer func() {
				wg.Done()
			}()
			options.ListOptions.Page = page
			ret, res, err1 := s.gitClient.Users.ListUsers(options)
			if err1 != nil {
				utility.Errorf("查询用户信息失败:%+v\n", err1.Error())
				return
			}
			if !utility.IsDebug(ctx, parse) {
				bar.Add(len(ret))
			}
			for _, v := range ret {
				gmap.Set(v.Username, v)
			}
			utility.Debugf(ctx, parse, "查询用户信息进度:%d/%d\n", page, res.TotalPages)
		}(page, &optionsCopy)
	}
	wg.Wait()
	if !utility.IsDebug(ctx, parse) {
		bar.Finish()
	}
	return
}

func (s *sGitlab) getProjectUserMap(ctx context.Context, parse *gcmd.Parser, pid int) (userMap map[string]*gitlab.ProjectUser) {
	userMap = make(map[string]*gitlab.ProjectUser)
	page := 1
	perPage := 100
	utility.Debugf(ctx, parse, "开始查询项目[%d]的用户信息\n", pid)
	for {
		ret, res, err := s.gitClient.Projects.ListProjectsUsers(pid, &gitlab.ListProjectUserOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			utility.Errorf("查询项目[%d]的用户信息失败:%+v\n", pid, err.Error())
			return
		}
		for _, v := range ret {
			userMap[v.Username] = v
		}
		utility.Debugf(ctx, parse, "查询项目[%d]的用户信息进度:%d/%d\n", pid, page, res.TotalPages)
		if page >= res.TotalPages {
			break
		}
		page++
	}
	return
}

func (s *sGitlab) getProjectCommits(ctx context.Context, parse *gcmd.Parser, pid int, options *gitlab.ListCommitsOptions) (commits []*gitlab.Commit) {
	page := 1
	perPage := 100
	utility.Debugf(ctx, parse, "开始查询项目[%d]的提交信息\n", pid)
	for {
		options.ListOptions.Page = page
		options.ListOptions.PerPage = perPage
		ret, res, err := s.gitClient.Commits.ListCommits(pid, options)
		if err != nil {
			utility.Errorf("查询项目[%d]提交信息失败:%+v\n", pid, err.Error())
			return
		}
		commits = append(commits, ret...)
		utility.Debugf(ctx, parse, "查询项目[%d]提交信息进度:%d/%d\n", pid, page, res.TotalPages)
		if page >= res.TotalPages {
			break
		}
		page++
	}
	return
}
