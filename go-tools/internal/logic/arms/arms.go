package arms

import (
	"context"
	arms20190808 "github.com/alibabacloud-go/arms-20190808/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
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
