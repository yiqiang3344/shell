package utility

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/util/gconv"
	"strings"
	"time"
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

var ExcelChar = []string{"", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

func ConvertNumToChar(num int) string {
	if num < 27 {
		return ExcelChar[num]
	}
	k := num % 26
	if k == 0 {
		k = 26
	}
	v := (num - k) / 26
	col := ConvertNumToChar(v)
	cols := col + ExcelChar[k]
	return cols
}

func FormatDuration(d time.Duration) string {
	var ret string
	if d > time.Hour {
		ret = fmt.Sprintf("%s%.0fh", ret, d.Hours())
		d = d - d.Truncate(time.Hour)
	}
	if d > time.Minute {
		ret = fmt.Sprintf("%s%.0fm", ret, d.Minutes())
		d = d - d.Truncate(time.Minute)
	}
	if d > time.Second {
		ret = fmt.Sprintf("%s%.0fs", ret, d.Seconds())
		d = d - d.Truncate(time.Second)
	}
	ret = fmt.Sprintf("%s%dms", ret, d.Milliseconds())
	return ret
}
