package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func CreateChatflow(ctx *gin.Context, orgID, name, desc, iconUri string) (*response.CozeWorkflowIDData, error) {
	url, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.CreateUri)
	ret := &response.CozeWorkflowIDResp{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"space_id":  orgID,
			"name":      name,
			"desc":      desc,
			"icon_uri":  iconUri,
			"flow_mode": "3",
		}).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_create", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_create", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_create", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func CreateChatflowConversation(ctx *gin.Context, userId, orgId, workflowId, conversationName string) (*response.OpenAPIChatflowCreateConversationResponse, error) {
	url, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.CreateChatflowConversationUri)
	ret := &response.CozeCreateConversationResponse{}

	// 1. 先查询是否已有 appId
	appInfo, _ := app.GetChatflowApplication(ctx.Request.Context(), &app_service.GetChatflowApplicationReq{
		OrgId:      orgId,
		UserId:     userId,
		WorkflowId: workflowId,
	})

	// 2. 构建请求
	body := map[string]any{
		"conversation_name": conversationName,
		"connector_id":      "1024",
		"draft_mode":        false,
		"get_or_create":     true,
		"workflow_id":       workflowId,
	}

	// 3. 如果已有 appId，传给接口复用
	if appInfo.GetApplicationId() != "" {
		body["app_id"] = appInfo.ApplicationId
	}

	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"space_id": orgId,
		}).
		SetBody(body).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_conversation_create", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_conversation_create", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_conversation_create", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}

	appId := ret.ConversationData.MetaData["appId"]

	_, err := app.CreateConversation(ctx.Request.Context(), &app_service.CreateConversationReq{
		AppId:            appId,
		AppType:          constant.AppTypeChatflow,
		ConversationId:   strconv.Itoa(int(ret.ConversationData.Id)),
		ConversationName: conversationName,
		UserId:           userId,
		OrgId:            orgId,
	})
	if err != nil {
		return nil, err
	}

	// 4. 首次创建时保存 workflowId 和 appId 的关联关系
	if appInfo.GetApplicationId() == "" {
		_, err = app.CreateChatflowApplication(ctx.Request.Context(), &app_service.CreateChatflowApplicationReq{
			WorkflowId:    workflowId,
			ApplicationId: appId,
			UserId:        userId,
			OrgId:         orgId,
		})
		if err != nil {
			return nil, err
		}
	}

	return &response.OpenAPIChatflowCreateConversationResponse{
		ConversationId: strconv.Itoa(int(ret.ConversationData.Id)),
	}, nil
}

func GetConversationMessageList(ctx *gin.Context, userId, orgId, appId, conversationId, limit string) (*response.OpenAPIChatflowGetConversationMessageListResponse, error) {
	url, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.GetConversationMessageListUri)
	ret := &response.CozeListMessageApiResponse{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParam("conversation_id", conversationId).
		SetBody(map[string]int64{
			"limit": util.MustI64(limit),
		}).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_conversation_message_list", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_conversation_message_list", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_conversation_message_list", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return &response.OpenAPIChatflowGetConversationMessageListResponse{
		Messages: ret.Messages,
		HasMore:  *ret.HasMore,
		FirstID:  *ret.FirstID,
		LastID:   *ret.LastID,
	}, nil
}

