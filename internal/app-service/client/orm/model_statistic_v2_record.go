package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	luaUpdateStatisticModel = `
local key = KEYS[1]
local field = KEYS[2]
local delta = cjson.decode(ARGV[1])
local expire = tonumber(ARGV[2])

local current = redis.call('HGET', key, field)

if current then
	local record = cjson.decode(current)
	record.promptTokens = record.promptTokens + delta.promptTokens
	record.completionTokens = record.completionTokens + delta.completionTokens
	record.totalTokens = record.totalTokens + delta.totalTokens
	record.firstTokenLatency = record.firstTokenLatency + delta.firstTokenLatency
	record.costs = record.costs + delta.costs
	record.callCount = record.callCount + delta.callCount
	record.callFailure = record.callFailure + delta.callFailure
	record.streamCount = record.streamCount + delta.streamCount
	record.streamFailure = record.streamFailure + delta.streamFailure
	record.nonStreamCount = record.nonStreamCount + delta.nonStreamCount
	record.nonStreamFailure = record.nonStreamFailure + delta.nonStreamFailure
	redis.call('HSET', key, field, cjson.encode(record))
else
	redis.call('HSET', key, field, ARGV[1])
end

redis.call('EXPIRE', key, expire)
return 1
`
)

// StatisticModelRecordStats Redis 聚合值
type StatisticModelRecordStats struct {
	Model              string `json:"model"`
	Provider           string `json:"provider"`
	ModelType          string `json:"modelType"`
	ModelCreatorUserID string `json:"modelCreatorUserId"`
	ModelCreatorOrgID  string `json:"modelCreatorOrgId"`
	PromptTokens       int64  `json:"promptTokens"`
	CompletionTokens   int64  `json:"completionTokens"`
	TotalTokens        int64  `json:"totalTokens"`
	CallCount          int32  `json:"callCount"`
	CallFailure        int32  `json:"callFailure"`
	StreamCount        int32  `json:"streamCount"`
	NonStreamCount     int32  `json:"nonStreamCount"`
	StreamFailure      int32  `json:"streamFailure"`
	NonStreamFailure   int32  `json:"nonStreamFailure"`
	FirstTokenLatency  int64  `json:"firstTokenLatency"`
	Costs              int64  `json:"costs"`
}

// RecordModelStatisticV2 记录新版模型统计（Redis 聚合 + 明细表）。
// 写入顺序与 V1 API Key 一致：先 Redis 聚合，再 MySQL 明细；任一步失败仅打日志，不阻断调用方。
func (c *Client) RecordModelStatisticV2(ctx context.Context, req *RecordModelStatisticV2Input) *errs.Status {
	if req == nil || req.ModelID == "" {
		return toErrStatus("app_model_statistic_v2_record", "invalid request: modelId required")
	}
	if err := recordStatisticModelToRedis(ctx, req); err != nil {
		log.Errorf("record statistic model to redis traceId=%s modelId=%s err: %v", req.TraceID, req.ModelID, err)
	}
	if err := c.db.WithContext(ctx).Create(buildModelCallRecord(req)).Error; err != nil {
		log.Errorf("create model call record traceId=%s modelId=%s err: %v", req.TraceID, req.ModelID, err)
	}
	return nil
}

