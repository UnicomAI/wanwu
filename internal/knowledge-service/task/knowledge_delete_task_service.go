package task

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	async_task_pkg "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

var knowledgeDeleteTask = &KnowledgeDeleteTask{Del: true}

type KnowledgeDeleteTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(knowledgeDeleteTask)
}

func (t *KnowledgeDeleteTask) BuildServiceType() uint32 {
	return async_task_pkg.KnowledgeDeleteTaskType
}

func (t *KnowledgeDeleteTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return knowledgeDeleteTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *KnowledgeDeleteTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "KnowledgeDeleteTask", t.BuildServiceType(), string(paramsStr), true)
	log.Infof("delete knowledge task %d", taskId)
	return err
}

func (t *KnowledgeDeleteTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	reportCh := make(chan async_task.IReport)
	t.Wg.Add(1)
	go func() {
		defer t.Wg.Wait()
		defer t.Wg.Done()
		defer close(reportCh)

		r := &report{phase: async_task.RunPhaseNormal, del: t.Del, ctx: taskCtx}
		defer func() {
			reportCh <- r.clone()
		}()

		//执行知识库删除
		systemStop, err := t.runStep(ctx, taskCtx, stop)
		if systemStop {
			log.Infof("system stop")
			return
		}
		if err != nil {
			log.Errorf("executeKnowledgeDeleteTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *KnowledgeDeleteTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *KnowledgeDeleteTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- deleteKnowledgeByKnowledgeId(ctx, taskCtx)
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

// deleteKnowledgeByKnowledgeId 根据知识库id 删除知识库
func deleteKnowledgeByKnowledgeId(ctx context.Context, taskCtx string) Result {
	log.Infof("KnowledgeDeleteTask execute task %s", taskCtx)
	var params = &async_task_pkg.KnowledgeDeleteParams{}
	err := json.Unmarshal([]byte(taskCtx), params)
	if err != nil {
		return Result{Error: err}
	}

	//1.查询知识库信息
	knowledge, err := orm.SelectKnowledgeByIdNoDeleteCheck(ctx, params.KnowledgeId, "", "")
	if err != nil {
		return Result{Error: err}
	}

	//2.查询所有doc详情
	docList, err := orm.GetDocListByKnowledgeIdNoDeleteCheck(ctx, "", "", params.KnowledgeId)
	if err != nil {
		return Result{Error: err}
	}

	//3.查询所有导出任务
	exportTasks, err := orm.SelectExportTaskByKnowledgeIdNoDeleteCheck(ctx, params.KnowledgeId, "", "")
	if err != nil {
		return Result{Error: err}
	}

	//4.事务执行删除数据
	err = db.GetClient().DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(docList) > 0 {
			err := BatchDeleteAllDoc(ctx, tx, knowledge, docList)
			if err != nil {
				return err
			}
		}
		// 删除导出文件
		if len(exportTasks) > 0 {
			err = batchMinioDeleteExportFiles(ctx, exportTasks)
			if err != nil {
				return err
			}
		}
		err := service.RagKnowledgeDelete(ctx, &service.RagDeleteParams{
			UserId:            knowledge.UserId,
			KnowledgeBaseName: knowledge.RagName,
			KnowledgeId:       knowledge.KnowledgeId,
		})
		if err != nil {
			return err
		}
		err = orm.ExecuteDeleteKnowledgeMeta(tx, knowledge.KnowledgeId)
		if err != nil {
			return err
		}
		err = orm.ExecuteDeleteKnowledge(tx, knowledge.Id)
		if err != nil {
			return err
		}
		err = orm.DeleteImportTaskByKnowledgeId(tx, knowledge.KnowledgeId)
		if err != nil {
			return err
		}
		// 删除知识库导出任务
		err = orm.DeleteExportTaskByKnowledgeId(tx, knowledge.KnowledgeId)
		if err != nil {
			return err
		}
		//删除相关权限
		err1 := orm.AsyncDeletePermissionByKnowledgeId(knowledge.KnowledgeId)
		if err1 != nil {
			log.Errorf("deleteKnowledgeIdPermission err: %s", err1)
		}
		return nil
	})
	return Result{Error: err}
}
