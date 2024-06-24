package gitlab

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/schollz/progressbar/v3"
	"github.com/xanzy/go-gitlab"
	"github.com/xuri/excelize/v2"
	"go-tools/internal/service"
	"go-tools/internal/utility"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type sGitlab struct {
	gitClient *gitlab.Client
}

func New() *sGitlab {
	return &sGitlab{}
}

func init() {
	service.RegisterGitlab(New())
}

var (
	AccessLevelMap = map[string]gitlab.AccessLevelValue{
		"无":    gitlab.NoPermissions,
		"最小访问": gitlab.MinimalAccessPermissions,
		"访客":   gitlab.GuestPermissions,
		"报告者":  gitlab.ReporterPermissions,
		"开发人员": gitlab.DeveloperPermissions,
		"主程序员": gitlab.MaintainerPermissions,
		"所有者":  gitlab.OwnerPermissions,
		"管理员":  gitlab.AdminPermissions,
	}
)

func (s *sGitlab) initClient(ctx context.Context, parser *gcmd.Parser) (err error) {
	var (
		token = utility.GetArgString(ctx, parser, "gitlab.token", "token")
		url   = utility.GetArgString(ctx, parser, "gitlab.url", "url")
	)
	if token == "" {
		err = gerror.New("token不能为空")
		return
	}
	if url == "" {
		err = gerror.New("url不能为空")
		return
	}
	s.gitClient, err = gitlab.NewClient(
		token,
		gitlab.WithBaseURL(url),
	)
	return
}

func (s *sGitlab) SetProjectsMember(ctx context.Context, parser *gcmd.Parser) {
	var (
		user     *gitlab.User
		projects *gmap.ListMap
		err      error
	)

	err = s.initClient(ctx, parser)
	if err != nil {
		utility.Errorf("客户端初始化失败:%+v\n", err.Error())
		return
	}

	user = s.FindUserByUsername(ctx, parser)
	projects = s.FindProjectsByNames(ctx, parser)
	accessLevel := s.InputAccessLevel(ctx, parser)
	accessLevelTmp := AccessLevelMap[s.InputAccessLevel(ctx, parser)]
	for _, v := range projects.Values() {
		project := v.(*gitlab.Project)
		var (
			projectMember *gitlab.ProjectMember
		)
		//先检查人员是否存在，不存在则新增，存在则修改权限
		projectMember, _, err = s.gitClient.ProjectMembers.GetProjectMember(project.ID, user.ID)
		if projectMember != nil {
			_, _, err = s.gitClient.ProjectMembers.EditProjectMember(project.ID, user.ID, &gitlab.EditProjectMemberOptions{
				AccessLevel: &accessLevelTmp,
			})
			if err != nil {
				utility.Errorf("仓库[%+v]编辑用户[%+v]权限[%s]失败:%+v\n", project.PathWithNamespace, user.Username, accessLevel, err.Error())
				continue
			}
			utility.Debugf(ctx, parser, "仓库[%+v]编辑用户[%+v]权限[%s]成功\n", project.PathWithNamespace, user.Username, accessLevel)
		} else {
			_, _, err = s.gitClient.ProjectMembers.AddProjectMember(project.ID, &gitlab.AddProjectMemberOptions{
				UserID:      user.ID,
				AccessLevel: &accessLevelTmp,
			})
			if err != nil {
				utility.Errorf("仓库[%+v]添加用户[%+v]权限[%s]失败:%+v\n", project.PathWithNamespace, user.Username, accessLevel, err.Error())
				continue
			}
			utility.Debugf(ctx, parser, "仓库[%+v]添加用户[%+v]权限[%s]成功\n", project.PathWithNamespace, user.Username, accessLevel)
		}

		fmt.Printf("仓库[%+v]设置用户[%+v]权限[%s]成功\n", project.PathWithNamespace, user.Username, accessLevel)
	}
}

