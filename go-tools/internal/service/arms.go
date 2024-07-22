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
	IArms interface {
		ExportAlertHistory(ctx context.Context, parse *gcmd.Parser)
		ExportPromAlerts(ctx context.Context, parse *gcmd.Parser)
	}
)

var (
	localArms IArms
)

func Arms() IArms {
	if localArms == nil {
		panic("implement not found for interface IArms, forgot register?")
	}
	return localArms
}

func RegisterArms(i IArms) {
	localArms = i
}
