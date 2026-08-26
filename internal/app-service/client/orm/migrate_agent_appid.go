package orm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/app-service/client/assistant"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

const migrateAgentAppIDToUUIDFlagKey = "v0.6.4_app_agent_appid_to_uuid"

// assistant 自增老 id < 10⁹；uuid（snowflake 18 位）远超此值，isOldAppID 自动排除。
const oldAppIDUpperBound = 1000000000

// migrateAgentAppIDToUUID 一次性清洗 app-service 中 agent 的 app_id（assistant 自增老 id → uuid）。
//
// 两阶段执行，共享同一份 uuidMap（老 id → uuid 映射），阶段间无依赖、可任意顺序执行：
//   - MySQL：11 张表 app_id 列覆盖回写，聚合表撞唯一键时合并累加
//   - Redis：统计 Hash field 中的 app_id 替换为 uuid，field 重命名 + value 累加
//
// 失败策略：
//   - 标记已存在 → 直接返回（幂等）
//   - assistant gRPC 调用失败 → 返回 err（启动期硬要求）
//   - 单个老 id 查不到 uuid → 保留原值 + 告警，不阻断
//   - MySQL 聚合表撞唯一键 → 合并累加列到 uuid 行 + 删老行（事务），保证数据不丢
//   - Redis 单 key 失败 → 告警跳过，不阻断整体
func migrateAgentAppIDToUUID(ctx context.Context, db *gorm.DB, assistantCli assistant.IClient) error {
	// 幂等门控
	var meta Metadata
	err := db.Where(&Metadata{MetaKey: migrateAgentAppIDToUUIDFlagKey}).First(&meta).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("query migrate flag failed: %w", err)
	}

	// 扫描 MySQL 各表收集 agent 老 app_id（去重）
	oldIDSet := make(map[string]struct{})
	for _, table := range agentAppIDTables {
		var ids []string
		if err := db.Table(table).
			Where("app_type = ?", constant.AppTypeAgent).
			Distinct("app_id").
			Pluck("app_id", &ids).Error; err != nil {
			return fmt.Errorf("scan old app_id from %s failed: %w", table, err)
		}
		for _, id := range ids {
			if isOldAppID(id) {
				oldIDSet[id] = struct{}{}
			}
		}
	}

	// 批量查老 id → uuid，构建共享 uuidMap
	uuidMap := make(map[string]string)
	if len(oldIDSet) > 0 {
		oldIDs := make([]string, 0, len(oldIDSet))
		for id := range oldIDSet {
			oldIDs = append(oldIDs, id)
		}
		uuidMap, err = assistantCli.GetAppUUIDByOldIDs(ctx, oldIDs)
		if err != nil {
			return fmt.Errorf("query assistant app uuid failed: %w", err)
		}
		for oldID := range oldIDSet {
			if _, ok := uuidMap[oldID]; !ok {
				log.Warnf("migrate_agent_appid: agent old app_id %q has no uuid, kept as-is", oldID)
			}
		}
	}

	if err := migrateMySQLAgentAppID(db, uuidMap); err != nil {
		return err
	}

	if err := migrateRedisStatsAgentAppID(ctx, uuidMap, assistantCli); err != nil {
		return err
	}

	if err := db.Create(&Metadata{MetaKey: migrateAgentAppIDToUUIDFlagKey}).Error; err != nil {
		return fmt.Errorf("write migrate flag failed: %w", err)
	}
	return nil
}

// ===== MySQL =====

// migrateMySQLAgentAppID 清洗 MySQL 11 张表的 agent app_id（老 id → uuid）。
// 普通索引表直接 UPDATE；聚合表唯一键含 app_id，UPDATE 可能撞键，需合并累加 + 删老行。
func migrateMySQLAgentAppID(db *gorm.DB, uuidMap map[string]string) error {
	// 普通索引表：app_id 仅为普通索引，直接 UPDATE 无冲突风险
	for _, table := range agentAppIDTables {
		if isStatisticTable(table) {
			continue
		}
		if err := migratePlainTable(db, table, uuidMap); err != nil {
			return err
		}
	}

	// 聚合表：唯一键含 app_id，UPDATE 可能撞键
	if err := migrateStatisticApp(db, uuidMap); err != nil {
		return err
	}
	if err := migrateStatisticModel(db, uuidMap); err != nil {
		return err
	}
	return nil
}