func (s *sGitlab) InputAccessLevel(ctx context.Context, parser *gcmd.Parser) (accessLevel string) {
	accessLevel = utility.GetArgString(ctx, parser, "gitlab.setProjectMember.accessLevel", "accessLevel")
	for {
		if accessLevel == "" {
			accessLevel = utility.Scanf("输入权限类型:")
		}
		if strings.Trim(accessLevel, "") == "" {
			utility.Warnln("权限类型不能为空")
			accessLevel = ""
			continue
		}

		if _, ok := AccessLevelMap[accessLevel]; !ok {
			utility.Warnf("权限类型不合法:%s\n", accessLevel)
			accessLevel = ""
			continue
		}
		break
	}
	return
}

func (s *sGitlab) FindUserByUsername(ctx context.Context, parser *gcmd.Parser) (user *gitlab.User) {
	var (
		userId     string
		users      []*gitlab.User
		inputMatch bool
		err        error
	)

	//根据用户名获取用户ID
	username := utility.GetArgString(ctx, parser, "gitlab.setProjectMember.username", "username")
	for {
		if username == "" {
			username = utility.Scanf("输入gitlab用户名:")
		}
		if strings.Trim(username, "") == "" {
			utility.Warnln("用户名不能为空")
			username = ""
			continue
		}
		users, _, err = s.gitClient.Users.ListUsers(&gitlab.ListUsersOptions{
			Username: &username,
		})
		if err != nil {
			utility.Errorf("查询用户报错:%+v\n", err.Error())
			username = ""
			continue
		}
		if len(users) == 0 {
			utility.Warnln("无匹配的用户")
			username = ""
			continue
		}
		break
	}

	//输出查询到的用户信息，确认用户
	fmt.Printf("序号 | id | username | email | name\n")
	for i, v := range users {
		fmt.Printf("%+v | %+v | %+v | %+v | %+v\n", i, v.ID, v.Username, v.Email, v.Name)
	}

	//确认用户
	for {
		userId = utility.Scanf("确认用户序号:")
		//检查选中的用户序号是否存在
		inputMatch, err = regexp.MatchString("[0-"+gconv.String(len(users)-1)+"]", userId)
		if !inputMatch {
			utility.Warnf("选择的用户序号[%+v]不在可选范围内\n", userId)
			continue
		}
		break
	}

	user = users[gconv.Int(userId)]
	return
}

func (s *sGitlab) FindProjectsByNames(ctx context.Context, parser *gcmd.Parser) (projects *gmap.ListMap) {
	var (
		tmpProjects    *gmap.ListMap
		err            error
		projectIds     []string
		simpleRepoInfo = true
	)
	//根据仓库名确认仓库ID
	projectNames := utility.GetArgString(ctx, parser, "gitlab.setProjectMember.projectNames", "projectNames")
	for {
		if strings.Trim(projectNames, "") == "" {
			projectNames = utility.Scanf("输入仓库名,多个则以逗号分割:")
		}
		if strings.Trim(projectNames, "") == "" {
			utility.Warnln("仓库名不能为空")
			projectNames = ""
			continue
		}
		projectNameList := strings.Split(projectNames, ",")
		flag := true
		tmpProjects = gmap.NewListMap()
		for _, v := range projectNameList {
			if strings.Trim(v, "") == "" {
				utility.Warnln("仓库名不能为空")
				projectNames = ""
				flag = false
				break
			}
			var projectsTmp []*gitlab.Project
			projectsTmp, _, err = s.gitClient.Projects.ListProjects(&gitlab.ListProjectsOptions{
				Simple: &simpleRepoInfo,
				Search: &v,
			})
			if err != nil {
				utility.Errorf("拉取仓库列表报错:%+v\n", err.Error())
				projectNames = ""
				flag = false
				break
			}
			if len(projectsTmp) == 0 {
				utility.Warnf("无对应仓库:%+v\n", v)
				projectNames = ""
				flag = false
				break
			}
			for _, v := range projectsTmp {
				tmpProjects.Set(v.ID, v)
			}
		}
		if !flag {
			continue
		}
		if tmpProjects.Size() == 0 {
			utility.Warnln("未查询到任何仓库")
			projectNames = ""
			continue
		}
		break
	}

	fmt.Printf("id | name | Path | description\n")
	for _, v := range tmpProjects.Values() {
		vTmp := v.(*gitlab.Project)
		fmt.Printf("%+v | %+v | %+v | %+v \n", vTmp.ID, vTmp.Name, vTmp.PathWithNamespace, vTmp.Description)
	}

	for {
		projectIdsTmp := utility.Scanf("选择仓库ID,多个则以逗号分割:")
		if strings.Trim(projectIdsTmp, "") == "" {
			utility.Warnln("仓库ID列表不能为空")
			continue
		}
		projectIds = strings.Split(projectIdsTmp, ",")
		//检查选择的ID是否存在
		flag := true
		projects = gmap.NewListMap()
		for _, v := range projectIds {
			id := gconv.Int(v)
			//检查选中ID是否存在
			if !tmpProjects.Contains(id) {
				utility.Warnf("选择的仓库序号[%+v]不在可选范围内\n", id)
				flag = false
				continue
			}
			projects.Set(id, tmpProjects.Get(id))
		}
		if !flag {
			continue
		}
		break
	}

	return
}

