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
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const luaUpdateAPIKeyStatsV2 = `
local key = KEYS[1]
local field = KEYS[2]
local delta = cjson.decode(ARGV[1])
local expire = tonumber(ARGV[2])

local current = redis.call('HGET', key, field)

if current then
	local record = cjson.decode(current)
	record.callCount = record.callCount + delta.callCount
	record.callFailure = record.callFailure + delta.callFailure
	record.streamCount = record.streamCount + delta.streamCount
	record.nonStreamCount = record.nonStreamCount + delta.nonStreamCount
	record.streamFailure = record.streamFailure + delta.streamFailure
	record.nonStreamFailure = record.nonStreamFailure + delta.nonStreamFailure
	record.firstTokenLatency = record.firstTokenLatency + delta.firstTokenLatency
	record.costs = record.costs + delta.costs
	redis.call('HSET', key, field, cjson.encode(record))
else
	redis.call('HSET', key, field, ARGV[1])
end

redis.call('EXPIRE', key, expire)
return 1
`

type APIKeyRecordV2Stats struct {
	APIKeyID          string `json:"apiKeyId"`
	MethodPath        string `json:"methodPath"`
	CallCount         int32  `json:"callCount"`
	CallFailure       int32  `json:"callFailure"`
	StreamCount       int32  `json:"streamCount"`
	NonStreamCount    int32  `json:"nonStreamCount"`
	StreamFailure     int32  `json:"streamFailure"`
	NonStreamFailure  int32  `json:"nonStreamFailure"`
	FirstTokenLatency int64  `json:"firstTokenLatency"`
	Costs             int64  `json:"costs"`
}

func getRedisAPIKeyStatsV2Key(date string) string {
	return fmt.Sprintf("statistic|api|%s", date)
}

func getRedisAPIKeyStatsV2Field(apiKeyId, userId, orgId, methodPath string) string {
	return fmt.Sprintf("%s|%s|%s|%s", apiKeyId, userId, orgId, methodPath)
}

func parseRedisAPIKeyStatsV2Field(field string) (apiKeyId, userId, orgId, methodPath string, ok bool) {
	parts := strings.Split(field, "|")
	if len(parts) != 4 {
		return "", "", "", "", false
	}
	return parts[0], parts[1], parts[2], parts[3], true
}

func getRedisAPIKeyStatsV2Value(value string) (*APIKeyRecordV2Stats, error) {
	record := &APIKeyRecordV2Stats{}
	if err := json.Unmarshal([]byte(value), record); err != nil {
		return nil, fmt.Errorf("unmarshal api key v2 value %v err: %v", value, err)
	}
	return record, nil
}

// RecordAPIKeyStatisticV2 记录 API Key 统计 V2（先 Redis 日聚合，再写 api_key_record_v2 明细）。
func (c *Client) RecordAPIKeyStatisticV2(ctx context.Context, req *RecordAPIKeyStatisticV2Input) *errs.Status {
	if req == nil || req.APIKeyID == "" || req.UserID == "" {
		return toErrStatus("app_api_key_statistic_v2_record", "invalid request: apiKeyId and userId required")
	}
	if err := recordAPIKeyStatisticV2ToRedis(ctx, req); err != nil {
		log.Errorf("record api key statistic v2 to redis apiKeyId=%s err: %v", req.APIKeyID, err)
	}
	if err := c.db.WithContext(ctx).Create(&model.APIKeyRecordV2{
		// 与 App/Model 一致：明细 CreatedAt 取记录时刻（GORM auto），不沿用请求开始时间，避免跨日分桶与聚合表错位。
		CreatedAt:           time.Now().UnixMilli(),
		OrgID:               req.OrgID,
		UserID:              req.UserID,
		APIKeyID:            req.APIKeyID,
		MethodPath:          req.MethodPath,
		TraceID:             req.TraceID,
		Source:              req.Source,
		Module:              req.Module,
		ModuleCreatorUserID: req.ModuleCreatorUserID,
		ModuleCreatorOrgID:  req.ModuleCreatorOrgID,
		AppID:               req.AppID,
		AppType:             req.AppType,
		IsStream:            req.IsStream,
		Costs:               req.Costs,
		FirstTokenLatency:   req.FirstTokenLatency,
		StatusCode:          req.StatusCode,
		FailureReason:       req.FailureReason,
		RequestBody:         req.RequestBody,
		ResponseBody:        req.ResponseBody,
	}).Error; err != nil {
		log.Errorf("create api key record v2 apiKeyId=%s err: %v", req.APIKeyID, err)
	}
	return nil
}

