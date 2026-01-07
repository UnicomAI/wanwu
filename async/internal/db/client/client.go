package client

import (
	"context"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/pkg/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/UnicomAI/wanwu/async/internal/async/config"
	"github.com/UnicomAI/wanwu/async/internal/db/model"
	"github.com/UnicomAI/wanwu/async/internal/db/trans"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_component"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
)

// ITaskClient goroutine safe
type ITaskClient interface {
	CreateTask(ctx context.Context, user, group string, taskTyp uint32, taskCtx string, autoRun bool) (uint32, error)
	GetTask(ctx context.Context, taskID uint32) (*model.AsyncTask, error)
	GetTasks(ctx context.Context, user, group string, taskTypes []uint32, status []trans.TaskStatus, offset, limit int32) ([]*model.AsyncTask, error)

	ChangeTaskGroup(ctx context.Context, taskID uint32, group string) error

	SelectOneRun(ctx context.Context, taskTypes []uint32) (*model.AsyncTask, error)
	SelectOneDelete(ctx context.Context, taskTypes []uint32) (*model.AsyncTask, error)

	TransStatus(ctx context.Context, taskID uint32, event trans.TaskEvent) error

	CheckStop(ctx context.Context, taskID uint32) (bool, error)
	UpdateHeartbeat(ctx context.Context, taskID uint32) error
	UpdateContext(ctx context.Context, taskID uint32, taskCtx string) error
	Delete(ctx context.Context, taskID uint32) error

	Clean(ctx context.Context, timeout int) error
}

type client struct {
	log async_config.Logger
	db  *gorm.DB

	pendingRun async_component.IQueue
	pendingDel async_component.IQueue
}

func NewClient(db *gorm.DB, cfg config.Config) (ITaskClient, error) {
	if err := db.AutoMigrate(
		model.AsyncTask{},
	); err != nil {
		return nil, err
	}
	return &client{
		log:        cfg.Log,
		db:         db,
		pendingRun: cfg.PendingRun,
		pendingDel: cfg.PendingDel,
	}, nil
}

func (c *client) CreateTask(ctx context.Context, user, group string, taskTyp uint32, taskCtx string, autoRun bool) (uint32, error) {
	dbTask := &model.AsyncTask{
		User:  user,
		Group: group,
		Type:  taskTyp,
		State: trans.TaskStateInit,
		Mark:  trans.TaskMarkNone,
		Ctx:   db.LongText(taskCtx),
	}
	if err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(dbTask).Error; err != nil {
			return err
		}
		if autoRun {
			return c.transStatus(ctx, tx, dbTask.ID, trans.EventUserRun)
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return dbTask.ID, nil
}

func (c *client) GetTask(ctx context.Context, taskID uint32) (*model.AsyncTask, error) {
	dbTask := &model.AsyncTask{}
	if err := c.db.WithContext(ctx).Where("id = ?", taskID).First(dbTask).Error; err != nil {
		return nil, err
	}
	return dbTask, nil
}

func (c *client) GetTasks(ctx context.Context, user, group string, taskTypes []uint32, status []trans.TaskStatus, offset, limit int32) ([]*model.AsyncTask, error) {
	db := c.db.WithContext(ctx)
	if user != "" {
		db = db.Where("user = ?", user)
	}
	if group != "" {
		db = db.Where("`group` = ?", group)
	}
	if len(taskTypes) > 0 {
		db = db.Where("type IN ?", taskTypes)
	}
	if len(status) > 0 {
		var query string
		var args []interface{}
		for i, s := range status {
			if i == 0 {
				query = query + "(state = ? AND mark = ?)"
			} else {
				query = query + " OR (state = ? AND mark = ?)"
			}
			args = append(args, s.S, s.M)
		}
		db = db.Where(query, args...)
	}
	if offset < 0 {
		offset = 0
	}
	if limit < 0 {
		limit = -1
	}
	var dbTasks []*model.AsyncTask
	if err := db.Offset(int(offset)).Limit(int(limit)).Order("id desc").Find(&dbTasks).Error; err != nil {
		return nil, err
	}
	return dbTasks, nil
}

func (c *client) ChangeTaskGroup(ctx context.Context, taskID uint32, group string) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbTask := &model.AsyncTask{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", taskID).First(dbTask).Error; err != nil {
			return err
		}
		if dbTask.Group == group {
			return nil
		}
		var isPending bool // 用于标记是否需要记录错误日志；pending queue是外部组件，执行错误事务中断可能引起内外数据不一致，需要人工介入
		if dbTask.State == trans.TaskStatePending && dbTask.Mark == trans.TaskMarkRun {
			isPending = true
			if err := c.pendingRun.DelTask(ctx, taskID); err != nil {
				c.log.Errorf("async task %v change group %v -> %v out pending run queue err: %v", taskID, dbTask.Group, group, err)
				return err
			}
			if err := c.pendingRun.AddTask(ctx, dbTask.User, group, taskID, dbTask.Type); err != nil {
				c.log.Errorf("async task %v change group %v -> %v in pending run queue err: %v", taskID, dbTask.Group, group, err)
				return err
			}
		}
		if dbTask.State == trans.TaskStatePending && dbTask.Mark == trans.TaskMarkDelete {
			isPending = true
			if err := c.pendingDel.DelTask(ctx, taskID); err != nil {
				c.log.Errorf("async task %v change group %v -> %v out pending del queue err: %v", taskID, dbTask.Group, group, err)
				return err
			}
			if err := c.pendingDel.AddTask(ctx, dbTask.User, group, taskID, dbTask.Type); err != nil {
				c.log.Errorf("async task %v change group %v -> %v in pending del queue err: %v", taskID, dbTask.Group, group, err)
				return err
			}
		}
		if err := tx.Model(&model.AsyncTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"group": group,
		}).Error; err != nil {
			if isPending {
				c.log.Errorf("async task %v change group %v -> %v err: %v", taskID, dbTask.Group, group, err)
			}
			return err
		}
		return nil
	})
}

