// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/gogf/gf/v2/os/gcmd"
)

type (
	IGitlab interface {
		// GetUserCommitStats 获取用户提交统计信息
		GetUserCommitStats(ctx context.Context, parse *gcmd.Parser)
		Clone(ctx context.Context, parse *gcmd.Parser)
		// SetProjectsMember 设置仓库的用户权限
		SetProjectsMember(ctx context.Context, parser *gcmd.Parser)
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
