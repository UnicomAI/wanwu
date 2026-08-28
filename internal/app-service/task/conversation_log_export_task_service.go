package task

import (
	"context"
	"encoding/json"
	"errors"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/config"
	async_task_pkg "github.com/UnicomAI/wanwu/internal/app-service/pkg/async-task"
	csv_util "github.com/UnicomAI/wanwu/pkg/csv-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	async "github.com/gromitlee/go-async"
	"github.com/gromitlee/go-async/pkg/async/async_task"
	"strconv"
	"strings"
	"sync"
)

const (
	exportLocalDir = "static/export/"
	csvSuffix      = ".csv"
)

var conversationLogExportCsvHeader = []string{
	"来源", "使用者", "标题", "会话 ID", "创建时间", "最后对话时间", "消息总数", "用户反馈", "平均响应时长", "报错数量", "版本名", "对话详情",
}

var conversationLogExportTask = &ConversationLogExportTask{Del: true}

type ConversationLogExportTask struct {
	Wg  sync.WaitGroup
	Del bool // 是否需要自动清理
}

type Result struct {
	Error error
}

func init() {
	async_task_pkg.AddContainer(conversationLogExportTask)
}

func (t *ConversationLogExportTask) BuildServiceType() uint32 {
	return async_task_pkg.ConversationLogExportTaskType
}

func (t *ConversationLogExportTask) InitTask() error {
	if err := async.RegisterTask(t.BuildServiceType(), func() async_task.ITask {
		return conversationLogExportTask
	}); err != nil {
		return err
	}
	return nil
}

func (t *ConversationLogExportTask) SubmitTask(ctx context.Context, params interface{}) (err error) {
	if params == nil {
		return errors.New("参数不能为空")
	}
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var taskId uint32
	taskId, err = async.CreateTask(ctx, "", "ConversationLogExportTask", t.BuildServiceType(), string(paramsStr), true)
	log.Infof("conversation log export task %d", taskId)
	return err
}

func (t *ConversationLogExportTask) Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
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

		// 执行对话日志导出
		systemStop, err := t.runStep(ctx, taskCtx, stop)
		if systemStop {
			log.Infof("system stop")
			return
		}
		if err != nil {
			log.Errorf("executeConversationLogExportTask err: %s", err)
			r.phase = async_task.RunPhaseFailed
			return
		} else {
			r.phase = async_task.RunPhaseFinished
			return
		}

	}()

	return reportCh
}

func (t *ConversationLogExportTask) Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan async_task.IReport {
	return CommonDeleting(ctx, taskCtx, stop, &t.Wg)
}

