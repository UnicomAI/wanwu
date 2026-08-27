package app

import (
	"context"
	"encoding/json"
	"path/filepath"

	"google.golang.org/protobuf/types/known/emptypb"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
)

func (s *Service) GetConversationLogList(ctx context.Context, req *app_service.GetConversationLogListReq) (*app_service.GetConversationLogListResp, error) {
	list, total, err := s.cli.GetConversationLogList(ctx, req.OrgIds, req.UserIds, req.AppId, req.AppType, req.Name, req.Source, req.StartDate, req.EndDate, req.OrderBy, req.OrderType, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(err_code.Code_AppConversationLog, err)
	}
	return convertConversationLogList(list, total), nil
}

// GetConversationLog 根据 appId、appType、conversationId 查询单条会话日志。
func (s *Service) GetConversationLog(ctx context.Context, req *app_service.GetConversationLogReq) (*common.ConversationLog, error) {
	log, err := s.cli.GetConversationLog(ctx, req.AppId, req.AppType, req.ConversationId)
	if err != nil {
		return nil, errStatus(err_code.Code_AppConversationLog, err)
	}
	return toCommonConversationLog(log), nil
}

// GetConversationLogByLogIds 根据 logId 列表批量查询会话日志。
func (s *Service) GetConversationLogByLogIds(ctx context.Context, req *app_service.GetConversationLogByLogIdsReq) (*app_service.GetConversationLogByLogIdsResp, error) {
	logs, err := s.cli.GetConversationLogByLogIds(ctx, req.LogIds)
	if err != nil {
		return nil, errStatus(err_code.Code_AppConversationLog, err)
	}
	items := make([]*common.ConversationLog, 0, len(logs))
	for _, log := range logs {
		items = append(items, toCommonConversationLog(log))
	}
	return &app_service.GetConversationLogByLogIdsResp{Items: items}, nil
}

// GetConversationLogUserSelect 获取会话日志使用者列表（去重后的 userId）
func (s *Service) GetConversationLogUserSelect(ctx context.Context, req *app_service.GetConversationLogUserSelectReq) (*app_service.GetConversationLogUserSelectResp, error) {
	userIds, err := s.cli.GetConversationLogUserIds(ctx, req.AppId, req.AppType, req.OrgIds, req.UserIds)
	if err != nil {
		return nil, errStatus(err_code.Code_AppConversationLog, err)
	}
	return &app_service.GetConversationLogUserSelectResp{
		UserIds: userIds,
	}, nil
}

func convertConversationLogList(list []*model.ConversationLog, total int64) *app_service.GetConversationLogListResp {
	items := make([]*common.ConversationLog, 0, len(list))
	for _, item := range list {
		items = append(items, toCommonConversationLog(item))
	}
	return &app_service.GetConversationLogListResp{
		Items: items,
		Total: total,
	}
}

// toCommonConversationLog 将数据库模型转换为 proto 的 ConversationLog。
func toCommonConversationLog(item *model.ConversationLog) *common.ConversationLog {
	return &common.ConversationLog{
		LogId:             item.LogId,
		AppId:             item.AppId,
		AppType:           item.AppType,
		Source:            item.Source,
		Version:           item.Version,
		UserId:            item.UserId,
		OrgId:             item.OrgId,
		Title:             item.Title,
		ConversationId:    item.ConversationId,
		MessageCount:      int64(item.MessageCount),
		CreateAt:          item.CreatedAt,
		UpdateAt:          item.UpdatedAt,
		Costs:             item.Costs,
		FirstTokenLatency: item.FirstTokenLatency,
		LikeCount:         int64(item.LikeCount),
		DislikeCount:      int64(item.DisLikeCount),
		ErrorCount:        int64(item.ErrorCount),
		Ext:               item.Ext,
	}
}

// RecordConversationLog 记录会话日志（appId+appType+conversationId 不存在则新建，否则更新）
func (s *Service) RecordConversationLog(ctx context.Context, req *common.ConversationLog) (*emptypb.Empty, error) {
	if err := s.cli.RecordConversationLog(ctx, toModelConversationLog(req)); err != nil {
		return nil, errStatus(err_code.Code_AppConversationLog, err)
	}
	return &emptypb.Empty{}, nil
}

