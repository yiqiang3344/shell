package arms

import (
	"context"
	"fmt"
	arms20190808 "github.com/alibabacloud-go/arms-20190808/v5/client"
	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
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

	for _, v := range projectsInput {
		list, err := s.getAllPromAlerts()
		if err != nil {
			utility.Errorf("查询接口调用异常:%v", err)
			return
		}
		alertsMap.Set(v, list)
	}
	return

	//输出excel
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			utility.Errorf("excel关闭异常:%v", err)
		}
	}()
	// 创建一个工作表
	excelFilepath := fmt.Sprintf("%s/告警规则%s.xlsx", gfile.Pwd(), gtime.Now().Format("YmdHis"))

	type cell struct {
		Name  string
		Width float64
	}
	sheet1Head := []cell{
		{Name: "ID", Width: 22},
		{Name: "名称", Width: 60},
		{Name: "状态", Width: 10},
		{Name: "检查频率", Width: 15},
		{Name: "tags", Width: 15},
		{Name: "告警阈值条件:严重", Width: 40},
		{Name: "告警阈值条件:高", Width: 40},
		{Name: "告警阈值条件:中", Width: 40},
		{Name: "标签", Width: 100},
		{Name: "标注", Width: 100},
		{Name: "恢复通知", Width: 10},
		{Name: "告警策略", Width: 80},
		{Name: "创建时间", Width: 18},
	}

	hasActive := false
	for kA, vA := range alertsMap.Map() {
		list := vA.([]*sls20201230.Alert)
		sheetName := kA.(string)
		sheetIndex, err := f.NewSheet(sheetName)
		if err != nil {
			utility.Errorf("sheet[%s]创建异常:%v", sheetName, err)
			return
		}
		// 设置头部
		for k, v := range sheet1Head {
			f.SetCellValue(sheetName, utility.ConvertNumToChar(k+1)+gconv.String(1), v.Name)
			f.SetColWidth(sheetName, utility.ConvertNumToChar(k+1), utility.ConvertNumToChar(k+1), v.Width)
		}
		// 设置值
		n := 2
		//排序
		sortList := garray.NewSortedArray(func(a, b interface{}) int {
			if *a.(*sls20201230.Alert).CreateTime > *b.(*sls20201230.Alert).CreateTime {
				return 1
			} else if *a.(*sls20201230.Alert).CreateTime < *b.(*sls20201230.Alert).CreateTime {
				return -1
			} else {
				return 0
			}
		})
		for _, v := range list {
			sortList.Add(v)
		}
		for _, vi := range sortList.Slice() {
			v := vi.(*sls20201230.Alert)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(1)+gconv.String(n), *v.Name)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(2)+gconv.String(n), *v.DisplayName)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(3)+gconv.String(n), *v.Status)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(4)+gconv.String(n), fmt.Sprintf("%s%s", gconv.String(v.Schedule.CronExpression), gconv.String(v.Schedule.Interval)))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(5)+gconv.String(n), strings.Join(gconv.Strings(v.Configuration.Tags), ","))
			m := map[int32]string{}
			for _, v1 := range v.Configuration.SeverityConfigurations {
				if gconv.String(v1.EvalCondition.Condition) == "" && gconv.String(v1.EvalCondition.CountCondition) == "" {
					m[*v1.Severity] = "有数据"
				} else {
					m[*v1.Severity] = fmt.Sprintf("%s%s", gconv.String(v1.EvalCondition.Condition), gconv.String(v1.EvalCondition.CountCondition))
				}
			}
			if v1, ok := m[10]; ok {
				f.SetCellValue(sheetName, utility.ConvertNumToChar(6)+gconv.String(n), v1)
			}
			if v1, ok := m[8]; ok {
				f.SetCellValue(sheetName, utility.ConvertNumToChar(7)+gconv.String(n), v1)
			}
			if v1, ok := m[6]; ok {
				f.SetCellValue(sheetName, utility.ConvertNumToChar(8)+gconv.String(n), v1)
			}
			sT, _ := gjson.Marshal(v.Configuration.Labels)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(9)+gconv.String(n), gconv.String(sT))
			sT, _ = gjson.Marshal(v.Configuration.Annotations)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(10)+gconv.String(n), gconv.String(sT))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(11)+gconv.String(n), *v.Configuration.SendResolved)
			sT, _ = gjson.Marshal(v.Configuration.PolicyConfiguration)
			f.SetCellValue(sheetName, utility.ConvertNumToChar(12)+gconv.String(n), gconv.String(sT))
			f.SetCellValue(sheetName, utility.ConvertNumToChar(13)+gconv.String(n), gtime.NewFromTimeStamp(*v.CreateTime).Format("Y/m/d H:i:s"))
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

func (s *sArms) getAllPromAlerts() (list []*arms20190808.GetAlertRulesResponseBodyPageBeanAlertRules, err error) {
	var page int64 = 1
	var size int64 = 10
	var total int64
	var pageCnt int64 = 1
	runtime := &util.RuntimeOptions{}
	listAlertsRequest := &arms20190808.GetAlertRulesRequest{
		Size:     &size,
		Page:     &page,
		RegionId: tea.String("cn-beijing"),
	}
	ret, err := s.client.GetAlertRulesWithOptions(listAlertsRequest, runtime)
	if err != nil {
		return
	}
	if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.AlertRules != nil {
		list = append(list, ret.Body.PageBean.AlertRules...)
	}
	g.Dump(list)
	return
	total = *ret.Body.PageBean.Total
	pageCnt = int64(math.Ceil(float64(total / size)))
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
