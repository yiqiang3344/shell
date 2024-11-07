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
	IDing interface {
		SendMsgStats(ctx context.Context, parse *gcmd.Parser)
	}
)

var (
	localDing IDing
)

func Ding() IDing {
	if localDing == nil {
		panic("implement not found for interface IDing, forgot register?")
	}
	return localDing
}

func RegisterDing(i IDing) {
	localDing = i
}
