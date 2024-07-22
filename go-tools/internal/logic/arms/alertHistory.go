package arms

import (
	"context"
	"fmt"
	arms20190808 "github.com/alibabacloud-go/arms-20190808/v5/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
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
	FirstHandleDesc   string      //首次处理描述
	NotifyRobots      string      //通知机器人
	DispatchRuleName  string      //通知策略名称
	DispatchRuleId    string      //通知策略ID
	CreateTime        *gtime.Time //创建时间
	ClaimTime         int32       //首次认领耗时(分钟)
	HandleCostTime    int32       //首次处理耗时(分钟)
	RecoverCostTime   int32       //恢复耗时(分钟)
}

var stateMap = map[int64]string{
	0: "待处理",
	1: "处理中",
	2: "已处理",
}

func (s *sArms) ExportAlertHistory(ctx context.Context, parse *gcmd.Parser) {
	var (
		alerts []*Alert
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

	//开始时间为当前时间2小时前整点小时时间
	startTime := gtime.Now().Add(-2 * time.Hour).Format("Y-m-d H:00:00")
	//截止时间为当前时间1小时前整点小时时间
	endTime := gtime.Now().Add(-1 * time.Hour).Format("Y-m-d H:00:00")

	list, err := s.getAllAlertHistory(regionId, startTime, endTime)
	if err != nil {
		utility.Errorf("获取告警历史异常:%v", err)
		return
	}

	//获取联系人组数据
	contactGroups, err := s.getAllContactGroups(regionId, true)
	if err != nil {
		utility.Errorf("查询联系人组接口调用异常:%v", err)
		return
	}
	contactGroupMap := map[int64]*arms20190808.DescribeContactGroupsResponseBodyPageBeanAlertContactGroups{}
	for _, v := range contactGroups {
		contactGroupMap[int64(*v.ContactGroupId)] = v
	}

	//获取通知策略数据
	notifies, err := s.getAllNotificationPolicies(regionId, true)
	if err != nil {
		utility.Errorf("查询通知策略接口调用异常:%v", err)
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
		//过滤 todo
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
			FirstHandleDesc:   "",
			NotifyRobots:      gconv.String(v.NotifyRobots),
			DispatchRuleName:  gconv.String(v.DispatchRuleName),
			DispatchRuleId:    gconv.String(v.DispatchRuleId),
			CreateTime:        gtime.NewFromStr(*v.CreateTime),
		}
		//恢复时间
		if v.RecoverTime != nil {
			a.RecoverTime = a.CreateTime.Add(time.Duration(*v.RecoverTime) * time.Second)
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
		}
		//获取关闭信息
		if activeMap[4].Len() > 0 {
			v2, _ := activeMap[4].Reverse().Get(0) //倒序排序获取第一个
			a.FirstHandleTime = gtime.NewFromStr(*v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).Time)
			a.FirstHandleMember = *v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).HandlerName
			a.FirstHandleDesc = *v2.(*arms20190808.ListAlertsResponseBodyPageBeanListAlertsActivities).Description
		}

		alerts = append(alerts, a)
	}

	g.Dump(alerts[0:2])
	//todo 追加到excel中，每周生成一个新的excel
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
	pageCnt = int64(math.Ceil(float64(total / size)))
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
