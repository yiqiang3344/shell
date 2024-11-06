package dingtalk

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xuri/excelize/v2"
	"go-tools/internal/service"
	"go-tools/internal/utility"
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

func (s *sDing) Query(ctx context.Context, parse *gcmd.Parser) (err error) {
	// 先登录钉钉管理后台，再进入如下页面
	// https://open-dev.dingtalk.com/fe/sourceUseDetail?type=event&hash=%23%2F#/
	// 然后打开开发者工具，可看到是通过如下接口获取数据
	// https://open-dev.dingtalk.com/resource/queryWebhookUseInfo?pageSize=100&pageNum=1&statisticType=&startTime=20241101&endTime=20241106&access_token=token
	// 调整一下参数，可直接在浏览器中调用，获取全量数据
	j, err := gjson.Load("resource/tmp/dingtalkSendMsgStats.json")
	if err != nil {
		utility.Errorf("数据导入异常:%v", err)
		return
	}
	var list []*DingData
	j.Get("data.data.data").Scan(&list)

	//输出excel
	f := excelize.NewFile()
	defer func() {
		if err = f.Close(); err != nil {
			utility.Errorf("excel关闭异常:%v", err)
		}
	}()
	// 创建一个工作表
	excelFilepath := fmt.Sprintf("%s/钉钉消息发送统计%s.xlsx", gfile.Pwd(), gtime.Now().Format("YmdHis"))

	type cell struct {
		Name  string
		Width float64
	}

	sheetName := "Sheet1"
	sheet1Head := []cell{
		{Name: "发送者", Width: 40},
		{Name: "类型", Width: 10},
		{Name: "钉钉群", Width: 60},
		{Name: "群主", Width: 30},
		{Name: "总数", Width: 10},
		{Name: "有效数", Width: 10},
		{Name: "异常数", Width: 10},
	}
	// 设置头部
	for k, v := range sheet1Head {
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
		f.SetCellValue(sheetName, utility.ConvertNumToChar(5)+gconv.String(n), v.TotalCnt)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(6)+gconv.String(n), v.MeterCnt)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(7)+gconv.String(n), v.ErrCnt)
		n++
	}
	// 根据指定路径保存文件
	if err = f.SaveAs(excelFilepath); err != nil {
		utility.Errorf("excel保存异常:%v", err)
	}
	fmt.Printf("excel文件生成完毕，地址:%s\n", excelFilepath)
	return
}