func ChatflowChat(ctx *gin.Context, userId, orgId, workflowId, conversationId, message string, parameters map[string]any) (err error) {
	startTime := time.Now()
	var firstTokenLatency int64
	var firstTokenRecorded bool
	var hasErr bool
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	defer func() {
		go func() {
			defer util.PrintPanicStack()
			RecordAppStatistic(detachedCtx, userId, orgId, workflowId, constant.AppTypeChatflow, !hasErr, true, firstTokenLatency, 0, constant.AppStatisticSourceOpenAPI)
		}()
	}()

	url, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ChatflowRunByOpenapiUri)
	p, err := json.Marshal(parameters)
	if err != nil {
		hasErr = true
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_chat", err.Error())
	}
	cvInfo, err := app.GetConversationByID(ctx.Request.Context(), &app_service.GetConversationByIDReq{
		ConversionId: conversationId,
	})
	if err != nil {
		hasErr = true
		return err
	}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetDoNotParseResponse(true).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "text/event-stream").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("Connection", "keep-alive").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"space_id": orgId,
		}).
		SetBody(map[string]any{
			"additional_messages": []map[string]any{
				{
					"role":         "user",
					"content_type": "text",
					"content":      message,
				},
			},
			"parameters":      string(p),
			"connector_id":    "1024",
			"workflow_id":     workflowId,
			"app_id":          cvInfo.AppId,
			"conversation_id": conversationId,
			"ext": map[string]any{
				"_caller": "CANVAS",
				"user_id": "",
			},
			"suggest_reply_info": map[string]any{},
		}).
		Post(url)

	if err != nil {
		hasErr = true
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_chat", err.Error())
	}
	if resp.StatusCode() >= 300 {
		hasErr = true
		b, err := io.ReadAll(resp.RawResponse.Body)
		if err != nil {
			return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_chat", fmt.Sprintf("[%v] %v", resp.StatusCode(), err))
		}
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_chat", fmt.Sprintf("[%v] %v", resp.StatusCode(), string(b)))
	}
	defer func() { _ = resp.RawBody().Close() }()

	ctx.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")

	scan := util.NewScanner(resp.RawBody())

	// 设置适当的缓冲区大小以避免扫描错误
	const (
		initialBufferSize = 64 * 1024
		maxBufferSize     = 10 * 1024 * 1024
	)
	scan.Buffer(make([]byte, initialBufferSize), maxBufferSize)

	for scan.Scan() {
		// 记录首 token 时延
		if !firstTokenRecorded {
			firstTokenLatency = time.Since(startTime).Milliseconds()
			firstTokenRecorded = true
			ctx.Set(gin_util.FIRST_RESP_LATENCY, firstTokenLatency)
		}
		// 写入数据到响应体（添加双换行符符合SSE格式）
		if _, err := ctx.Writer.Write([]byte(scan.Text() + "\n")); err != nil {
			log.Errorf("chatflow id [%v]chat conversationId [%v]: failed to write to client: %v", workflowId, conversationId, err)
			break
		}
		// 刷新缓冲区，确保数据立即发送到客户端
		ctx.Writer.Flush()
	}
	// 检查扫描错误（排除正常的EOF）
	if err := scan.Err(); err != nil && !errors.Is(err, io.EOF) {
		// 如果是客户端断开连接，记录info级别日志
		if errors.Is(err, context.Canceled) {
			log.Debugf("chatflow id [%v]chat conversationId [%v]: client disconnected: %v", workflowId, conversationId, err)
		} else {
			hasErr = true
			log.Errorf("chatflow id [%v]chat conversationId [%v]: failed to scan response body: %v", workflowId, conversationId, err)
		}
		return nil
	}
	return nil
}