func recordStatisticModelToRedis(ctx context.Context, req *RecordModelStatisticV2Input) error {
	today := util.Time2Date(time.Now().UnixMilli())
	key := getStatisticModelRedisKey(today)
	field := getStatisticModelRedisField(
		req.ModelID, req.UserID, req.OrgID,
		req.Source, req.Module, req.AppID, req.AppType,
		req.APIKey,
		req.MethodPath,
		req.ModuleCreatorUserID, req.ModuleCreatorOrgID,
	)

	callFailure, streamCount, streamFailure, nonStreamCount, nonStreamFailure := buildCallCounters(req.StatusCode, req.IsStream)
	ftl, costs := statisticCostsForAvg(req.StatusCode, req.FirstTokenLatency, req.Costs)

	delta := &StatisticModelRecordStats{
		Model:              req.Model,
		Provider:           req.Provider,
		ModelType:          req.ModelType,
		ModelCreatorUserID: req.ModelCreatorUserID,
		ModelCreatorOrgID:  req.ModelCreatorOrgID,
		PromptTokens:       req.PromptTokens,
		CompletionTokens:   req.CompletionTokens,
		TotalTokens:        req.TotalTokens,
		CallCount:          1,
		CallFailure:        int32(callFailure),
		StreamCount:        int32(streamCount),
		StreamFailure:      int32(streamFailure),
		NonStreamCount:     int32(nonStreamCount),
		NonStreamFailure:   int32(nonStreamFailure),
		FirstTokenLatency:  ftl,
		Costs:              costs,
	}
	deltaJSON, err := json.Marshal(delta)
	if err != nil {
		return fmt.Errorf("marshal statistic model delta err: %v", err)
	}
	_, err = redis.App().Eval(ctx, luaUpdateStatisticModel, []string{key, field}, string(deltaJSON), redisStatsExpireSeconds)
	if err != nil {
		return fmt.Errorf("redis eval statistic model err: %v", err)
	}
	return nil
}

func buildCallCounters(statusCode int64, isStream bool) (callFailure, streamCount, streamFailure, nonStreamCount, nonStreamFailure int) {
	if isStream {
		streamCount = 1
		if !statistic.IsSuccess(statusCode) {
			streamFailure = 1
		}
	} else {
		nonStreamCount = 1
		if !statistic.IsSuccess(statusCode) {
			nonStreamFailure = 1
		}
	}
	if !statistic.IsSuccess(statusCode) {
		callFailure = 1
	}
	return
}

func getStatisticModelRedisKey(date string) string {
	return fmt.Sprintf("statistic|model|%s", date)
}

func getStatisticModelRedisField(modelID, userID, orgID, source, module, appID, appType, apiKey, methodPath, moduleCreatorUserID, moduleCreatorOrgID string) string {
	return strings.Join([]string{
		modelID, userID, orgID, source, module, appID, appType, apiKey, methodPath, moduleCreatorUserID, moduleCreatorOrgID,
	}, "|")
}

func parseStatisticModelRedisField(field string) (modelID, userID, orgID, source, module, appID, appType, apiKey, methodPath, moduleCreatorUserID, moduleCreatorOrgID string, ok bool) {
	parts := strings.Split(field, "|")
	if len(parts) != 11 {
		return "", "", "", "", "", "", "", "", "", "", "", false
	}
	return parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], parts[6], parts[7], parts[8], parts[9], parts[10], true
}

func getStatisticModelRecordStats(value string) (*StatisticModelRecordStats, error) {
	record := &StatisticModelRecordStats{}
	if err := json.Unmarshal([]byte(value), record); err != nil {
		return nil, fmt.Errorf("unmarshal statistic model value %v err: %v", value, err)
	}
	return record, nil
}

func syncStatisticModelStats(ctx context.Context, date string, db *gorm.DB) error {
	key := getStatisticModelRedisKey(date)
	resultMap, err := redis.App().HGetAll(ctx, key)
	if err != nil {
		return fmt.Errorf("redis HGetAll key %v failed: %v", key, err)
	}
	for _, item := range resultMap {
		modelID, userID, orgID, source, module, appID, appType, apiKey, methodPath, moduleCreatorUserID, moduleCreatorOrgID, ok := parseStatisticModelRedisField(item.K)
		if !ok {
			log.Errorf("parse statistic model field %v failed", item.K)
			continue
		}
		record, err := getStatisticModelRecordStats(item.V)
		if err != nil {
			log.Errorf("get statistic model item %v err: %v", item.K, err)
			continue
		}
		if err := upsertStatisticModelByRecord(ctx, db, modelID, userID, orgID, source, module, appID, appType, apiKey, methodPath, moduleCreatorUserID, moduleCreatorOrgID, date, record); err != nil {
			log.Errorf("upsert statistic model date %v field %v err: %v", date, item.K, err)
		}
	}
	return nil
}

