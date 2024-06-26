package gitlab

import (
	"github.com/xanzy/go-gitlab"
	"go-tools/internal/service"
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
