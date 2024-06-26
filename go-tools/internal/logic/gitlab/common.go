package gitlab

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/xanzy/go-gitlab"
	"go-tools/internal/utility"
)

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
