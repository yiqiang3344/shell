package utility

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/util/gconv"
	"strings"
)

func MapFromList(list []interface{}) map[string]int {
	_map := map[string]int{}
	for _, v := range list {
		_v := strings.Trim(gconv.String(v), "")
		if _v == "" {
			continue
		}
		_map[_v] = 1
	}
	return _map
}

func Scanf(str string) string {
	return gcmd.Scanf("%c[1;0;32m%s%c[0m\n", 0x1B, str, 0x1B)
}

// GetArgString 获取字符串参数，优先从parser获取，其次从配置中获取
func GetArgString(ctx context.Context, parser *gcmd.Parser, cfgPattern string, parserOpt string) string {
	ret := parser.GetOpt(parserOpt).String()
	if ret != "" {
		return ret
	}
	return g.Cfg().MustGet(ctx, cfgPattern).String()
}

func IsDebug(ctx context.Context, parser *gcmd.Parser) bool {
	//开启了打印debug信息开关才打印，包括配置文件和命令选项方式
	return parser.GetOpt("debug") != nil || g.Cfg().MustGet(ctx, "debug").Bool()
}

// Debugf 打印debug信息
func Debugf(ctx context.Context, parser *gcmd.Parser, format string, a ...any) (n int, err error) {
	if IsDebug(ctx, parser) {
		return fmt.Printf("[debug]"+format, a...)
	}
	return
}

// Debugln 打印debug信息
func Debugln(ctx context.Context, parser *gcmd.Parser, format string) (n int, err error) {
	if IsDebug(ctx, parser) {
		return fmt.Println("[debug]" + format)
	}
	return
}

func Errorln(format string) (n int, err error) {
	return fmt.Printf("%c[1;0;31m%s%c[0m\n", 0x1B, "[error]"+format, 0x1B)
}
func Errorf(format string, args ...any) (n int, err error) {
	return fmt.Printf("%c[1;0;31m%s%c[0m", 0x1B, fmt.Sprintf("[error]"+format, args...), 0x1B)
}

func Warnln(format string) (n int, err error) {
	return fmt.Printf("%c[1;0;33m%s%c[0m\n", 0x1B, "[warning]"+format, 0x1B)
}
func Warnf(format string, args ...any) (n int, err error) {
	return fmt.Printf("%c[1;0;33m%s%c[0m", 0x1B, fmt.Sprintf("[warning]"+format, args...), 0x1B)
}
