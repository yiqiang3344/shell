package arms

import (
	"context"
	"fmt"
	arms20190808 "github.com/alibabacloud-go/arms-20190808/v5/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xuri/excelize/v2"
	"go-tools/internal/utility"
	"math"
	"strings"
	"time"
)

type Alert struct {
	Id                string      //ID
	Severity          string      //等级
	Name              string      //名称
	State             string      //状态
	RecoverTime       *gtime.Time //恢复时间
	NotifyTime        *gtime.Time //通知时间
	NotifyContent     string      //通知内容
	NotifyObject      string      //通知名单
	FirstClaimMember  string      //首次认领人
	FirstClaimTime    *gtime.Time //首次认领时间
	FirstHandleMember string      //首次处理人
	FirstHandleTime   *gtime.Time //首次处理时间
	Solution          string      //解决方案
	NotifyRobots      string      //通知机器人
	DispatchRuleName  string      //通知策略名称
	DispatchRuleId    string      //通知策略ID
	CreateTime        *gtime.Time //创建时间
	ClaimCostTime     string      //首次认领耗时(分钟)
	HandleCostTime    string      //首次处理耗时(分钟)
	RecoverCostTime   string      //恢复耗时(分钟)
}

var stateMap = map[int64]string{
	0: "待处理",
	1: "处理中",
	2: "已处理",
}

func (s *sArms) ExportAlertHistoryHourCron(ctx context.Context, parse *gcmd.Parser) {
	var (
		err error
	)
	_, err = gcron.AddSingleton(ctx, "0 0 * * * *", func(ctx context.Context) {
		s.ExportAlertHistory(ctx, parse)
	})
	if err != nil {
		panic(err)
	}
	select {}
}