// agentAppIDTables 所有需清洗 agent app_id 的表名。
// GORM 默认复数命名，但 inflection 不对以数字结尾的词加 's'，
// 所以 *_v2 结尾的表名是单数（app_record_v2 而非 app_record_v2s）。
var agentAppIDTables = []string{
	"apps",
	"app_urls",
	"app_conversations",
	"app_favorites",
	"app_histories",
	"api_keys",
	"app_record_v2",
	"api_key_record_v2",
	"model_record_v2",
	"statistic_apps",   // 聚合表
	"statistic_models", // 聚合表
}

func isStatisticTable(table string) bool {
	return table == "statistic_apps" || table == "statistic_models"
}

func migratePlainTable(db *gorm.DB, table string, uuidMap map[string]string) error {
	for oldID, uuid := range uuidMap {
		if err := db.Table(table).
			Where("app_type = ? AND app_id = ?", constant.AppTypeAgent, oldID).
			Update("app_id", uuid).Error; err != nil {
			log.Errorf("update %s app_id failed old=%s: %w", table, oldID, err)
			continue
		}
	}
	return nil
}

// statistic_apps 累加列
var statisticAppAccumCols = []string{
	"call_count", "call_failure", "stream_count", "non_stream_count",
	"stream_failure", "non_stream_failure", "first_token_latency", "costs",
}

// statistic_models 累加列
var statisticModelAccumCols = []string{
	"prompt_tokens", "completion_tokens", "total_tokens",
	"first_token_latency", "costs",
	"call_count", "stream_count", "non_stream_count",
	"call_failure", "stream_failure", "non_stream_failure",
}

// migrateStatisticApp 清洗 statistic_apps：先 UPDATE，撞唯一键则按分组列合并累加 + 删老行。
// 唯一键（除 app_id）：org_id, user_id, source, module, app_type,
// module_creator_user_id, module_creator_org_id, date
func migrateStatisticApp(db *gorm.DB, uuidMap map[string]string) error {
	groupCols := []string{"org_id", "user_id", "source", "module", "app_type",
		"module_creator_user_id", "module_creator_org_id", "date"}
	for oldID, uuid := range uuidMap {
		if err := migrateStatisticTable(db, "statistic_apps", oldID, uuid, groupCols, statisticAppAccumCols); err != nil {
			return fmt.Errorf("migrate statistic_apps failed old=%s: %w", oldID, err)
		}
	}
	return nil
}

// migrateStatisticModel 清洗 statistic_models：先 UPDATE，撞唯一键则合并累加 + 删老行。
// 唯一键（除 app_id）：model_id, date, user_id, org_id, source, module, app_type,
// api_key, method_path, module_creator_user_id, module_creator_org_id
func migrateStatisticModel(db *gorm.DB, uuidMap map[string]string) error {
	groupCols := []string{"model_id", "date", "user_id", "org_id", "source", "module", "app_type",
		"api_key", "method_path", "module_creator_user_id", "module_creator_org_id"}
	for oldID, uuid := range uuidMap {
		if err := migrateStatisticTable(db, "statistic_models", oldID, uuid, groupCols, statisticModelAccumCols); err != nil {
			return fmt.Errorf("migrate statistic_models failed old=%s: %w", oldID, err)
		}
	}
	return nil
}

