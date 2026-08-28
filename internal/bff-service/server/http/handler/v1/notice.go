package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetNoticeUnreadCount
//
//	@Tags			notice
//	@Summary		未读总数+分类角标
//	@Description	未读总数+分类角标，驱动头像红点与各 Tab 角标；消息按账号+组织双维度隔离，切换组织需重新拉取
//	@Security		JWT
//	@Produce		json
//	@Param			X-Org-Id	header		string	true	"当前组织上下文ID"
//	@Success		200			{object}	response.Response{data=response.NoticeUnreadCountResp}
//	@Router			/notice/unread/count [get]
func GetNoticeUnreadCount(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	resp, err := service.GetNoticeUnreadCount(ctx, userId, orgId)
	gin_util.Response(ctx, resp, err)
}

// ListNotice
//
//	@Tags			notice
//	@Summary		消息中心整页列表
//	@Description	已读与未读混排，每条带 isRead；悬浮面板的未读列表也复用本接口（onlyUnread=true）
//	@Security		JWT
//	@Produce		json
//	@Param			X-Org-Id	header		string					true	"当前组织上下文ID"
//	@Param			data		query		request.NoticeListReq	true	"消息列表请求参数"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=response.NoticeItem}}
//	@Router			/notice/list [get]
func ListNotice(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.NoticeListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ListNotice(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// ReadNotice
//
//	@Tags			notice
//	@Summary		消息已读（单条）
//	@Description	把单条消息标记为已读；对已读消息重复调用是幂等的
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Org-Id	header		string					true	"当前组织上下文ID"
//	@Param			data		body		request.NoticeReadReq	true	"单条已读请求参数"
//	@Success		200			{object}	response.Response
//	@Router			/notice/read [put]
func ReadNotice(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.NoticeReadReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ReadNotice(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// ReadAllNotice
//
//	@Tags			notice
//	@Summary		一键已读
//	@Description	把当前账号在当前组织上下文的全部可见未读标记为已读；不分类别，列表筛选与勾选均不影响作用范围
//	@Security		JWT
//	@Produce		json
//	@Param			X-Org-Id	header		string	true	"当前组织上下文ID"
//	@Success		200			{object}	response.Response
//	@Router			/notice/read-all [put]
func ReadAllNotice(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	err := service.ReadAllNotice(ctx, userId, orgId)
	gin_util.Response(ctx, nil, err)
}

// DeleteNotice
//
//	@Tags			notice
//	@Summary		批量删除消息
//	@Description	按勾选的 messageId 从当前账号在当前组织上下文的列表中移除，不影响其他收件人
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Org-Id	header		string					true	"当前组织上下文ID"
//	@Param			data		body		request.NoticeDeleteReq	true	"批量删除请求参数"
//	@Success		200			{object}	response.Response{data=response.NoticeDeleteResp}
//	@Router			/notice [delete]
func DeleteNotice(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.NoticeDeleteReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.DeleteNotice(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}
