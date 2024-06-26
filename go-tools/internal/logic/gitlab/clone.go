package gitlab

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/schollz/progressbar/v3"
	"github.com/xanzy/go-gitlab"
	"go-tools/internal/utility"
	"os/exec"
	"strings"
	"sync"
)

// Clone 批量克隆仓库
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
