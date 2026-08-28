package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"sort"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// FilterOwnerUser 根据owner、普通使用者 过滤userId,orgId
func FilterOwnerUser(ctx *gin.Context, userId, orgId, appId, appType string, filterUserIds []string) (bool, []string, []string, error) {
	var userIds, orgIds []string
	ownerUserId, ownerOrgId, err := OwnerInfo(ctx, appType, appId)
	if err != nil {
		return false, nil, nil, err
	}
	// owner
	if ownerUserId == userId && ownerOrgId == orgId {
		if len(filterUserIds) > 0 {
			userIds = append(userIds, filterUserIds...)
		}
		return true, userIds, orgIds, nil
	}

	// 普通使用者
	//userIds = append(userIds, userId)
	//orgIds = append(orgIds, orgId)
	return false, userIds, orgIds, nil
}

// GetConversationLogList 获取会话日志列表
func GetConversationLogList(ctx *gin.Context, userId, orgId string, req request.GetConversationLogListRequest) (response.PageResult, error) {
	_, userIds, orgIds, err := FilterOwnerUser(ctx, userId, orgId, req.AppId, req.AppType, req.UserIds)
	if err != nil {
		return response.PageResult{}, err
	}
	resp, err := app.GetConversationLogList(ctx.Request.Context(), &app_service.GetConversationLogListReq{
		AppId:     req.AppId,
		AppType:   req.AppType,
		Name:      req.Name,
		Source:    req.Source,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		OrgIds:    orgIds,
		UserIds:   userIds,
		PageNo:    int32(req.PageNo),
		PageSize:  int32(req.PageSize),
		OrderBy:   req.OrderBy,
		OrderType: req.OrderType,
	})
	if err != nil {
		return response.PageResult{}, err
	}
	userNameMap := buildConversationLogUserNameMap(ctx, resp.Items)
	list := make([]response.ConversationLogInfo, 0, len(resp.Items))
	for _, item := range resp.Items {
		list = append(list, response.ConversationLogInfo{
			LogId:                item.LogId,
			Source:               item.Source,
			UserId:               item.UserId,
			UserName:             userNameMap[item.UserId],
			Title:                item.Title,
			ConversationId:       item.ConversationId,
			MessageCount:         item.MessageCount,
			CreateAt:             util.Time2Str(item.CreateAt),
			UpdateAt:             util.Time2Str(item.UpdateAt),
			AvgCosts:             float64(item.Costs),
			AvgFirstTokenLatency: float64(item.FirstTokenLatency),
			LikeCount:            item.LikeCount,
			DislikeCount:         item.DislikeCount,
			ErrorCount:           item.ErrorCount,
			Version:              item.Version,
		})
	}
	return response.PageResult{
		List:     list,
		Total:    resp.Total,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// buildConversationLogUserNameMap 批量查询会话日志列表中 userId 对应的用户名称。
func buildConversationLogUserNameMap(ctx *gin.Context, items []*common.ConversationLog) map[string]string {
	userNameMap := make(map[string]string)
	userIdMap := make(map[string]bool)
	for _, item := range items {
		if item.UserId != "" {
			userIdMap[item.UserId] = true
		}
	}
	userInfoMap, _ := searchUserAndOrgInfo(ctx, userIdMap, nil)
	for userId, info := range userInfoMap {
		if info != nil {
			userNameMap[userId] = info.Name
		}
	}
	// webURL 来源查不到用户名时，回退用 userId 代替。
	for _, item := range items {
		if item.Source == "webURL" && item.UserId != "" {
			if userNameMap[item.UserId] == "" {
				userNameMap[item.UserId] = item.UserId
			}
		}
	}
	return userNameMap
}

// GetAssistantConversationLogDetail 获取会话日志详情
func GetAssistantConversationLogDetail(ctx *gin.Context, userId, orgId string, req request.GetConversationLogDetailRequest) (response.PageResult, error) {
	isOwner, userIds, orgIds, err := FilterOwnerUser(ctx, userId, orgId, req.AppId, req.AppType, nil)
	if err != nil {
		return response.PageResult{}, err
	}
	conversationLog, err := app.GetConversationLog(ctx, &app_service.GetConversationLogReq{AppId: req.AppId, AppType: req.AppType, ConversationId: req.ConversationId})
	if err != nil {
		return response.PageResult{}, err
	}
	if !isOwner && !slices.Contains(userIds, conversationLog.UserId) && !slices.Contains(orgIds, conversationLog.OrgId) {
		return response.PageResult{}, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Errorf("user org permit valid fail").Error())
	}
	req.ConversationId = resolveLegacyConversationId(conversationLog)
	return fetchConversationDetailList(ctx, conversationLog.UserId, conversationLog.OrgId, req)
}

// resolveLegacyConversationId 根据 conversationLog.Ext 中的 conversation_mark 判断是否为旧会话，
// 返回用于查询 ES 的 conversationId：旧会话返回旧自增主键（conversation_id_mark），否则返回原 conversationId。
func resolveLegacyConversationId(conversationLog *common.ConversationLog) string {
	if conversationLog.Ext != "" {
		var ext struct {
			ConversationIDMark uint32 `json:"conversation_id_mark"`
			ConversationMark   uint32 `json:"conversation_mark"`
		}
		if json.Unmarshal([]byte(conversationLog.Ext), &ext) == nil && ext.ConversationMark == 1 && ext.ConversationIDMark > 0 {
			return util.Int2Str(ext.ConversationIDMark)
		}
	}
	return conversationLog.ConversationId
}

// fetchConversationDetailList 按 conversationId 查询 ES 对话详情并转换为响应结构（不做权限校验，由调用方保证）。
func fetchConversationDetailList(ctx *gin.Context, userId, orgId string, req request.GetConversationLogDetailRequest) (response.PageResult, error) {
	resp, err := assistant.InternalGetConversationDetailList(ctx.Request.Context(), &assistant_service.GetConversationDetailListReq{
		ConversationId: req.ConversationId,
		PageSize:       int32(req.PageSize),
		PageNo:         int32(req.PageNo),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		ExcludeDeleted: false,
	})
	if err != nil {
		return response.PageResult{}, err
	}

	// 转换resp.Data为自定义的ConversionDetailInfo结构体数组
	var convertedList []response.ConversationDetailInfo
	for _, item := range resp.Data {
		convertedItem := response.ConversationDetailInfo{
			Id:                  item.Id,
			AssistantId:         item.AssistantId,
			ConversationId:      item.ConversationId,
			Prompt:              item.Prompt,
			SysPrompt:           item.SysPrompt,
			Response:            item.Response,
			ResponseList:        buildResponseList(item.ConversationResponse),
			QaType:              item.QaType,
			CreatedBy:           item.CreatedBy,
			CreatedAt:           item.CreatedAt,
			UpdatedAt:           item.UpdatedAt,
			RequestFiles:        transRequestFiles(item.RequestFiles),
			FileSize:            item.FileSize,
			FileName:            item.FileName,
			SubConversationList: buildSubConversationList(item.SubConversationList),
			ResponseFiles:       transResponseFiles(item.ResponseFiles),
			Feedback:            item.Feedback,
			FeedbackContent:     item.FeedbackContent,
		}

		// 将SearchList从string转换为interface{}
		convertedItem.SearchList = buildSearchList(item.SearchList)

		convertedList = append(convertedList, convertedItem)

		// 对切片进行排序
		sort.Slice(convertedList, func(i, j int) bool {
			// CreatedAt值小的时间更早，排在前面
			return convertedList[i].CreatedAt < convertedList[j].CreatedAt
		})
	}

	return response.PageResult{Total: resp.Total, List: convertedList, PageNo: req.PageNo, PageSize: req.PageSize}, nil
}

// GetConversationLogUserSelect 获取会话日志使用者列表
func GetConversationLogUserSelect(ctx *gin.Context, userId, orgId string, req request.GetConversationLogUserSelectRequest) (*response.Users, error) {
	isOwner, userIds, _, err := FilterOwnerUser(ctx, userId, orgId, req.AppId, req.AppType, nil)
	if err != nil {
		return nil, err
	}
	// 仅应用 owner 可见全部使用者；非 owner 仅返回自己
	if isOwner {
		resp, err := app.GetConversationLogUserSelect(ctx.Request.Context(), &app_service.GetConversationLogUserSelectReq{
			AppId:   req.AppId,
			AppType: req.AppType,
		})
		if err != nil {
			return nil, err
		}
		userIds = resp.UserIds
	}

	if len(userIds) == 0 {
		return &response.Users{Users: []response.IDNameWithAvatar{}}, nil
	}
	resp, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
		UserIds: userIds,
	})
	if err != nil {
		return nil, err
	}
	var users []response.IDNameWithAvatar
	for _, u := range resp.Selects {
		users = append(users, response.IDNameWithAvatar{
			ID:     u.Id,
			Name:   u.Name,
			Avatar: cacheUserAvatar(u.AvatarPath),
		})
	}
	return &response.Users{
		Users: users,
	}, nil
}