func ChatflowApplicationList(ctx *gin.Context, userId, orgId, workflowId string) (*response.CozeDraftIntelligenceListData, error) {
	// 1.先获取workflow-wanwu的draft intelligence list信息
	getDraftUrl, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.GetDraftIntelligenceListUri)
	// 构造referer url
	baseURL := config.Cfg().Server.WebBaseUrl // e.g. "http://172.25.214.210:8081"

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_list", err.Error())
	}
	u.Path = "/workflow" // 设置路径
	q := u.Query()
	q.Set("workflow_id", workflowId)
	u.RawQuery = q.Encode()
	getDraftRet := &response.CozeGetDraftIntelligenceListResponse{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Referer", u.String()).
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]any{
			"space_id": orgId,
		}).
		SetResult(getDraftRet).
		Post(getDraftUrl); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_list", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_list", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), getDraftRet.Code, getDraftRet.Msg))
	} else if getDraftRet.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_list", fmt.Sprintf("code %v msg %v", getDraftRet.Code, getDraftRet.Msg))
	}
	// 2.查询app-service, 看是否有对应的chatflow_application记录
	appRet, err := app.GetChatflowApplication(ctx.Request.Context(), &app_service.GetChatflowApplicationReq{
		OrgId:      orgId,
		UserId:     userId,
		WorkflowId: workflowId,
	})
	if err != nil {
		return nil, err
	}
	if appRet.ApplicationId != "" {
		getDraftRet.Data.Intelligences[0].BasicInfo.ID, _ = strconv.ParseInt(appRet.ApplicationId, 10, 64)
		// 构造 DraftIntelligenceListData
		return &response.CozeDraftIntelligenceListData{
			Intelligences: getDraftRet.Data.Intelligences,
			Total:         1,
			HasMore:       false,
			NextCursorID:  "",
		}, nil
	}
	// 3.如果没有记录，则通过workflow接口创建一条，并且替换掉返回值中的ID
	url, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.GetProjectConversationUri)
	ret := &response.CozeCreateProjectConversationDefResponse{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]any{
			"space_id":          orgId,
			"project_id":        workflowId,
			"conversation_name": getDraftRet.Data.Intelligences[0].BasicInfo.Name,
		}).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_list", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_list", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_list", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	_, err = app.CreateChatflowApplication(ctx.Request.Context(), &app_service.CreateChatflowApplicationReq{
		WorkflowId:    workflowId,
		UserId:        userId,
		OrgId:         orgId,
		ApplicationId: ret.UniqueID,
	})
	if err != nil {
		return nil, err
	}
	getDraftRet.Data.Intelligences[0].BasicInfo.ID, _ = strconv.ParseInt(ret.UniqueID, 10, 64)
	// 构造 DraftIntelligenceListData
	return &response.CozeDraftIntelligenceListData{
		Intelligences: getDraftRet.Data.Intelligences,
		Total:         1,
		HasMore:       false,
		NextCursorID:  "",
	}, nil
}

func ChatflowApplicationInfo(ctx *gin.Context, userId, orgId string, req request.ChatflowApplicationInfoReq) (*response.CozeGetDraftIntelligenceInfoData, error) {
	// 先去app通过applicationId和userId查出workflowId
	resp, err := app.GetChatflowByApplicationID(ctx.Request.Context(), &app_service.GetChatflowByApplicationIDReq{
		OrgId:         orgId,
		UserId:        userId,
		ApplicationId: req.IntelligenceID,
	})
	if err != nil {
		return nil, err
	}
	// 再去workflow接口获取draft intelligence info
	url, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.GetDraftIntelligenceInfoUri)
	// 构造请求
	getDraftInfoResp := &response.CozeGetDraftIntelligenceInfoResponse{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"intelligence_id":   resp.WorkflowId,
			"intelligence_type": strconv.Itoa(int(req.IntelligenceType)),
		}).
		SetResult(getDraftInfoResp).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_info", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_info", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), getDraftInfoResp.Code, getDraftInfoResp.Msg))
	} else if getDraftInfoResp.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_chatflow_application_info", fmt.Sprintf("code %v msg %v", getDraftInfoResp.Code, getDraftInfoResp.Msg))
	}
	// 把workflow返回的intelligence id替换成app-service的application id
	getDraftInfoResp.Data.BasicInfo.ID, _ = strconv.ParseInt(req.IntelligenceID, 10, 64)
	return getDraftInfoResp.Data, nil
}

func DeleteChatflowConversation(ctx *gin.Context, orgId, projectId, uniqueId string) error {
	url, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.DeleteConversationUri)
	ret := &response.CozeDeleteProjectConversationDefResponse{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]string{
			"space_id":   orgId,
			"project_id": projectId,
			"unique_id":  uniqueId,
		}).
		SetResult(ret).
		Post(url); err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_delete_chatflow_conversation", err.Error())
	} else if resp.StatusCode() >= 300 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_delete_chatflow_conversation", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_delete_chatflow_conversation", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	if ret.Success {
		return nil
	}
	return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_delete_chatflow_conversation", "delete chatflow conversation failed")
}