func (s *sGitlab) Clone(ctx context.Context, parse *gcmd.Parser) {
	var (
		codePath       string
		searchKey      string
		simpleRepoInfo = true
		err            error
		projects       []*gitlab.Project
		page           int
		perPage        int
		cfgMap         map[string]int
		bar            *progressbar.ProgressBar
	)

	err = s.initClient(ctx, parse)
	if err != nil {
		utility.Errorf("初始化gitlab客户端报错:%+v\n", err.Error())
		return
	}

	//1.检查是否有本地存储的目录，没有的话交互输入
	codePath = utility.GetArgString(ctx, parse, "gitlab.clone.codeDir", "codeDir")
	for {
		if strings.Trim(codePath, "") == "" {
			codePath = utility.Scanf("请输入代码存储的目录:")
			continue
		}
		break
	}
	//1.1.代码存储目录是否存在，不存在则创建目录
	if !gfile.IsDir(codePath) {
		err = gfile.Mkdir(codePath)
		if err != nil {
			utility.Errorf("创建代码目录报错:%+v\n", err)
			return
		}
	}
	utility.Debugf(ctx, parse, "代码存储目录:%+v\n", codePath)

	//2.检查是否有clone查询关键字，没有则交互提示输入，输入空则获取所有
	for {
		searchKey = utility.GetArgString(ctx, parse, "gitlab.clone.searchKey", "searchKey")
		if strings.Trim(searchKey, "") == "" {
			searchKey = utility.Scanf("请输入项目查询关键字，为空则表示所有:")
		}
		projects = []*gitlab.Project{}
		page = 1
		perPage = 100
		for {
			var (
				projectsTmp []*gitlab.Project
			)
			projectsTmp, _, err = s.gitClient.Projects.ListProjects(&gitlab.ListProjectsOptions{
				Search: &searchKey,
				Simple: &simpleRepoInfo,
				ListOptions: gitlab.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			})
			if err != nil {
				utility.Errorf("查询仓库列表报错:%+v\n", err)
				return
			}
			if len(projectsTmp) == 0 {
				break
			}
			projects = append(projects, projectsTmp...)
			page++
		}

		//3.检查是否有过滤项目配置，有的话过滤
		//3.1.检查只需要的仓库配置
		cfgMap = utility.MapFromList(g.Cfg().MustGet(ctx, "gitlab.clone.filter.expectProjects").Slice())
		if len(cfgMap) > 0 {
			utility.Debugln(ctx, parse, "3.1.检查只需要的仓库配置")
			var projectsTmp []*gitlab.Project
			for _, p := range projects {
				if _, ok := cfgMap[p.Name]; ok {
					utility.Debugf(ctx, parse, "命中:%+v\n", p.PathWithNamespace)
					projectsTmp = append(projectsTmp, p)
				}
			}
			projects = projectsTmp
		}
		//3.2.检查只需要的仓库组配置
		cfgMap = utility.MapFromList(g.Cfg().MustGet(ctx, "gitlab.clone.filter.expectGroups").Slice())
		if len(cfgMap) > 0 {
			utility.Debugln(ctx, parse, "3.2.检查只需要的仓库组配置")
			var projectsTmp []*gitlab.Project
			for _, p := range projects {
				if _, ok := cfgMap[p.Namespace.FullPath]; ok {
					utility.Debugf(ctx, parse, "命中:%+v\n", p.PathWithNamespace)
					projectsTmp = append(projectsTmp, p)
				}
			}
			projects = projectsTmp
		}
		//3.3.检查只需要的顶级仓库组配置
		cfgMap = utility.MapFromList(g.Cfg().MustGet(ctx, "gitlab.clone.filter.expectTopGroups").Slice())
		if len(cfgMap) > 0 {
			utility.Debugln(ctx, parse, "3.3.检查只需要的顶级仓库组配置")
			var projectsTmp []*gitlab.Project
			for _, p := range projects {
				if _, ok := cfgMap[strings.Split(p.Namespace.FullPath, "/")[0]]; ok {
					utility.Debugf(ctx, parse, "命中:%+v\n", p.PathWithNamespace)
					projectsTmp = append(projectsTmp, p)
				}
			}
			projects = projectsTmp
		}
		//3.4.检查要过滤的仓库配置
		cfgMap = utility.MapFromList(g.Cfg().MustGet(ctx, "gitlab.clone.filter.ignoreProjects").Slice())
		if len(cfgMap) > 0 {
			utility.Debugln(ctx, parse, "3.4.检查要过滤的仓库配置")
			for i, p := range projects {
				if _, ok := cfgMap[p.Name]; ok {
					utility.Debugf(ctx, parse, "命中:%+v\n", p.PathWithNamespace)
					if len(projects) > 1 {
						projects = append(projects[:i], projects[i+1:]...)
					} else {
						projects = projects[0:0]
					}
				}
			}
		}
		//3.5.检查要过滤的仓库组配置
		cfgMap = utility.MapFromList(g.Cfg().MustGet(ctx, "gitlab.clone.filter.ignoreGroups").Slice())
		if len(cfgMap) > 0 {
			utility.Debugln(ctx, parse, "3.5.检查要过滤的仓库组配置")
			for i, p := range projects {
				if _, ok := cfgMap[p.Namespace.FullPath]; ok {
					utility.Debugf(ctx, parse, "命中:%+v\n", p.PathWithNamespace)
					if len(projects) > 1 {
						projects = append(projects[:i], projects[i+1:]...)
					} else {
						projects = projects[0:0]
					}
				}
			}
		}
		//3.6.检查要过滤的顶级仓库组配置
		cfgMap = utility.MapFromList(g.Cfg().MustGet(ctx, "gitlab.clone.filter.ignoreTopGroups").Slice())
		if len(cfgMap) > 0 {
			utility.Debugln(ctx, parse, "3.6.检查要过滤的顶级仓库组配置")
			for i, p := range projects {
				if _, ok := cfgMap[strings.Split(p.Namespace.FullPath, "/")[0]]; ok {
					utility.Debugf(ctx, parse, "命中:%+v\n", p.PathWithNamespace)
					if len(projects) > 1 {
						projects = append(projects[:i], projects[i+1:]...)
					} else {
						projects = projects[0:0]
					}
				}
			}
		}

		//4.仓库数小于20则列出仓库信息，大于则只输出仓库数，让用户确认是否clone
		if len(projects) <= 20 {
			for i, v := range projects {
				if i == 0 {
					fmt.Printf("列表详情:\nid | name | Path | description\n")
				}
				fmt.Printf("%+v | %+v | %+v | %+v \n", v.ID, v.Name, v.PathWithNamespace, v.Description)
			}
		}
		fmt.Printf("总数:%+v\n", len(projects))

		flag := false
		for {
			flag = false
			confirm := utility.Scanf("是否确认[y:确认,d:查看列表详情,其他:重新查询]")
			if confirm == "y" {
				flag = true
				break
			} else if confirm == "d" {
				for i, v := range projects {
					if i == 0 {
						fmt.Printf("列表详情:\nid | name | Path | description\n")
					}
					fmt.Printf("%+v | %+v | %+v | %+v \n", v.ID, v.Name, v.PathWithNamespace, v.Description)
				}
				continue
			} else {
				break
			}
		}
		if flag {
			break
		}
	}

	//5.并行clone
	if !utility.IsDebug(ctx, parse) {
		bar = progressbar.Default(int64(len(projects)), "cloning")
	}
	var wg sync.WaitGroup
	for _, p := range projects {
		wg.Add(1)
		go func(p *gitlab.Project) {
			defer wg.Done()
			var (
				projectDir  = codePath + "/" + p.Namespace.FullPath
				projectPath = projectDir + "/" + p.Name
				stdout      []byte
				time        *gtime.Time
			)
			err = gfile.Mkdir(projectDir)
			if err != nil {
				utility.Errorf("%+v创建目录报错:%+v\n", p.PathWithNamespace, err)
				if !utility.IsDebug(ctx, parse) {
					bar.Add(1)
				}
				return
			}
			cmd := exec.Command("git", "clone", p.SSHURLToRepo, projectPath)
			utility.Debugf(ctx, parse, "执行命令:%+v\n", cmd.Args)
			time = gtime.Now()
			stdout, err = cmd.CombinedOutput()
			if err != nil {
				utility.Errorf("%+v执行报错:%+v\n", p.PathWithNamespace, strings.TrimRight(string(stdout), "\n"))
				if !utility.IsDebug(ctx, parse) {
					bar.Add(1)
				}
				return
			}
			utility.Debugf(ctx, parse, "[%+v]clone完成,耗时%.2f秒\n", p.PathWithNamespace, gtime.Now().Sub(time).Seconds())
			if !utility.IsDebug(ctx, parse) {
				bar.Add(1)
			}
		}(p)
	}
	wg.Wait()
	utility.Debugln(ctx, parse, "完成")
	fmt.Printf("代码存储目录:\n%s\n", codePath)
}

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
	Projects           map[int]projectInfo
	TotalCommitStats   *gitlab.CommitStats
	ProjectCommitStats map[int]*gitlab.CommitStats
	ProjectCommit      map[int][]*gitlab.Commit
}

