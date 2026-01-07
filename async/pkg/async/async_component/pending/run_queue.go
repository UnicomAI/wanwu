package pending

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/UnicomAI/wanwu/async/pkg/async/async_component"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
)

func NewPendingRun(db *gorm.DB, log async_config.Logger) (async_component.IQueue, error) {
	if err := db.AutoMigrate(
		asyncPendingRunTask{},
	); err != nil {
		return nil, err
	}
	return &pendingRun{
		log: log,
		db:  db,
	}, nil
}

type pendingRun struct {
	log async_config.Logger
	db  *gorm.DB
}

type asyncPendingRunTask struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;not null"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli;not null"`
	// 用户
	User string `gorm:"index:idx_async_pending_run_task_user"`
	// 任务组
	Group string `gorm:"index:idx_async_pending_run_task_group"`
	// 任务ID
	TaskID uint32 `gorm:"index:idx_async_pending_run_task_task_id;not null"`
	// 任务类型
	Type uint32 `gorm:"index:idx_async_pending_run_task_type;not null"`
}

func (r *pendingRun) AddTask(ctx context.Context, user, group string, taskID, taskType uint32) error {
	return r.db.WithContext(ctx).Create(&asyncPendingRunTask{
		User:   user,
		Group:  group,
		TaskID: taskID,
		Type:   taskType,
	}).Error
}

func (r *pendingRun) DelTask(ctx context.Context, taskID uint32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// select
		dbTask := &asyncPendingRunTask{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("task_id = ?", taskID).First(dbTask).Error; err != nil {
			return err
		}
		// delete
		return tx.Unscoped().Where("task_id = ?", dbTask.TaskID).Delete(&asyncPendingRunTask{}).Error
	})
}

func (r *pendingRun) PopOne(ctx context.Context, taskTypes []uint32) (uint32, error) {
	var taskID uint32
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dbTasks []*asyncPendingRunTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("type IN ?", taskTypes).
			Order("id").Limit(1).Find(&dbTasks).Error; err != nil {
			return err
		} else if len(dbTasks) == 0 {
			return nil
		} else {
			taskID = dbTasks[0].TaskID
		}
		return tx.Unscoped().Where("task_id = ?", taskID).Delete(&asyncPendingRunTask{}).Error
	}); err != nil {
		return 0, err
	}
	return taskID, nil
}
