package orm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/UnicomAI/wanwu/internal/operate-service/client/assistant"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

// 迁移版本标记：存在即视为本版本迁移已执行，不再重跑（仿 app-service 的 initFlagKey）。
const migrateAgentAppIDToUUIDFlagKey = "v0.6.4_notice_action_agent_appid_to_uuid"

// Metadata 迁移状态标记表。
// 与 app-service/orm/client.go 的 Metadata 同名同构，但各服务独立维护自己的迁移版本，
// 不跨服务共享语义——operate-service 只用它门控本服务的一次性数据清洗。
type Metadata struct {
	MetaKey   string `gorm:"primaryKey;column:key"`
	MetaValue string `gorm:"column:value"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
}

// migrateAgentAppIDToUUID 一次性清洗 notice_messages.actions：
// 对 msgType=agent 的 action，把 actionParams.appId（assistant 自增老 id）换为 uuid 覆盖回写。
//
// 背景：历史写入的 agent appId 是 assistant 自增老 id，读侧期望 uuid。
// assistant-service 的 GetAssistantBriefByPrimaryIds 提供老 id → uuid 的批量查询。
//
// 失败策略：
//   - 标记已存在 → 直接返回（幂等，不重跑）
//   - assistant 调用失败 → 返回 err（启动期硬要求，与 app-service 一致）
//   - 单个老 id 查不到 → 保留原值 + 告警，不阻断
//   - 单行 JSON 解析失败 → 告警跳过该行，不阻断整体
func migrateAgentAppIDToUUID(db *gorm.DB, assistantCli assistant.IClient) error {
	// 幂等门控：已迁移过则跳过
	var meta Metadata
	err := db.Where(&Metadata{MetaKey: migrateAgentAppIDToUUIDFlagKey}).First(&meta).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("query migrate flag failed: %w", err)
	}

	// 分批游标扫描，解析 actions，收集所有 agent action 的老 appId（去重）
	const batchSize = 100
	oldIDSet := make(map[string]struct{})
	parsed := make(map[int64][]Action)
	var lastID int64
	for {
		var messages []model.NoticeMessage
		if err := db.Select("id", "actions").Where("id > ?", lastID).Order("id").Limit(batchSize).Find(&messages).Error; err != nil {
			return fmt.Errorf("scan notice_messages failed: %w", err)
		}
		if len(messages) == 0 {
			break
		}
		for _, m := range messages {
			lastID = m.ID
			acts := ParseActions(m.Actions)
			parsed[m.ID] = acts
			for _, a := range acts {
				if a.MsgType != MsgTypeAgent {
					continue
				}
				if id, ok := a.ActionParams["appId"].(string); ok && id != "" {
					oldIDSet[id] = struct{}{}
				}
			}
		}
	}

	if len(oldIDSet) == 0 {
		// 无 agent action 需清洗：直接落标记，避免下次重复全表扫描
		return writeMigrateFlag(db)
	}

	// 批量查老 id → uuid
	oldIDs := make([]string, 0, len(oldIDSet))
	for id := range oldIDSet {
		oldIDs = append(oldIDs, id)
	}
	uuidMap, err := assistantCli.GetAppUUIDByOldIDs(context.Background(), oldIDs)
	if err != nil {
		return fmt.Errorf("query assistant app uuid failed: %w", err)
	}

	// 逐行重写有变更的 actions
	for msgID, acts := range parsed {
		if len(acts) == 0 {
			continue
		}
		newJSON, changed := rewriteAgentAppID(acts, uuidMap)
		if !changed {
			continue
		}
		if err := db.Model(&model.NoticeMessage{}).Where("id = ?", msgID).Update("actions", newJSON).Error; err != nil {
			return fmt.Errorf("update notice_messages actions failed id=%d: %w", msgID, err)
		}
	}

	return writeMigrateFlag(db)
}

// rewriteAgentAppID 把 agent action 的老 appId 换为 uuid，返回新 JSON 与是否有变更。
// 输出格式对齐 sanitizeActions（notice_action.go）：[{msgType,actionType,actionParams}]，
// 保证读侧 ParseActions 能正确解析。
func rewriteAgentAppID(acts []Action, uuidMap map[string]string) (string, bool) {
	changed := false
	out := make([]map[string]interface{}, 0, len(acts))
	for _, a := range acts {
		params := a.ActionParams
		if params == nil {
			params = map[string]interface{}{}
		}
		if a.MsgType == MsgTypeAgent {
			if oldID, ok := params["appId"].(string); ok && oldID != "" {
				if uuid, hit := uuidMap[oldID]; hit && uuid != "" {
					params["appId"] = uuid
					changed = true
				} else {
					// 查不到（应用已删等）：保留原值，告警
					log.Warnf("notice_action_migrate: agent appId %q has no uuid, kept as-is", oldID)
				}
			}
		}
		out = append(out, map[string]interface{}{
			"msgType":      a.MsgType,
			"actionType":   a.ActionType,
			"actionParams": params,
		})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", false
	}
	return string(b), changed
}

func writeMigrateFlag(db *gorm.DB) error {
	if err := db.Create(&Metadata{MetaKey: migrateAgentAppIDToUUIDFlagKey}).Error; err != nil {
		return fmt.Errorf("write migrate flag failed: %w", err)
	}
	return nil
}