func upsertStatisticModelByRecord(ctx context.Context, db *gorm.DB, modelID, userID, orgID, source, module, appID, appType, apiKey, methodPath, moduleCreatorUserID, moduleCreatorOrgID, date string, record *StatisticModelRecordStats) error {
	stat := &model.StatisticModel{
		ModelID:             modelID,
		Date:                date,
		UserID:              userID,
		OrgID:               orgID,
		Source:              source,
		Module:              module,
		AppID:               appID,
		AppType:             appType,
		APIKey:              apiKey,
		MethodPath:          methodPath,
		ModuleCreatorUserID: moduleCreatorUserID,
		ModuleCreatorOrgID:  moduleCreatorOrgID,
		Provider:            record.Provider,
		Model:               record.Model,
		ModelType:           record.ModelType,
		ModelCreatorUserID:  record.ModelCreatorUserID,
		ModelCreatorOrgID:   record.ModelCreatorOrgID,
		PromptTokens:        record.PromptTokens,
		CompletionTokens:    record.CompletionTokens,
		TotalTokens:         record.TotalTokens,
		FirstTokenLatency:   record.FirstTokenLatency,
		Costs:               record.Costs,
		CallCount:           record.CallCount,
		StreamCount:         record.StreamCount,
		NonStreamCount:      record.NonStreamCount,
		CallFailure:         record.CallFailure,
		StreamFailure:       record.StreamFailure,
		NonStreamFailure:    record.NonStreamFailure,
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "model_id"},
			{Name: "date"},
			{Name: "user_id"},
			{Name: "org_id"},
			{Name: "source"},
			{Name: "module"},
			{Name: "app_id"},
			{Name: "app_type"},
			{Name: "api_key"},
			{Name: "method_path"},
			{Name: "module_creator_user_id"},
			{Name: "module_creator_org_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"provider", "model", "model_type", "model_creator_user_id", "model_creator_org_id",
			"prompt_tokens", "completion_tokens", "total_tokens",
			"first_token_latency", "costs", "call_count", "stream_count", "non_stream_count",
			"call_failure", "stream_failure", "non_stream_failure",
		}),
	}).Create(stat).Error
}

func buildModelCallRecord(req *RecordModelStatisticV2Input) *model.ModelRecordV2 {
	return &model.ModelRecordV2{
		TraceID:             req.TraceID,
		OrgID:               req.OrgID,
		UserID:              req.UserID,
		Source:              req.Source,
		Module:              req.Module,
		ModuleCreatorUserID: req.ModuleCreatorUserID,
		ModuleCreatorOrgID:  req.ModuleCreatorOrgID,
		AppID:               req.AppID,
		AppType:             req.AppType,
		ModelID:             req.ModelID,
		Model:               req.Model,
		ModelCreatorUserID:  req.ModelCreatorUserID,
		ModelCreatorOrgID:   req.ModelCreatorOrgID,
		Provider:            req.Provider,
		ModelType:           req.ModelType,
		IsStream:            req.IsStream,
		APIKeyID:            req.APIKeyID,
		APIKey:              req.APIKey,
		MethodPath:          req.MethodPath,
		PromptTokens:        req.PromptTokens,
		CompletionTokens:    req.CompletionTokens,
		TotalTokens:         req.TotalTokens,
		Costs:               req.Costs,
		FirstTokenLatency:   req.FirstTokenLatency,
		StatusCode:          req.StatusCode,
		RequestBody:         req.RequestBody,
		ResponseBody:        req.ResponseBody,
		FinishReason:        req.FinishReason,
		FailureReason:       req.FailureReason,
	}
}
