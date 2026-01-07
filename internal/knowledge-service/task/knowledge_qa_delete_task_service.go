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
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

var qaDeleteTask = &QADeleteTask{Del: true}

type QADeleteTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(qaDeleteTask)
}

func (t *QADeleteTask) BuildServiceType() uint32 {
	return async_task_pkg.QADeleteTaskType
}

func (t *QADeleteTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return qaDeleteTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *QADeleteTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "QADeleteTask", t.BuildServiceType(), string(paramsStr), true)
	log.Infof("delete knowledge task %d", taskId)
	return err
}

func (t *QADeleteTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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
			log.Errorf("executeQADeleteTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *QADeleteTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *QADeleteTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- deleteKnowledgeQAByKnowledgeId(ctx, taskCtx)
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

// deleteKnowledgeQAByKnowledgeId 根据问答库id 删除问答库
func deleteKnowledgeQAByKnowledgeId(ctx context.Context, taskCtx string) Result {
	log.Infof("QADeleteTask execute task %s", taskCtx)
	var params = &async_task_pkg.KnowledgeDeleteParams{}
	err := json.Unmarshal([]byte(taskCtx), params)
	if err != nil {
		return Result{Error: err}
	}

	//1.查询问答库信息
	knowledge, err := orm.SelectKnowledgeByIdNoDeleteCheck(ctx, params.KnowledgeId, "", "")
	if err != nil {
		return Result{Error: err}
	}

	//2.查询所有qa pair详情
	qaPairList, err := orm.GetQAPairListByKnowledgeIdNoDeleteCheck(ctx, "", "", params.KnowledgeId)
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
		// 删除导出文件
		if len(exportTasks) > 0 {
			err = batchMinioDeleteExportFiles(ctx, exportTasks)
			if err != nil {
				return err
			}
		}
		// 删除问答库
		err = orm.ExecuteDeleteKnowledge(tx, knowledge.Id)
		if err != nil {
			return err
		}
		// 删除问答库导入任务
		err = orm.DeleteQAImportTaskByKnowledgeId(tx, knowledge.KnowledgeId)
		if err != nil {
			return err
		}
		// 删除问答库导出任务
		err = orm.DeleteExportTaskByKnowledgeId(tx, knowledge.KnowledgeId)
		if err != nil {
			return err
		}
		// 删除相关权限
		err1 := orm.AsyncDeletePermissionByKnowledgeId(knowledge.KnowledgeId)
		if err1 != nil {
			log.Errorf("deleteKnowledgeIdPermission err: %s", err1)
		}
		// 删除全部qa pair
		if len(qaPairList) > 0 {
			err := BatchDeleteAllQAPair(ctx, tx, knowledge, qaPairList)
			if err != nil {
				return err
			}
		}
		// 删除元数据
		err = orm.ExecuteDeleteKnowledgeMeta(tx, knowledge.KnowledgeId)
		if err != nil {
			return err
		}
		err := service.RagQADelete(ctx, &service.RagQADeleteParams{
			UserId: knowledge.UserId,
			QABase: knowledge.RagName,
			QAId:   knowledge.KnowledgeId,
		})
		if err != nil {
			return err
		}
		return nil
	})
	return Result{Error: err}
}

// batchMinioDeleteExportFiles 批量删除minio导出文件
func batchMinioDeleteExportFiles(ctx context.Context, taskList []*model.KnowledgeExportTask) error {
	for _, task := range taskList {
		filePath := "http://" + config.GetConfig().Minio.EndPoint + "/" + task.ExportFilePath
		err := service.DeleteFile(ctx, filePath)
		if err != nil {
			log.Errorf("batchMinioDelete error %v", err)
		}
	}
	return nil
}
