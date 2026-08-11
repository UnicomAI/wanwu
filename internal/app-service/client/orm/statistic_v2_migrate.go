package orm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	statisticV2LegacyMigratedFlagKey = "v0.6.3_statistic_v2_legacy_migrated"
	statisticV2MigrateBatchSize      = 500
	// 迁移进度状态机：pending=失败待续；running=进行中；done=完成。
	statisticV2MigrateStatusPending = "pending"
	statisticV2MigrateStatusRunning = "running"
	statisticV2MigrateStatusDone    = "done"
)

// statisticV2MigrateProgress 记录迁移进度，序列化为 Metadata.MetaValue。
type statisticV2MigrateProgress struct {
	Status string `json:"status"`  // pending | running | done
	LastID uint32 `json:"last_id"` // 仅 api_key_records 续迁使用
}

func parseStatisticV2MigrateProgress(raw string) (*statisticV2MigrateProgress, error) {
	var p statisticV2MigrateProgress
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, fmt.Errorf("parse migrate progress failed: %w", err)
	}
	if p.Status != statisticV2MigrateStatusPending &&
		p.Status != statisticV2MigrateStatusRunning &&
		p.Status != statisticV2MigrateStatusDone {
		return nil, fmt.Errorf("invalid migrate progress status: %q", p.Status)
	}
	return &p, nil
}

func (p *statisticV2MigrateProgress) encode() (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("encode migrate progress failed: %w", err)
	}
	return string(b), nil
}

func saveStatisticV2MigrateProgress(tx *gorm.DB, status string, lastID uint32) error {
	raw, err := (&statisticV2MigrateProgress{Status: status, LastID: lastID}).encode()
	if err != nil {
		return err
	}
	return tx.Model(&Metadata{}).
		Where(&Metadata{MetaKey: statisticV2LegacyMigratedFlagKey}).
		Update("value", raw).Error
}

// migrateStatisticV2FromLegacy 在启动时将旧统计表幂等迁入 V2，保证多副本安全。
//
// 抢锁：首启 Create+OnConflict{DoNothing}（RowsAffected=1 抢到）；
// 续迁 CAS pending→running（WHERE value=旧值，RowsAffected=1 抢到）。
// done 放行；running 报错（别的副本在跑或崩溃残留，需人工置 pending）。
//
// 迁移顺序：三个聚合表（upsert 幂等）→ 明细表（按 last_id 续迁）。
// lastID==0 时从头跑聚合+明细；>0 跳过聚合仅续迁明细。
// 优雅失败置 pending 保留 last_id；进程崩溃留 running 需人工干预。
func migrateStatisticV2FromLegacy(ctx context.Context, db *gorm.DB) error {
	var lastID uint32

	var meta Metadata
	err := db.Where(&Metadata{MetaKey: statisticV2LegacyMigratedFlagKey}).First(&meta).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("query migrate progress failed: %w", err)
		}
		// 首启抢锁：Create+OnConflict{DoNothing}，RowsAffected=1 抢到。
		raw, err := (&statisticV2MigrateProgress{Status: statisticV2MigrateStatusRunning}).encode()
		if err != nil {
			return err
		}
		res := db.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&Metadata{MetaKey: statisticV2LegacyMigratedFlagKey, MetaValue: raw})
		if res.Error != nil {
			return fmt.Errorf("init migrate progress failed: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("statistic v2 migrate already running")
		}
	} else {
		lc, err := parseStatisticV2MigrateProgress(meta.MetaValue)
		if err != nil {
			return err
		}
		switch lc.Status {
		case statisticV2MigrateStatusDone:
			return nil
		case statisticV2MigrateStatusRunning:
			return fmt.Errorf("statistic v2 migrate already running")
		}
		// pending → CAS：WHERE value=旧值，RowsAffected=0 说明被别的副本抢先。
		lc.Status = statisticV2MigrateStatusRunning
		raw, err := lc.encode()
		if err != nil {
			return err
		}
		res := db.Model(&Metadata{}).
			Where("meta_key = ? AND value = ?", statisticV2LegacyMigratedFlagKey, meta.MetaValue).
			Update("value", raw)
		if res.Error != nil {
			return fmt.Errorf("claim migrate progress failed: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("statistic v2 migrate already running")
		}
		lastID = lc.LastID
	}

	// 聚合三步（lastID>0 时跳过，仅续迁明细）。
	if lastID == 0 {
		for _, step := range []struct {
			name string
			fn   func(context.Context, *gorm.DB) (int64, error)
		}{
			{"model_statistics", migrateLegacyModelStats},
			{"app_statistics", migrateLegacyAppStats},
			{"api_key_statistics", migrateLegacyAPIKeyStats},
		} {
			total, err := step.fn(ctx, db)
			if err != nil {
				if saveErr := saveStatisticV2MigrateProgress(db, statisticV2MigrateStatusPending, 0); saveErr != nil {
					log.Errorf("statistic v2 migrate %s failed (%v) and mark pending also failed: %v", step.name, err, saveErr)
				}
				return err
			}
			log.Infof("statistic v2 migrate %s done: %d rows", step.name, total)
		}
	} else {
		log.Infof("statistic v2 migrate resume from last_id=%d, skip aggregate tables", lastID)
	}

	// 明细续迁，每批「插入+写游标」同事务原子。
	lastID, err = migrateLegacyAPIKeyRecords(ctx, db, lastID, func(tx *gorm.DB, batchLastID uint32) error {
		return saveStatisticV2MigrateProgress(tx, statisticV2MigrateStatusRunning, batchLastID)
	})
	if err != nil {
		if saveErr := saveStatisticV2MigrateProgress(db, statisticV2MigrateStatusPending, lastID); saveErr != nil {
			log.Errorf("statistic v2 migrate api_key_records failed (%v) and mark pending also failed: %v", err, saveErr)
		}
		return err
	}
	return saveStatisticV2MigrateProgress(db, statisticV2MigrateStatusDone, lastID)
}

