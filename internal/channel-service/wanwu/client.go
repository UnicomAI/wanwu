package wanwu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	clientconv "github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// Client 万悟平台代理客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建万悟代理客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- 数据结构 ---

// wanwuResponse 万悟 BFF 统一响应格式
type wanwuResponse struct {
	Code int64           `json:"code"`
	Data json.RawMessage `json:"data"`
	Msg  string          `json:"msg"`
}

// ChatRequest 智能体对话请求
type ChatRequest struct {
	UUID           string `json:"uuid"`
	ConversationID string `json:"conversation_id"`
	Query          string `json:"query"`
	Stream         bool   `json:"stream"`
}

// CreateConversationRequest 创建会话请求
type CreateConversationRequest struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
}

// CreateConversationResponse 创建会话响应
// 注意：JSON tag 必须与 bff 返回一致（bff OpenAPIAgentCreateConversationResponse 用 conversation_id），
// 否则反序列化拿到空串，会导致 channel_conversations 表的 conversation_id 落空。
type CreateConversationResponse struct {
	ConversationID string `json:"conversation_id"`
}

// ConversationManager 会话管理器
// 维护 (channelID + platformUserID + appType) -> conversationId 的映射，持久化到 DB
// appType=agent 时存 conversationId；appType=wga 时存 threadId。重启进程后从 DB 恢复。
type ConversationManager struct {
	cli client.IClient
}

// NewConversationManager 创建会话管理器
func NewConversationManager(cli client.IClient) *ConversationManager {
	return &ConversationManager{cli: cli}
}

// GetConversationID 获取已有的会话 ID（从 DB 读取），未找到返回 ok=false。
func (m *ConversationManager) GetConversationID(ctx context.Context, channelID, userID, appType string) (string, bool) {
	conv, err := m.cli.GetConversation(ctx, channelID, userID, appType)
	if err != nil || conv == nil {
		return "", false
	}
	return conv.ConversationID, true
}

// SetConversationID 设置会话 ID 映射（落 DB，按 channel+user+appType 复用同一行）
func (m *ConversationManager) SetConversationID(ctx context.Context, channelID, userID, appType, conversationID string) {
	if err := m.cli.UpsertConversation(ctx, &clientconv.ChannelConversation{
		ChannelID:      channelID,
		UserID:         userID,
		AppType:        appType,
		ConversationID: conversationID,
	}); err != nil {
		log.Errorf("failed to persist conversation mapping channel=%s user=%s appType=%s: %v", channelID, userID, appType, err)
	}
}

// --- API 方法 ---

// CreateConversation 调用智能体创建对话接口
func (c *Client) CreateConversation(ctx context.Context, apiKey string, req *CreateConversationRequest) (*CreateConversationResponse, error) {
	reqURL := fmt.Sprintf("%s/openapi/v1/agent/conversation", c.baseURL)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create conversation request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call create conversation api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create conversation api returned status %d: %s", resp.StatusCode, string(body))
	}

	var wanwuResp wanwuResponse
	if err := json.Unmarshal(body, &wanwuResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if wanwuResp.Code != 0 {
		return nil, fmt.Errorf("create conversation api error: code=%d, msg=%s", wanwuResp.Code, wanwuResp.Msg)
	}

	var convResp CreateConversationResponse
	if err := json.Unmarshal(wanwuResp.Data, &convResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation data: %w", err)
	}

	return &convResp, nil
}

