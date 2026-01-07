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
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var knowledgeReportTask = &KnowledgeReportTask{Del: true}

type KnowledgeReportTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(knowledgeReportTask)
}

func (t *KnowledgeReportTask) BuildServiceType() uint32 {
	return async_task_pkg.KnowledgeReportTaskType
}

func (t *KnowledgeReportTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return knowledgeReportTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *KnowledgeReportTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "KnowledgeReportTask", t.BuildServiceType(), string(paramsStr), true)
	log.Infof("knowledge report import task %d", taskId)
	return err
}

func (t *KnowledgeReportTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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

		//执行知识库报告导入
		systemStop, err := t.runStep(ctx, taskCtx, stop)
		if systemStop {
			log.Infof("system stop")
			return
		}
		if err != nil {
			log.Errorf("executeKnowledgeReportTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *KnowledgeReportTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *KnowledgeReportTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- importKnowledgeReport(ctx, taskCtx)
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

// importKnowledgeReport 根据知识库id 删除知识库
func importKnowledgeReport(ctx context.Context, taskCtx string) Result {
	log.Infof("KnowledgeReportTask execute task %s", taskCtx)
	var params = &async_task_pkg.KnowledgeReportImportTaskParams{}
	err := json.Unmarshal([]byte(taskCtx), params)
	if err != nil {
		return Result{Error: err}
	}

	//1.查询知识库信息
	knowledgeReport, err := orm.SelectKnowledgeReportById(ctx, params.TaskId)
	if err != nil {
		return Result{Error: err}
	}

	//2.查询所有doc详情
	knowledge, err := orm.SelectKnowledgeById(ctx, knowledgeReport.KnowledgeId, "", "")
	if err != nil {
		return Result{Error: err}
	}

	var importTaskParams = model.KnowledgeReportImportParams{}
	err = json.Unmarshal([]byte(knowledgeReport.ImportParams), &importTaskParams)
	if err != nil {
		log.Errorf("knowledge report import params err: %s", err)
		return Result{Error: err}
	}

	//3.更新状态处理中
	err = orm.UpdateReportImportTaskStatus(ctx, params.TaskId, model.KnowledgeReportImportImporting, "", 0)
	if err != nil {
		log.Errorf("UpdateDocSegmentImportTaskStatus err: %s", err)
		return Result{Error: err}
	}
	//4.执行导入
	lineCount, err := doKnowledgeReportImport(ctx, knowledge, &importTaskParams, knowledgeReport)
	if err != nil {
		log.Errorf("knowledge report import err: %s, lineCount %d", err, lineCount)
		return Result{Error: err}
	}
	log.Infof("knowledge report import  lineCount %d", lineCount)
	return Result{Error: err}
}

// PrintPanicStackWithCall 执行文件导入
func doKnowledgeReportImport(ctx context.Context, knowledgeBase *model.KnowledgeBase, importTaskParams *model.KnowledgeReportImportParams, importTask *model.KnowledgeReportImportTask) (lineCount int, err error) {
	defer util.PrintPanicStackWithCall(func(panicOccur bool, err2 error) {
		if panicOccur {
			log.Errorf("do knowledge report import task panic: %v", err2)
			err = fmt.Errorf("文件导入异常")
		}
		var status = model.KnowledgeReportImportSuccess
		var errMsg string
		if err != nil {
			status = model.KnowledgeReportImportFail
			errMsg = err.Error()
		}
		if lineCount == 0 {
			status = model.KnowledgeReportImportFail
			errMsg = "文件所有行全部处理失败"
		}
		//更新状态和数量
		err = orm.UpdateReportImportTaskStatus(ctx, importTask.ImportId, status, errMsg, lineCount)
	})
	lineCount, err = processCsvFileLine(ctx, importTaskParams.FileUrl, buildKnowledgeReportLineProcessor(knowledgeBase, importTask, importTaskParams))
	return
}

// csv 文件行处理器
func buildKnowledgeReportLineProcessor(knowledgeBase *model.KnowledgeBase, importTask *model.KnowledgeReportImportTask, importParams *model.KnowledgeReportImportParams) func(ctx context.Context, strings []string) error {
	return func(ctx context.Context, lineData []string) error {
		var chunks []*service.RagAddReportItem
		chunks = append(chunks, &service.RagAddReportItem{
			Title:   lineData[0],
			Content: lineData[1],
		})

		return orm.CreateOneKnowledgeReport(ctx, importTask, &service.RagAddReportParams{
			UserId:            knowledgeBase.UserId,
			KnowledgeBaseName: knowledgeBase.RagName,
			KnowledgeId:       knowledgeBase.KnowledgeId,
			ReportItem:        chunks,
		})
	}
}
