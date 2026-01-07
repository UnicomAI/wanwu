package model

import (
	"github.com/UnicomAI/wanwu/async/internal/db/trans"
	"github.com/UnicomAI/wanwu/pkg/db"
)

// AsyncTask 异步任务 DB Model
// target变化不会导致state变化
type AsyncTask struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;not null"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli;not null"`
	// 用户
	User string `gorm:"index:idx_async_task_user"`
	// 任务组
	Group string `gorm:"index:idx_async_task_group"`
	// 任务类型
	Type uint32 `gorm:"index:idx_async_task_type;not null"`
	// 状态
	State trans.TaskState `gorm:"index:idx_async_task_state;not null"`
	// 标记
	Mark trans.TaskMark `gorm:"index:idx_async_task_mark;not null"`
	// 结束时间戳（finished/failed）
	DoneAt int64 `gorm:"not null"`
	// 序列化上下文
	Ctx db.LongText `gorm:"not null"`
}