func DeleteChatflowConversationByConversationId(ctx *gin.Context, userId, orgId, workflowId, conversationId string) error {
	// 1. 获取 applicationId (projectId)
	appInfo, err := app.GetChatflowApplication(ctx.Request.Context(), &app_service.GetChatflowApplicationReq{
		OrgId:      orgId,
		UserId:     userId,
		WorkflowId: workflowId,
	})
	if err != nil {
		return err
	}
	if appInfo.ApplicationId == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_delete_chatflow_conversation_by_conv_id", "application not found")
	}

	// 2. 获取会话列表，通过 conversationId 找到 uniqueId
	conversations, err := getChatflowProjectConversationList(ctx, orgId, appInfo.ApplicationId)
	if err != nil {
		return err
	}

	var uniqueId string
	for _, conv := range conversations {
		if conv.ConversationID == conversationId {
			uniqueId = conv.UniqueID
			break
		}
	}
	if uniqueId == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_delete_chatflow_conversation_by_conv_id", "conversation not found")
	}

	// 3. 调用删除接口
	return DeleteChatflowConversation(ctx, orgId, appInfo.ApplicationId, uniqueId)
}

func GetChatflowConversationList(ctx *gin.Context, userId, orgId, workflowId string) (*response.OpenAPIChatflowConversationListResponse, error) {
	// 1. 获取 applicationId (projectId)
	appInfo, err := app.GetChatflowApplication(ctx.Request.Context(), &app_service.GetChatflowApplicationReq{
		OrgId:      orgId,
		UserId:     userId,
		WorkflowId: workflowId,
	})
	if err != nil {
		return nil, err
	}
	if appInfo.ApplicationId == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_get_chatflow_conversation_list", "application not found")
	}

	// 2. 获取会话列表
	conversations, err := getChatflowProjectConversationList(ctx, orgId, appInfo.ApplicationId)
	if err != nil {
		return nil, err
	}

	// 3. 转换为 OpenAPI 响应格式
	result := &response.OpenAPIChatflowConversationListResponse{
		Conversations: make([]*response.OpenAPIChatflowConversationItem, 0, len(conversations)),
	}
	for _, conv := range conversations {
		result.Conversations = append(result.Conversations, &response.OpenAPIChatflowConversationItem{
			ConversationID:   conv.ConversationID,
			ConversationName: conv.ConversationName,
		})
	}
	return result, nil
}

// --- internal ---

func getChatflowProjectConversationList(ctx *gin.Context, orgId, projectId string) ([]*response.CozeProjectConversationItem, error) {
	url, _ := url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.GetProjectConversationListUri)
	ret := &response.CozeListProjectConversationResponse{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"project_id":    projectId,
			"create_method": "2",
			"create_env":    "2",
			"limit":         "1000",
			"connector_id":  "1024",
			"space_id":      orgId,
		}).
		SetResult(ret).
		Get(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_get_project_conversation_list", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_get_project_conversation_list", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_get_project_conversation_list", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func cozeChatflowInfo2Model(chatflowInfo *response.CozeWorkflowListDataWorkflow) response.AppBriefInfo {
	return response.AppBriefInfo{
		AppId:     chatflowInfo.WorkflowId,
		AppType:   constant.AppTypeChatflow,
		Name:      chatflowInfo.Name,
		Desc:      chatflowInfo.Desc,
		Avatar:    cacheWorkflowAvatar(chatflowInfo.URL, constant.AppTypeChatflow),
		CreatedAt: util.Time2Str(chatflowInfo.CreateTime * 1000),
		UpdatedAt: util.Time2Str(chatflowInfo.UpdateTime * 1000),
	}
}