// migrateStatisticTable 尝试 UPDATE app_id=oldID→uuid；
// 撞唯一键（Duplicate entry 1062）时按 groupCols 把老行累加列合并到 uuid 行后删除老行。
// MySQL 遇 1062 时整条 UPDATE 回滚（不更新任何行），所以老行仍完整保留供合并。
func migrateStatisticTable(db *gorm.DB, table, oldID, uuid string, groupCols, accumCols []string) error {
	res := db.Table(table).
		Where("app_type = ? AND app_id = ?", constant.AppTypeAgent, oldID).
		Update("app_id", uuid)
	if res.Error == nil {
		return nil
	}
	if !isDuplicateEntry(res.Error) {
		return res.Error
	}

	log.Warnf("migrate_agent_appid: %s duplicate entry on app_id %s → %s, merging", table, oldID, uuid)

	// 查出所有老 id 行，逐行合并到对应 uuid 行
	var oldRows []map[string]interface{}
	if err := db.Table(table).
		Where("app_type = ? AND app_id = ?", constant.AppTypeAgent, oldID).
		Find(&oldRows).Error; err != nil {
		return fmt.Errorf("scan old rows failed: %w", err)
	}

	for _, oldRow := range oldRows {
		if err := mergeStatisticRow(db, table, oldRow, uuid, groupCols, accumCols); err != nil {
			return err
		}
	}
	return nil
}

// mergeStatisticRow 把一行老 id 数据合并到对应 uuid 行：累加列相加，无 uuid 行则改 app_id。
// 每行独立事务，保证单行合并的原子性。
func mergeStatisticRow(db *gorm.DB, table string, oldRow map[string]interface{}, uuid string, groupCols, accumCols []string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 按分组列定位同维度的 uuid 行
		query := tx.Table(table).Where("app_id = ?", uuid)
		for _, col := range groupCols {
			query = query.Where(col+" = ?", oldRow[col])
		}

		var uuidRow map[string]interface{}
		// 用 Take 而非 First:First 会按主键排序,但此处经 Table() 构建无 Schema,
		// 渲染 PrimaryKey 时触发 ErrModelValueRequired;Take 不带主键排序,语义亦仅需取一条。
		err := query.Take(&uuidRow).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 无 uuid 行：直接把这行老 id 改成 uuid（该分组下原本只有老 id 行，无冲突）
			id, ok := oldRow["id"]
			if !ok {
				return fmt.Errorf("old row missing id")
			}
			return tx.Table(table).Where("id = ?", id).Update("app_id", uuid).Error
		}
		if err != nil {
			return fmt.Errorf("query uuid row failed: %w", err)
		}

		// 累加列相加到 uuid 行，然后删除老 id 行
		updates := make(map[string]interface{}, len(accumCols))
		for _, col := range accumCols {
			updates[col] = toInt64(oldRow[col]) + toInt64(uuidRow[col])
		}
		uuidID := uuidRow["id"]
		if err := tx.Table(table).Where("id = ?", uuidID).Updates(updates).Error; err != nil {
			return fmt.Errorf("update uuid row accum cols failed: %w", err)
		}
		if err := tx.Table(table).Where("id = ?", oldRow["id"]).Delete(nil).Error; err != nil {
			return fmt.Errorf("delete old row failed: %w", err)
		}
		return nil
	})
}

func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "Duplicate entry") || strings.Contains(s, "1062")
}

// ===== Redis =====

// redisStatsSpec 描述一组统计 Hash 的迁移规格。
type redisStatsSpec struct {
	keyPrefix  string   // Redis key 前缀，如 "statistic|app|"
	fieldParts int      // pipe 分隔的 field 段数
	appIDIdx   int      // appID 在 field 中的段索引
	accumCols  []string // value JSON 中需累加的字段名（camelCase）
}

// 两组统计 Hash 的迁移规格。appID 分别在 field index 4（app）和 5（model）。
// statistic|api| 的 field 不含 appID，无需迁移。
var redisStatsSpecs = []redisStatsSpec{
	{
		keyPrefix:  "statistic|app|",
		fieldParts: 8,
		appIDIdx:   4,
		accumCols: []string{
			"callCount", "callFailure", "streamCount", "nonStreamCount",
			"streamFailure", "nonStreamFailure", "firstTokenLatency", "costs",
		},
	},
	{
		keyPrefix:  "statistic|model|",
		fieldParts: 11,
		appIDIdx:   5,
		accumCols: []string{
			"promptTokens", "completionTokens", "totalTokens",
			"firstTokenLatency", "costs",
			"callCount", "streamCount", "nonStreamCount",
			"callFailure", "streamFailure", "nonStreamFailure",
		},
	},
}

