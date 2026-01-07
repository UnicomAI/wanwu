package orm

import (
	"context"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) (*Client, error) {
	// auto migrate
	if err := db.AutoMigrate(
		model.Assistant{},
		model.Conversation{},
		model.AssistantWorkflow{},
		model.AssistantMCP{},
		model.AssistantTool{},
		model.CustomPrompt{},
		model.ConversationDetails{},
		model.AssistantSnapshot{},
	); err != nil {
		return nil, err
	}

	if err := initAssistantUUID(db); err != nil {
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
