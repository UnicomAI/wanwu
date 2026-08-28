package knowledge_log

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"go.uber.org/zap"
)

// 打印等级 Panic > Error > Warn > Info > Debug

var rpcSLog *zap.SugaredLogger

var rpcLog = RpcLog{}

type RpcLog struct {
}

func init() {
	pkg.AddContainer(rpcLog)
}

func (c RpcLog) LoadType() string {
	return "rpc-log-config"
}

func (c RpcLog) Load() error {
	configInfo := config.GetConfig()
	logConfig := configInfo.RpcLog
	logger, err := log.InitLogCore(logConfig.Std, logConfig.Level, logConfig.Logs...)
	if err != nil {
		return err
	}
	rpcSLog = logger
	return nil
}

func (c RpcLog) StopPriority() int {
	return pkg.DefaultPriority
}

func (c RpcLog) Stop() error {
	return nil
}

func LogRpcJsonNoParams(business string, method string, err error, starTimestamp int64) {
	LogRpcJson(context.Background(), business, method, nil, nil, err, starTimestamp)
}

func LogRpcJson(ctx context.Context, business string, method string, params interface{}, result interface{}, err error, starTimestamp int64) {
	defer func() {
		if err1 := recover(); err1 != nil {
			fmt.Println(err1)
		}
	}()
	var success = 1
	if err != nil {
		success = 0
	}
	var paramsStr = Convert2LogString(params)
	var resultStr = Convert2LogString(result)
	requestId := trace_util.GetTraceID(ctx)
	rpcSLog.Infof("%s|%s|%s|%d|%d|%+v|%+v", requestId, business, method, success, time.Now().UnixMilli()-starTimestamp, paramsStr, resultStr)
}

func Convert2LogString(object interface{}) string {
	if object == nil {
		return "-"
	}
	switch obj := object.(type) {
	case string:
		return obj
	case []byte:
		return string(obj)
	default:
		bytes, err := json.Marshal(object)
		if err != nil {
			return "-"
		}
		return string(bytes)
	}
}
