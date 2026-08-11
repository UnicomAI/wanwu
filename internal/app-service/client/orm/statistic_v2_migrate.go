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
	statisticV2LegacyMigratedFlagKey = "v0.6.2_statistic_v2_legacy_migrated"
	statisticV2MigrateBatchSize      = 500
	// 迁移进度状态机：
	// pending=失败待续（仅续迁入口，非首启初态）；running=进行中/首启抢锁目标态；done=完成。
	statisticV2MigrateStatusPending = "pending"
	statisticV2MigrateStatusRunning = "running"
	statisticV2MigrateStatusDone    = "done"
)

// errStatisticV2MigrateDone 标记持锁后发现已是 done：外层应直接放行，跳过迁移步骤。
var errStatisticV2MigrateDone = errors.New("statistic v2 migrate already done")

// statisticV2MigrateProgress 记录迁移进度，支持失败后从 last_id 续迁。
// status: pending=失败待续；running=进行中；done=完成。
// 首启靠 Create(status=running)+OnConflict{DoNothing} 抢锁：
// 抢到的副本（RowsAffected=1）迁移；抢不到的（RowsAffected=0）报错退出。
// 续迁路径靠 CAS pending→running 抢锁，语义对称。
// 其余遇到 running 报错，阻止启动对外服务；done 放行。
// 优雅失败置回 pending 让下家续迁；进程崩溃（OOM/panic/kill）跑不了代码，
// status 留 running，需人工改 pending 后下个副本续迁。
// 续迁按 LastID 分流：<=0 三个聚合表 upsert 幂等从头全量重跑 + 明细从 0 起；
// >0 跳过聚合表，仅明细按 last_id 续迁。last_id 仅明细表使用。
type statisticV2MigrateProgress struct {
	Status string `json:"status"`  // pending | running | done
	LastID uint32 `json:"last_id"` // 仅 api_key_records 使用
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

func saveStatisticV2MigrateProgress(tx *gorm.DB, p *statisticV2MigrateProgress) error {
	raw, err := p.encode()
	if err != nil {
		return err
	}
	return tx.Model(&Metadata{}).
		Where(&Metadata{MetaKey: statisticV2LegacyMigratedFlagKey}).
		Update("value", raw).Error
}

// errStatisticV2MigrateAlreadyRunning 表示别的副本正在迁移（或崩溃残留 running），
// 本副本不能对外提供服务。调用方可 errors.Is 判断。
// migrateStatisticV2FromLegacy 多副本安全迁移入口。
// 首启（行不存在）：Create(status=running)+OnConflict{DoNothing} 抢锁，
// RowsAffected=1 的副本迁移，=0 的报错退出。
// 行已存在：done 放行；running 报错（别的副本在跑，或崩了等人工置 pending），
// 避免未迁完就对外提供服务；pending 用 CAS pending→running 抢占（续迁路径），
// 抢到的副本迁移，抢不到的同样报错。
// 迁移步骤成功置 done；优雅失败置回 pending（保留已提交 last_id）让下家续迁；
// 进程崩溃跑不了代码，status 留 running，需人工改 pending。
// 续迁按 LastID 分流：<=0 三个聚合表 upsert 幂等从头全量重跑 + 明细从 0 起；
// >0 跳过聚合表，仅明细按 last_id 续迁。明细每批「插入+写游标」同事务原子。
func migrateStatisticV2FromLegacy(db *gorm.DB) error {
	ctx := context.Background()
	var progress *statisticV2MigrateProgress

	var meta Metadata
	err := db.Where(&Metadata{MetaKey: statisticV2LegacyMigratedFlagKey}).First(&meta).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("query migrate progress failed: %w", err)
		}
		// 首启：Create(status=running)+OnConflict{DoNothing} 抢锁。
		// 并发首启只有一个 RowsAffected=1，抢到的副本迁移；其余报错退出。
		progress = &statisticV2MigrateProgress{Status: statisticV2MigrateStatusRunning}
		raw, err := progress.encode()
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
	}
	// 行已存在：done 放行；running 报错；pending 用 CAS pending→running 抢锁（与首启语义对称）。
	// pending：事务内 SELECT ... FOR UPDATE 抢行锁，串行化并发副本，再 UPDATE 置 running。
	err = db.Transaction(func(tx *gorm.DB) error {
		lc, err := parseStatisticV2MigrateProgress(meta.MetaValue)
		if err != nil {
			return err
		}
		switch lc.Status {
		case statisticV2MigrateStatusDone:
			return errStatisticV2MigrateDone
		case statisticV2MigrateStatusRunning:
			return fmt.Errorf("statistic v2 migrate already running")
		}
		// 2. 在锁保护下置 running（用持锁重读的 lc 编码，避免外层旧读的 LastID 回退游标）。
		lc.Status = statisticV2MigrateStatusRunning
		raw, err := lc.encode()
		if err != nil {
			return err
		}
		if err := tx.Model(&Metadata{}).
			Where("meta_key = ?", statisticV2LegacyMigratedFlagKey).
			Update("value", raw).Error; err != nil {
			return fmt.Errorf("claim migrate progress failed: %w", err)
		}
		progress = lc
		return nil
	})
	if err != nil {
		if errors.Is(err, errStatisticV2MigrateDone) {
			return nil
		}
		return err
	}

	// 聚合三步（LastID>0 时跳过，仅续迁明细）。
	if progress.LastID == 0 {
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
				saveStatisticV2MigrateProgress(db, &statisticV2MigrateProgress{Status: statisticV2MigrateStatusPending})
				return err
			}
			log.Infof("statistic v2 migrate %s done: %d rows", step.name, total)
		}
	} else {
		log.Infof("statistic v2 migrate resume from last_id=%d, skip aggregate tables", progress.LastID)
	}

	// 明细续迁，每批「插入+写游标」同事务原子。
	recordTotal, err := migrateLegacyAPIKeyRecords(ctx, db, progress.LastID, func(tx *gorm.DB, lastID uint32) error {
		progress.LastID = lastID
		return saveStatisticV2MigrateProgress(tx, progress)
	})
	if err != nil {
		saveStatisticV2MigrateProgress(db, &statisticV2MigrateProgress{Status: statisticV2MigrateStatusPending})
		return err
	}
	log.Infof("statistic v2 migrate api_key_records done: %d rows", recordTotal)
	return saveStatisticV2MigrateProgress(db, &statisticV2MigrateProgress{Status: statisticV2MigrateStatusDone, LastID: progress.LastID})
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
		for _, row := range rows {
			stat := mapLegacyModelStat(row)
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
			}).Create(stat).Error; err != nil {
				return total, fmt.Errorf("upsert statistic_models id=%d failed: %w", row.ID, err)
			}
			total++
			lastID = row.ID
		}
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
		for _, row := range rows {
			stat := mapLegacyAppStat(row)
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
			}).Create(stat).Error; err != nil {
				return total, fmt.Errorf("upsert statistic_apps id=%d failed: %w", row.ID, err)
			}
			total++
			lastID = row.ID
		}
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
		for _, row := range rows {
			stat := mapLegacyAPIKeyStat(row)
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
			}).Create(stat).Error; err != nil {
				return total, fmt.Errorf("upsert statistic_api_keys id=%d failed: %w", row.ID, err)
			}
			total++
			lastID = row.ID
		}
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

// migrateLegacyAPIKeyRecords 从 lastID 续迁明细。
// 每批「插入 + onBatch 回调」在同一事务内执行，保证游标与数据一致：
// 要么整批+游标都提交，要么都回滚，避免插入成功但游标未存导致重启重复插入。
func migrateLegacyAPIKeyRecords(ctx context.Context, db *gorm.DB, lastID uint32, onBatch func(tx *gorm.DB, lastID uint32) error) (int64, error) {
	if !db.Migrator().HasTable(&model.LegacyAPIKeyRecord{}) {
		log.Infof("statistic v2 migrate skip api key record: table api_key_records not found")
		return 0, nil
	}
	var total int64
	for {
		var rows []model.LegacyAPIKeyRecord
		if err := db.WithContext(ctx).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(statisticV2MigrateBatchSize).
			Find(&rows).Error; err != nil {
			return total, fmt.Errorf("read api_key_records failed: %w", err)
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
			return total, err
		}
		lastID = batchLastID
		total += int64(len(batch))
	}
	return total, nil
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
