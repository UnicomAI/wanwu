package examples

import (
	"context"

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

func asyncInit(del bool, failRate int, options ...async.AsyncOption) error {
	// init
	if err := async.Init(context.TODO(), getDB(dbMysql, dbName), options...); err != nil {
		return err
	}
	// register
	if err := async.RegisterTask(taskTypeDot, func() async_task.ITask {
		return &taskDot{del: del, failRate: failRate}
	}); err != nil {
		return err
	}
	if err := async.RegisterTask(taskTypeMM, func() async_task.ITask {
		return &taskMM{del: del}
	}); err != nil {
		return err
	}
	return nil
}