// migrateRedisStatsAgentAppID 清洗 Redis 统计 Hash field 中的 agent app_id（老 id → uuid）。
//
// Redis 数据结构：key=statistic|{app|model}|{date}，value=Hash{field=维度组合, value=JSON计数器}。
// appID 嵌在 field 的 pipe 分隔段中（不在 key 名也不在 value JSON 中）。
// 迁移操作：HDEL 老 field + HSET 新 field（appID 替换为 uuid）。
// 若新 field 已存在（同维度的 uuid 行已在 Redis），value 累加合并而非覆盖。
//
// uuidMap 复用 MySQL 阶段结果；Redis 中可能存在 MySQL 不有的老 id（Redis 写成功但 MySQL 写失败），
// 差集部分单独调 assistant-service 补查后合并进 uuidMap。
func migrateRedisStatsAgentAppID(ctx context.Context, uuidMap map[string]string, assistantCli assistant.IClient) error {
	if redis.App() == nil {
		return nil
	}

	dates := lastNDates(30)

	// 扫描所有日期 key，缓存 HGetAll 结果，收集 uuidMap 中缺失的老 id
	type cached struct {
		spec  redisStatsSpec
		key   string
		items []redis.HashItem
	}
	var cachedKeys []cached
	missingIDs := make(map[string]struct{})

	for _, spec := range redisStatsSpecs {
		for _, date := range dates {
			key := spec.keyPrefix + date
			items, err := redis.App().HGetAll(ctx, key)
			if err != nil {
				log.Warnf("migrate_agent_appid: redis HGetAll %s failed: %v", key, err)
				continue
			}
			if len(items) == 0 {
				continue
			}
			for _, item := range items {
				parts := strings.Split(item.K, "|")
				if len(parts) != spec.fieldParts {
					continue
				}
				oldID := parts[spec.appIDIdx]
				if !isOldAppID(oldID) {
					continue
				}
				if _, ok := uuidMap[oldID]; !ok {
					missingIDs[oldID] = struct{}{}
				}
			}
			cachedKeys = append(cachedKeys, cached{spec, key, items})
		}
	}

	// 补查差集老 id → uuid，合并进 uuidMap
	if len(missingIDs) > 0 {
		missingList := make([]string, 0, len(missingIDs))
		for id := range missingIDs {
			missingList = append(missingList, id)
		}
		supplement, err := assistantCli.GetAppUUIDByOldIDs(ctx, missingList)
		if err != nil {
			return fmt.Errorf("redis migrate: query assistant app uuid failed: %w", err)
		}
		for k, v := range supplement {
			uuidMap[k] = v
		}
		for id := range missingIDs {
			if _, ok := uuidMap[id]; !ok {
				log.Warnf("migrate_agent_appid: redis old id %q has no uuid, skipped", id)
			}
		}
	}

	// 逐 key 执行 field 重命名 + value 合并
	for _, c := range cachedKeys {
		if err := migrateRedisHashKey(ctx, c.key, c.spec, c.items, uuidMap); err != nil {
			log.Warnf("migrate_agent_appid: redis %s failed: %v", c.key, err)
		}
	}

	return nil
}

