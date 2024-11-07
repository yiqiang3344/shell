package dingtalk

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xuri/excelize/v2"
	"go-tools/internal/service"
	"go-tools/internal/utility"
	"sort"
)

type sDing struct {
}

func New() *sDing {
	return &sDing{}
}

func init() {
	service.RegisterDing(New())
}

type DingData struct {
	StatisticType     string    `json:"statisticType" dc:"类型"`
	MeterCnt          float64   `json:"meterCnt" dc:"计费数"`
	TotalCnt          float64   `json:"totalCnt" dc:"总数" dc:""`
	ErrCnt            float64   `json:"errCnt" dc:"错误数" dc:""`
	StatisticKey      string    `json:"statisticKey" dc:""`
	StatisticName     string    `json:"statisticName" dc:"介质"`
	StatisticTypeName string    `json:"statisticTypeName" dc:"类型"`
	IsCorpAuthApp     bool      `json:"isCorpAuthApp" dc:""`
	SecondKey         string    `json:"secondKey" dc:""`
	DingGroup         DingGroup `json:"groupInfoVO" dc:"钉钉群信息"`
}

type DingGroup struct {
	Id        string  `json:"id" dc:"ID"`
	Name      string  `json:"name" dc:"钉钉群名"`
	OwnerId   float64 `json:"ownerId" dc:"群主ID"`
	OwnerName string  `json:"ownerName" dc:"群主"`
}

func querySendMsgStatsByPage(ctx context.Context, startTime, endTime *gtime.Time, token, cookie string, pageNum, pageSize int) (total int, ret []*DingData, err error) {
	c := g.Client()
	c.SetPrefix("https://open-dev.dingtalk.com")
	c.SetHeader("Cookie", cookie)
	res := c.GetContent(ctx, "/resource/queryWebhookUseInfo", g.Map{
		"pageSize":  pageSize,
		"pageNum":   pageNum,
		"token":     token,
		"startTime": startTime.Format("Ymd"),
		"endTime":   endTime.Format("Ymd"),
	})
	j, err := gjson.DecodeToJson(res)
	if err != nil {
		return
	}
	total = j.Get("data.data.totalCount").Int()
	err = j.Get("data.data.data").Scan(&ret)
	return
}

func getSendMsgStatsByTime(ctx context.Context, startTime, endTime *gtime.Time, token, cookie string) (list []*DingData, err error) {
	page := 1
	pageSize := 1000
	for {
		total, list1, err1 := querySendMsgStatsByPage(ctx, startTime, endTime, token, cookie, page, pageSize)
		if err1 != nil {
			return nil, err1
		}
		list = append(list, list1...)
		if total <= len(list) {
			break
		}
		page++
	}
	return
}