func (t *ConversationLogExportTask) runStep(ctx context.Context, taskCtx string, stop <-chan struct{}) (bool, error) {
	ret := make(chan Result, 1)
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		ret <- exportConversationLog(ctx, taskCtx)
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

// exportConversationLog 执行对话日志导出主流程。
func exportConversationLog(ctx context.Context, taskCtx string) Result {
	log.Infof("ConversationLogExportTask execute task %s", taskCtx)
	var params = &async_task_pkg.ConversationLogExportTaskParams{}
	if err := json.Unmarshal([]byte(taskCtx), params); err != nil {
		return Result{Error: err}
	}

	// 1. 乐观锁更新 0-任务待处理 -> 1-任务导出中
	affected, err := ormClient.TrySetConvLogExpTask(ctx, params.TaskId)
	if err != nil {
		return Result{Error: err}
	}
	// 没有数据（已被删除）或 已执行任务
	if affected == 0 {
		log.Infof("conversation log export task %s already taken or finished, skip", params.TaskId)
		return Result{Error: nil}
	}

	// 2.查询导出任务
	task, err := ormClient.SelectConversationLogExportTaskById(ctx, params.TaskId)
	if err != nil {
		return Result{Error: err}
	}

	// 3.执行导出
	if err = doExportConversationLog(ctx, task); err != nil {
		log.Errorf("conversation log export err: %s", err)
		return Result{Error: err}
	}
	return Result{Error: nil}
}

// doExportConversationLog 执行文件导出，defer 中终态化任务状态（成功写 file，失败写 errorMsg）。
func doExportConversationLog(ctx context.Context, exportTask *model.ConversationLogExportTask) (err error) {
	filePath := ""
	fileSize := int64(0)
	totalCount := 0
	successCount := 0
	defer util.PrintPanicStackWithCall(func(panicOccur bool, err2 error) {
		if panicOccur {
			log.Errorf("do conversation log export task panic: %v", err2)
			err = errors.New("文件导出异常")
		}
		status := model.ConversationLogExportSuccess
		var errMsg string
		if err != nil {
			status = model.ConversationLogExportFail
			errMsg = err.Error()
		}
		if err = ormClient.UpdateConversationLogExportTask(ctx, exportTask.ExportId, status, errMsg, totalCount, successCount, filePath, fileSize); err != nil {
			log.Errorf("update conversation log export task final status err: %s", err)
		}
	})
	csvResult, err := exportConversationLogCsvFile(ctx, exportTask)
	if err != nil {
		return
	}
	totalCount, successCount, filePath, fileSize = csvResult.TotalCount, csvResult.SuccessCount, csvResult.MinioPath, csvResult.FileSize
	return
}

func exportConversationLogCsvFile(ctx context.Context, exportTask *model.ConversationLogExportTask) (*csv_util.ExportCsvResult, error) {
	dir := config.Cfg().Minio.AppLogExportDir + "/" + strings.Split(util.GenUUID(), "-")[0]
	exportParams := &csv_util.ExportCsvParams{
		ExportLocalDir:  exportLocalDir,
		CsvHeader:       conversationLogExportCsvHeader,
		MinioBucketDir:  dir,
		MinioBucketName: config.Cfg().Minio.PublicExportBucket,
	}
	dataGetter := buildDataGetter(ctx, exportTask)
	return csv_util.ExportCsvFile[*model.ConversationLog](ctx, exportParams, dataGetter, buildLineProcessor())
}

func buildDataGetter(ctx context.Context, exportTask *model.ConversationLogExportTask) func(ctx context.Context) (csv_util.CsvData[*model.ConversationLog], error) {
	return func(ctx context.Context) (csv_util.CsvData[*model.ConversationLog], error) {
		// 查询对话日志基本信息 根据exportTask.ExportParams 查询
		logs, err := getConversationLogsByExportParams(ctx, exportTask.ExportParams)
		if err != nil {
			return csv_util.CsvData[*model.ConversationLog]{}, err
		}
		return csv_util.CsvData[*model.ConversationLog]{
			DataList: logs,
		}, nil
	}
}

func buildLineProcessor() func(ctx context.Context, c csv_util.CsvData[*model.ConversationLog], item *model.ConversationLog) []string {
	return func(ctx context.Context, c csv_util.CsvData[*model.ConversationLog], item *model.ConversationLog) []string {
		dataCon, err := getConversationByLogItem(ctx, item)
		log.Infof("item %v", item)
		if err != nil {
			log.Errorf("get last conversation by log item err: %s", err)
			dataCon = ""
		}
		return conversationLogToRecord(item, dataCon)
	}
}

// getConversationByLogItem 按 appType 分发到对应处理器。
func getConversationByLogItem(ctx context.Context, lg *model.ConversationLog) (string, error) {
	if lg == nil || lg.ConversationId == "" || lg.AppType == "" || lg.AppId == "" {
		return "", errors.New("conversation_id、 app_type or app_id is empty")
	}
	handler, ok := conversationDetailHandlers[lg.AppType]
	if !ok {
		log.Errorf("no conversationDetailHandler registered for appType %s, appId %s, conversationId %s", lg.AppType, lg.AppId, lg.ConversationId)
		return "", nil
	}
	return handler.GetConversationDetail(ctx, lg)
}

func conversationLogToRecord(item *model.ConversationLog, conversationDetail string) []string {
	return []string{
		item.Source,
		item.UserId,
		item.Title,
		item.ConversationId,
		strconv.FormatInt(item.CreatedAt, 10),
		strconv.FormatInt(item.UpdatedAt, 10),
		strconv.Itoa(item.MessageCount),
		"good : " + strconv.Itoa(item.LikeCount) + ", bad : " + strconv.Itoa(item.DisLikeCount),
		strconv.FormatInt(item.Costs, 10),
		strconv.Itoa(item.ErrorCount),
		item.Version,
		conversationDetail,
	}
}

func getConversationLogsByExportParams(ctx context.Context, exportParams string) ([]*model.ConversationLog, error) {
	// 权限由 bff 在创建导出任务时通过 FilterOwnerUser 计算并落库到 params.OrgIds/UserIds：
	//   1. 请求者 appid 权限 —— owner 全量(orgIds/userIds 空)、非 owner 仅自己
	//   2. logids 属于对应 appid —— SQL 加 app_id/app_type 过滤
	//   3. 该用户有 logid 权限 —— SQL 加 org_id/user_id 过滤
	var params model.ConversationLogExportTaskParams
	if err := json.Unmarshal([]byte(exportParams), &params); err != nil {
		return nil, err
	}
	// 有 LogIds 时只按 LogIds 查（同时校验 app 归属与可见范围）；否则按 AppId/AppType 查全量
	var (
		logs   []*model.ConversationLog
		status *errs.Status
	)
	if len(params.LogIds) > 0 {
		logs, status = ormClient.GetConversationLogListByLogIds(ctx, params.LogIds, params.AppId, params.AppType, params.OrgIds, params.UserIds)
	} else {
		// 全量导出：limit 传 -1 表示不限（GORM 中 Limit(0) 会生成 SQL `LIMIT 0` 查出 0 条），offset 传 0。
		logs, _, status = ormClient.GetConversationLogList(ctx, params.OrgIds, params.UserIds, params.AppId, params.AppType, "", nil, "", "", "", "", 0, -1)
	}
	if status != nil {
		log.Errorf("query conversation log for export err: %v", status)
		return nil, errors.New(status.GetTextKey())
	}
	return logs, nil
}