// ChatWithAgent 调用智能体对话接口（SSE 流式）
func (c *Client) ChatWithAgent(ctx context.Context, apiKey string, chatReq *ChatRequest) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s/openapi/v1/agent/chat", c.baseURL)

	bodyBytes, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	// SSE 流式不设置超时
	req.Header.Set("Accept", "text/event-stream")

	// 使用独立的 HTTP 客户端，不设置超时（用于 SSE 流式响应）
	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call chat api: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("chat api returned status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// --- WGA 通用智能体数据结构 ---

// WGACreateConversationRequest WGA 创建对话请求
type WGACreateConversationRequest struct {
	Title     string `json:"title"`
	ModelUuid string `json:"modelUuid"`
}

// WGACreateConversationResponse WGA 创建对话响应
type WGACreateConversationResponse struct {
	ThreadID string `json:"threadId"`
}

// WGAChatRequest WGA 对话请求
type WGAChatRequest struct {
	ThreadID  string       `json:"threadId"`
	Messages  []WGAMessage `json:"messages"`
	ModelUuid string       `json:"modelUuid,omitempty"`
	AgentId   string       `json:"agentId,omitempty"` // 直连子智能体（留空走 Supervisor 默认路由）
}

// WGAMessage WGA 消息
type WGAMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// WGAMessageContentPart WGA 多模态消息内容片段
// type=text 时填 Text；type=binary 时填 URL/FileName/MimeType（URL 为上传后的 minio 文件路径）
type WGAMessageContentPart struct {
	Type     string `json:"type"`               // text / binary
	Text     string `json:"text,omitempty"`     // type=text 时
	MimeType string `json:"mimeType,omitempty"` // type=binary 时
	URL      string `json:"url,omitempty"`      // type=binary 时，上传后的文件 URL
	FileName string `json:"fileName,omitempty"` // type=binary 时，文件名
}

// --- WGA question（人机交互）数据结构 ---

// WGAQuestionOption WGA question 选项
type WGAQuestionOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// WGAQuestion WGA question 中的单个问题
type WGAQuestion struct {
	Header   string              `json:"header"`
	Question string              `json:"question"`
	Multiple bool                `json:"multiple"` // 是否多选
	Custom   bool                `json:"custom"`   // 是否支持自定义输入
	Options  []WGAQuestionOption `json:"options"`
}

// WGAQuestionContent ACTIVITY_SNAPSHOT(activityType=question) 事件的 content 字段
type WGAQuestionContent struct {
	QuestionID string        `json:"questionId"`
	RunID      string        `json:"runId"`
	ThreadID   string        `json:"threadId"`
	Status     string        `json:"status"` // pending / answered / rejected
	Questions  []WGAQuestion `json:"questions"`
}

// WGAQuestionReplyRequest 回答 WGA question 请求
// Answers 为二维数组：Answers[i] 对应 Questions[i]，每个元素是选中选项的 label（多选多个，自定义文本也并入）
type WGAQuestionReplyRequest struct {
	RunID      string     `json:"runId"`
	QuestionID string     `json:"questionId"`
	Answers    [][]string `json:"answers"`
}

// WGAFileNode WGA 工作区文件/目录树节点（递归）
type WGAFileNode struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"` // file / directory
	Size     int64          `json:"size,omitempty"`
	MimeType string         `json:"mimeType,omitempty"`
	Children []*WGAFileNode `json:"children,omitempty"`
}

// WGAWorkspace WGA 工作区目录树响应
type WGAWorkspace struct {
	ThreadID  string         `json:"threadId"`
	RunID     string         `json:"runId"`
	FileCount int32          `json:"fileCount"`
	TotalSize int64          `json:"totalSize"`
	IsDisplay bool           `json:"isDisplay"`
	Path      string         `json:"path"`
	Files     []*WGAFileNode `json:"files"`
}

// WGAUploadFile WGA 文件上传结果（代理 bff /file/upload/direct 后解析）
type WGAUploadFile struct {
	FileName string `json:"fileName"`
	FileId   string `json:"fileId"`
	FilePath string `json:"filePath"` // minio 完整 URL，可作为 WGA 多模态 binary.url
	FileSize int64  `json:"fileSize"`
}

// --- WGA API 方法 ---

// CreateWGAConversation 调用 WGA 创建对话接口
func (c *Client) CreateWGAConversation(ctx context.Context, apiKey string, req *WGACreateConversationRequest) (*WGACreateConversationResponse, error) {
	reqURL := fmt.Sprintf("%s/openapi/v1/wga/conversation", c.baseURL)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wga create conversation request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call wga create conversation api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wga create conversation api returned status %d: %s", resp.StatusCode, string(body))
	}

	var wanwuResp wanwuResponse
	if err := json.Unmarshal(body, &wanwuResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if wanwuResp.Code != 0 {
		return nil, fmt.Errorf("wga create conversation api error: code=%d, msg=%s", wanwuResp.Code, wanwuResp.Msg)
	}

	var convResp WGACreateConversationResponse
	if err := json.Unmarshal(wanwuResp.Data, &convResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wga conversation data: %w", err)
	}

	return &convResp, nil
}

// ReplyQuestion 回答 WGA question（人机交互）。
// 调用后 WGA 会继续在原对话 SSE 流上推送后续事件（工具调用/产物生成/RUN_FINISHED）。
func (c *Client) ReplyQuestion(ctx context.Context, apiKey string, req *WGAQuestionReplyRequest) error {
	reqURL := fmt.Sprintf("%s/openapi/v1/wga/conversation/question/reply", c.baseURL)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal wga question reply request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	return c.doQuestionAction(ctx, httpReq, "question reply")
}