func migrateLegacyModelStats(ctx context.Context, db *gorm.DB) (int64, error) {
	if !db.Migrator().HasTable(&model.LegacyModelStatistic{}) {
		log.Infof("statistic v2 migrate skip model: table model_statistics not found")
		return 0, nil
	}
	var total int64
	var lastID uint32
	for {
		var rows []model.LegacyModelStatistic
		if err := db.WithContext(ctx).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(statisticV2MigrateBatchSize).
			Find(&rows).Error; err != nil {
			return total, fmt.Errorf("read model_statistics failed: %w", err)
		}
		if len(rows) == 0 {
			break
		}
		batch := make([]*model.StatisticModel, 0, len(rows))
		for _, row := range rows {
			batch = append(batch, mapLegacyModelStat(row))
			lastID = row.ID
		}
		if err := db.WithContext(ctx).Clauses(clause.OnConflict{
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
				"provider", "model", "model_type",
				"model_creator_user_id", "model_creator_org_id",
				"prompt_tokens", "completion_tokens", "total_tokens",
				"first_token_latency", "costs",
				"call_count", "stream_count", "non_stream_count",
				"call_failure", "stream_failure", "non_stream_failure",
			}),
		}).CreateInBatches(batch, statisticV2MigrateBatchSize).Error; err != nil {
			return total, fmt.Errorf("upsert statistic_models lastID=%d failed: %w", lastID, err)
		}
		total += int64(len(batch))
	}
	return total, nil
}

func mapLegacyModelStat(row model.LegacyModelStatistic) *model.StatisticModel {
	return &model.StatisticModel{
		ModelID:             row.ModelID,
		Date:                row.Date,
		UserID:              row.UserID,
		OrgID:               row.OrgID,
		Source:              constant.BizSourceWeb,
		Module:              constant.BizModuleModel,
		AppID:               "",
		AppType:             "",
		APIKey:              "",
		MethodPath:          "",
		ModuleCreatorUserID: row.UserID,
		ModuleCreatorOrgID:  row.OrgID,
		Provider:            row.Provider,
		Model:               row.Model,
		ModelType:           row.ModelType,
		ModelCreatorUserID:  row.UserID,
		ModelCreatorOrgID:   row.OrgID,
		PromptTokens:        row.PromptTokens,
		CompletionTokens:    row.CompletionTokens,
		TotalTokens:         row.TotalTokens,
		FirstTokenLatency:   row.FirstTokenLatency,
		Costs:               row.Costs,
		CallCount:           row.CallCount,
		StreamCount:         row.StreamCount,
		NonStreamCount:      row.NonStreamCount,
		CallFailure:         row.CallFailure,
		StreamFailure:       row.StreamFailure,
		NonStreamFailure:    row.NonStreamFailure,
	}
}

