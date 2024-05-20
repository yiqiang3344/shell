// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/xanzy/go-gitlab"
)

type (
	IGitlab interface {
		SetProjectsMember(ctx context.Context)
		FindUserByUsername(ctx context.Context) (user *gitlab.User)
		FindProjectsByNames(ctx context.Context) (projects []*gitlab.Project)
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