func (c *client) SelectOneRun(ctx context.Context, taskTypes []uint32) (*model.AsyncTask, error) {
	var dbTask *model.AsyncTask
	if err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var taskID uint32
		// 1. 优先查找 TaskStatus{S: TaskStatePause, M: TaskMarkRun} 的任务
		var dbTasks []*model.AsyncTask
		if err := tx.Where("state = ? AND mark = ?", trans.TaskStatePause, trans.TaskMarkRun).
			Where("type IN ?", taskTypes).Order("updated_at").Limit(1).Find(&dbTasks).Error; err != nil {
			return err
		} else if len(dbTasks) > 0 {
			taskID = dbTasks[0].ID
		}
		// 2. 再查找pending.runQueue中的任务
		if taskID == 0 {
			if pendingID, err := c.pendingRun.PopOne(ctx, taskTypes); err != nil {
				return err
			} else if pendingID == 0 {
				return nil
			} else {
				taskID = pendingID
			}
		}

		if err := c.transStatus(ctx, tx, taskID, trans.EventSysExecute); err != nil {
			return err
		}
		dbTask = &model.AsyncTask{}
		return tx.Where("id = ?", taskID).First(dbTask).Error
	}); err != nil {
		return nil, err
	}
	return dbTask, nil
}

func (c *client) SelectOneDelete(ctx context.Context, taskTypes []uint32) (*model.AsyncTask, error) {
	var dbTask *model.AsyncTask
	if err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var taskID uint32
		// 1. 优先查找 TaskStatus{S: TaskStatePause, M: TaskMarkDelete} 的任务
		var dbTasks []*model.AsyncTask
		if err := tx.Where("state = ? AND mark = ?", trans.TaskStatePause, trans.TaskMarkDelete).
			Where("type IN ?", taskTypes).Order("updated_at").Limit(1).Find(&dbTasks).Error; err != nil {
			return err
		} else if len(dbTasks) > 0 {
			taskID = dbTasks[0].ID
		}
		// 2. 再查找pending.deleteQueue中的任务
		if taskID == 0 {
			if pendingID, err := c.pendingDel.PopOne(ctx, taskTypes); err != nil {
				return err
			} else if pendingID == 0 {
				return nil
			} else {
				taskID = pendingID
			}
		}

		if err := c.transStatus(ctx, tx, taskID, trans.EventSysExecute); err != nil {
			return err
		}
		dbTask = &model.AsyncTask{}
		return tx.Where("id = ?", taskID).First(dbTask).Error
	}); err != nil {
		return nil, err
	}
	return dbTask, nil
}

func (c *client) TransStatus(ctx context.Context, taskID uint32, event trans.TaskEvent) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return c.transStatus(ctx, tx, taskID, event)
	})
}

func (c *client) CheckStop(ctx context.Context, taskID uint32) (bool, error) {
	dbTask := &model.AsyncTask{}
	if err := c.db.WithContext(ctx).Where("id = ?", taskID).First(dbTask).Error; err != nil {
		return false, err
	}
	return dbTask.State == trans.TaskStateRunning && (dbTask.Mark == trans.TaskMarkDelete || dbTask.Mark == trans.TaskMarkPause), nil
}

func (c *client) UpdateHeartbeat(ctx context.Context, taskID uint32) error {
	return c.db.WithContext(ctx).Model(&model.AsyncTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"updated_at": time.Now().UnixMilli(),
	}).Error
}

func (c *client) UpdateContext(ctx context.Context, taskID uint32, taskCtx string) error {
	return c.db.WithContext(ctx).Model(&model.AsyncTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"ctx": taskCtx,
	}).Error
}

func (c *client) Delete(ctx context.Context, taskID uint32) error {
	return c.db.WithContext(ctx).Unscoped().Where("id = ?", taskID).Delete(&model.AsyncTask{}).Error
}

