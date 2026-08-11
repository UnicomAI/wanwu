package orm

import (
	"context"
	"errors"
	"fmt"
	"regexp"

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
	initBuiltinToolsFlagKey = "v0.6.10_builtin_tools_initialized"
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

	if err := initConversationType(db); err != nil {
		return nil, err
	}

	if err := initBuiltinTools(db); err != nil {
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
