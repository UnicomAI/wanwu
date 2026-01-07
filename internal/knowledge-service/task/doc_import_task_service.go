package task

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	async_task_pkg "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	import_service "github.com/UnicomAI/wanwu/internal/knowledge-service/task/import-service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var docImportTask = &DocImportTask{Del: true}

type DocImportTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

type Result struct {
	Error error
}

func init() {
	async_task_pkg.AddContainer(docImportTask)
}

func (t *DocImportTask) BuildServiceType() uint32 {
	return async_task_pkg.DocImportTaskType
}

func (t *DocImportTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return docImportTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *DocImportTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "DocImportTask", t.BuildServiceType(), string(paramStr), true)
	log.Infof("doc import task %d ", taskId)
	return err
}

func (t *DocImportTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	reportCh := make(chan async_task.IReport)
	t.Wg.Add(1)
	go func() {
		defer util.PrintPanicStack()
		defer t.Wg.Wait()
		defer t.Wg.Done()
		defer close(reportCh)

		r := &report{phase: async_task.RunPhaseNormal, del: t.Del, ctx: taskCtx}
		defer func() {
			reportCh <- r.clone()
		}()

		//执行文件导入
		systemStop, err := t.runStep(ctx, taskCtx, stop)
		if systemStop {
			log.Infof("system stop")
			return
		}
		if err != nil {
			log.Errorf("executeDataCleanTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *DocImportTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *DocImportTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- importDoc(ctx, taskCtx)
	}()
	for {
		select {
		case <-ctx.Done():
			return false, nil
		case <-stop:
			return true, nil
		case result := <-ret:
			return false, result.Error
		}
	}
}

func importDoc(ctx context.Context, taskCtx string) Result {
	var docImportTaskParams = &async_task_pkg.DocImportTaskParams{}
	err := json.Unmarshal([]byte(taskCtx), docImportTaskParams)
	if err != nil {
		log.Errorf("unmarshal json err: %s", err)
		return Result{Error: err}
	}
	importTask, err := orm.SelectKnowledgeImportTaskById(ctx, docImportTaskParams.TaskId)
	if err != nil {
		log.Errorf("select knowledge import task err: %s", err)
		return Result{Error: err}
	}
	//状态校验
	if importTask.Status == model.KnowledgeImportFinish || importTask.Status == model.KnowledgeImportError {
		log.Infof("knowledge import task not need process : %s status %d", importTask.ImportId, importTask.Status)
		return Result{Error: err}
	}
	//执行导入
	list, err := import_service.DoDocImport(ctx, importTask)
	if len(list) > 0 {
		log.Infof("import task success : %s status %d, doc list %v", importTask.ImportId, importTask.Status, list)
	}
	return Result{Error: err}
}
