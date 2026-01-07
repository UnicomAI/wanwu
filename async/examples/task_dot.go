package examples

import (
	"context"
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/async/internal/tools"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

const (
	taskTypeDot uint32 = 1
)

// taskDot 向量点积任务
// taskCtx json格式：{"I":2,"Sum":14,"A":[1,2,3],"B":[4,5,6]}
// A记录第一个向量，B记录第二个向量，I记录已经计算了第几个元素的乘积(I从1开始，0表示还未计算)，Sum记录已经计算的乘积累加和
type taskDot struct {
	wg sync.WaitGroup

	del bool // 是否需要自动清理

	failRate  int // 每次tick fail概率(%)，小于等于0不会fail
	panicRate int // 每次tick panic概率(%)，小于等于0不会panic
}

type dotCtx struct {
	I   int
	Sum int
	A   []int
	B   []int
}

func (dot *dotCtx) String() string {
	b, _ := json.Marshal(dot)
	return string(b)
}

func (t *taskDot) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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
		dot := &dotCtx{}
		if err := json.Unmarshal([]byte(taskCtx), dot); err != nil {
			r.phase = async_task.RunPhaseFailed
			return
		}
		if len(dot.A) != len(dot.B) {
			r.phase = async_task.RunPhaseFailed
			return
		}

		rn := rand.New(rand.NewSource(time.Now().UnixNano()))
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			if dot.I >= len(dot.A) {
				r.phase = async_task.RunPhaseFinished
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				if rn.Intn(100) < t.panicRate {
					panic("task panic test")
				}
				if rn.Intn(100) < t.failRate {
					r.phase = async_task.RunPhaseFailed
					return
				}
				dot.Sum = dot.Sum + dot.A[dot.I]*dot.B[dot.I]
				dot.I = dot.I + 1
				r.ctx = dot.String()
				reportCh <- r.clone()
			}
		}
	}()

	return reportCh
}

func (t *taskDot) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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
