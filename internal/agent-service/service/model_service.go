package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/http"
	service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	successCode = 0
)

// SearchModel 查询model信息
func SearchModel(ctx context.Context, modelId string) (*service_model.ModelInfo, error) {
	bffServer := config.GetConfig().BffServer
	url := bffServer.Endpoint + bffServer.SearchModelUri + modelId
	result, err := http.GetClient().Get(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Timeout:    time.Duration(bffServer.Timeout) * time.Second,
		MonitorKey: "search_model",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp service_model.BffResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if resp.Code != successCode {
		if strings.Contains(resp.Msg, "模型查询异常:record not found") {
			return nil, errors.New("智能体模型查询失败，请检查模型是否已下线，重新选择模型。")
		}
		return nil, errors.New(resp.Msg)
	}
	return resp.Data, nil
}
