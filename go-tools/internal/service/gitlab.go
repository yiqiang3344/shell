// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/xanzy/go-gitlab"
)

type (
	IGitlab interface {
		SetProjectsMember(ctx context.Context, parser *gcmd.Parser)
		InputAccessLevel(ctx context.Context, parser *gcmd.Parser) (accessLevel string)
		FindUserByUsername(ctx context.Context, parser *gcmd.Parser) (user *gitlab.User)
		FindProjectsByNames(ctx context.Context, parser *gcmd.Parser) (projects *gmap.ListMap)
		Clone(ctx context.Context, parse *gcmd.Parser)
		GetUserCommitStats(ctx context.Context, parse *gcmd.Parser)
	}
)

var (
	localGitlab IGitlab
)

func Gitlab() IGitlab {
	if localGitlab == nil {
		panic("implement not found for interface IGitlab, forgot register?")
	}
	return localGitlab
}

func RegisterGitlab(i IGitlab) {
	localGitlab = i
}
