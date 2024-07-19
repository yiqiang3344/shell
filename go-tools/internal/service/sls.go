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
	ISls interface {
		ExportAlerts(ctx context.Context, parse *gcmd.Parser)
	}
)

var (
	localSls ISls
)

func Sls() ISls {
	if localSls == nil {
		panic("implement not found for interface ISls, forgot register?")
	}
	return localSls
}

func RegisterSls(i ISls) {
	localSls = i
}