func (s *sArms) ExportAlertHistory(ctx context.Context, parse *gcmd.Parser) {
	var (
		alerts = garray.NewSortedArray(func(a, b interface{}) int {
			if a.(*Alert).Id > b.(*Alert).Id {
				return 1
			} else if a.(*Alert).Id < b.(*Alert).Id {
				return -1
			} else {
				return 0
			}
		})
		sTime     = gtime.Now()
		f         *excelize.File
		rowNum    = 2
		rows      [][]string
		startTime string
		//截止时间为当前时间1小时前整点小时时间
		endTime = gtime.Now().Add(-2 * time.Hour).Format("Y-m-d H:59:59")
		//追加到excel中，每周生成一个新的excel
		excelFilepath = fmt.Sprintf("%s/所有告警记录%s.xlsx", gfile.Pwd(), gtime.NewFromStr(endTime).Format("Y年第W周"))
		sheetName     = "Sheet1"
		err           error
	)

	if err = s.initClient(ctx, parse); err != nil {
		utility.Errorf("%s 初始化客户端异常:%v", sTime.Format("Y-m-d H:i:s"), err)
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

	if gfile.Exists(excelFilepath) {
		//如果excel存在则查询新写入的行数
		f, err = excelize.OpenFile(excelFilepath)
		if err != nil {
			utility.Errorf("%s 打开excel异常:%v", sTime.Format("Y-m-d H:i:s"), err)
			return
		}
		rows, err = f.GetRows("Sheet1")
		if err != nil {
			utility.Errorf("%s 获取excel行数异常:%v", sTime.Format("Y-m-d H:i:s"), err)
			return
		}
		rowNum = len(rows) + 1
	} else {
		f = excelize.NewFile()
		defer func() {
			if err = f.Close(); err != nil {
				utility.Errorf("%s excel关闭异常:%v", sTime.Format("Y-m-d H:i:s"), err)
			}
		}()
		//不存在则设置头部
		type cell struct {
			Name  string
			Width float64
		}
		sheetHead := []cell{
			{Name: "创建时间", Width: 18},
			{Name: "ID", Width: 10},
			{Name: "等级", Width: 5},
			{Name: "名称", Width: 35},
			{Name: "状态", Width: 6},
			{Name: "恢复时间", Width: 18},
			{Name: "通知时间", Width: 18},
			{Name: "通知对象", Width: 50},
			{Name: "首次认领时间", Width: 18},
			{Name: "首次认领人", Width: 10},
			{Name: "首次处理时间", Width: 18},
			{Name: "首次处理人", Width: 12},
			{Name: "解决方案", Width: 20},
			{Name: "恢复耗时", Width: 10},
			{Name: "认领耗时", Width: 10},
			{Name: "处理耗时", Width: 10},
			{Name: "通知机器人", Width: 30},
			{Name: "通知策略", Width: 30},
			{Name: "通知策略ID", Width: 10},
			{Name: "告警内容", Width: 100},
		}
		// 设置头部
		for k, v := range sheetHead {
			f.SetCellValue(sheetName, utility.ConvertNumToChar(k+1)+gconv.String(1), v.Name)
			f.SetColWidth(sheetName, utility.ConvertNumToChar(k+1), utility.ConvertNumToChar(k+1), v.Width)
		}
	}

	startTime = utility.GetArgString(ctx, parse, "arms.startTime", "startTime")
	if strings.Trim(startTime, " ") == "" {
		if rowNum == 2 {
			lastWeekFilePath := fmt.Sprintf("%s/所有告警记录%s.xlsx", gfile.Pwd(), gtime.NewFromStr(endTime).AddDate(0, 0, -1).Format("Y年第W周"))
			if lastWeekFilePath != excelFilepath && gfile.Exists(lastWeekFilePath) {
				//跨周日，且上周有数据，开始时间为当天0点
				startTime = gtime.Now().Format("Y-m-d 00:00:00")
			} else {
				//即不是跨周且上周数据，之前也没有数据，则说明是第一次执行，开始时间为当前时间2天前0点
				startTime = gtime.Now().Add(-2 * 24 * time.Hour).Format("Y-m-d 00:00:00")
			}
		} else {
			//如果有现有数据，开始时间为最后一条数据的创建时间+1秒
			startTime = gtime.NewFromStr(rows[len(rows)-1][0]).Add(1 * time.Second).Format("Y-m-d H:i:s")
		}
	}

	//如果截止时间小于开始时间则退出
	if gtime.NewFromStr(endTime).Before(gtime.NewFromStr(startTime)) {
		utility.Errorf("%s 截止时间%s小于开始时间%s", sTime.Format("Y-m-d H:i:s"), endTime, startTime)
		return
	}

	list, err := s.getAllAlertHistory(regionId, startTime, endTime)
	if err != nil {
		utility.Errorf("%s 获取告警历史异常:%v", sTime.Format("Y-m-d H:i:s"), err)
		return
	}

	//获取联系人组数据
	contactGroups, err := s.getAllContactGroups(regionId, true)
	if err != nil {
		utility.Errorf("%s 查询联系人组接口调用异常:%v", sTime.Format("Y-m-d H:i:s"), err)
		return
	}
	contactGroupMap := map[int64]*arms20190808.DescribeContactGroupsResponseBodyPageBeanAlertContactGroups{}
	for _, v := range contactGroups {
		contactGroupMap[int64(*v.ContactGroupId)] = v
	}

	//获取通知策略数据
	notifies, err := s.getAllNotificationPolicies(regionId, true)
	if err != nil {
		utility.Errorf("%s 查询通知策略接口调用异常:%v", sTime.Format("Y-m-d H:i:s"), err)
		return
	}
	notifyMap := map[int64]string{}
	for _, v := range notifies {
		if utility.InArray(v.NotifyRule.NotifyChannels, tea.String("dingTalk")) {
			continue
		}
		var sArr []string
		for _, v1 := range v.NotifyRule.NotifyObjects {
			switch *v1.NotifyObjectType {
			case "CONTACT":
				sArr = append(sArr, fmt.Sprintf("%s[%s]", *v1.NotifyObjectName, strings.Join(gconv.Strings(v1.NotifyChannels), ",")))
			case "CONTACT_GROUP":
				if v3, ok := contactGroupMap[*v1.NotifyObjectId]; ok {
					for _, v4 := range v3.Contacts {
						sArr = append(sArr, fmt.Sprintf("%s[%s]", *v4.ContactName, strings.Join(gconv.Strings(v1.NotifyChannels), ",")))
					}
				}
			case "CONTACT_SCHEDULE":
			}
		}
		notifyMap[*v.Id] = strings.Join(sArr, ",")
	}

	//整理数据
	for _, v := range list {
		if v.DispatchRuleId == nil {
			continue
		}

		a := &Alert{
			Id:                gconv.String(v.AlertId),
			Severity:          gconv.String(v.Severity),
			Name:              gconv.String(v.AlertName),
			State:             stateMap[*v.State],
			RecoverTime:       nil,
			NotifyTime:        nil,
			NotifyContent:     "",
			NotifyObject:      "",
			FirstClaimMember:  "",
			FirstClaimTime:    nil,
			FirstHandleMember: "",
			FirstHandleTime:   nil,
			Solution:          gconv.String(v.Solution),
			NotifyRobots:      gconv.String(v.NotifyRobots),
			DispatchRuleName:  gconv.String(v.DispatchRuleName),
			DispatchRuleId:    gconv.String(v.DispatchRuleId),
			CreateTime:        gtime.NewFromStr(*v.CreateTime),
			ClaimCostTime:     "",
			RecoverCostTime:   "",
			HandleCostTime:    "",
		}
		//恢复时间
		if v.RecoverTime != nil {
			a.RecoverTime = a.CreateTime.Add(time.Duration(*v.RecoverTime) * time.Second)
			a.RecoverCostTime = fmt.Sprintf("%d分", *v.RecoverTime/60)
		}
		//获取通知、认领、处理信息
		activeMap := map[int64]*garray.Array{
			1: garray.NewArray(false),
			2: garray.NewArray(false),
			3: garray.NewArray(false),
			4: garray.NewArray(false),
			5: garray.NewArray(false),
		}
		for _, v1 := range v.Activities {
			if _, ok := activeMap[*v1.Type]; !ok {
				continue
			}
			activeMap[*v1.Type].Append(v1)
		}
		//获取通知信息
		if activeMap[5].Len() > 0 {
			v2, _ := activeMap[5].Reverse().Get(0) //倒序排序获取第一个
			a.NotifyTime = gtime.NewFromStr(*v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).Time)
			a.NotifyContent = *v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).Content
			if v.DispatchRuleId != nil {
				a.NotifyObject = notifyMap[int64(*v.DispatchRuleId)]
			}
		}
		//获取认领信息
		if activeMap[1].Len() > 0 {
			v2, _ := activeMap[1].Reverse().Get(0) //倒序排序获取第一个
			a.FirstClaimTime = gtime.NewFromStr(*v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).Time)
			a.FirstClaimMember = *v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).HandlerName
			a.ClaimCostTime = fmt.Sprintf("%.0f分", a.FirstClaimTime.Sub(a.CreateTime).Minutes())
		}
		//获取关闭信息
		if activeMap[4].Len() > 0 {
			v2, _ := activeMap[4].Reverse().Get(0) //倒序排序获取第一个
			a.FirstHandleTime = gtime.NewFromStr(*v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).Time)
			a.FirstHandleMember = *v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).HandlerName
			a.HandleCostTime = fmt.Sprintf("%.0f分", a.FirstHandleTime.Sub(a.CreateTime).Minutes())
		}

		alerts.Append(a)
	}

	for k, vv := range alerts.Slice() {
		v := vv.(*Alert)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(1)+gconv.String(rowNum+k), v.CreateTime.Format("Y-m-d H:i:s"))
		f.SetCellValue(sheetName, utility.ConvertNumToChar(2)+gconv.String(rowNum+k), v.Id)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(3)+gconv.String(rowNum+k), v.Severity)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(4)+gconv.String(rowNum+k), v.Name)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(5)+gconv.String(rowNum+k), v.State)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(6)+gconv.String(rowNum+k), v.RecoverTime.Format("Y-m-d H:i:s"))
		f.SetCellValue(sheetName, utility.ConvertNumToChar(7)+gconv.String(rowNum+k), v.NotifyTime.Format("Y-m-d H:i:s"))
		f.SetCellValue(sheetName, utility.ConvertNumToChar(8)+gconv.String(rowNum+k), v.NotifyObject)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(9)+gconv.String(rowNum+k), v.FirstClaimTime.Format("Y-m-d H:i:s"))
		f.SetCellValue(sheetName, utility.ConvertNumToChar(10)+gconv.String(rowNum+k), v.FirstClaimMember)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(11)+gconv.String(rowNum+k), v.FirstHandleTime.Format("Y-m-d H:i:s"))
		f.SetCellValue(sheetName, utility.ConvertNumToChar(12)+gconv.String(rowNum+k), v.FirstHandleMember)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(13)+gconv.String(rowNum+k), v.Solution)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(14)+gconv.String(rowNum+k), v.RecoverCostTime)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(15)+gconv.String(rowNum+k), v.ClaimCostTime)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(16)+gconv.String(rowNum+k), v.HandleCostTime)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(17)+gconv.String(rowNum+k), v.NotifyRobots)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(18)+gconv.String(rowNum+k), v.DispatchRuleName)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(19)+gconv.String(rowNum+k), v.DispatchRuleId)
		f.SetCellValue(sheetName, utility.ConvertNumToChar(20)+gconv.String(rowNum+k), v.NotifyContent)
	}
	// 根据指定路径保存文件
	if err = f.SaveAs(excelFilepath); err != nil {
		utility.Errorf("%s excel保存异常:%v", sTime.Format("Y-m-d H:i:s"), err)
	}
	fmt.Printf("excel文件生成完毕，地址:%s\n", excelFilepath)
	fmt.Printf("%s开始，统计%s到%s的数据，总耗时:%s\n", sTime.Format("Y-m-d H:i:s"), startTime, endTime, utility.FormatDuration(gtime.Now().Sub(sTime)))
}