// DeleteConversationLogByAppId 删除指定应用下的全部会话日志（应用删除时清场）。
func (s *Service) DeleteConversationLogByAppId(ctx context.Context, req *app_service.DeleteConversationLogByAppIdReq) (*emptypb.Empty, error) {
	if err := s.cli.DeleteConversationLogByAppId(ctx, req.AppId, req.AppType); err != nil {
		return nil, errStatus(err_code.Code_AppConversationLog, err)
	}
	return &emptypb.Empty{}, nil
}

// ExportConversationLog 创建对话日志导出任务
func (s *Service) ExportConversationLog(ctx context.Context, req *app_service.ExportConversationLogReq) (*emptypb.Empty, error) {
	exportParam, err := json.Marshal(model.ConversationLogExportTaskParams{AppId: req.AppId, AppType: req.AppType, LogIds: req.LogIds, OrgIds: req.OrgIds, UserIds: req.UserIds})
	if err != nil {
		return nil, errStatus(err_code.Code_AppGeneral, toErrStatus("conversation_log_export_param", req.AppId, req.AppType, err.Error()))
	}
	exportTask := &model.ConversationLogExportTask{
		ExportId:     util.NewID(),
		AppId:        req.AppId,
		AppType:      req.AppType,
		Status:       model.ConversationLogExportInit,
		ExportParams: string(exportParam),
		UserId:       req.UserId,
		OrgId:        req.OrgId,
	}
	if status := s.cli.CreateConversationLogExportTask(ctx, exportTask); status != nil {
		return nil, errStatus(err_code.Code_AppGeneral, status)
	}
	return &emptypb.Empty{}, nil
}

// GetConversationLogExportRecordList 分页查询导出记录列表。
func (s *Service) GetConversationLogExportRecordList(ctx context.Context, req *app_service.GetConversationLogExportRecordListReq) (*app_service.GetConversationLogExportRecordListResp, error) {
	list, total, status := s.cli.GetConversationLogExportTaskList(ctx, req.AppId, req.AppType, req.StartDate, req.EndDate, req.Title, req.UserIds, req.OrgIds, req.PageSize, req.PageNum)
	if status != nil {
		return nil, errStatus(err_code.Code_AppGeneral, status)
	}
	return &app_service.GetConversationLogExportRecordListResp{
		Total:             total,
		PageNum:           req.PageNum,
		PageSize:          req.PageSize,
		ExportRecordInfos: toProtoConversationLogExportRecordInfos(list),
	}, nil
}

// DeleteConversationLogExportRecord 删除导出记录 + 删 MinIO 文件。
func (s *Service) DeleteConversationLogExportRecord(ctx context.Context, req *app_service.DeleteConversationLogExportRecordReq) (*emptypb.Empty, error) {
	if status := s.cli.DeleteConversationLogExportTaskByIds(ctx, req.ExportRecordIds, req.UserId, req.OrgId); status != nil {
		return nil, errStatus(err_code.Code_AppGeneral, status)
	}
	return &emptypb.Empty{}, nil
}

// --- internal ---

// toModelConversationLog 将 proto 的 ConversationLog 转换为数据库模型。
func toModelConversationLog(req *common.ConversationLog) *model.ConversationLog {
	return &model.ConversationLog{
		LogId:             req.LogId,
		ConversationId:    req.ConversationId,
		AppId:             req.AppId,
		AppType:           req.AppType,
		Source:            req.Source,
		Version:           req.Version,
		Title:             req.Title,
		MessageCount:      int(req.MessageCount),
		Costs:             req.Costs,
		FirstTokenLatency: req.FirstTokenLatency,
		LikeCount:         int(req.LikeCount),
		DisLikeCount:      int(req.DislikeCount),
		ErrorCount:        int(req.ErrorCount),
		UserId:            req.UserId,
		OrgId:             req.OrgId,
		Ext:               req.Ext,
	}
}

func toProtoConversationLogExportRecordInfos(list []*model.ConversationLogExportTask) []*app_service.ConversationLogExportRecordInfo {
	infos := make([]*app_service.ConversationLogExportRecordInfo, 0, len(list))
	for _, item := range list {
		infos = append(infos, &app_service.ConversationLogExportRecordInfo{
			ExportRecordId: item.ExportId,
			Status:         int32(item.Status),
			FilePath:       item.ExportFilePath,
			FileName:       filepath.Base(item.ExportFilePath),
			ErrorMsg:       item.ErrorMsg,
			ExportTime:     util.Time2Str(item.CreatedAt),
			UserId:         item.UserId,
		})
	}
	return infos
}

func toErrStatus(key string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: key,
		Args:    args,
	}
}