func migrateLegacyAppStats(ctx context.Context, db *gorm.DB) (int64, error) {
	if !db.Migrator().HasTable(&model.LegacyAppStatistic{}) {
		log.Infof("statistic v2 migrate skip app: table app_statistics not found")
		return 0, nil
	}
	var total int64
	var lastID uint32
	for {
		var rows []model.LegacyAppStatistic
		if err := db.WithContext(ctx).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(statisticV2MigrateBatchSize).
			Find(&rows).Error; err != nil {
			return total, fmt.Errorf("read app_statistics failed: %w", err)
		}
		if len(rows) == 0 {
			break
		}
		batch := make([]*model.StatisticApp, 0, len(rows))
		for _, row := range rows {
			batch = append(batch, mapLegacyAppStat(row))
			lastID = row.ID
		}
		if err := db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "org_id"},
				{Name: "user_id"},
				{Name: "source"},
				{Name: "module"},
				{Name: "app_id"},
				{Name: "app_type"},
				{Name: "module_creator_user_id"},
				{Name: "module_creator_org_id"},
				{Name: "date"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"call_count", "call_failure", "stream_count", "non_stream_count",
				"stream_failure", "non_stream_failure",
				"first_token_latency", "costs",
			}),
		}).CreateInBatches(batch, statisticV2MigrateBatchSize).Error; err != nil {
			return total, fmt.Errorf("upsert statistic_apps lastID=%d failed: %w", lastID, err)
		}
		total += int64(len(batch))
	}
	return total, nil
}

func mapLegacyAppStat(row model.LegacyAppStatistic) *model.StatisticApp {
	return &model.StatisticApp{
		OrgID:               row.OrgID,
		UserID:              row.UserID,
		Source:              constant.BizSourceWeb,
		Module:              appTypeToModule(row.AppType),
		AppID:               row.AppID,
		AppType:             row.AppType,
		ModuleCreatorUserID: row.UserID,
		ModuleCreatorOrgID:  row.OrgID,
		Date:                row.Date,
		CallCount:           row.CallCount,
		CallFailure:         row.CallFailure,
		StreamCount:         row.StreamCount,
		NonStreamCount:      row.NonStreamCount,
		StreamFailure:       row.StreamFailure,
		NonStreamFailure:    row.NonStreamFailure,
		FirstTokenLatency:   row.StreamCosts,
		Costs:               row.NonStreamCosts,
	}
}

func appTypeToModule(appType string) string {
	switch appType {
	case constant.AppTypeAgent:
		return constant.BizModuleAppAgent
	case constant.AppTypeRag:
		return constant.BizModuleAppRag
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		return constant.BizModuleAppWorkflow
	default:
		return appType
	}
}

func migrateLegacyAPIKeyStats(ctx context.Context, db *gorm.DB) (int64, error) {
	if !db.Migrator().HasTable(&model.LegacyAPIKeyStatistic{}) {
		log.Infof("statistic v2 migrate skip api key stat: table api_key_statistics not found")
		return 0, nil
	}
	var total int64
	var lastID uint32
	for {
		var rows []model.LegacyAPIKeyStatistic
		if err := db.WithContext(ctx).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(statisticV2MigrateBatchSize).
			Find(&rows).Error; err != nil {
			return total, fmt.Errorf("read api_key_statistics failed: %w", err)
		}
		if len(rows) == 0 {
			break
		}
		batch := make([]*model.StatisticApiKey, 0, len(rows))
		for _, row := range rows {
			batch = append(batch, mapLegacyAPIKeyStat(row))
			lastID = row.ID
		}
		if err := db.WithContext(ctx).Clauses(clause.OnConflict{
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
		}).CreateInBatches(batch, statisticV2MigrateBatchSize).Error; err != nil {
			return total, fmt.Errorf("upsert statistic_api_keys lastID=%d failed: %w", lastID, err)
		}
		total += int64(len(batch))
	}
	return total, nil
}

