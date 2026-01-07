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

var docDeleteTask = &DocDeleteTask{Del: true}

type DocDeleteTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(docDeleteTask)
}

func (t *DocDeleteTask) BuildServiceType() uint32 {
	return async_task_pkg.DocDeleteTaskType
}

func (t *DocDeleteTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return docDeleteTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *DocDeleteTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "DocDeleteTask", t.BuildServiceType(), string(paramStr), true)
	log.Infof("create doc delete task task %d ", taskId)
	return err
}

func (t *DocDeleteTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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

func (t *DocDeleteTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *DocDeleteTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- deleteDocByIds(ctx, taskCtx)
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

func deleteDocByIds(ctx context.Context, taskCtx string) Result {
	var params = &async_task_pkg.DocDeleteParams{}
	err := json.Unmarshal([]byte(taskCtx), params)
	if err != nil {
		return Result{Error: err}
	}
	//1.查询所有doc详情
	list, err := orm.GetDocListByIdListNoDeleteCheck(ctx, "", "", params.DocIdList)
	if err != nil {
		return Result{Error: err}
	}
	if len(list) == 0 {
		return Result{Error: nil}
	}
	//2.查询知识库信息
	knowledge, err := orm.SelectKnowledgeByIdNoDeleteCheck(ctx, list[0].KnowledgeId, "", "")
	if err != nil {
		return Result{Error: err}
	}
	//3.事务执行删除数据
	err = db.GetClient().DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return BatchDeleteAllDoc(ctx, tx, knowledge, list)
	})
	return Result{Error: err}
}

// BatchDeleteAllDoc 批量删除所有文档
func BatchDeleteAllDoc(ctx context.Context, tx *gorm.DB, knowledge *model.KnowledgeBase, docList []*model.KnowledgeDoc) error {
	var docIdList []uint32
	for _, doc := range docList {
		docIdList = append(docIdList, doc.Id)
	}
	//1.删除minio
	err := batchMinioDelete(ctx, docList)
	if err != nil {
		//只打印，不阻塞
		log.Errorf("batchMinioDelete error %v", err)
	}
	//2.删除db数据
	err = orm.ExecuteDeleteDocByIdList(tx, docIdList)
	if err != nil {
		log.Errorf("ExecuteDeleteDocByIdList error %v", err)
		return err
	}
	//3.删除元数据
	err = orm.DeleteMetaDataByDocIdList(tx, knowledge.KnowledgeId, buildDocIdList(docList))
	if err != nil {
		//只打印，不阻塞
		log.Errorf("DeleteMetaDataByDocIdList error %v", err)
	}
	//4.删除底层数据
	err = batchRagDelete(ctx, knowledge, docList)
	if err != nil {
		//只打印，不阻塞
		log.Errorf("batchRagDelete error %v", err)
	}
	return nil
}

// batchRagDelete 批量rag删除
func batchRagDelete(ctx context.Context, knowledge *model.KnowledgeBase, docList []*model.KnowledgeDoc) error {
	for _, doc := range docList {
		var fileName = service.RebuildFileName(doc.DocId, doc.FileType, doc.Name)
		err := service.RagDeleteDoc(ctx, &service.RagDeleteDocParams{
			UserId:        knowledge.UserId,
			KnowledgeBase: knowledge.RagName,
			FileName:      fileName,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// batchMinioDelete 批量minio url 删除
func batchMinioDelete(ctx context.Context, docList []*model.KnowledgeDoc) error {
	for _, doc := range docList {
		if doc.FileType == "url" {
			//url 类型没有上传minio，跳过
			continue
		}
		err := service.DeleteFile(ctx, doc.FilePath)
		if err != nil {
			log.Errorf("batchMinioDelete error %v", err)
		}
	}
	return nil
}

func buildDocIdList(docList []*model.KnowledgeDoc) []string {
	var retList []string
	for _, doc := range docList {
		retList = append(retList, doc.DocId)
	}
	return retList
}
