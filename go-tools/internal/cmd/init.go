package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"go-tools/internal/service"
)

var (
	_init = gcmd.Command{
		Usage:       "./go-tools 工具命令",
		Description: "go版工具箱",
	}

	demo = &gcmd.Command{
		Name:        "demo",
		Usage:       "./go-tools demo",
		Description: "示范demo",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			service.Demo().Demo()
			return
		},
	}

	setGitlabProjectsMember = &gcmd.Command{
		Name:        "setGitlabProjectsMember",
		Usage:       "./go-tools setGitlabProjectsMember",
		Description: "给指定用户名的gitlab用户批量设置指定仓库的报告者权限，通过交互方式输入用户名和仓库名。",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			service.Gitlab().SetProjectsMember(ctx)
			return
		},
	}
)

func Init() {
	err := _init.AddCommand(demo, setGitlabProjectsMember)
	if err != nil {
		panic(err)
	}
	_init.Run(gctx.New())
}
