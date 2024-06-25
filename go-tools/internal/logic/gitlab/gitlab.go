package gitlab

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/container/garray"
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
	"math"
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

	//遍历用户，获取用户项目
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
	for k, v := range allProjectMap {
		go func(pid int, project *gitlab.Project) {
			defer func() {
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

	withStats := true
	for _, uc := range userMap.Map() {
		//并发遍历项目，获取提交信息
		wg := sync.WaitGroup{}
		wg.Add(uc.(*UserCommitStats).Projects.Size())
		for _, p := range uc.(*UserCommitStats).Projects.Map() {
			go func(uc1 *UserCommitStats, project projectInfo) {
				defer func() {
					wg.Done()
				}()
				commits := s.getProjectCommits(ctx, parse, project.ID, &gitlab.ListCommitsOptions{
					Author:    &uc1.UserInfo.Username,
					Since:     &startTime.Time,
					Until:     &endTime.Time,
					WithStats: &withStats,
				})
				//遍历提交信息，获取代码行数
				for _, cmt := range commits {
					// 再次过滤，上面的查询过滤条件可能没生效
					if cmt.CommitterEmail != uc1.UserInfo.Email {
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
			}(uc.(*UserCommitStats), p.(projectInfo))
		}
		wg.Wait()
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
			for _, v := range ret {
				gmap.Set(v.ID, v)
			}
			utility.Debugf(ctx, parse, "查询项目仓库信息进度:%d/%d\n", page, res.TotalPages)
		}(page, &optionsCopy)
	}
	wg.Wait()
	return
}

func (s *sGitlab) getAllUserMap(ctx context.Context, parse *gcmd.Parser, options *gitlab.ListUsersOptions) (userMap map[string]*gitlab.User) {
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
			for _, v := range ret {
				gmap.Set(v.Username, v)
			}
			utility.Debugf(ctx, parse, "查询用户信息进度:%d/%d\n", page, res.TotalPages)
		}(page, &optionsCopy)
	}
	wg.Wait()
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
