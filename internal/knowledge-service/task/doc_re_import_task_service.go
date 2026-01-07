package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	async_task_pkg "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

var docReImportTask = &DocReImportTask{Del: true}

type DocReImportTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(docReImportTask)
}

func (t *DocReImportTask) BuildServiceType() uint32 {
	return async_task_pkg.DocReImportTaskType
}

func (t *DocReImportTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return docReImportTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *DocReImportTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "DocReImportTask", t.BuildServiceType(), string(paramStr), true)
	log.Infof("doc import task %d params %s", taskId, paramStr)
	return err
}

func (t *DocReImportTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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

func (t *DocReImportTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *DocReImportTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- reImportDoc(ctx, taskCtx)
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

func reImportDoc(ctx context.Context, taskCtx string) Result {
	var docReImportTaskParams = &async_task_pkg.DocReImportTaskParams{}
	err := json.Unmarshal([]byte(taskCtx), docReImportTaskParams)
	if err != nil {
		log.Errorf("unmarshal json err: %s", err)
		return Result{Error: err}
	}
	importTask, err := orm.SelectKnowledgeImportTaskById(ctx, docReImportTaskParams.TaskId)
	if err != nil {
		log.Errorf("select knowledge import task err: %s", err)
		return Result{Error: err}
	}
	//状态校验
	if importTask.Status == model.KnowledgeImportFinish || importTask.Status == model.KnowledgeImportError {
		log.Infof("knowledge import task not need process : %s status %d", importTask.ImportId, importTask.Status)
		return Result{Error: err}
	}
	detail, err := orm.GetDocDetail(ctx, "", "", docReImportTaskParams.DocId)
	if err != nil {
		log.Errorf("select knowledge doc task err: %s", err)
		return Result{Error: err}
	}
	//执行导入
	list, err := doDocImport(ctx, importTask, detail)
	if len(list) > 0 {
		log.Infof("import task success : %s status %d, doc list %v", importTask.ImportId, importTask.Status, list)
	}
	return Result{Error: err}
}

// doDocImport 执行文件导入
func doDocImport(ctx context.Context, task *model.KnowledgeImportTask, knowledgeDoc *model.KnowledgeDoc) (resultList []*model.DocInfo, err error) {
	defer util.PrintPanicStackWithCall(func(panicOccur bool, err2 error) {
		if panicOccur {
			log.Errorf("do doc import task panic: %v", err2)
			err = fmt.Errorf("文件导入异常")
		}
		var status = model.KnowledgeImportFinish
		var errMsg string
		if err != nil {
			status = model.KnowledgeImportError
			errMsg = err.Error()
		}
		//更新状态和数量
		err = db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
			err = orm.UpdateKnowledgeImportTaskStatus(ctx, tx, task.Id, status, errMsg)
			if err != nil {
				return err
			}
			if len(resultList) > 0 {
				return orm.UpdateKnowledgeFileInfo(tx, task.KnowledgeId, resultList)
			}
			return nil
		})
	})

	//1.更新导入任务状态
	err = orm.UpdateKnowledgeImportTaskStatus(ctx, nil, task.Id, model.KnowledgeImportAnalyze, "")
	if err != nil {
		log.Errorf("Update fail %v", err)
		return nil, err
	}
	//2.执行rag 导入
	err = orm.ReImportKnowledgeDoc(ctx, knowledgeDoc, task)
	if err != nil {
		log.Errorf("ReImportKnowledgeDoc fail %v", err)
		return nil, err
	}
	return
}
