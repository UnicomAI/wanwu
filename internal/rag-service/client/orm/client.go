package orm

import (
	"context"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"

	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"

	"gorm.io/gorm"
)

type Client struct {
	db *gorm.DB
}

func NewClient(ctx context.Context, db *gorm.DB) (*Client, error) {
	// auto migrate
	if err := db.AutoMigrate(
		model.RagInfo{},
		model.RagPublish{},
		model.RagConversation{},
	); err != nil {
		return nil, err
	}
	return &Client{
		db: db,
	}, nil
}

func toErrStatus(key string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: key,
		Args:    args,
	}
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
