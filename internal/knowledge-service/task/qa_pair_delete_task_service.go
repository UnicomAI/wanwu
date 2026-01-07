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
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

var qaPairDeleteTask = &QAPairDeleteTask{Del: true}

type QAPairDeleteTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(qaPairDeleteTask)
}

func (t *QAPairDeleteTask) BuildServiceType() uint32 {
	return async_task_pkg.QAPairDeleteTaskType
}

func (t *QAPairDeleteTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return qaPairDeleteTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *QAPairDeleteTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "QAPairDeleteTask", t.BuildServiceType(), string(paramStr), true)
	log.Infof("create qa pair delete task task %d ", taskId)
	return err
}

func (t *QAPairDeleteTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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

		//执行数据清理
		systemStop, err := t.runStep(ctx, taskCtx, stop)
		if systemStop {
			log.Infof("system stop")
			return
		}
		if err != nil {
			log.Errorf("executeDocDeleteTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *QAPairDeleteTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *QAPairDeleteTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- deleteQAPairByQAPairId(ctx, taskCtx)
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

func deleteQAPairByQAPairId(ctx context.Context, taskCtx string) Result {
	var params = &async_task_pkg.QAPairDeleteParams{}
	err := json.Unmarshal([]byte(taskCtx), params)
	if err != nil {
		return Result{Error: err}
	}
	//1.查询问答对详情
	qaPair, err := orm.GetQAPairByQAPairIdNoDeleteCheck(ctx, "", "", params.QaPairId)
	if err != nil {
		return Result{Error: err}
	}
	if qaPair == nil {
		return Result{Error: nil}
	}
	//2.查询知识库信息
	knowledge, err := orm.SelectKnowledgeByIdNoDeleteCheck(ctx, qaPair.KnowledgeId, "", "")
	if err != nil {
		return Result{Error: err}
	}
	//3.事务执行删除数据
	err = db.GetClient().DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return BatchDeleteAllQAPair(ctx, tx, knowledge, []*model.KnowledgeQAPair{qaPair})
	})
	return Result{Error: err}
}

// BatchDeleteAllQAPair 批量删除所有问答对
func BatchDeleteAllQAPair(ctx context.Context, tx *gorm.DB, knowledge *model.KnowledgeBase, qaPairList []*model.KnowledgeQAPair) error {
	var qaPairIdList []string
	for _, qaPair := range qaPairList {
		qaPairIdList = append(qaPairIdList, qaPair.QAPairId)
	}
	//1.删除db数据
	err := orm.ExecuteDeleteQAPairByQAPairIdList(tx, qaPairIdList)
	if err != nil {
		log.Errorf("ExecuteDeleteQAPairByQAPairIdList error %v", err)
		return err
	}
	//2.删除元数据
	err = orm.DeleteMetaDataByDocIdList(tx, knowledge.KnowledgeId, qaPairIdList)
	if err != nil {
		//只打印，不阻塞
		log.Errorf("DeleteMetaDataByDocIdList error %v", err)
	}
	//3.删除底层数据
	err = service.RagDeleteQAPair(ctx, &service.RagDeleteQAPairParams{
		UserId:     knowledge.UserId,
		QAId:       knowledge.KnowledgeId,
		QABaseName: knowledge.RagName,
		QAPairIds:  qaPairIdList,
	})
	if err != nil {
		//只打印，不阻塞
		log.Errorf("RagDeleteQAPair error %v", err)
	}
	return nil
}
