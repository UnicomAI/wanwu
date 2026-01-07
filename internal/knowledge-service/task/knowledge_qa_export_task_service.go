package task

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	async_task_pkg "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	file_util "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	exportBatchSize = 20
	exportLocalDir  = "static/export/"
)

var knowledgeQAPairExportTask = &KnowledgeQAPairExportTask{Del: true}

type KnowledgeQAPairExportTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(knowledgeQAPairExportTask)
}

func (t *KnowledgeQAPairExportTask) BuildServiceType() uint32 {
	return async_task_pkg.KnowledgeQAPairExportTaskType
}

func (t *KnowledgeQAPairExportTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return knowledgeQAPairExportTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *KnowledgeQAPairExportTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "KnowledgeQAPairExportTask", t.BuildServiceType(), string(paramsStr), true)
	log.Infof("knowledge qa pair Export task %d", taskId)
	return err
}

func (t *KnowledgeQAPairExportTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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

		//执行问答库导出
		systemStop, err := t.runStep(ctx, taskCtx, stop)
		if systemStop {
			log.Infof("system stop")
			return
		}
		if err != nil {
			log.Errorf("executeKnowledgeQAPairExportTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *KnowledgeQAPairExportTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *KnowledgeQAPairExportTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- exportKnowledgeQAPair(ctx, taskCtx)
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

// exportKnowledgeQAPair 导出问答对
func exportKnowledgeQAPair(ctx context.Context, taskCtx string) Result {
	log.Infof("KnowledgeQAPairExportTask execute task %s", taskCtx)
	var params = &async_task_pkg.KnowledgeQAPairExportTaskParams{}
	err := json.Unmarshal([]byte(taskCtx), params)
	if err != nil {
		return Result{Error: err}
	}

	//1.查询问答库导出任务
	task, err := orm.SelectKnowledgeExportTaskById(ctx, params.TaskId)
	if err != nil {
		return Result{Error: err}
	}

	//2.更新状态处理中
	err = orm.UpdateKnowledgeExportTask(ctx, params.TaskId, model.KnowledgeExportExporting, "", 0, 0, "", 0)
	if err != nil {
		log.Errorf("UpdateQAPairExportTaskStatus err: %s", err)
		return Result{Error: err}
	}
	//3.执行导出
	lineCount, successCount, err := doKnowledgeQAPairExport(ctx, task)
	if err != nil {
		log.Errorf("knowledge qa pair export err: %s, lineCount %d successCount %d", err, lineCount, successCount)
		return Result{Error: err}
	}
	log.Infof("knowledge qa pair export lineCount %d successCount %d", lineCount, successCount)
	return Result{Error: err}
}

// PrintPanicStackWithCall 执行文件导出
func doKnowledgeQAPairExport(ctx context.Context, exportTask *model.KnowledgeExportTask) (lineCount int64, successCount int64, err error) {
	filePath := ""
	fileSize := int64(0)
	defer util.PrintPanicStackWithCall(func(panicOccur bool, err2 error) {
		if panicOccur {
			log.Errorf("do knowledge qa pair export task panic: %v", err2)
			err = fmt.Errorf("文件导出异常")
		}
		var status = model.KnowledgeExportSuccess
		var errMsg string
		if err != nil {
			status = model.KnowledgeExportFail
			errMsg = err.Error()
		}
		if lineCount == 0 {
			status = model.KnowledgeExportFail
			errMsg = "文件所有行全部处理失败"
		}
		//更新状态和数量
		err = orm.UpdateKnowledgeExportTask(ctx, exportTask.ExportId, status, errMsg, lineCount, successCount, filePath, fileSize)
	})
	lineCount, successCount, filePath, fileSize, err = exportCsvFile(ctx, exportTask.KnowledgeId)
	if err != nil {
		log.Errorf("knowledge qa pair export err: %s, knowledgeId %v doclineCount %d docSuccessCount %d", err, exportTask.KnowledgeId, lineCount, successCount)
		return
	}
	return
}

func exportCsvFile(ctx context.Context, knowledgeId string) (int64, int64, string, int64, error) {
	exportFilePath := file_util.BuildFilePath(exportLocalDir, ".csv")
	err := os.MkdirAll(filepath.Dir(exportFilePath), 0755)
	if err != nil {
		return 0, 0, "", 0, err
	}
	file, err := os.OpenFile(exportFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0644)
	if err != nil {
		return 0, 0, "", 0, err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Infof("Error closing file: %v", err)
		}
		err = os.Remove(exportFilePath)
		if err != nil {
			log.Infof("Error remove file: %v", err)
		}
	}()
	// 创建CSV写入器
	writer := csv.NewWriter(file)

	// 根据需要配置CSV写入器
	writer.Comma = ',' // 设置分隔符，默认为逗号

	// 写入表头行
	err = writer.Write([]string{"question", "answer"})
	if err != nil {
		return 0, 0, "", 0, err
	}
	var lineCount, successCount, batchNo int64 = 0, 0, 0
	for {
		batchNo++
		qaPairs, total, err := orm.GetQAPairList(ctx, "", "", knowledgeId, "", model.KnowledgeExportSuccess, nil, exportBatchSize, int32(batchNo))
		if err != nil {
			return 0, 0, "", 0, err
		}
		if len(qaPairs) <= 0 {
			break
		}
		lineCount = total
		var records [][]string
		for _, qaPair := range qaPairs {
			records = append(records, []string{string(qaPair.Question), string(qaPair.Answer)})
		}
		err = writer.WriteAll(records)
		if err != nil {
			return 0, 0, "", 0, err
		}
		successCount += int64(len(records))
	}
	currentDate := time.Now().Format("2006-01-02")
	dir := config.GetConfig().Minio.QAExportDir + "/" + currentDate
	bucketName := config.GetConfig().Minio.PublicExportBucket
	_, minioFilePath, fileSize, err := service.UploadLocalFile(ctx, dir, bucketName, filepath.Base(exportFilePath), exportFilePath)
	if err != nil {
		log.Errorf("upload file err: %v", err)
		return 0, 0, "", 0, err
	}
	bucket, objectName, _ := service.SplitFilePath(minioFilePath)
	filePath := bucket + "/" + objectName
	return lineCount, successCount, filePath, fileSize, nil
}
