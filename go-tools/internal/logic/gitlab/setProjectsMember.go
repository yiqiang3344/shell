package gitlab

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xanzy/go-gitlab"
	"go-tools/internal/utility"
	"regexp"
	"strings"
)

// SetProjectsMember 设置仓库的用户权限
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

	user = s.findUserByUsername(ctx, parser)
	projects = s.findProjectsByNames(ctx, parser)
	accessLevel := s.inputAccessLevel(ctx, parser)
	accessLevelTmp := AccessLevelMap[s.inputAccessLevel(ctx, parser)]
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

func (s *sGitlab) findUserByUsername(ctx context.Context, parser *gcmd.Parser) (user *gitlab.User) {
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

func (s *sGitlab) findProjectsByNames(ctx context.Context, parser *gcmd.Parser) (projects *gmap.ListMap) {
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

			//如果带了仓库组名，则先以仓库名搜索，再通过组名过滤
			group := ""
			if strings.Contains(v, "/") {
				arr := strings.Split(v, "/")
				group = strings.Join(arr[0:len(arr)-1], "/")
				v = arr[len(arr)-1]
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
			for _, v1 := range projectsTmp {
				if group != "" && v1.Namespace.FullPath != group {
					continue
				}
				tmpProjects.Set(v1.ID, v1)
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

func (s *sGitlab) inputAccessLevel(ctx context.Context, parser *gcmd.Parser) (accessLevel string) {
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