func (s *sGitlab) StatsUserCodeLines(ctx context.Context, parse *gcmd.Parser) {
	var (
		err       error
		userInput []string
		userMap   = make(map[string]*UserCommitStats)
		startTime time.Time
		endTime   time.Time
	)
	//要统计的用户列表
	userInput = []string{
		"fengshutong",
	}

	//获取起止时间
	startTime = gtime.NewFromStr("2024-06-14").Time
	endTime = gtime.NewFromStr("2024-06-24").Time

	err = s.initClient(ctx, parse)
	if err != nil {
		utility.Errorf("客户端初始化失败:%+v\n", err.Error())
		return
	}

	//获取用户信息
	allUserMap := s.getAllUserMap(ctx, parse, &gitlab.ListUsersOptions{})

	//遍历用户，获取用户项目
	for _, v := range userInput {
		if _, ok := allUserMap[v]; !ok {
			utility.Debugf(ctx, parse, "用户[%s]不存在\n", v)
			continue
		}
		uInfo := userInfo{}
		gconv.ConvertWithRefer(allUserMap[v], &uInfo)
		userMap[v] = &UserCommitStats{
			UserInfo:           uInfo,
			Projects:           make(map[int]projectInfo),
			ProjectCommitStats: make(map[int]*gitlab.CommitStats),
			TotalCommitStats:   &gitlab.CommitStats{},
			ProjectCommit:      make(map[int][]*gitlab.Commit),
		}
	}

	// 遍历所有项目，查询用户加入的项目列表
	//search := "lendtrade"
	allProjectMap := s.getAllProjectMap(ctx, parse, &gitlab.ListProjectsOptions{
		LastActivityAfter:  &startTime,
		LastActivityBefore: &endTime,
		//Search:             &search,
	})
	for _, v := range allProjectMap {
		projectUserMap := s.getProjectUserMap(ctx, parse, v.ID)
		for _, v1 := range userMap {
			if _, ok := projectUserMap[v1.UserInfo.Username]; ok {
				pInfo := projectInfo{}
				gconv.ConvertWithRefer(v, &pInfo)
				v1.Projects[v.ID] = pInfo
			}
		}
	}

	withStats := true
	for _, v := range userMap {
		//遍历项目，获取提交
		for _, v1 := range v.Projects {
			commits := s.getProjectCommits(ctx, parse, v1.ID, &gitlab.ListCommitsOptions{
				Author:    &v.UserInfo.Username,
				Since:     &startTime,
				Until:     &endTime,
				WithStats: &withStats,
			})
			//遍历提交，获取代码行数
			for _, v2 := range commits {
				// 再次过滤，上面的查询过滤条件可能没生效
				if v2.CommitterEmail != v.UserInfo.Email {
					continue
				}
				g.Dump("命中", v2)
				//记录项目提交明细
				if c, ok := v.ProjectCommit[v1.ID]; ok {
					c = append(c, v2)
					v.ProjectCommit[v1.ID] = c
				} else {
					v.ProjectCommit[v1.ID] = []*gitlab.Commit{v2}
				}
				//统计用户每项目的代码行数
				if c, ok := v.ProjectCommitStats[v1.ID]; ok {
					c.Total += v2.Stats.Total
					c.Deletions += v2.Stats.Deletions
					c.Additions += v2.Stats.Additions
				} else {
					v.ProjectCommitStats[v1.ID] = v2.Stats
				}
				//统计用户总代码行数
				v.TotalCommitStats.Total += v2.Stats.Total
				v.TotalCommitStats.Deletions += v2.Stats.Deletions
				v.TotalCommitStats.Additions += v2.Stats.Additions
			}
		}
	}

	//输出excel
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			utility.Errorf("excel关闭异常:%v", err)
		}
	}()
	// 创建一个工作表
	excelFilepath := "./用户提交统计.xlsx"
	sheet1Name := "用户维度"
	sheet1Head := []string{"姓名", "增加数", "删除数", "总数"}
	sheet2Name := "用户项目维度"
	sheet2Head := []string{"姓名", "项目", "增加数", "删除数", "总数"}
	sheet1Index, err := f.NewSheet(sheet1Name)
	if err != nil {
		utility.Errorf("excel创建异常:%v", err)
		return
	}
	_, err = f.NewSheet(sheet2Name)
	if err != nil {
		utility.Errorf("excel创建异常:%v", err)
		return
	}
	// 设置sheet1的头部
	for k, v := range sheet1Head {
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(k+1)+gconv.String(1), v)
	}
	// 设置sheet1的值
	for _, v := range userMap {
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(1)+gconv.String(2), v.UserInfo.Name)
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(2)+gconv.String(2), v.TotalCommitStats.Additions)
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(3)+gconv.String(2), v.TotalCommitStats.Deletions)
		f.SetCellValue(sheet1Name, utility.ConvertNumToChar(4)+gconv.String(2), v.TotalCommitStats.Total)
	}
	// 设置sheet2的头部
	for k, v := range sheet2Head {
		f.SetCellValue(sheet2Name, utility.ConvertNumToChar(k+1)+gconv.String(1), v)
	}
	// 设置sheet2的值
	for _, v := range userMap {
		n := 2
		for k1, v1 := range v.ProjectCommitStats {
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(1)+gconv.String(n), v.UserInfo.Name)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(2)+gconv.String(n), v.Projects[k1].PathWithNamespace)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(3)+gconv.String(n), v1.Additions)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(4)+gconv.String(n), v1.Deletions)
			f.SetCellValue(sheet2Name, utility.ConvertNumToChar(5)+gconv.String(n), v1.Total)
			n++
		}
	}
	// 设置工作簿的默认工作表
	f.SetActiveSheet(sheet1Index)
	f.DeleteSheet("Sheet1")
	// 根据指定路径保存文件
	if err := f.SaveAs(excelFilepath); err != nil {
		utility.Errorf("excel保存异常:%v", err)
	}
	fmt.Printf("excel文件生成完毕，地址:%s\n", excelFilepath)
	return
}

