package demo

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"go-tools/internal/service"
)

type sDemo struct{}

func New() *sDemo {
	return &sDemo{}
}

func init() {
	service.RegisterDemo(New())
}

func (s *sDemo) Demo(ctx context.Context, parser *gcmd.Parser) {
	g.Dump(parser.GetOptAll(), parser.GetArgAll())
}
