// Package assistant 是 operate-service 访问 assistant-service 的 grpc 客户端封装。
//
// 当前仅用于消息中心 actions 一次性清洗：把 notice_messages.actions 中
// msgType=agent 的 actionParams.appId（assistant 自增老 id）批量换为 uuid。
// 抽成接口是为了 orm 层迁移函数可 mock、不必真起 assistant。
package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
)

type IClient interface {
	// GetAppUUIDByOldIDs 批量查老 id（assistant 自增 id 字符串）→ uuid。
	// 查不到的老 id 不在返回 map 中，调用方据此保留原值。
	GetAppUUIDByOldIDs(ctx context.Context, oldIDs []string) (map[string]string, error)
}

type Client struct {
	cli assistant_service.AssistantServiceClient
}

func NewClient(host string) (*Client, error) {
	conn, err := trace_util.NewGrpcTracerConn(host, nil)
	if err != nil {
		return nil, err
	}
	return &Client{cli: assistant_service.NewAssistantServiceClient(conn)}, nil
}

func (c *Client) GetAppUUIDByOldIDs(ctx context.Context, oldIDs []string) (map[string]string, error) {
	resp, err := c.cli.GetAssistantBriefByPrimaryIds(ctx, &assistant_service.GetAssistantBriefByPrimaryIdsReq{
		AssistantPrimaryIds: oldIDs,
	})
	if err != nil {
		return nil, err
	}
	ret := make(map[string]string, len(resp.GetBriefMap()))
	for oldID, brief := range resp.GetBriefMap() {
		if brief == nil || brief.GetInfo() == nil {
			continue
		}
		if uuid := brief.GetInfo().GetAppId(); uuid != "" {
			ret[oldID] = uuid
		}
	}
	return ret, nil
}
