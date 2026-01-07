package examples

import (
	"context"
	"testing"
	"time"

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/internal/tools"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_component/pending"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

func TestTaskMM_All(t *testing.T) {
	TestTaskMM_Default(t)
	TestTaskMM_WithPendingRunQueue(t)
	TestTaskMM_WithPendingDelQueue(t)
	TestTaskMM_WithPendingRunAndDelQueue(t)
}

func TestTaskMM_Default(t *testing.T) {
	TestTaskMM_Finished(t)
	TestTaskMM_SysPauseAndRestart(t)
}

func TestTaskMM_WithPendingRunQueue(t *testing.T) {
	pendingRun, err := pending.NewPendingRun(getDB(dbMysql, dbName), tools.DefaultLog())
	if err != nil {
		t.Fatal(err)
	}
	options = []async.AsyncOption{
		async.WithPendingRunQueue(pendingRun),
	}

	TestTaskMM_Default(t)
}

func TestTaskMM_WithPendingDelQueue(t *testing.T) {
	pendingDel, err := pending.NewPendingDel(getDB(dbMysql, dbName), tools.DefaultLog())
	if err != nil {
		t.Fatal(err)
	}
	options = []async.AsyncOption{
		async.WithPendingDelQueue(pendingDel),
	}

	TestTaskMM_Default(t)
}

func TestTaskMM_WithPendingRunAndDelQueue(t *testing.T) {
	pendingRun, err := pending.NewPendingRun(getDB(dbMysql, dbName), tools.DefaultLog())
	if err != nil {
		t.Fatal(err)
	}
	pendingDel, err := pending.NewPendingDel(getDB(dbMysql, dbName), tools.DefaultLog())
	if err != nil {
		t.Fatal(err)
	}
	options = []async.AsyncOption{
		async.WithPendingRunQueue(pendingRun),
		async.WithPendingDelQueue(pendingDel),
	}

	TestTaskMM_Default(t)
}

func TestTaskMM_Finished(t *testing.T) {
	// init
	if err := asyncInit(false, 30, options...); err != nil {
		t.Fatal(err)
	}
	// create & run
	taskCtx := "{\"A\":[[0,1,2,3,4],[5,6,7,8,9],[10,11,12,13,14],[15,16,17,18,19],[20,21,22,23,24]]," +
		"\"B\":[[0,1,2],[3,4,5],[6,7,8],[9,10,11],[12,13,14]]}"
	taskID, err := async.CreateTask(context.TODO(), "", "taskMM", taskTypeMM, taskCtx, true)
	if err != nil {
		t.Fatal(err)
	}
	checkTicker := time.NewTicker(123 * time.Millisecond)
	defer checkTicker.Stop()
	var stop bool
	for {
		select {
		case <-checkTicker.C:
			if task, err := async.GetTask(context.TODO(), taskID); err != nil {
				t.Fatal(err)
			} else if task.State == async_task.StateFinished {
				t.Logf("%+v", task)
				stop = true
			}
		}
		if stop {
			break
		}
	}
	// stop
	async.Stop()
}

func TestTaskMM_SysPauseAndRestart(t *testing.T) {
	// init
	if err := asyncInit(false, 0, options...); err != nil {
		t.Fatal(err)
	}
	// create & run
	taskCtx := "{\"A\":[[0,1,2,3,4],[5,6,7,8,9],[10,11,12,13,14],[15,16,17,18,19],[20,21,22,23,24]]," +
		"\"B\":[[0,1,2],[3,4,5],[6,7,8],[9,10,11],[12,13,14]]}"
	taskID, err := async.CreateTask(context.TODO(), "", "taskMM", taskTypeMM, taskCtx, true)
	if err != nil {
		t.Fatal(err)
	}

	stopTicker := time.NewTicker(time.Second * 10)
	defer stopTicker.Stop()
	checkTicker := time.NewTicker(time.Millisecond * 123)
	defer checkTicker.Stop()
	var stop bool
	for {
		select {
		case <-stopTicker.C:
			async.Stop()
			if err := asyncInit(false, 0, options...); err != nil {
				t.Fatal(err)
			}
		case <-checkTicker.C:
			if task, err := async.GetTask(context.TODO(), taskID); err != nil {
				t.Fatal(err)
			} else if task.State == async_task.StateFinished {
				t.Logf("%+v", task)
				stop = true
			}
		}
		if stop {
			break
		}
	}
	// stop
	async.Stop()
}