func (s *sArms) getAllAlertHistory(RegionId string, startTime string, endTime string) (list []*arms20190808.ListAlertsResponseBodyPageBeanListAlerts, err error) {
	var page int64 = 1
	var size int64 = 100
	var total int64
	var pageCnt int64 = 1
	runtime := &util.RuntimeOptions{}
	listAlertsRequest := &arms20190808.ListAlertsRequest{
		Size:           &size,
		Page:           &page,
		RegionId:       &RegionId,
		ShowActivities: tea.Bool(true),
		StartTime:      &startTime,
		EndTime:        &endTime,
	}
	ret, err := s.client.ListAlertsWithOptions(listAlertsRequest, runtime)
	if err != nil {
		return
	}
	if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.ListAlerts != nil {
		list = append(list, ret.Body.PageBean.ListAlerts...)
	}
	total = *ret.Body.PageBean.Total
	pageCnt = int64(math.Ceil(float64(total) / float64(size)))
	page = 2
	for page <= pageCnt {
		listAlertsRequest.Page = &page
		ret, err = s.client.ListAlertsWithOptions(listAlertsRequest, runtime)
		if err != nil {
			return
		}
		if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.ListAlerts != nil {
			list = append(list, ret.Body.PageBean.ListAlerts...)
		}
		page++
	}
	return
}
