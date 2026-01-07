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

	async "github.com/UnicomAI/wanwu/async"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	async_task_pkg "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	batchSize = 20
	lineLimit = 1000
)

var knowledgeQAPairImportTask = &KnowledgeQAPairImportTask{Del: true}

type KnowledgeQAPairImportTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

func init() {
	async_task_pkg.AddContainer(knowledgeQAPairImportTask)
}

func (t *KnowledgeQAPairImportTask) BuildServiceType() uint32 {
	return async_task_pkg.KnowledgeQAPairImportTaskType
}

func (t *KnowledgeQAPairImportTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return knowledgeQAPairImportTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *KnowledgeQAPairImportTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "KnowledgeQAPairImportTask", t.BuildServiceType(), string(paramsStr), true)
	log.Infof("knowledge qa pair import task %d", taskId)
	return err
}

func (t *KnowledgeQAPairImportTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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

		//执行问答库导入
		systemStop, err := t.runStep(ctx, taskCtx, stop)
		if systemStop {
			log.Infof("system stop")
			return
		}
		if err != nil {
			log.Errorf("executeKnowledgeQAPairImportTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *KnowledgeQAPairImportTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *KnowledgeQAPairImportTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- importKnowledgeQAPair(ctx, taskCtx)
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

// importKnowledgeQAPair 导入问答对
func importKnowledgeQAPair(ctx context.Context, taskCtx string) Result {
	log.Infof("KnowledgeQAPairImportTask execute task %s", taskCtx)
	var params = &async_task_pkg.KnowledgeQAPairImportTaskParams{}
	err := json.Unmarshal([]byte(taskCtx), params)
	if err != nil {
		return Result{Error: err}
	}

	//1.查询问答库导入任务
	task, err := orm.SelectKnowledgeQAPairImportTaskById(ctx, params.TaskId)
	if err != nil {
		return Result{Error: err}
	}
	var importDocInfo = model.DocImportInfo{}
	err = json.Unmarshal([]byte(task.DocInfo), &importDocInfo)
	if err != nil {
		log.Errorf("Unmarshal fail %v", err)
		return Result{Error: err}
	}

	//2.查询问答库详情
	knowledge, err := orm.SelectKnowledgeById(ctx, task.KnowledgeId, "", "")
	if err != nil {
		return Result{Error: err}
	}

	//3.更新状态处理中
	err = orm.UpdateKnowledgeQAPairImportTaskStatus(ctx, nil, params.TaskId, model.KnowledgeQAPairImportImporting, "", 0, 0)
	if err != nil {
		log.Errorf("UpdateQAPairImportTaskStatus err: %s", err)
		return Result{Error: err}
	}
	//4.执行导入
	lineCount, successCount, err := doKnowledgeQAPairImport(ctx, knowledge, &importDocInfo, task)
	if err != nil {
		log.Errorf("knowledge qa pair import err: %s, lineCount %d successCount %d", err, lineCount, successCount)
		return Result{Error: err}
	}
	log.Infof("knowledge qa pair import lineCount %d successCount %d", lineCount, successCount)
	return Result{Error: err}
}

// PrintPanicStackWithCall 执行文件导入
func doKnowledgeQAPairImport(ctx context.Context, knowledgeBase *model.KnowledgeBase, importTaskParams *model.DocImportInfo, importTask *model.KnowledgeQAPairImportTask) (lineCount int64, successCount int64, err error) {
	defer util.PrintPanicStackWithCall(func(panicOccur bool, err2 error) {
		if panicOccur {
			log.Errorf("do knowledge qa pair import task panic: %v", err2)
			err = fmt.Errorf("文件导入异常")
		}
		var status = model.KnowledgeQAPairImportSuccess
		var errMsg string
		if err != nil {
			status = model.KnowledgeQAPairImportFail
			errMsg = err.Error()
		}
		if successCount == 0 {
			status = model.KnowledgeQAPairImportFail
			errMsg = "文件所有行全部处理失败"
		}
		//更新状态和数量
		err = orm.UpdateKnowledgeQAPairImportTaskStatusAndCount(ctx, importTask.ImportId, status, errMsg, lineCount, successCount, knowledgeBase.KnowledgeId)
		if err != nil {
			log.Errorf("UpdateQAPairImportTaskStatus err: %s", err)
			return
		}
	})
	for _, docInfo := range importTaskParams.DocInfoList {
		docLineCount, docSuccessCount, err := processCsvFileBatchLine(ctx, docInfo, buildQAPairBatchProcessor(knowledgeBase, importTask))
		if err != nil {
			log.Errorf("knowledge qa pair import err: %s, doc %v doclineCount %d docSuccessCount %d", err, docInfo, lineCount, successCount)
			continue
		}
		lineCount += docLineCount
		successCount += docSuccessCount
	}
	return
}

func processCsvFileBatchLine(ctx context.Context, docInfo *model.DocInfo,
	batchProcessor func(context.Context, [][]string) (int64, error)) (int64, int64, error) {

	//下载url，循环调用rag
	object, err := service.DownloadFileObject(ctx, docInfo.DocUrl)
	if err != nil {
		log.Errorf("download file err: %s", err)
		return 0, 0, err
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
			return 0, 0, nil
		}
		log.Errorf("read header line err: %v", err)
	}

	var lineCount int64
	var successCount int64
	batchRecord := make([][]string, 0)
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
		if lineCount > lineLimit {
			break
		}
		if len(record) < 2 {
			err = fmt.Errorf("line data not ok lineCount %d", lineCount)
			// 可以选择记录错误并继续，或者直接返回错误
			log.Errorf("解析CSV行时出错: %v", err)
			continue
		}
		batchRecord = append(batchRecord, record)
		if lineCount%batchSize == 0 {
			batchSuccessCount, err := batchProcessor(ctx, batchRecord)
			if err != nil {
				log.Errorf("process csv batch lineCount %d err: %s", lineCount, err)
				continue
			}
			successCount += batchSuccessCount
			batchRecord = make([][]string, 0)
		}
	}
	if len(batchRecord) > 0 {
		batchSuccessCount, err := batchProcessor(ctx, batchRecord)
		if err != nil {
			log.Errorf("process csv batch lineCount %d err: %s", lineCount, err)
		}
		successCount += batchSuccessCount
	}
	return lineCount, successCount, nil
}

// csv 文件批量处理器
func buildQAPairBatchProcessor(knowledgeBase *model.KnowledgeBase, importTask *model.KnowledgeQAPairImportTask) func(ctx context.Context, strings [][]string) (int64, error) {
	return func(ctx context.Context, batchData [][]string) (int64, error) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("knowledge qa pair batch process panic %+v", r)
			}
		}()
		var chunks []*service.RagQAPairItem
		var QAPairs []*model.KnowledgeQAPair
		var successCount int64 = 0
		for _, lineData := range batchData {
			qaPairId := util.NewID()
			question := strings.Trim(lineData[0], " ")
			answer := strings.Trim(lineData[1], " ")
			questionMD5 := util.MD5([]byte(question))
			err := orm.CheckKnowledgeQAPairQuestion(ctx, "", knowledgeBase.KnowledgeId, questionMD5)
			if err != nil {
				continue
			}
			chunks = append(chunks, &service.RagQAPairItem{
				QAPairId: qaPairId,
				Question: question,
				Answer:   answer,
			})
			QAPairs = append(QAPairs, &model.KnowledgeQAPair{
				QAPairId:     qaPairId,
				ImportTaskId: importTask.ImportId,
				KnowledgeId:  knowledgeBase.KnowledgeId,
				Question:     db.LongText(question),
				Answer:       db.LongText(answer),
				Status:       model.KnowledgeQAPairImportSuccess,
				Switch:       true,
				QuestionMd5:  questionMD5,
				UserId:       importTask.UserId,
				OrgId:        importTask.OrgId,
			})
			successCount++
		}
		err := orm.CreateKnowledgeQAPair(ctx, QAPairs, &service.RagAddQAPairParams{
			UserId:     knowledgeBase.UserId,
			QAId:       knowledgeBase.KnowledgeId,
			QABaseName: knowledgeBase.RagName,
			QAPairs:    chunks,
		})
		if err != nil {
			log.Errorf("knowledge qa pair batch process err: %s", err)
			return 0, err
		}
		return successCount, nil
	}
}
