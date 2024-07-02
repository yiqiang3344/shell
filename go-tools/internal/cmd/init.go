package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"go-tools/internal/service"
)

var (
	gitlabCommonArs = []gcmd.Argument{
		{
			Name:   "url",
			Short:  "l",
			Brief:  "gitlab链接地址，如：https://gitlab.com。也可以在配置文件中设置，优先使用命令行参数。",
			IsArg:  false,
			Orphan: false,
		},
		{
			Name:   "token",
			Short:  "t",
			Brief:  "gitlab token。也可以在配置文件中设置，优先使用命令行参数。",
			IsArg:  false,
			Orphan: false,
		},
		{
			Name:   "debug",
			Short:  "v",
			Brief:  "打印debug信息。也可以在配置文件中设置，优先使用命令行参数。",
			IsArg:  false,
			Orphan: true,
		},
	}

	_init = gcmd.Command{
		Usage:       "./go-tools 工具命令",
		Description: "go版工具箱",
	}

	demo = &gcmd.Command{
		Name:        "demo",
		Usage:       "./go-tools demo",
		Description: "示范demo",
		Arguments: []gcmd.Argument{
			{
				Name:   "argsA",
				Short:  "a",
				Brief:  "参数A",
				IsArg:  false,
				Orphan: false,
			},
			{
				Name:   "argsB",
				Short:  "b",
				Brief:  "参数B",
				IsArg:  false,
				Orphan: true,
			},
			{
				Name:   "argsC",
				Short:  "c",
				Brief:  "参数C",
				IsArg:  true,
				Orphan: false,
			},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			service.Demo().Demo(ctx, parser)
			return
		},
	}

	setGitlabProjectsMember = &gcmd.Command{
		Name:        "setGitlabProjectsMember",
		Usage:       "./go-tools setGitlabProjectsMember",
		Description: "给指定用户名的gitlab用户批量设置指定仓库的报告者权限，通过交互方式输入用户名和仓库名。",
		Arguments: append(gitlabCommonArs, []gcmd.Argument{
			{
				Name:   "username",
				Short:  "u",
				Brief:  "gitlab用户名。也可以在配置文件中设置，优先使用命令行参数。",
				IsArg:  false,
				Orphan: false,
			},
			{
				Name:   "projectNames",
				Short:  "p",
				Brief:  "gitlab仓库名（可以是完整组名加项目名，也可以是项目名，项目名可以模糊匹配，完整组名会精确匹配），多个则逗号分割，如: projectName1,group/projectName2。也可以在配置文件中设置，优先使用命令行参数。",
				IsArg:  false,
				Orphan: false,
			},
			{
				Name:   "accessLevel",
				Short:  "a",
				Brief:  "项目权限：访客, 报告者, 开发人员, 主程序员。也可以在配置文件中设置，优先使用命令行参数。",
				IsArg:  false,
				Orphan: false,
			},
		}...),
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			service.Gitlab().SetProjectsMember(ctx, parser)
			return
		},
	}

	gitClone = &gcmd.Command{
		Name:        "gitClone",
		Usage:       "./go-tools gitClone",
		Description: "批量clone仓库的代码到指定目录。",
		Arguments: append(gitlabCommonArs, []gcmd.Argument{
			{
				Name:   "codeDir",
				Short:  "c",
				Brief:  "代码目录。也可以在配置文件中设置，优先使用命令行参数。",
				IsArg:  false,
				Orphan: false,
			},
			{
				Name:   "searchKey",
				Short:  "k",
				Brief:  "搜索条件，匹配仓库名、描述。也可以在配置文件中设置，优先使用命令行参数。",
				IsArg:  false,
				Orphan: false,
			},
		}...),
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			service.Gitlab().Clone(ctx, parser)
			return
		},
	}

	gitlabCommitStats = &gcmd.Command{
		Name:        "gitlabCommitStats",
		Usage:       "./go-tools gitlabCommitStats",
		Description: "统计指定gitlab用户指定时间范围的提交统计信息",
		Arguments: append(gitlabCommonArs, []gcmd.Argument{
			{
				Name:   "usernames",
				Short:  "u",
				Brief:  "gitlab用户名列表，逗号分割。也可以在配置文件中设置，优先使用命令行参数。",
				IsArg:  false,
				Orphan: false,
			},
			{
				Name:   "startTime",
				Short:  "s",
				Brief:  "开始时间。也可以在配置文件中设置，优先使用命令行参数。",
				IsArg:  false,
				Orphan: false,
			},
			{
				Name:   "endTime",
				Short:  "e",
				Brief:  "截止时间。也可以在配置文件中设置，优先使用命令行参数。",
				IsArg:  false,
				Orphan: false,
			},
		}...),
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			service.Gitlab().GetUserCommitStats(ctx, parser)
			return
		},
	}
)

func Init() {
	err := _init.AddCommand(demo, setGitlabProjectsMember, gitClone, gitlabCommitStats)
	if err != nil {
		panic(err)
	}
	_init.Run(gctx.New())
}