func mapLegacyAPIKeyStat(row model.LegacyAPIKeyStatistic) *model.StatisticApiKey {
	return &model.StatisticApiKey{
		OrgID:             row.OrgID,
		UserID:            row.UserID,
		APIKeyID:          row.APIKeyID,
		MethodPath:        row.MethodPath,
		Date:              row.Date,
		CallCount:         row.CallCount,
		CallFailure:       row.CallFailure,
		StreamCount:       row.StreamCount,
		NonStreamCount:    row.NonStreamCount,
		StreamFailure:     row.StreamFailure,
		NonStreamFailure:  row.NonStreamFailure,
		FirstTokenLatency: row.StreamCosts,
		Costs:             row.NonStreamCosts,
	}
}

// migrateLegacyAPIKeyRecords 从 lastID 续迁明细，每批「插入+写游标」同事务原子。
// 失败时返回最后一个已提交批次的 lastID，保证游标与 DB 一致。
func migrateLegacyAPIKeyRecords(ctx context.Context, db *gorm.DB, lastID uint32, onBatch func(tx *gorm.DB, lastID uint32) error) (uint32, error) {
	if !db.Migrator().HasTable(&model.LegacyAPIKeyRecord{}) {
		log.Infof("statistic v2 migrate skip api key record: table api_key_records not found")
		return lastID, nil
	}
	for {
		var rows []model.LegacyAPIKeyRecord
		if err := db.WithContext(ctx).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(statisticV2MigrateBatchSize).
			Find(&rows).Error; err != nil {
			return lastID, fmt.Errorf("read api_key_records failed: %w", err)
		}
		if len(rows) == 0 {
			break
		}
		batch := make([]*model.APIKeyRecordV2, 0, len(rows))
		batchLastID := lastID
		for _, row := range rows {
			batch = append(batch, mapLegacyAPIKeyRecord(row))
			batchLastID = row.ID
		}
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.CreateInBatches(batch, statisticV2MigrateBatchSize).Error; err != nil {
				return fmt.Errorf("insert api_key_record_v2 batch lastID=%d failed: %w", batchLastID, err)
			}
			return onBatch(tx, batchLastID)
		}); err != nil {
			return lastID, err
		}
		lastID = batchLastID
	}
	return lastID, nil
}

func mapLegacyAPIKeyRecord(row model.LegacyAPIKeyRecord) *model.APIKeyRecordV2 {
	createdAt := row.CallTime
	if createdAt == 0 {
		createdAt = row.CreatedAt
	}
	updatedAt := row.UpdatedAt
	if updatedAt == 0 {
		updatedAt = createdAt
	}
	var costs, firstTokenLatency int64
	if row.IsStream {
		firstTokenLatency = row.StreamCosts
	} else {
		costs = row.NonStreamCosts
	}
	return &model.APIKeyRecordV2{
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		OrgID:               row.OrgID,
		UserID:              row.UserID,
		APIKeyID:            row.APIKeyID,
		MethodPath:          row.MethodPath,
		TraceID:             "",
		Source:              constant.BizSourceOpenAPI,
		Module:              "",
		ModuleCreatorUserID: "",
		ModuleCreatorOrgID:  "",
		AppID:               "",
		AppType:             "",
		IsStream:            row.IsStream,
		Costs:               costs,
		FirstTokenLatency:   firstTokenLatency,
		// 旧表无调用状态可迁，统一按成功计：查询侧 IsSuccess 与聚合 CASE WHEN 均以 200 为成功口径。
		StatusCode:    200,
		FailureReason: "",
		RequestBody:   row.RequestBody,
		ResponseBody:  row.ResponseBody,
	}
}
