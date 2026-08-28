package task

import (
	"context"
	"sync"

	"github.com/gromitlee/go-async/pkg/async/async_task"
)

// report impl IReport
type report struct {
	phase async_task.RunPhase
	del   bool
	ctx   string
}

func (r *report) Phase() (async_task.RunPhase, bool) {
	return r.phase, r.del
}

func (r *report) Context() string {
	return r.ctx
}

func (r *report) clone() *report {
	return &report{
		phase: r.phase,
		del:   r.del,
		ctx:   r.ctx,
	}
}

func CommonDeleting(ctx context.Context, taskCtx string, stop <-chan struct{}, wg *sync.WaitGroup) <-chan async_task.IReport {
	reportCh := make(chan async_task.IReport)
	wg.Add(1)
	go func() {
		defer wg.Wait()
		defer wg.Done()
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
