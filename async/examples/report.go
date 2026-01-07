package examples

import "github.com/UnicomAI/wanwu/async/pkg/async/async_task"

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