// RejectQuestion 拒绝（放弃）WGA question。
func (c *Client) RejectQuestion(ctx context.Context, apiKey, runID, questionID string) error {
	reqURL := fmt.Sprintf("%s/openapi/v1/wga/conversation/question/reject", c.baseURL)

	bodyBytes, err := json.Marshal(map[string]string{
		"runId":      runID,
		"questionId": questionID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal wga question reject request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	return c.doQuestionAction(ctx, httpReq, "question reject")
}

// doQuestionAction 执行 question reply/reject 请求的公共逻辑：发请求、校验状态码与 wanwuResponse.code。
func (c *Client) doQuestionAction(ctx context.Context, httpReq *http.Request, action string) error {
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call wga %s api: %w", action, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wga %s api returned status %d: %s", action, resp.StatusCode, string(body))
	}

	var wanwuResp wanwuResponse
	if err := json.Unmarshal(body, &wanwuResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if wanwuResp.Code != 0 {
		return fmt.Errorf("wga %s api error: code=%d, msg=%s", action, wanwuResp.Code, wanwuResp.Msg)
	}

	return nil
}

// ChatWithWGA 调用 WGA 对话接口（SSE 流式 AG-UI 协议）
func (c *Client) ChatWithWGA(ctx context.Context, apiKey string, chatReq *WGAChatRequest) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s/openapi/v1/wga/conversation/chat", c.baseURL)

	bodyBytes, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wga chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// 使用独立的 HTTP 客户端，不设置超时（用于 SSE 流式响应）
	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call wga chat api: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("wga chat api returned status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// WGAWorkspace 调用 WGA 工作区目录树接口
func (c *Client) WGAWorkspace(ctx context.Context, apiKey, threadID, runID string) (*WGAWorkspace, error) {
	reqURL := fmt.Sprintf("%s/openapi/v1/wga/conversation/workspace?threadId=%s&runId=%s",
		c.baseURL, urlQuery(threadID), urlQuery(runID))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call wga workspace api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wga workspace api returned status %d: %s", resp.StatusCode, string(body))
	}

	var wanwuResp wanwuResponse
	if err := json.Unmarshal(body, &wanwuResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if wanwuResp.Code != 0 {
		return nil, fmt.Errorf("wga workspace api error: code=%d, msg=%s", wanwuResp.Code, wanwuResp.Msg)
	}

	var ws WGAWorkspace
	if err := json.Unmarshal(wanwuResp.Data, &ws); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wga workspace data: %w", err)
	}
	return &ws, nil
}

// WGAWorkspaceDownload 调用 WGA 工作区文件下载接口，返回原始 HTTP 响应（二进制流）。
// 调用方负责读取并关闭 resp.Body。path 为空时下载整个工作区（ZIP）。
func (c *Client) WGAWorkspaceDownload(ctx context.Context, apiKey, threadID, runID, path string) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s/openapi/v1/wga/conversation/workspace/download?threadId=%s&runId=%s",
		c.baseURL, urlQuery(threadID), urlQuery(runID))
	if path != "" {
		reqURL += "&path=" + urlQuery(path)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// 下载用独立客户端，不设超时（文件可能较大）
	dlClient := &http.Client{}
	resp, err := dlClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call wga workspace download api: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("wga workspace download api returned status %d: %s", resp.StatusCode, string(body))
	}
	return resp, nil
}

// UploadFile 上传文件到万悟 bff（/file/upload/direct），返回 minio 文件路径等。
// 用于 WGA 多模态对话前把文件传到 minio，拿到 filePath 作为 binary.url。
func (c *Client) UploadFile(ctx context.Context, apiKey, fileName, mimeType string, data []byte) (*WGAUploadFile, error) {
	reqURL := fmt.Sprintf("%s/openapi/v1/file/upload/direct", c.baseURL)

	// multipart/form-data，字段名 files（对齐 bff DirectUploadFiles 实现）
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("files", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call upload api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload api returned status %d: %s", resp.StatusCode, string(body))
	}

	var wanwuResp wanwuResponse
	if err := json.Unmarshal(body, &wanwuResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if wanwuResp.Code != 0 {
		return nil, fmt.Errorf("upload api error: code=%d, msg=%s", wanwuResp.Code, wanwuResp.Msg)
	}

	var uploadResp struct {
		Files []*WGAUploadFile `json:"files"`
	}
	if err := json.Unmarshal(wanwuResp.Data, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload data: %w", err)
	}
	if len(uploadResp.Files) == 0 {
		return nil, fmt.Errorf("upload api returned no files")
	}
	return uploadResp.Files[0], nil
}

// urlQuery 对查询参数做 URL 编码（避免特殊字符破坏 URL）
func urlQuery(s string) string {
	return url.QueryEscape(s)
}
