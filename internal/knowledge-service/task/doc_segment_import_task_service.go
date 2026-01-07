package task

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode/utf8"

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	async_task_pkg "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var docSegmentImportTask = &DocSegmentImportTask{Del: true}

type DocSegmentImportTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(docSegmentImportTask)
}

func (t *DocSegmentImportTask) BuildServiceType() uint32 {
	return async_task_pkg.DocSegmentImportTaskType
}

func (t *DocSegmentImportTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return docSegmentImportTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *DocSegmentImportTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "DocSegmentImportTask", t.BuildServiceType(), string(paramStr), true)
	log.Infof("doc segment import task %d ", taskId)
	return err
}

func (t *DocSegmentImportTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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
			log.Errorf("execute DocSegmentImportTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *DocSegmentImportTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *DocSegmentImportTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- importDocSegment(ctx, taskCtx)
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

func importDocSegment(ctx context.Context, taskCtx string) Result {
	var docSegmentImportTaskParams = &async_task_pkg.DocSegmentImportTaskParams{}
	err := json.Unmarshal([]byte(taskCtx), docSegmentImportTaskParams)
	if err != nil {
		log.Errorf("unmarshal json err: %s", err)
		return Result{Error: err}
	}
	importTask, err := orm.SelectDocSegmentImportTaskById(ctx, docSegmentImportTaskParams.TaskId)
	if err != nil {
		log.Errorf("select doc segment import task err: %s", err)
		return Result{Error: err}
	}
	//状态校验
	if importTask.Status != model.DocSegmentImportInit {
		log.Infof("knowledge import task not need process : %s status %d", importTask.ImportId, importTask.Status)
		return Result{Error: err}
	}
	var importTaskParams = model.DocSegmentImportParams{}
	err = json.Unmarshal([]byte(importTask.ImportParams), &importTaskParams)
	if err != nil {
		log.Errorf("doc segment import params err: %s", err)
		return Result{Error: err}
	}
	//更新状态处理中
	err = orm.UpdateDocSegmentImportTaskStatus(ctx, docSegmentImportTaskParams.TaskId, model.DocSegmentImportImporting, "", 0)
	if err != nil {
		log.Errorf("UpdateDocSegmentImportTaskStatus err: %s", err)
		return Result{Error: err}
	}
	//执行导入
	lineCount, err := doDocSegmentImport(ctx, &importTaskParams, importTask)
	if err != nil {
		log.Errorf("doc segment file download err: %s, lineCount %d", err, lineCount)
		return Result{Error: err}
	}
	log.Infof("doc segment file download  lineCount %d", lineCount)
	return Result{Error: err}
}

// doDocSegmentImport 执行文件导入
func doDocSegmentImport(ctx context.Context, importTaskParams *model.DocSegmentImportParams, importTask *model.DocSegmentImportTask) (lineCount int, err error) {
	defer util.PrintPanicStackWithCall(func(panicOccur bool, err2 error) {
		if panicOccur {
			log.Errorf("do doc import task panic: %v", err2)
			err = fmt.Errorf("文件导入异常")
		}
		var status = model.DocSegmentImportSuccess
		var errMsg string
		if err != nil {
			status = model.DocSegmentImportFail
			errMsg = err.Error()
		}
		if lineCount == 0 {
			status = model.DocSegmentImportFail
			errMsg = "文件所有行全部处理失败"
		}
		//更新状态和数量
		err = orm.UpdateDocSegmentImportTaskStatus(ctx, importTask.ImportId, status, errMsg, lineCount)
	})
	lineCount, err = processCsvFileLine(ctx, importTaskParams.FileUrl, buildLineProcessor(importTask, importTaskParams))
	return
}

// csv 文件行处理器
func buildLineProcessor(importTask *model.DocSegmentImportTask, importParams *model.DocSegmentImportParams) func(ctx context.Context, strings []string) error {
	return func(ctx context.Context, lineData []string) error {
		if utf8.RuneCountInString(lineData[0]) > importParams.MaxSentenceSize {
			return errors.New("line exceeds max sentence")
		}

		var chunks []*service.NewChunkItem
		chunks = append(chunks, &service.NewChunkItem{
			Content: lineData[0],
			Labels:  strings.Split(lineData[1], ","),
		})

		return orm.CreateOneDocSegment(ctx, importTask, &service.RagCreateDocSegmentParams{
			UserId:           importParams.KnowledgeCreatorId,
			KnowledgeBase:    buildKnowledgeBase(importParams),
			KnowledgeId:      importParams.KnowledgeId,
			FileName:         importParams.FileName,
			MaxSentenceSize:  importParams.MaxSentenceSize,
			Chunks:           chunks,
			SplitType:        service.RebuildSplitType(importParams.SegmentMethod),
			ChildChunkConfig: service.RebuildChildChunkConfig(importParams.SegmentMethod, importParams.SubMaxSplitter, importParams.SubSplitter),
		})
	}
}

func buildKnowledgeBase(importParams *model.DocSegmentImportParams) string {
	if len(importParams.KnowledgeRagName) > 0 {
		return importParams.KnowledgeRagName
	}
	return importParams.KnowledgeName
}

func processCsvFileLine(ctx context.Context, csvUrl string,
	lineProcessor func(context.Context, []string) error) (int, error) {

	//下载url，循环调用rag
	object, err := service.DownloadFileObject(ctx, csvUrl)
	if err != nil {
		log.Errorf("download file err: %s", err)
		return 0, err
	}

	defer func() {
		err2 := object.Close()
		if err2 != nil {
			log.Errorf("close file err: %s", err2)
		}
	}()

	// 创建CSV读取器
	reader := csv.NewReader(object)

	// 根据需要配置CSV读取器
	reader.Comma = ','          // 设置分隔符，默认为逗号
	reader.Comment = '#'        // 设置注释字符
	reader.FieldsPerRecord = -1 // 允许可变字段数量

	// 读取并跳过表头行
	_, err = reader.Read()
	if err != nil {
		if err == io.EOF {
			log.Errorf("csv file is empty")
			return 0, nil
		}
		log.Errorf("read header line err: %v", err)
	}

	var lineCount = 0
	// 逐行读取CSV内容
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// 可以选择记录错误并继续，或者直接返回错误
			log.Errorf("解析CSV行时出错: %v, lineCount %d", err, lineCount)
			continue
		}
		lineCount++
		if len(record) < 2 {
			err = fmt.Errorf("line data not ok lineCount %d", lineCount)
			// 可以选择记录错误并继续，或者直接返回错误
			log.Errorf("解析CSV行时出错: %v", err)
			continue
		}

		err = lineProcessor(ctx, record)
		if err != nil {
			log.Errorf("process csv line lineCount %d err: %s", lineCount, err)
			continue
		}
	}

	return lineCount, nil
}
