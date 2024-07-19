package sls

import (
	"context"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	"github.com/gogf/gf/v2/os/gcmd"
	"go-tools/internal/service"
	"go-tools/internal/utility"
	"strings"
)

type sSls struct {
	client *sls20201230.Client
}

func New() *sSls {
	return &sSls{}
}

func init() {
	service.RegisterSls(New())
}

func (s *sSls) initClient(ctx context.Context, parse *gcmd.Parser) (err error) {
	endpoint := utility.GetArgString(ctx, parse, "sls.endpoint", "endpoint")
	for {
		if strings.Trim(endpoint, "") == "" {
			endpoint = utility.Scanf("请输入endpoint:")
			continue
		}
		break
	}
	accessKeyId := utility.GetArgString(ctx, parse, "sls.accessKeyId", "accessKeyId")
	for {
		if strings.Trim(accessKeyId, "") == "" {
			accessKeyId = utility.Scanf("请输入accessKeyId:")
			continue
		}
		break
	}
	accessKeySecret := utility.GetArgString(ctx, parse, "sls.accessKeySecret", "accessKeySecret")
	for {
		if strings.Trim(accessKeySecret, "") == "" {
			accessKeySecret = utility.Scanf("请输入accessKeySecret:")
			continue
		}
		break
	}

	s.client, err = sls20201230.NewClient(&openapi.Config{
		AccessKeyId:     &accessKeyId,
		AccessKeySecret: &accessKeySecret,
		Endpoint:        &endpoint,
	})
	return
}