func (c *client) Clean(ctx context.Context, timeout int) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dbTasks []*model.AsyncTask
		if err := tx.Where("state IN ?", []trans.TaskState{trans.TaskStateRunning, trans.TaskStateDeleting}).Find(&dbTasks).Error; err != nil {
			return err
		}
		if len(dbTasks) == 0 {
			return nil
		}
		updateLimit := time.Now().Add(-time.Duration(timeout) * time.Second).UnixMilli()
		for _, task := range dbTasks {
			if task.UpdatedAt < updateLimit {
				if err := tx.Model(&model.AsyncTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
					"state": trans.TaskStateFailed,
				}).Error; err != nil {
					return err
				}
				if task.State == trans.TaskStateRunning {
					c.log.Errorf("async task %v running but cleaned, last updated at %v", task.ID, task.UpdatedAt)
				} else {
					c.log.Errorf("async task %v deleting but cleaned, last updated at %v", task.ID, task.UpdatedAt)
				}
			}
		}
		return nil
	})
}

func (c *client) transStatus(ctx context.Context, tx *gorm.DB, taskID uint32, event trans.TaskEvent) error {
	dbTask := &model.AsyncTask{}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", taskID).First(dbTask).Error; err != nil {
		return err
	}
	// check transfer
	state, del, err := trans.CheckTransfer(trans.TaskStatus{S: dbTask.State, M: dbTask.Mark}, event)
	if err != nil {
		return fmt.Errorf("async task %v %v", taskID, err.Error())
	}
	// check out pending.runQueue
	var isPending bool // 用于标记是否需要记录错误日志；pending queue是外部组件，执行错误事务中断可能引起内外数据不一致，需要人工介入
	if dbTask.State == trans.TaskStatePending && dbTask.Mark == trans.TaskMarkRun && event != trans.EventSysExecute {
		if event != trans.EventSysExecute { // event是系统执行时，任务已经出队列了
			isPending = true
			if err = c.pendingRun.DelTask(ctx, taskID); err != nil {
				c.log.Errorf("async task %v {state %v, mark %v} event %v transfer {state %v, mark %v} out pending run queue err: %v",
					taskID, dbTask.State, dbTask.Mark, event, state.S, state.M, err)
				return err
			}
		}
	}
	// check in pending.runQueue
	if state.S == trans.TaskStatePending && state.M == trans.TaskMarkRun {
		isPending = true
		if err = c.pendingRun.AddTask(ctx, dbTask.User, dbTask.Group, taskID, dbTask.Type); err != nil {
			c.log.Errorf("async task %v {state %v, mark %v} event %v transfer {state %v, mark %v} in pending run queue err: %v",
				taskID, dbTask.State, dbTask.Mark, event, state.S, state.M, err)
			return err
		}
	}
	// check out pending.delQueue
	if dbTask.State == trans.TaskStatePending && dbTask.Mark == trans.TaskMarkDelete && event != trans.EventSysExecute {
		if event != trans.EventSysExecute { // event是系统执行时，任务已经出队列了
			isPending = true
			if err = c.pendingDel.DelTask(ctx, taskID); err != nil {
				c.log.Errorf("async task %v {state %v, mark %v} event %v transfer {state %v, mark %v} out pending del queue err: %v",
					taskID, dbTask.State, dbTask.Mark, event, state.S, state.M, err)
				return err
			}
		}
	}
	// check in pending.delQueue
	if state.S == trans.TaskStatePending && state.M == trans.TaskMarkDelete {
		isPending = true
		if err = c.pendingDel.AddTask(ctx, dbTask.User, dbTask.Group, taskID, dbTask.Type); err != nil {
			c.log.Errorf("async task %v {state %v, mark %v} event %v transfer {state %v, mark %v} in pending del queue err: %v",
				taskID, dbTask.State, dbTask.Mark, event, state.S, state.M, err)
			return err
		}
	}
	// del or update
	if del {
		if err = tx.Unscoped().Where("id = ?", taskID).Delete(&model.AsyncTask{}).Error; err != nil {
			if isPending {
				c.log.Errorf("async task %v {state %v, mark %v} event %v transfer {state %v, mark %v} del err: %v",
					taskID, dbTask.State, dbTask.Mark, event, state.S, state.M, err)
			}
			return err
		}
		return nil
	}
	updates := map[string]interface{}{
		"state": state.S,
		"mark":  state.M,
	}
	if state.S == trans.TaskStateFinished || state.S == trans.TaskStateFailed {
		updates["done_at"] = time.Now().UnixMilli()
	}
	if err = tx.Model(&model.AsyncTask{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		if isPending {
			c.log.Errorf("async task %v {state %v, mark %v} event %v transfer {state %v, mark %v} update err: %v",
				taskID, dbTask.State, dbTask.Mark, event, state.S, state.M, err)
		}
		return err
	}
	return nil
}
