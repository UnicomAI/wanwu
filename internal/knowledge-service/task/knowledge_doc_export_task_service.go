package task

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const fileTypeUrl = "url"

var knowledgeDocExportTask = &KnowledgeDocExportTask{Del: true}

type KnowledgeDocExportTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(knowledgeDocExportTask)
}

func (t *KnowledgeDocExportTask) BuildServiceType() uint32 {
	return async_task_pkg.KnowledgeDocExportTaskType
}

func (t *KnowledgeDocExportTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return knowledgeDocExportTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *KnowledgeDocExportTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "KnowledgeDocExportTask", t.BuildServiceType(), string(paramsStr), true)
	log.Infof("knowledge doc Export task %d", taskId)
	return err
}

func (t *KnowledgeDocExportTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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
			log.Errorf("executeKnowledgeDocExportTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *KnowledgeDocExportTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *KnowledgeDocExportTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- exportKnowledgeDoc(ctx, taskCtx)
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

// exportKnowledgeDoc 导出知识库
func exportKnowledgeDoc(ctx context.Context, taskCtx string) Result {
	log.Infof("KnowledgeDocExportTask execute task %s", taskCtx)
	var params = &async_task_pkg.KnowledgeDocExportTaskParams{}
	err := json.Unmarshal([]byte(taskCtx), params)
	if err != nil {
		return Result{Error: err}
	}

	//1.查询知识库导出任务
	task, err := orm.SelectKnowledgeExportTaskById(ctx, params.TaskId)
	if err != nil {
		return Result{Error: err}
	}

	//2.查询知识库库详情
	knowledge, err := orm.SelectKnowledgeById(ctx, task.KnowledgeId, "", "")
	if err != nil {
		return Result{Error: err}
	}

	var exportTaskParams = model.KnowledgeExportTaskParams{}
	err = json.Unmarshal([]byte(task.ExportParams), &exportTaskParams)
	if err != nil {
		log.Errorf("knowledge export params err: %s", err)
		return Result{Error: err}
	}

	//2.更新状态处理中
	err = orm.UpdateKnowledgeExportTask(ctx, params.TaskId, model.KnowledgeExportExporting, "", 0, 0, "", 0)
	if err != nil {
		log.Errorf("UpdateDocExportTaskStatus err: %s", err)
		return Result{Error: err}
	}
	//3.执行导出
	totalCount, successCount, err := doKnowledgeDocExport(ctx, knowledge, task, &exportTaskParams)
	if err != nil {
		log.Errorf("knowledge doc export err: %s, totalCount %d successCount %d", err, totalCount, successCount)
		return Result{Error: err}
	}
	log.Infof("knowledge doc export totalCount %d successCount %d", totalCount, successCount)
	return Result{Error: err}
}

// PrintPanicStackWithCall 执行文件导出
func doKnowledgeDocExport(ctx context.Context, knowledge *model.KnowledgeBase, exportTask *model.KnowledgeExportTask, exportParams *model.KnowledgeExportTaskParams) (totalCount int64, successCount int64, err error) {
	filePath := ""
	fileSize := int64(0)
	defer util.PrintPanicStackWithCall(func(panicOccur bool, err2 error) {
		if panicOccur {
			log.Errorf("do knowledge doc export task panic: %v", err2)
			err = fmt.Errorf("文件导出异常")
		}
		var status = model.KnowledgeExportSuccess
		var errMsg string
		if err != nil {
			status = model.KnowledgeExportFail
			errMsg = err.Error()
		}
		if totalCount == 0 {
			status = model.KnowledgeExportFail
			errMsg = "文件全部处理失败"
		}
		//更新状态和数量
		err = orm.UpdateKnowledgeExportTask(ctx, exportTask.ExportId, status, errMsg, totalCount, successCount, filePath, fileSize)
	})
	totalCount, successCount, filePath, fileSize, err = exportDocFiles(ctx, knowledge, exportParams)
	if err != nil {
		log.Errorf("knowledge doc export err: %s, knowledgeId %v doclineCount %d docSuccessCount %d", err, exportTask.KnowledgeId, totalCount, successCount)
		return
	}
	return
}

func exportDocFiles(ctx context.Context, knowledge *model.KnowledgeBase, exportParams *model.KnowledgeExportTaskParams) (int64, int64, string, int64, error) {
	var totalCount, successCount int64
	var docInfos []*model.KnowledgeDoc
	var err error
	if len(exportParams.DocIdList) <= 0 {
		docInfos, err = orm.GetDocListByKnowledgeIdAndFileTypeFilter(ctx, "", "", exportParams.KnowledgeId, fileTypeUrl)
		if err != nil {
			log.Errorf("GetDocListByKnowledgeId err: %s", err)
			return 0, 0, "", 0, err
		}
		totalCount = int64(len(docInfos))
	} else {
		docInfos, err = orm.SelectDocByDocIdListAndFileTypeFilter(ctx, exportParams.DocIdList, "", "", fileTypeUrl)
		if err != nil {
			log.Errorf("SelectDocByDocIdList err: %s", err)
			return 0, 0, "", 0, err
		}
		totalCount = int64(len(docInfos))
	}
	exportZipFilePath := exportLocalDir + knowledge.Name + time.Now().Format("20060102150405") + ".zip"
	err = os.MkdirAll(filepath.Dir(exportZipFilePath), 0755)
	if err != nil {
		log.Infof("Error create directory: %v", err)
		return 0, 0, "", 0, err
	}
	exportZipFile, err := os.Create(exportZipFilePath)
	if err != nil {
		log.Infof("Error opening file: %v", err)
		return 0, 0, "", 0, err
	}
	defer func() {
		if err = exportZipFile.Close(); err != nil {
			log.Infof("Error closing file: %v", err)
		}
		if err = os.Remove(exportZipFilePath); err != nil {
			log.Infof("Error remove file: %v", err)
		}
	}()
	zipWriter := zip.NewWriter(exportZipFile)
	defer func() {
		if err = zipWriter.Close(); err != nil {
			log.Infof("Error closing zip writer: %v", err)
		}
	}()
	for _, docInfo := range docInfos {
		exportFilePath := exportLocalDir + "/" + docInfo.Name
		err = processExportFileDoc(ctx, docInfo.FilePath, exportFilePath, zipWriter)
		if err != nil {
			log.Errorf("processExportFileDoc err: %s", err)
			return 0, 0, "", 0, err
		}
		successCount++
	}
	err = zipWriter.Close()
	if err != nil {
		log.Errorf("zipWriter.Close err: %s", err)
		return 0, 0, "", 0, err
	}
	currentDate := time.Now().Format("2006-01-02")
	dir := config.GetConfig().Minio.KnowledgeExportDir + "/" + currentDate
	bucketName := config.GetConfig().Minio.PublicExportBucket
	_, minioFilePath, fileSize, err := service.UploadLocalFile(ctx, dir, bucketName, filepath.Base(exportZipFilePath), exportZipFilePath)
	if err != nil {
		log.Errorf("upload file err: %v", err)
		return 0, 0, "", 0, err
	}
	bucket, objectName, _ := service.SplitFilePath(minioFilePath)
	filePath := bucket + "/" + objectName
	return totalCount, successCount, filePath, fileSize, nil
}

func processExportFileDoc(ctx context.Context, srcFilePath string, exportFilePath string, zipWriter *zip.Writer) error {
	err := service.DownloadFileToLocal(ctx, srcFilePath, exportFilePath)
	if err != nil {
		log.Errorf("download file err: %s", err)
		return err
	}
	exportFile, err := os.Open(exportFilePath)
	if err != nil {
		log.Errorf("open file err: %s", err)
		return err
	}
	defer func() {
		if err = exportFile.Close(); err != nil {
			log.Infof("Error closing file: %v", err)
		}
		if err = os.Remove(exportFilePath); err != nil {
			log.Infof("Error remove file: %v", err)
		}
	}()
	fileName := filepath.Base(exportFilePath)
	writer, err := zipWriter.Create(fileName)
	if err != nil {
		log.Errorf("create file err: %s", err)
		return err
	}
	if _, err = io.Copy(writer, exportFile); err != nil {
		log.Infof("Error copy file: %v", err)
		return err
	}
	return nil
}
