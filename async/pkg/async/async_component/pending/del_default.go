package pending

import (
	"context"

	"gorm.io/gorm"

	"github.com/UnicomAI/wanwu/async/internal/db/model"
	"github.com/UnicomAI/wanwu/async/internal/db/trans"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_component"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
)

func NewPendingDelDefault(db *gorm.DB, log async_config.Logger) async_component.IQueue {
	return &pendingDelDefault{
		log: log,
		db:  db,
	}
}

type pendingDelDefault struct {
	log async_config.Logger
	db  *gorm.DB
}

func (d *pendingDelDefault) AddTask(ctx context.Context, user, group string, taskID, taskType uint32) error {
	return nil
}

func (d *pendingDelDefault) DelTask(ctx context.Context, taskID uint32) error {
	return nil
}

func (d *pendingDelDefault) PopOne(ctx context.Context, taskTypes []uint32) (uint32, error) {
	var dbTasks []*model.AsyncTask
	if err := d.db.WithContext(ctx).Where("state = ? AND mark = ?", trans.TaskStatePending, trans.TaskMarkDelete).
		Where("type IN ?", taskTypes).
		Order("updated_at").Limit(1).Find(&dbTasks).Error; err != nil {
		return 0, err
	} else if len(dbTasks) == 0 {
		return 0, nil
	} else {
		return dbTasks[0].ID, nil
	}
}
