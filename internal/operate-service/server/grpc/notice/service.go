// Package notice 是消息中心的 grpc 服务实现。
//
// 消息域复用 operate-service 宿主但保持三隔离：独立 proto service NoticeService
// （与 OperateService 并列双注册）、独立 model/DAO、独立表与索引。
// 未来消息域长大，按此边界整目录平移出去即可。
package notice

import (
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/operate-service/client"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
)

type Service struct {
	operate_service.UnimplementedNoticeServiceServer
	cli client.IClient
}

func NewService(cli client.IClient) *Service {
	return &Service{
		cli: cli,
	}
}

func errStatus(code errs.Code, status *errs.Status) error {
	return grpc_util.ErrorStatusWithKey(code, status.TextKey, status.Args...)
}
