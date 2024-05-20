package gitlab

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xanzy/go-gitlab"
	"go-tools/internal/service"
	"regexp"
	"strings"
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
	AccessLevelReporter = gitlab.ReporterPermissions
)

func (s *sGitlab) SetProjectsMember(ctx context.Context) {
	var (
		user     *gitlab.User
		projects []*gitlab.Project
		err      error
	)

	s.gitClient, err = gitlab.NewClient(g.Cfg().MustGet(ctx, "gitlab.token").String(), gitlab.WithBaseURL(g.Cfg().MustGet(context.Background(), "gitlab.url").String()))
	if err != nil {
		fmt.Println(err)
		return
	}

	user = s.FindUserByUsername(ctx)
	projects = s.FindProjectsByNames(ctx)

	for _, _project := range projects {
		_, _, err := s.gitClient.ProjectMembers.AddProjectMember(_project.ID, &gitlab.AddProjectMemberOptions{
			UserID:      user.ID,
			AccessLevel: &AccessLevelReporter,
		})
		if err != nil {
			fmt.Printf("仓库[%+v]添加用户[%+v]权限失败:%+v\n", _project.PathWithNamespace, user.Username, err.Error())
			continue
		}
		fmt.Printf("仓库[%+v]添加用户[%+v]权限成功\n", _project.PathWithNamespace, user.Username)
	}
}

func (s *sGitlab) FindUserByUsername(ctx context.Context) (user *gitlab.User) {
	var (
		userId     string
		users      []*gitlab.User
		inputMatch bool
		err        error
	)

	//根据用户名获取用户ID
	for {
		username := gcmd.Scanf("%c[1;0;32m%s%c[0m\n", 0x1B, "输入gitlab用户名:", 0x1B)
		//username := gcmd.Scanf("输入gitlab用户名:\n")
		if strings.Trim(username, "") == "" {
			fmt.Println("用户名不能为空")
			continue
		}
		users, _, err = s.gitClient.Users.ListUsers(&gitlab.ListUsersOptions{
			Username: &username,
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(users) == 0 {
			fmt.Println("无匹配的用户")
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
		userId = gcmd.Scanf("%c[1;0;32m%s%c[0m\n", 0x1B, "确认用户序号:", 0x1B)
		//检查选中的用户序号是否存在
		inputMatch, err = regexp.MatchString("[0-"+gconv.String(len(users)-1)+"]", userId)
		if !inputMatch {
			fmt.Printf("选择的用户序号[%+v]不在可选范围内\n", userId)
			continue
		}
		break
	}

	user = users[gconv.Int(userId)]
	return
}

func (s *sGitlab) FindProjectsByNames(ctx context.Context) (projects []*gitlab.Project) {
	var (
		tmpProjects []*gitlab.Project
		inputMatch  bool
		err         error
		projectIds  []string
	)
	//根据仓库名确认仓库ID
	for {
		projectName := gcmd.Scanf("%c[1;0;32m%s%c[0m\n", 0x1B, "输入仓库名,多个则以逗号分割:", 0x1B)
		if strings.Trim(projectName, "") == "" {
			fmt.Println("仓库名不能为空")
			continue
		}
		projectNames := strings.Split(projectName, ",")
		flag := true
		tmpProjects = []*gitlab.Project{}
		for _, v := range projectNames {
			if strings.Trim(v, "") == "" {
				fmt.Println("仓库名不能为空")
				flag = false
				break
			}
			var _projects []*gitlab.Project
			_projects, _, err = s.gitClient.Projects.ListProjects(&gitlab.ListProjectsOptions{
				Search: &v,
			})
			if err != nil {
				fmt.Println(err)
				flag = false
				break
			}
			if len(_projects) == 0 {
				fmt.Printf("无对应仓库:%+v\n", v)
				flag = false
				break
			}
			tmpProjects = append(tmpProjects, _projects...)
		}
		if !flag {
			continue
		}
		if len(tmpProjects) == 0 {
			fmt.Println("未查询到任何仓库")
			continue
		}
		break
	}

	fmt.Printf("序号 | id | name | Path | description\n")
	for i, v := range tmpProjects {
		fmt.Printf("%+v | %+v | %+v | %+v | %+v \n", i, v.ID, v.Name, v.PathWithNamespace, v.Description)
	}

	for {
		_projectIds := gcmd.Scanf("%c[1;0;32m%s%c[0m\n", 0x1B, "选择仓库序号,多个则以逗号分割:", 0x1B)
		if strings.Trim(_projectIds, "") == "" {
			fmt.Println("仓库序号列表不能为空")
			continue
		}
		projectIds = strings.Split(_projectIds, ",")
		//检查选择的序号是否存在
		flag := true
		projects = []*gitlab.Project{}
		for _, v := range projectIds {
			//检查选中的用户序号是否存在
			inputMatch, err = regexp.MatchString("[0-"+gconv.String(len(tmpProjects)-1)+"]", v)
			if !inputMatch {
				fmt.Printf("选择的仓库序号[%+v]不在可选范围内\n", v)
				flag = false
				continue
			}
			projects = append(projects, tmpProjects[gconv.Int(v)])
		}
		if !flag {
			continue
		}
		break
	}

	return
}