func (s *sDing) SendMsgStats(ctx context.Context, parse *gcmd.Parser) {
	var err error
	token := utility.GetArgString(ctx, parse, "dingtalk.sendMsgStats.token", "token")
	if token == "" {
		utility.Errorf("token不能为空")
		return
	}
	cookie := utility.GetArgString(ctx, parse, "dingtalk.sendMsgStats.cookie", "cookie")
	if cookie == "" {
		utility.Errorf("cookie不能为空")
		return
	}
	startTimeStr := utility.GetArgString(ctx, parse, "dingtalk.sendMsgStats.startTime", "startTime")
	endTimeStr := utility.GetArgString(ctx, parse, "dingtalk.sendMsgStats.endTime", "endTime")
	//默认为前7天到前1天的数据，当天数据无法查询
	startTime := gtime.NewFromStrFormat(startTimeStr, "Ymd")
	if startTimeStr == "" {
		startTime = gtime.Now().AddDate(0, 0, -7).FormatNew("Y-m-d 00:00:00")
	}
	endTime := gtime.NewFromStrFormat(endTimeStr, "Ymd")
	if startTimeStr == "" {
		endTime = gtime.Now().AddDate(0, 0, -1).FormatNew("Y-m-d 00:00:00")
	}

	utility.Debugf(ctx, parse, "开始处理:\n起始时间:%v\n截止时间:%v\n", startTime.Format("Ymd"), endTime.Format("Ymd"))

	//输出excel
	f := excelize.NewFile()
	defer func() {
		if err = f.Close(); err != nil {
			utility.Errorf("excel关闭异常:%v", err)
		}
	}()
	excelFilepath := fmt.Sprintf("%s/%s到%s钉钉消息发送统计.xlsx", gfile.Pwd(), startTime.Format("Ymd"), endTime.Format("Ymd"))
	sheetHead := []cell{
		{Name: "发送者", Width: 40},
		{Name: "类型", Width: 10},
		{Name: "钉钉群", Width: 60},
		{Name: "群主", Width: 30},
		{Name: "有效数", Width: 10},
		{Name: "异常数", Width: 10},
		{Name: "总数", Width: 10},
	}
	// 汇总统计
	list, err := getSendMsgStatsByTime(ctx, startTime, endTime, token, cookie)
	if err != nil {
		utility.Errorf("获取汇总数据失败:%v", err)
		return
	}
	//按有效数倒序排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].MeterCnt > list[j].MeterCnt
	})
	sheetIndex, err := writeSheet(f, "汇总", sheetHead, list)
	if err != nil {
		utility.Errorf("汇总数据写入excel异常:%v", err)
		return
	}
	f.SetActiveSheet(sheetIndex)
	utility.Debugf(ctx, parse, "汇总数据生成完毕\n")

	// 按天统计
	sTime := startTime
	for {
		list1, err1 := getSendMsgStatsByTime(ctx, sTime, sTime, token, cookie)
		if err1 != nil {
			utility.Errorf("获取%s数据失败:%v", sTime.Format("Ymd"), err1)
			return
		}
		//按有效数倒序排序
		sort.Slice(list1, func(i, j int) bool {
			return list1[i].MeterCnt > list1[j].MeterCnt
		})
		_, err1 = writeSheet(f, sTime.Format("Ymd"), sheetHead, list1)
		if err1 != nil {
			utility.Errorf("%s数据写入excel异常:%v", sTime.Format("Ymd"), err)
			return
		}
		utility.Debugf(ctx, parse, "%s数据处理完毕\n", sTime.Format("Ymd"))
		if sTime.Format("Ymd") == endTime.Format("Ymd") {
			break
		}
		sTime = sTime.AddDate(0, 0, 1)
	}
	utility.Debugf(ctx, parse, "整体数据处理完毕\n")
	f.DeleteSheet("Sheet1")
	// 根据指定路径保存文件
	if err = f.SaveAs(excelFilepath); err != nil {
		utility.Errorf("excel保存异常:%v", err)
	}
	fmt.Printf("excel文件生成完毕，地址:%s\n", excelFilepath)
	return
}

type cell struct {
	Name  string
	Width float64
}

func writeSheet(f *excelize.File, sheetName string, sheetHead []cell, list []*DingData) (sheetIndex int, err error) {
	// 添加sheet
	sheetIndex, err = f.NewSheet(sheetName)
	if err != nil {
		return
	}
	// 设置头部
	for k, v := range sheetHead {
		f.SetCellValue(sheetName, utility.ConvertNumToChar(k+1)+gconv.String(1), v.Name)
		f.SetColWidth(sheetName, utility.ConvertNumToChar(k+1), utility.ConvertNumToChar(k+1), v.Width)
	}
	// 设置值
	n := 2
	for _, v := range list {
		f.SetCellValue(sheetName, utility.ConvertNumToChar(1)+gconv.String(n), v.StatisticName)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(2)+gconv.String(n), v.StatisticTypeName)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(3)+gconv.String(n), v.DingGroup.Name)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(4)+gconv.String(n), v.DingGroup.OwnerName)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(5)+gconv.String(n), v.MeterCnt)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(6)+gconv.String(n), v.ErrCnt)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(7)+gconv.String(n), v.TotalCnt)
		n++
	}
	return
}
