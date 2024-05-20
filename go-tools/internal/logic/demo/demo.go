package demo

import (
	"github.com/gogf/gf/v2/frame/g"
	"go-tools/internal/service"
)

type sDemo struct{}

func New() *sDemo {
	return &sDemo{}
}

func init() {
	service.RegisterDemo(New())
}

func (s *sDemo) Demo() {
	g.Dump("OK")
}
