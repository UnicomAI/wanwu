package examples

import (
	"context"
	"sync"
	"testing"
	"time"

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/internal/tools"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_component/pending"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

var options []async.AsyncOption

func TestTaskDot_All(t *testing.T) {
	TestTaskDot_Default(t)
	TestTaskDot_WithPendingRunQueue(t)
	TestTaskDot_WithPendingDelQueue(t)
	TestTaskDot_WithPendingRunAndDelQueue(t)
}

func TestTaskDot_Default(t *testing.T) {
	TestTaskDot_Finished(t)
	TestTaskDot_UserDelete(t)
	TestTaskDot_FailedAndUserRestart(t)
	TestTaskDot_FailedOrUserPauseAndUserRestart(t)
	TestTaskDot_FailedAndDelete(t)
}

func TestTaskDot_WithPendingRunQueue(t *testing.T) {
	pendingRun, err := pending.NewPendingRun(getDB(dbMysql, dbName), tools.DefaultLog())
	if err != nil {
		t.Fatal(err)
	}
	options = []async.AsyncOption{
		async.WithPendingRunQueue(pendingRun),
	}

	TestTaskDot_Default(t)
}

func TestTaskDot_WithPendingDelQueue(t *testing.T) {
	pendingDel, err := pending.NewPendingDel(getDB(dbMysql, dbName), tools.DefaultLog())
	if err != nil {
		t.Fatal(err)
	}
	options = []async.AsyncOption{
		async.WithPendingDelQueue(pendingDel),
	}

	TestTaskDot_Default(t)
}

func TestTaskDot_WithPendingRunAndDelQueue(t *testing.T) {
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

	TestTaskDot_Default(t)
}

func TestTaskDot_Finished(t *testing.T) {
	// init
	if err := asyncInit(false, 0, options...); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	// create & run
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			taskCtx := "{\"A\":[0,1,2,3,4,5,6,7,8,9],\"B\":[0,10,20,30,40,50,60,70,80,90]}"
			taskID, err := async.CreateTask(context.TODO(), "", "taskDot", taskTypeDot, taskCtx, true)
			if err != nil {
				t.Error(err)
				return
			}
			checkTicker := time.NewTicker(time.Millisecond * 123)
			defer checkTicker.Stop()
			var stop bool
			for {
				select {
				case <-checkTicker.C:
					if task, err := async.GetTask(context.TODO(), taskID); err != nil {
						t.Error(err)
						return
					} else if task.State == async_task.StateFinished {
						t.Logf("%+v", task)
						stop = true
					}
				}
				if stop {
					break
				}
			}
		}()
	}
	// stop
	wg.Wait()
	async.Stop()
}

func TestTaskDot_UserDelete(t *testing.T) {
	// init
	if err := asyncInit(false, 0, options...); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	// create & run
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			taskCtx := "{\"A\":[0,1,2,3,4,5,6,7,8,9],\"B\":[0,10,20,30,40,50,60,70,80,90]}"
			taskID, err := async.CreateTask(context.TODO(), "", "taskDot", taskTypeDot, taskCtx, true)
			if err != nil {
				t.Error(err)
				return
			}
			checkTicker := time.NewTicker(time.Millisecond * 123)
			defer checkTicker.Stop()
			var stop bool
			for {
				select {
				case <-checkTicker.C:
					if task, err := async.GetTask(context.TODO(), taskID); err != nil {
						t.Error(err)
						return
					} else {
						switch task.State {
						case async_task.StateRunning:
							_ = async.DeleteTask(context.TODO(), taskID)
						case async_task.StateCanceling:
							stop = true
						}
					}
				}
				if stop {
					break
				}
			}
			time.Sleep(time.Second * 5)
			if _, err := async.GetTask(context.TODO(), taskID); err != async.ErrTaskNotFound {
				t.Error(err)
				return
			}
		}()
	}
	// stop
	wg.Wait()
	async.Stop()
}

func TestTaskDot_FailedAndUserRestart(t *testing.T) {
	// init
	if err := asyncInit(false, 30, options...); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	// create & run
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			taskCtx := "{\"A\":[0,1,2,3,4,5,6,7,8,9],\"B\":[0,10,20,30,40,50,60,70,80,90]}"
			taskID, err := async.CreateTask(context.TODO(), "", "taskDot", taskTypeDot, taskCtx, true)
			if err != nil {
				t.Error(err)
				return
			}
			checkTicker := time.NewTicker(time.Millisecond * 123)
			defer checkTicker.Stop()
			var stop bool
			for {
				select {
				case <-checkTicker.C:
					if task, err := async.GetTask(context.TODO(), taskID); err != nil {
						t.Error(err)
						return
					} else {
						switch task.State {
						case async_task.StateFinished:
							t.Logf("%+v", task)
							stop = true
						case async_task.StateFailed:
							if err := async.RunTask(context.TODO(), taskID); err != nil {
								t.Error(err)
								return
							}
						}
					}
				}
				if stop {
					break
				}
			}
		}()
	}
	// stop
	wg.Wait()
	async.Stop()
}

func TestTaskDot_FailedOrUserPauseAndUserRestart(t *testing.T) {
	// init
	if err := asyncInit(false, 30, options...); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	// create & run
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			taskCtx := "{\"A\":[0,1,2,3,4,5,6,7,8,9],\"B\":[0,10,20,30,40,50,60,70,80,90]}"
			taskID, err := async.CreateTask(context.TODO(), "", "taskDot", taskTypeDot, taskCtx, true)
			if err != nil {
				t.Error(err)
				return
			}
			checkTicker := time.NewTicker(time.Millisecond * 123)
			defer checkTicker.Stop()
			pauseTicker := time.NewTicker(time.Second * 5)
			defer pauseTicker.Stop()
			var stop bool
			for {
				select {
				case <-checkTicker.C:
					if task, err := async.GetTask(context.TODO(), taskID); err != nil {
						t.Error(err)
						return
					} else {
						switch task.State {
						case async_task.StateFinished:
							t.Logf("%+v", task)
							stop = true
						case async_task.StatePause, async_task.StateFailed:
							_ = async.RunTask(context.TODO(), taskID)
						}
					}
				case <-pauseTicker.C:
					if task, err := async.GetTask(context.TODO(), taskID); err != nil {
						t.Error(err)
						return
					} else if task.State == async_task.StateRunning {
						_ = async.PauseTask(context.TODO(), taskID)
					}
				}
				if stop {
					break
				}
			}
		}()
	}
	// stop
	wg.Wait()
	async.Stop()
}

func TestTaskDot_FailedAndDelete(t *testing.T) {
	// init
	if err := asyncInit(true, 50, options...); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	// create & run
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			taskCtx := "{\"A\":[0,1,2,3,4,5,6,7,8,9],\"B\":[0,10,20,30,40,50,60,70,80,90]}"
			taskID, err := async.CreateTask(context.TODO(), "", "taskDot", taskTypeDot, taskCtx, true)
			if err != nil {
				t.Error(err)
				return
			}
			checkTicker := time.NewTicker(time.Millisecond * 123)
			defer checkTicker.Stop()
			var stop bool
			for {
				select {
				case <-checkTicker.C:
					if _, err := async.GetTask(context.TODO(), taskID); err != nil {
						if err != async.ErrTaskNotFound {
							t.Error(err)
							return
						} else {
							stop = true
						}
					}
				}
				if stop {
					break
				}
			}
		}()
	}
	// stop
	wg.Wait()
	async.Stop()
}
