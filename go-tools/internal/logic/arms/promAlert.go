package arms

import (
	"context"
	"fmt"
	arms20190808 "github.com/alibabacloud-go/arms-20190808/v5/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xuri/excelize/v2"
	"go-tools/internal/utility"
	"math"
	"strings"
)

func (s *sArms) ExportPromAlerts(ctx context.Context, parse *gcmd.Parser) {
	var (
		sTime         = gtime.Now()
		projectsInput = []string{"default"}
		alertsMap     = gmap.NewListMap()
	)

	if err := s.initClient(ctx, parse); err != nil {
		utility.Errorf("初始化客户端异常:%v", err)
		return
	}

	regionId := utility.GetArgString(ctx, parse, "arms.regionId", "regionId")
	for {
		if strings.Trim(regionId, "") == "" {
			regionId = utility.Scanf("请输入regionId:")
			continue
		}
		break
	}

	notificationPolices, err := s.getAllNotificationPolicies(regionId, false)
	if err != nil {
		utility.Errorf("查询通知策略接口调用异常:%v", err)
		return
	}
	notificationPoliceMap := map[string]string{}
	for _, v := range notificationPolices {
		notificationPoliceMap[gconv.String(*v.Id)] = *v.Name
	}

	for _, v := range projectsInput {
		list, err := s.getAllPromAlerts(regionId)
		if err != nil {
			utility.Errorf("查询告警规则接口调用异常:%v", err)
			return
		}
		alertsMap.Set(v, list)
	}

	//输出excel
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			utility.Errorf("excel关闭异常:%v", err)
		}
	}()
	// 创建一个工作表
	excelFilepath := fmt.Sprintf("%s/prom告警规则%s.xlsx", gfile.Pwd(), gtime.Now().Format("YmdHis"))

	type cell struct {
		Name  string
		Width float64
	}
	sheetHead := []cell{
		{Name: "ID", Width: 10},
		{Name: "名称", Width: 60},
		{Name: "实例ID", Width: 35},
		{Name: "状态", Width: 10},
		{Name: "检测类型", Width: 10},
		{Name: "等级", Width: 5},
		{Name: "触发条件", Width: 100},
		{Name: "持续分钟", Width: 8},
		{Name: "通知策略", Width: 30},
		{Name: "标签", Width: 50},
		{Name: "标注", Width: 50},
		{Name: "告警内容", Width: 100},
		{Name: "创建时间", Width: 18},
	}

	hasActive := false
	for kA, vA := range alertsMap.Map() {
		list := vA.([]*arms20190808.GetAlertRulesResponseBodyPageBeanAlertRules)
		sheetName := kA.(string)
		sheetIndex, err := f.NewSheet(sheetName)
		if err != nil {
			utility.Errorf("sheet[%s]创建异常:%v", sheetName, err)
			return
		}
		// 设置头部
		for k, v := range sheetHead {
			f.SetCellValue(sheetName, utility.ConvertNumToChar(k+1)+gconv.String(1), v.Name)
			f.SetColWidth(sheetName, utility.ConvertNumToChar(k+1), utility.ConvertNumToChar(k+1), v.Width)
		}
		// 设置值
		n := 2
		//排序
		sortList := garray.NewSortedArray(func(a, b interface{}) int {
			if *a.(*arms20190808.GetAlertRulesResponseBodyPageBeanAlertRules).CreatedTime > *b.(*arms20190808.GetAlertRulesResponseBodyPageBeanAlertRules).CreatedTime {
				return 1
			} else if *a.(*arms20190808.GetAlertRulesResponseBodyPageBeanAlertRules).CreatedTime < *b.(*arms20190808.GetAlertRulesResponseBodyPageBeanAlertRules).CreatedTime {
				return -1
			} else {
				return 0
			}
		})
		for _, v := range list {
			sortList.Add(v)
		}
		for _, vi := range sortList.Slice() {
			v := vi.(*arms20190808.GetAlertRulesResponseBodyPageBeanAlertRules)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(1)+gconv.String(n), gconv.String(v.AlertId))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(2)+gconv.String(n), gconv.String(v.AlertName))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(3)+gconv.String(n), gconv.String(v.ClusterId))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(4)+gconv.String(n), gconv.String(v.AlertStatus))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(5)+gconv.String(n), gconv.String(v.AlertCheckType))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(6)+gconv.String(n), gconv.String(v.Level))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(7)+gconv.String(n), gconv.String(v.PromQL))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(8)+gconv.String(n), gconv.String(v.Duration))
			if v1, ok := notificationPoliceMap[gconv.String(v.NotifyStrategy)]; ok {
				f.SetCellValue(sheetName, utility.ConvertNumToChar(9)+gconv.String(n), v1)
			}
			sT, _ := gjson.Marshal(v.Tags)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(10)+gconv.String(n), gconv.String(sT))
			sT, _ = gjson.Marshal(v.Annotations)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(11)+gconv.String(n), gconv.String(sT))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(12)+gconv.String(n), gconv.String(v.Message))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(13)+gconv.String(n), gtime.NewFromTimeStamp(*v.CreatedTime).Format("Y/m/d H:i:s"))
			n++
		}
		// 设置工作簿的默认工作表
		if !hasActive {
			f.SetActiveSheet(sheetIndex)
			hasActive = true
		}
	}

	f.DeleteSheet("Sheet1")
	// 根据指定路径保存文件
	if err := f.SaveAs(excelFilepath); err != nil {
		utility.Errorf("excel保存异常:%v", err)
	}
	fmt.Printf("excel文件生成完毕，地址:%s\n", excelFilepath)
	fmt.Printf("总耗时:%s\n", utility.FormatDuration(gtime.Now().Sub(sTime)))
}

func (s *sArms) getAllPromAlerts(RegionId string) (list []*arms20190808.GetAlertRulesResponseBodyPageBeanAlertRules, err error) {
	var page int64 = 1
	var size int64 = 100
	var total int64
	var pageCnt int64 = 1
	runtime := &util.RuntimeOptions{}
	listAlertsRequest := &arms20190808.GetAlertRulesRequest{
		Size:      &size,
		Page:      &page,
		RegionId:  &RegionId,
		AlertType: tea.String("PROMETHEUS_MONITORING_ALERT_RULE"),
	}
	ret, err := s.client.GetAlertRulesWithOptions(listAlertsRequest, runtime)
	if err != nil {
		return
	}
	if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.AlertRules != nil {
		list = append(list, ret.Body.PageBean.AlertRules...)
	}
	total = *ret.Body.PageBean.Total
	pageCnt = int64(math.Ceil(float64(total) / float64(size)))
	page = 2
	for page <= pageCnt {
		listAlertsRequest.Page = &page
		ret, err = s.client.GetAlertRulesWithOptions(listAlertsRequest, runtime)
		if err != nil {
			return
		}
		if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.AlertRules != nil {
			list = append(list, ret.Body.PageBean.AlertRules...)
		}
		page++
	}
	return
}