func RecordConversationLog(ctx context.Context, log *common.ConversationLog) error {
	if _, err := app.RecordConversationLog(ctx, log); err != nil {
		return err
	}
	return nil
}

// ExportConversationLog 创建对话日志导出任务（异步）。
func ExportConversationLog(ctx *gin.Context, userId, orgId string, req *request.ConversationLogExportReq) error {
	// 权限范围：应用 owner 可全量导出（orgIds/userIds 为空=不限）；非 owner 仅导自己。
	_, userIds, orgIds, err := FilterOwnerUser(ctx, userId, orgId, req.AppId, req.AppType, nil)
	if err != nil {
		return err
	}
	_, err = app.ExportConversationLog(ctx.Request.Context(), &app_service.ExportConversationLogReq{
		AppId:   req.AppId,
		AppType: req.AppType,
		LogIds:  req.LogIds,
		UserId:  userId,
		OrgId:   orgId,
		OrgIds:  orgIds,
		UserIds: userIds,
	})
	if err != nil {
		log.Errorf("对话日志导出失败(创建导出任务 失败(%v) ", err)
		return err
	}
	return nil
}

// GetConversationLogExportRecordList 查询对话日志导出记录列表。
func GetConversationLogExportRecordList(ctx *gin.Context, userId, orgId string, req *request.ConversationLogExportRecordListReq) (*response.PageResult, error) {
	_, userIds, orgIds, err := FilterOwnerUser(ctx, userId, orgId, req.AppId, req.AppType, req.UserIds)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetConversationLogExportRecordList(ctx.Request.Context(), &app_service.GetConversationLogExportRecordListReq{
		AppId:     req.AppId,
		AppType:   req.AppType,
		Title:     req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		UserIds:   userIds,
		OrgIds:    orgIds,
		PageSize:  int32(req.PageSize),
		PageNum:   int32(req.PageNo),
	})
	if err != nil {
		return nil, err
	}
	return &response.PageResult{
		List:     buildConversationLogExportRecordRespList(ctx, resp.ExportRecordInfos),
		Total:    resp.Total,
		PageNo:   int(resp.PageNum),
		PageSize: int(resp.PageSize),
	}, nil
}

