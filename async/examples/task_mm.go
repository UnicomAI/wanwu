package examples

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	async "github.com/UnicomAI/wanwu/async"
	async2 "github.com/UnicomAI/wanwu/async/internal/async"
	"github.com/UnicomAI/wanwu/async/internal/tools"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

const (
	taskTypeMM uint32 = 2
)

// taskMM 矩阵相乘并行计算，最后计算结果矩阵的元素总和
// taskCtx json格式：{"Sum":0,"TaskID":123,"A":[[],[],...,[]],"B":[[],[],...,[]],"TaskIDs":[123,456,...]}
// A记录第一个矩阵，B记录第二个矩阵，TaskID记录已经计算了累加和的任务ID，TaskIDs记录所有任务ID
type taskMM struct {
	wg sync.WaitGroup

	del bool // 是否需要自动清理
}

type mmCtx struct {
	Sum     int
	TaskID  int
	A       [][]int
	B       [][]int
	TaskIDs []int
}

func (mm *mmCtx) String() string {
	b, _ := json.Marshal(mm)
	return string(b)
}

func (t *taskMM) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	reportCh := make(chan async_task.IReport)
	t.wg.Add(1)
	go func() {
		defer tools.PrintPanicStack()
		defer t.wg.Wait()
		defer t.wg.Done()
		defer close(reportCh)

		r := &report{phase: async_task.RunPhaseNormal, del: t.del, ctx: taskCtx}
		defer func() {
			reportCh <- r.clone()
		}()

		// check
		mm := &mmCtx{}
		if err := json.Unmarshal([]byte(taskCtx), mm); err != nil {
			r.phase = async_task.RunPhaseFailed
			return
		}
		aH, bW, ok := check(mm.A, mm.B)
		if !ok {
			r.phase = async_task.RunPhaseFailed
			return
		}

		// create sub taskDot tasks
		createTicker := time.NewTicker(time.Millisecond * 100)
		defer createTicker.Stop()
		if n := len(mm.TaskIDs); n < aH*bW {
			for i := 0; i < aH; i++ {
				for j := 0; j < bW; j++ {
					select {
					case <-ctx.Done():
						return
					case <-stop:
						return
					case <-createTicker.C:
						var B []int
						for _, b := range mm.B {
							B = append(B, b[j])
						}
						if i*bW+j == n {
							dot := &dotCtx{A: mm.A[i], B: B}
							if taskID, err := async.CreateTask(ctx, "", "taskDot", taskTypeDot, dot.String(), true); err != nil {
								log.Printf("taskMM create sub task err: %v", err)
								r.phase = async_task.RunPhaseFailed
								return
							} else if err := async.ChangeTaskGroup(ctx, taskID, "taskMM_taskDot"); err != nil {
								log.Printf("taskMM change sub task %v group err: %v", taskID, err)
								r.phase = async_task.RunPhaseFailed
								return
							} else {
								mm.TaskIDs = append(mm.TaskIDs, int(taskID))
								r.ctx = mm.String()
								reportCh <- r.clone()
							}
							n++
						}
					}
				}
			}
		}

		// check sub tasks and sum
		checkTicker := time.NewTicker(time.Millisecond * 100)
		defer checkTicker.Stop()
		for _, taskID := range mm.TaskIDs {
			if taskID <= mm.TaskID {
				continue
			}
			var next bool
			for {
				select {
				case <-ctx.Done():
					return
				case <-stop:
					return
				case <-checkTicker.C:
					task, err := async.GetTask(ctx, uint32(taskID))
					if err != nil {
						if err != async2.ErrMgrAlreadyStop {
							log.Printf("taskMM get sub task %v err: %v", taskID, err)
							r.phase = async_task.RunPhaseFailed
							return
						}
						continue
					}
					switch task.State {
					case async_task.StateFinished:
						dot := &dotCtx{}
						if err := json.Unmarshal([]byte(task.Ctx), dot); err != nil {
							log.Printf("taskMM unmarshal sub task %v ctx err: %v", taskID, err)
							r.phase = async_task.RunPhaseFailed
							return
						}
						mm.Sum = mm.Sum + dot.Sum
						mm.TaskID = taskID
						r.ctx = mm.String()
						reportCh <- r.clone()
						next = true
					case async_task.StateFailed:
						if err := async.RunTask(ctx, uint32(taskID)); err != nil {
							if err != async2.ErrMgrAlreadyStop {
								log.Printf("taskMM restart sub task %v err: %v", taskID, err)
								r.phase = async_task.RunPhaseFailed
								return
							}
						}
					default:
					}
				}
				if next {
					break
				}
			}
		}
		// finished
		r.phase = async_task.RunPhaseFinished
	}()
	return reportCh
}

func (t *taskMM) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	reportCh := make(chan async_task.IReport)
	t.wg.Add(1)
	go func() {
		defer tools.PrintPanicStack()
		defer t.wg.Wait()
		defer t.wg.Done()
		defer close(reportCh)

		select {
		case <-ctx.Done():
			return
		case <-stop:
			return
		case reportCh <- &report{phase: async_task.RunPhaseFinished, ctx: taskCtx}:
		}
	}()
	return reportCh
}

func check(A, B [][]int) (int, int, bool) {
	aH := len(A)
	if aH == 0 {
		return 0, 0, false
	}
	aW := len(A[0])
	if aW == 0 {
		return 0, 0, false
	}
	for _, a := range A {
		if len(a) != aW {
			return 0, 0, false
		}
	}
	bH := len(B)
	if bH == 0 || bH != aW {
		return 0, 0, false
	}
	bW := len(B[0])
	if bW == 0 {
		return 0, 0, false
	}
	for _, b := range B {
		if len(b) != bW {
			return 0, 0, false
		}
	}
	return aH, bW, true
}
