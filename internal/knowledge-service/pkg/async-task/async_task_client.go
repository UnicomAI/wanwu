package async_task

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	async "github.com/gromitlee/go-async"
	"github.com/gromitlee/go-async/pkg/async/async_component/pending"
)

var asyncTaskClient = AsyncTaskClient{}

type AsyncTaskClient struct {
}

func init() {
	pkg.AddContainer(asyncTaskClient)
}

func (c AsyncTaskClient) LoadType() string {
	return "async-task"
}

func (c AsyncTaskClient) Load() error {
	pendingRun, err := pending.NewPendingRun(db.GetClient().DB, nil)
	if err != nil {
		return err
	}
	options := []async.AsyncOption{
		async.WithPendingRunQueue(pendingRun),
	}
	// init
	if err = async.Init(context.Background(), db.GetClient().DB, options...); err != nil {
		return err
	}
	if err = InitAllService(); err != nil {
		return err
	}
	return nil
}

func (c AsyncTaskClient) Stop() error {
	async.Stop()
	return nil
}

func (c AsyncTaskClient) StopPriority() int {
	return pkg.AsyncTaskPriority
}