// DeleteConversationLogExportRecord 删除对话日志导出记录。
func DeleteConversationLogExportRecord(ctx *gin.Context, userId, orgId string, req *request.DeleteConversationLogExportRecordReq) error {
	_, err := app.DeleteConversationLogExportRecord(ctx.Request.Context(), &app_service.DeleteConversationLogExportRecordReq{
		ExportRecordIds: req.ExportRecordIds,
		UserId:          userId,
		OrgId:           orgId,
	})
	if err != nil {
		log.Errorf("删除对话日志导出记录 失败(%v) ", err)
		return err
	}
	return nil
}

// buildConversationLogExportRecordRespList 构造对话日志导出记录返回列表。
func buildConversationLogExportRecordRespList(ctx *gin.Context, dataList []*app_service.ConversationLogExportRecordInfo) []*response.ConversationLogExportRecordResp {
	retList := make([]*response.ConversationLogExportRecordResp, 0, len(dataList))
	authorMap := buildConversationLogExportAuthorMap(ctx, dataList)
	for _, data := range dataList {
		filePath, _ := url.JoinPath(config.Cfg().Minio.DownloadURL, data.FilePath)
		filePath = urlDecode(filePath)
		retList = append(retList, &response.ConversationLogExportRecordResp{
			ExportRecordId: data.ExportRecordId,
			ExportTime:     data.ExportTime,
			FilePath:       filePath,
			FileName:       data.FileName,
			Author:         authorMap[data.UserId],
			Status:         int(data.Status),
			ErrorMsg:       gin_util.I18nKey(ctx, data.ErrorMsg),
		})
	}
	return retList
}

func buildConversationLogExportAuthorMap(ctx *gin.Context, dataList []*app_service.ConversationLogExportRecordInfo) map[string]string {
	authorMap := make(map[string]string)
	userIdMap := make(map[string]bool)
	for _, data := range dataList {
		if data.UserId != "" {
			userIdMap[data.UserId] = true
			authorMap[data.UserId] = ""
		}
	}
	if len(userIdMap) == 0 {
		return authorMap
	}
	userInfoMap, _ := searchUserAndOrgInfo(ctx, userIdMap, nil)
	for userId, info := range userInfoMap {
		if info != nil {
			authorMap[userId] = info.Name
		}
	}
	return authorMap
}