func (s *sGitlab) getAllProjectMap(ctx context.Context, parse *gcmd.Parser, options *gitlab.ListProjectsOptions) (projectMap map[int]*gitlab.Project) {
	projectMap = make(map[int]*gitlab.Project)
	page := 1
	perPage := 100
	searchSimple := true
	utility.Debugf(ctx, parse, "开始查询所有仓库项目信息\n")
	for {
		options.ListOptions.Page = page
		options.ListOptions.PerPage = perPage
		options.Simple = &searchSimple
		ret, res, err := s.gitClient.Projects.ListProjects(options)
		if err != nil {
			utility.Errorf("查询项目信息失败:%+v\n", err.Error())
			return
		}
		for _, v := range ret {
			projectMap[v.ID] = v
		}
		utility.Debugf(ctx, parse, "查询进度:%d/%d\n", page, res.TotalPages)
		if page >= res.TotalPages {
			break
		}
		page++
	}
	return
}

func (s *sGitlab) getAllUserMap(ctx context.Context, parse *gcmd.Parser, options *gitlab.ListUsersOptions) (userMap map[string]*gitlab.User) {
	userMap = make(map[string]*gitlab.User)
	page := 1
	perPage := 100
	searchActive := true
	utility.Debugf(ctx, parse, "开始查询所有用户信息\n")
	for {
		options.ListOptions.Page = page
		options.ListOptions.PerPage = perPage
		options.Active = &searchActive
		ret, res, err := s.gitClient.Users.ListUsers(options)
		if err != nil {
			utility.Errorf("查询用户信息失败:%+v\n", err.Error())
			return
		}
		for _, v := range ret {
			userMap[v.Username] = v
		}
		utility.Debugf(ctx, parse, "查询进度:%d/%d\n", page, res.TotalPages)
		if page >= res.TotalPages {
			break
		}
		page++
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
			utility.Errorf("查询项目用户信息失败:%+v\n", err.Error())
			return
		}
		for _, v := range ret {
			userMap[v.Username] = v
		}
		utility.Debugf(ctx, parse, "查询进度:%d/%d\n", page, res.TotalPages)
		if page >= res.TotalPages {
			break
		}
		page++
	}
	return
}

func (s *sGitlab) getProjectBranch(ctx context.Context, parse *gcmd.Parser, pid int) (branches []*gitlab.Branch) {
	page := 1
	perPage := 100
	utility.Debugf(ctx, parse, "开始查询项目[%d]的分支信息\n", pid)
	for {
		ret, res, err := s.gitClient.Branches.ListBranches(pid, &gitlab.ListBranchesOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			utility.Errorf("查询项目[%d]分支信息失败:%+v\n", pid, err.Error())
			return
		}
		branches = append(branches, ret...)
		utility.Debugf(ctx, parse, "查询进度:%d/%d\n", page, res.TotalPages)
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
		utility.Debugf(ctx, parse, "查询进度:%d/%d\n", page, res.TotalPages)
		if page >= res.TotalPages {
			break
		}
		page++
	}
	return
}