// migrateRedisHashKey 处理单个 Redis Hash key 的老 id field 迁移。
//
// 对每个含老 id 的 field：构建新 field（替换 appID 为 uuid），value 与已有 uuid field 累加合并，
// 然后 pipeline 批量 HDEL 旧 field + HSET 新 field。
//
// 合并优先级：已计算的同 newField > Redis 中已有的 uuid field > 直接取老 field 值。
// 多个老 id field 映射到同一 uuid field 时，值逐一累加。
func migrateRedisHashKey(ctx context.Context, key string, spec redisStatsSpec, items []redis.HashItem, uuidMap map[string]string) error {
	// 构建 field → value 查找表，用于判断 uuid field 是否已存在
	existing := make(map[string]string, len(items))
	for _, item := range items {
		existing[item.K] = item.V
	}

	var oldFields []string
	newValues := make(map[string]string)

	for _, item := range items {
		parts := strings.Split(item.K, "|")
		if len(parts) != spec.fieldParts {
			continue
		}
		oldID := parts[spec.appIDIdx]
		if !isOldAppID(oldID) {
			continue
		}
		uuid, ok := uuidMap[oldID]
		if !ok || uuid == "" {
			continue
		}

		// 构建新 field（替换 appID 段为 uuid）
		parts[spec.appIDIdx] = uuid
		newField := strings.Join(parts, "|")
		oldFields = append(oldFields, item.K)

		// value 合并：同一 newField 多次命中时逐一累加
		if cur, ok := newValues[newField]; ok {
			newValues[newField] = mergeJSONStats(cur, item.V, spec.accumCols)
		} else if redisVal, ok := existing[newField]; ok {
			newValues[newField] = mergeJSONStats(redisVal, item.V, spec.accumCols)
		} else {
			newValues[newField] = item.V
		}
	}

	if len(oldFields) == 0 {
		return nil
	}

	// pipeline 批量执行：先 HDEL 旧 field，再 HSET 新 field
	pipe := redis.App().Cli().Pipeline()
	pipe.HDel(ctx, key, oldFields...)
	for newField, val := range newValues {
		pipe.HSet(ctx, key, newField, val)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline exec %s failed: %w", key, err)
	}

	log.Infof("migrate_agent_appid: redis %s migrated %d fields (%d merged)",
		key, len(oldFields), len(newValues))
	return nil
}

// mergeJSONStats 把 src JSON 中的 accumCols 字段累加到 dst JSON，返回合并后的 JSON。
// 非 accumCols 字段取 dst 值（维度信息在相同 newField 下应一致）。
func mergeJSONStats(dst, src string, accumCols []string) string {
	var dstMap, srcMap map[string]interface{}
	if err := json.Unmarshal([]byte(dst), &dstMap); err != nil {
		return dst
	}
	if err := json.Unmarshal([]byte(src), &srcMap); err != nil {
		return dst
	}
	for _, col := range accumCols {
		dstMap[col] = toInt64(dstMap[col]) + toInt64(srcMap[col])
	}
	b, err := json.Marshal(dstMap)
	if err != nil {
		return dst
	}
	return string(b)
}

// lastNDates 返回最近 n 天的日期字符串列表（含今天，YYYY-MM-DD），按时间正序。
// 用 util.Time2Date 保证时区（UTC+8）和格式与写入侧一致。
func lastNDates(n int) []string {
	now := time.Now().UnixMilli()
	dayMs := int64(24 * time.Hour / time.Millisecond)
	dates := make([]string, 0, n)
	for i := n - 1; i >= 0; i-- {
		dates = append(dates, util.Time2Date(now-int64(i)*dayMs))
	}
	return dates
}

// ===== Shared =====

// toInt64 把 interface{}（gorm Find 到 map / JSON unmarshal 到 map 时数值可能为 []byte/string/int64/float64 等）安全转 int64。
func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		i, _ := strconv.ParseInt(string(n), 10, 64)
		return i
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	case nil:
		return 0
	default:
		return 0
	}
}

// isOldAppID 判定是否为 assistant 自增老 id：纯数字且 < 10⁹。
// uuid 无法解析为 int，自动跳过。
func isOldAppID(appID string) bool {
	if appID == "" {
		return false
	}
	n, err := strconv.Atoi(appID)
	if err != nil {
		return false
	}
	return n < oldAppIDUpperBound
}
