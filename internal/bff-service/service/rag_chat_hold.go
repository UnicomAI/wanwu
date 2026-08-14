package service

import (
	"context"
	"fmt"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	sse_connector "github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector"
	sse_model "github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/model"
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/session"
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/store"
	"github.com/gin-gonic/gin"
)

// 知识问答链接保持：客户端断开后上游不中断，会话内已产出的 AG-UI 行留在内存会话里，
// 重连时先补发历史再续接新消息。会话键为 conversationId + X-Client-ID。
// ext 里存的问题与附件供"运行中会话"接口回显。
const (
	ragSessionExtQuestion = "question"
	ragSessionExtFileInfo = "fileInfo"
)

// newRagSSESession 初始化 sse 链接保持器；未开启（openapi）或缺 clientId 时返回 invalid 管理器，各方法均为空实现
func newRagSSESession(ctx *gin.Context, clientId string, req request.ChatRagRequest) *session.Manager {
	mgr := sse_connector.NewSSESessionValid(ctx, &sse_model.Session{
		ConversationID: req.ConversationID,
		ClientID:       clientId,
	}, store.NewMemoryStore(), req.SseHold)
	mgr.AddExt(map[string]interface{}{
		ragSessionExtQuestion: req.Question,
		ragSessionExtFileInfo: req.FileInfo,
	})
	return mgr
}

// ragUpstreamCtx 链接保持生效时用会话后台 ctx（客户端断开不打断上游），否则跟随请求 ctx
func ragUpstreamCtx(ctx *gin.Context, mgr *session.Manager) context.Context {
	if mgr != nil && !mgr.Invalid {
		return mgr.GetBgContext()
	}
	return ctx.Request.Context()
}

// publishRagChatStream 把最终 AG-UI 行先发布到会话再转发给当前连接。
// 发布在前保证断线期间的内容也能进会话；转发用 clientCtx 兜底，客户端断开后停止转发但继续消费上游。
// onFinish 在这一轮真正结束时回调（客户端可能早已断开），用于落统计。
func publishRagChatStream(clientCtx context.Context, mgr *session.Manager, in <-chan string, onFinish func()) <-chan string {
	out := make(chan string, 128)
	safe_go_util.SafeGo(func() {
		outClosed := false
		closeOut := func() {
			if outClosed {
				return
			}
			outClosed = true
			close(out)
		}
		defer func() {
			closeOut()
			// 上游结束即会话结束：重连侧收到 channel 关闭后正常收尾
			if err := mgr.Cancel(); err != nil {
				log.Errorf("[RAG-Stream] close sse session err: %v", err)
			}
			if onFinish != nil {
				onFinish()
			}
		}()
		for line := range in {
			if err := mgr.Publish(&sse_model.Message{Data: line}, nil); err != nil {
				log.Errorf("[RAG-Stream] publish sse message err: %v", err)
			}
			if outClosed { // 客户端已断开，只落会话不再转发
				continue
			}
			select {
			case out <- line:
			case <-clientCtx.Done():
				closeOut()
			}
		}
	})
	return out
}

// GetRagPendingConversation 查询知识问答运行中会话，供页面刷新后决定是否重连
func GetRagPendingConversation(ctx *gin.Context, userId, orgId, clientId string, req request.RagPendingConversationReq) (*response.PendingConversationResp, error) {
	conversationId, err := getRagConversationID(ctx, userId, orgId, req)
	if err != nil {
		return nil, err
	}
	if conversationId == "" {
		return &response.PendingConversationResp{HasPendingConversation: false}, nil
	}
	sess := sse_connector.GetSession(&sse_model.Session{ConversationID: conversationId, ClientID: clientId})
	if sess == nil {
		return &response.PendingConversationResp{
			ConversationId:         conversationId,
			HasPendingConversation: false,
		}, nil
	}
	question, files := parseRagSessionExt(sess.GetExt())
	return &response.PendingConversationResp{
		ConversationId:         conversationId,
		HasPendingConversation: true,
		Prompt:                 question,
		RequestFiles:           files,
	}, nil
}

// ChatRagStreamConnect 知识问答流式问答断开后重连：补发会话内已产出的 AG-UI 行并续接后续消息
func ChatRagStreamConnect(ctx *gin.Context, userId, orgId, clientId string, req request.RagStreamConnectReq) error {
	if err := checkRagPublishAccessIfNotDraft(ctx, userId, orgId, req.RagID, req.ConversationID); err != nil {
		return err
	}
	chatCh, err := sse_connector.Connect(ctx, &sse_model.Session{
		ConversationID: req.ConversationID,
		ClientID:       clientId,
	}, func(data *sse_model.Message) string {
		line, _ := data.Data.(string)
		return line
	})
	if err != nil {
		return err
	}
	return sse_util.NewSSEWriter(ctx, fmt.Sprintf("[RAG-Stream] %v conversation %v user %v org %v reconnect", req.RagID, req.ConversationID, userId, orgId), "").
		WriteStream(chatCh, nil, buildRagChatResp(), nil)
}

// ChatRagStreamCancel 手动停止知识问答流式问答：掐断上游并清理会话
func ChatRagStreamCancel(ctx *gin.Context, userId, orgId, clientId string, req request.RagStreamCancelReq) error {
	conversationId, err := getRagConversationID(ctx, userId, orgId, req.RagPendingConversationReq)
	if err != nil {
		return err
	}
	if conversationId == "" {
		return nil
	}
	return sse_connector.Close(&sse_model.Session{ConversationID: conversationId, ClientID: clientId})
}

// getRagConversationID 草稿态每个知识问答仅一条会话，按 ragId 反查（已按 Identity 隔离）；
// 已发布态由调用方传 conversationId，须先校验归属
func getRagConversationID(ctx *gin.Context, userId, orgId string, req request.RagPendingConversationReq) (string, error) {
	if !req.Draft {
		if err := CheckRagConversationAccess(ctx, req.ConversationID, req.RagID, userId, orgId); err != nil {
			return "", err
		}
		if err := CheckAppPublishAccess(ctx, req.RagID, constant.AppTypeRag, userId, orgId); err != nil {
			return "", err
		}
		return req.ConversationID, nil
	}
	return GetDraftRagConversationId(ctx, userId, orgId, req.RagID)
}

func checkRagPublishAccessIfNotDraft(ctx *gin.Context, userId, orgId, ragId, conversationId string) error {
	draftConversationId, err := GetDraftRagConversationId(ctx, userId, orgId, ragId)
	if err != nil {
		return err
	}
	if conversationId != "" && conversationId == draftConversationId {
		return nil
	}
	return CheckAppPublishAccess(ctx, ragId, constant.AppTypeRag, userId, orgId)
}

// parseRagSessionExt 取会话扩展里的问题与附件，缺失或类型不符时降级为空值
func parseRagSessionExt(ext map[string]interface{}) (string, []response.AssistantRequestFile) {
	if len(ext) == 0 {
		return "", nil
	}
	question, _ := ext[ragSessionExtQuestion].(string)
	files, _ := ext[ragSessionExtFileInfo].([]request.ConversionStreamFile)
	var requestFiles []response.AssistantRequestFile
	for _, file := range files {
		requestFiles = append(requestFiles, response.AssistantRequestFile{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  file.FileUrl,
		})
	}
	return question, requestFiles
}