func recordAPIKeyStatisticV2ToRedis(ctx context.Context, req *RecordAPIKeyStatisticV2Input) error {
	// 与 App/Model 一致：按记录时刻（而非请求开始时间）分日桶，避免跨午夜长调用与聚合表不在同一天。
	today := util.Time2Date(time.Now().UnixMilli())
	key := getRedisAPIKeyStatsV2Key(today)
	field := getRedisAPIKeyStatsV2Field(req.APIKeyID, req.UserID, req.OrgID, req.MethodPath)

	callFailure, streamCount, streamFailure, nonStreamCount, nonStreamFailure := buildCallCounters(req.StatusCode, req.IsStream)
	ftl, costs := statisticCostsForAvg(req.StatusCode, req.FirstTokenLatency, req.Costs)
	delta := &APIKeyRecordV2Stats{
		APIKeyID:          req.APIKeyID,
		MethodPath:        req.MethodPath,
		CallCount:         1,
		CallFailure:       int32(callFailure),
		StreamCount:       int32(streamCount),
		StreamFailure:     int32(streamFailure),
		NonStreamCount:    int32(nonStreamCount),
		NonStreamFailure:  int32(nonStreamFailure),
		FirstTokenLatency: ftl,
		Costs:             costs,
	}
	deltaJSON, err := json.Marshal(delta)
	if err != nil {
		return fmt.Errorf("marshal api key v2 delta err: %v", err)
	}
	_, err = redis.App().Eval(ctx, luaUpdateAPIKeyStatsV2, []string{key, field}, string(deltaJSON), redisStatsExpireSeconds)
	if err != nil {
		return fmt.Errorf("redis eval api key v2 err: %v", err)
	}
	return nil
}

func syncAPIKeyStatisticV2Stats(ctx context.Context, date string, db *gorm.DB) error {
	key := getRedisAPIKeyStatsV2Key(date)
	resultMap, err := redis.App().HGetAll(ctx, key)
	if err != nil {
		return fmt.Errorf("redis HGetAll key %v failed: %v", key, err)
	}
	for _, item := range resultMap {
		apiKeyId, userId, orgId, methodPath, ok := parseRedisAPIKeyStatsV2Field(item.K)
		if !ok {
			log.Errorf("parse api key v2 field %v failed", item.K)
			continue
		}
		record, err := getRedisAPIKeyStatsV2Value(item.V)
		if err != nil {
			log.Errorf("get api key v2 item %v err: %v", item.K, err)
			continue
		}
		if err := upsertAPIKeyStatisticV2ByRecord(ctx, db, apiKeyId, userId, orgId, methodPath, date, record); err != nil {
			log.Errorf("upsert api key v2 date %v field %v err: %v", date, item.K, err)
		}
	}
	return nil
}

func upsertAPIKeyStatisticV2ByRecord(ctx context.Context, db *gorm.DB, apiKeyId, userId, orgId, methodPath, date string, record *APIKeyRecordV2Stats) error {
	stat := &model.StatisticApiKey{
		OrgID:             orgId,
		UserID:            userId,
		APIKeyID:          apiKeyId,
		MethodPath:        methodPath,
		Date:              date,
		CallCount:         record.CallCount,
		CallFailure:       record.CallFailure,
		StreamCount:       record.StreamCount,
		NonStreamCount:    record.NonStreamCount,
		StreamFailure:     record.StreamFailure,
		NonStreamFailure:  record.NonStreamFailure,
		FirstTokenLatency: record.FirstTokenLatency,
		Costs:             record.Costs,
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "org_id"},
			{Name: "user_id"},
			{Name: "api_key_id"},
			{Name: "method_path"},
			{Name: "date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"call_count", "call_failure", "stream_count", "non_stream_count",
			"stream_failure", "non_stream_failure", "first_token_latency", "costs",
		}),
	}).Create(stat).Error
}
