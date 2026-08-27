package orm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

const (
	// builtin_tools 初始化标记，保证历史数据清洗只执行一次
	initBuiltinToolsFlagKey = "v0.6.10_builtin_tools_initialized" // 写错了，其实是 v0.5.10 的数据兼容；)

	initWgaConfigAssistantListFlagKey = "v0.6.4_wga_config_assistant_list_to_uuid"
)

// Metadata 数据清洗幂等标记表
type Metadata struct {
	MetaKey   string `gorm:"primaryKey;column:key"`
	MetaValue string `gorm:"column:value"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
}

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) (*Client, error) {
	if err := db.AutoMigrate(&Metadata{}); err != nil {
		return nil, err
	}

	// auto migrate
	if err := db.AutoMigrate(
		model.Assistant{},
		model.Conversation{},
		model.AssistantWorkflow{},
		model.AssistantMCP{},
		model.AssistantTool{},
		model.AssistantSkill{},
		model.CustomPrompt{},
		model.AssistantSnapshot{},
		model.MultiAgentRelation{},
		model.WgaConversationConfig{},
		model.WgaConfig{},
		model.DigitalEmployeeConversationConfig{},
	); err != nil {
		return nil, err
	}

	if err := initAssistantUUID(db); err != nil {
		return nil, err
	}

	if err := initConversationUUID(db); err != nil {
		return nil, err
	}

	if err := initConversationType(db); err != nil {
		return nil, err
	}

	if err := initBuiltinTools(db); err != nil {
		return nil, err
	}

	if err := initWgaConfigAssistantListUUID(db); err != nil {
		return nil, err
	}

	return &Client{
		db: db,
	}, nil
}

func initAssistantUUID(dbClient *gorm.DB) error {
	const batchSize = 100

	for {
		var ids []uint32
		if err := dbClient.Model(&model.Assistant{}).Select("id").Where("uuid = ? OR uuid IS NULL", "").Limit(batchSize).Find(&ids).Error; err != nil {
			return err
		}

		if len(ids) == 0 {
			break
		}

		caseWhen := "CASE id "
		var args []interface{}
		for _, id := range ids {
			caseWhen += "WHEN ? THEN ? "
			args = append(args, id, util.NewID())
		}
		caseWhen += "END"

		if err := dbClient.Model(&model.Assistant{}).
			Where("id IN ?", ids).
			UpdateColumn("uuid", gorm.Expr(caseWhen, args...)).Error; err != nil {
			log.Errorf("init assistant uuid batch update error: %v", err)
			return err
		}
	}

	return nil
}

// initConversationUUID 为历史 Conversation 行补 conversation_id（新增字段，老数据为 NULL），
// 用 snowflake 生成，分批 + CASE WHEN 更新；填完后无 NULL 行，天然幂等。
func initConversationUUID(dbClient *gorm.DB) error {
	const batchSize = 100

	for {
		var ids []uint32
		if err := dbClient.Model(&model.Conversation{}).Select("id").Where("conversation_id = ? OR conversation_id IS NULL", "").Limit(batchSize).Find(&ids).Error; err != nil {
			return err
		}

		if len(ids) == 0 {
			break
		}

		caseWhen := "CASE id "
		var args []interface{}
		for _, id := range ids {
			caseWhen += "WHEN ? THEN ? "
			args = append(args, id, util.NewID())
		}
		caseWhen += "END"

		if err := dbClient.Model(&model.Conversation{}).
			Where("id IN ?", ids).
			UpdateColumns(map[string]interface{}{
				"conversation_id":   gorm.Expr(caseWhen, args...),
				"conversation_mark": model.ConversationMarkOld,
			}).Error; err != nil {
			log.Errorf("init conversation uuid batch update error: %v", err)
			return err
		}
	}

	return nil
}

func initConversationType(dbClient *gorm.DB) error {
	const batchSize = 100
	numericRegex := regexp.MustCompile(`^\d+$`)

	for {
		var conversations []model.Conversation
		if err := dbClient.Model(&model.Conversation{}).
			Select("id", "user_id").
			Where("conversation_type = ? OR conversation_type IS NULL", "").
			Limit(batchSize).
			Find(&conversations).Error; err != nil {
			return err
		}

		if len(conversations) == 0 {
			break
		}

		caseWhen := "CASE id "
		var args []interface{}
		var ids []uint32

		for _, conv := range conversations {
			ids = append(ids, conv.ID)
			newType := constant.ConversationTypeWebURL
			if numericRegex.MatchString(conv.UserId) {
				newType = constant.ConversationTypePublished
			}
			caseWhen += "WHEN ? THEN ? "
			args = append(args, conv.ID, newType)
		}
		caseWhen += "END"

		if err := dbClient.Model(&model.Conversation{}).
			Where("id IN ?", ids).
			UpdateColumn("conversation_type", gorm.Expr(caseWhen, args...)).Error; err != nil {
			log.Errorf("init conversation type batch update error: %v", err)
			return err
		}
	}

	return nil
}

// initBuiltinTools 为历史智能体补绑定配置中的内置工具，幂等执行一次
func initBuiltinTools(db *gorm.DB) error {
	builtinTools := config.Cfg().BuiltinTools
	if len(builtinTools) == 0 {
		return nil
	}

	// 幂等检查：已执行过则跳过
	var meta Metadata
	if err := db.Where(&Metadata{MetaKey: initBuiltinToolsFlagKey}).First(&meta).Error; err == nil {
		return nil
	}

	// 查询所有智能体
	var assistants []model.Assistant
	if err := db.Select("id, user_id, org_id").Find(&assistants).Error; err != nil {
		return fmt.Errorf("initBuiltinTools query assistants failed: %w", err)
	}

	for _, a := range assistants {
		for _, tool := range builtinTools {
			var count int64
			if err := db.Model(&model.AssistantTool{}).
				Where("assistant_id = ? AND tool_id = ? AND tool_type = ? AND action_name = ?",
					a.ID, tool.ToolId, tool.ToolType, tool.ActionName).
				Count(&count).Error; err != nil {
				log.Warnf("initBuiltinTools check tool %s for assistant %d failed: %v", tool.ToolId, a.ID, err)
				continue
			}
			if count > 0 {
				continue
			}
			if err := db.Create(&model.AssistantTool{
				AssistantId: a.ID,
				ToolId:      tool.ToolId,
				ToolType:    tool.ToolType,
				ActionName:  tool.ActionName,
				Enable:      true,
				UserId:      a.UserId,
				OrgId:       a.OrgId,
			}).Error; err != nil {
				log.Warnf("initBuiltinTools bind tool %s for assistant %d failed: %v", tool.ToolId, a.ID, err)
				continue
			}
		}
	}

	if err := db.Create(&Metadata{MetaKey: initBuiltinToolsFlagKey}).Error; err != nil {
		return fmt.Errorf("initBuiltinTools set flag failed: %w", err)
	}
	return nil
}

// initWgaConfigAssistantListUUID 一次性清洗 wga_configs.assistant_list JSON：
// 把历史存入的 assistant 自增老 id（数字字符串）替换为 uuid。
// 幂等：Metadata 标记存在则跳过。查不到 uuid 的老 id 保留原值。
func initWgaConfigAssistantListUUID(db *gorm.DB) error {
	var meta Metadata
	if err := db.Where(&Metadata{MetaKey: initWgaConfigAssistantListFlagKey}).First(&meta).Error; err == nil {
		return nil
	}

	type wgaAssistant struct {
		AssistantId   string `json:"assistantId"`
		AssistantType string `json:"assistantType"`
	}

	// 分批扫描 wga_configs，解析 assistant_list，收集老 id（去重）
	const batchSize = 100
	oldIDSet := make(map[string]struct{})
	parsed := make(map[uint32][]wgaAssistant)
	var lastID uint32
	for {
		var configs []model.WgaConfig
		if err := db.Select("id", "assistant_list").Where("id > ?", lastID).Order("id").Limit(batchSize).Find(&configs).Error; err != nil {
			return fmt.Errorf("initWgaConfigAssistantListUUID query configs failed: %w", err)
		}
		if len(configs) == 0 {
			break
		}
		for _, cfg := range configs {
			lastID = cfg.ID
			if cfg.AssistantList == "" || cfg.AssistantList == "null" {
				continue
			}
			var list []wgaAssistant
			if err := json.Unmarshal([]byte(cfg.AssistantList), &list); err != nil {
				log.Warnf("initWgaConfigAssistantListUUID parse config %d failed: %v", cfg.ID, err)
				continue
			}
			parsed[cfg.ID] = list
			for _, a := range list {
				if isOldAssistantID(a.AssistantId) {
					oldIDSet[a.AssistantId] = struct{}{}
				}
			}
		}
	}

	if len(oldIDSet) == 0 {
		if err := db.Create(&Metadata{MetaKey: initWgaConfigAssistantListFlagKey}).Error; err != nil {
			return fmt.Errorf("initWgaConfigAssistantListUUID set flag failed: %w", err)
		}
		return nil
	}

	// 批量查老 id → uuid（同库，无需 gRPC）
	oldIDs := make([]uint32, 0, len(oldIDSet))
	for idStr := range oldIDSet {
		id := util.MustU32(idStr)
		if id > 0 {
			oldIDs = append(oldIDs, id)
		}
	}
	var assistants []model.Assistant
	if err := db.Select("id", "uuid").Where("id IN ?", oldIDs).Find(&assistants).Error; err != nil {
		return fmt.Errorf("initWgaConfigAssistantListUUID query assistants failed: %w", err)
	}
	idToUUID := make(map[string]string, len(assistants))
	for _, a := range assistants {
		idToUUID[strconv.Itoa(int(a.ID))] = a.UUID
	}
	for idStr := range oldIDSet {
		if _, ok := idToUUID[idStr]; !ok {
			log.Warnf("initWgaConfigAssistantListUUID: assistant old id %q has no uuid, kept as-is", idStr)
		}
	}

	// 逐行重写有变更的 assistant_list
	for cfgID, list := range parsed {
		changed := false
		for i := range list {
			if !isOldAssistantID(list[i].AssistantId) {
				continue
			}
			if uuid, ok := idToUUID[list[i].AssistantId]; ok {
				list[i].AssistantId = uuid
				changed = true
			}
		}
		if !changed {
			continue
		}
		newJSON, err := json.Marshal(list)
		if err != nil {
			log.Warnf("initWgaConfigAssistantListUUID marshal config %d failed: %v", cfgID, err)
			continue
		}
		if err := db.Model(&model.WgaConfig{}).Where("id = ?", cfgID).Update("assistant_list", string(newJSON)).Error; err != nil {
			log.Warnf("initWgaConfigAssistantListUUID update config %d failed: %v", cfgID, err)
		}
	}

	if err := db.Create(&Metadata{MetaKey: initWgaConfigAssistantListFlagKey}).Error; err != nil {
		return fmt.Errorf("initWgaConfigAssistantListUUID set flag failed: %w", err)
	}
	return nil
}

// isOldAssistantID 判断是否为 assistant 自增老 id：纯数字且 < 10⁹。UUID 无法解析为 int，自动跳过。
func isOldAssistantID(id string) bool {
	if id == "" {
		return false
	}
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}
	return n < 1000000000
}

func (c *Client) transaction(ctx context.Context, fc func(tx *gorm.DB) *err_code.Status) *err_code.Status {
	var status *err_code.Status
	_ = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if status = fc(tx); status != nil {
			return errors.New(status.String())
		}
		return nil
	})
	return status
}

func toErrStatus(code string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: code,
		Args:    args,
	}
}

func ErrCode(code err_code.Code) error {
	return grpc_util.ErrorStatusWithKey(code, "")
}
