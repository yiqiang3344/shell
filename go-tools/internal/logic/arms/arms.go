package arms

import (
	"context"
	"fmt"
	arms20190808 "github.com/alibabacloud-go/arms-20190808/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/gogf/gf/v2/os/gcmd"
	"go-tools/internal/service"
	"go-tools/internal/utility"
	"strings"
)

type sArms struct {
	client *arms20190808.Client
}

func New() *sArms {
	return &sArms{}
}

func init() {
	service.RegisterArms(New())
}

func (s *sArms) initClient(ctx context.Context, parse *gcmd.Parser) (err error) {
	endpoint := utility.GetArgString(ctx, parse, "arms.endpoint", "endpoint")
	for {
		if strings.Trim(endpoint, "") == "" {
			endpoint = utility.Scanf("请输入endpoint:")
			continue
		}
		break
	}
	accessKeyId := utility.GetArgString(ctx, parse, "arms.accessKeyId", "accessKeyId")
	for {
		if strings.Trim(accessKeyId, "") == "" {
			accessKeyId = utility.Scanf("请输入accessKeyId:")
			continue
		}
		break
	}
	accessKeySecret := utility.GetArgString(ctx, parse, "arms.accessKeySecret", "accessKeySecret")
	for {
		if strings.Trim(accessKeySecret, "") == "" {
			accessKeySecret = utility.Scanf("请输入accessKeySecret:")
			continue
		}
		break
	}

	s.client, err = arms20190808.NewClient(&openapi.Config{
		AccessKeyId:     &accessKeyId,
		AccessKeySecret: &accessKeySecret,
		Endpoint:        &endpoint,
	})
	return
}

func (s *sArms) getAllNotificationPolicies(RegionId string, isDetail bool) (list []*arms20190808.ListNotificationPoliciesResponseBodyPageBeanNotificationPolicies, err error) {
	var page int64 = 1
	var size int64 = 10000
	runtime := &util.RuntimeOptions{}
	listAlertsRequest := &arms20190808.ListNotificationPoliciesRequest{
		Size:     &size,
		Page:     &page,
		RegionId: &RegionId,
		IsDetail: &isDetail,
	}
	ret, err := s.client.ListNotificationPoliciesWithOptions(listAlertsRequest, runtime)
	if err != nil {
		return
	}
	if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.Total != nil && *ret.Body.PageBean.Total > size {
		err = fmt.Errorf("总数大于%d，请调整size大小", size)
		return
	}
	if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.NotificationPolicies != nil {
		list = append(list, ret.Body.PageBean.NotificationPolicies...)
	}
	return
}

func (s *sArms) getAllContactGroups(regionId string, isDetail bool) (list []*arms20190808.DescribeContactGroupsResponseBodyPageBeanAlertContactGroups, err error) {
	var page int64 = 1
	var size int64 = 10000
	runtime := &util.RuntimeOptions{}
	listAlertsRequest := &arms20190808.DescribeContactGroupsRequest{
		RegionId: &regionId,
		Size:     &size,
		Page:     &page,
		IsDetail: &isDetail,
	}
	ret, err := s.client.DescribeContactGroupsWithOptions(listAlertsRequest, runtime)
	if err != nil {
		return
	}
	if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.Total != nil && *ret.Body.PageBean.Total > size {
		err = fmt.Errorf("总数大于%d，请调整size大小", size)
		return
	}
	if ret.Body != nil && ret.Body.PageBean != nil && ret.Body.PageBean.AlertContactGroups != nil {
		list = append(list, ret.Body.PageBean.AlertContactGroups...)
	}
	return
}
