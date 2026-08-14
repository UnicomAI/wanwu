package chat

import (
	"context"
	"sync"

	"github.com/UnicomAI/wanwu/internal/channel-service/wanwu"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// ChatflowInputCache 对话流开始节点输入参数缓存。
// key = channelID + ":" + workflowUUID（同一通道绑定的对话流画布固定，入参基本不变）。
// 进程重启后丢失——丢失时重新调 8999 schema 拉取，不致命。
// 画布改了入参后，需重启通道或等 TTL 过期才刷新（MVP 接受重启刷新）。
type ChatflowInputCache struct {
	mu    sync.Mutex
	store map[string][]wanwu.ChatflowInput
}

// NewChatflowInputCache 创建对话流入参缓存。
func NewChatflowInputCache() *ChatflowInputCache {
	return &ChatflowInputCache{store: make(map[string][]wanwu.ChatflowInput)}
}

// Get 获取已缓存的入参列表，未命中返回 ok=false。
func (c *ChatflowInputCache) Get(channelID, workflowUUID string) ([]wanwu.ChatflowInput, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.store[channelID+":"+workflowUUID]
	return v, ok
}

// Set 缓存入参列表（含空切片：表示画布无入参，避免重复拉取）。
func (c *ChatflowInputCache) Set(channelID, workflowUUID string, inputs []wanwu.ChatflowInput) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[channelID+":"+workflowUUID] = inputs
}

// getChatflowInputs 获取对话流开始节点输入参数（带缓存）。
// 缓存命中直接返回；未命中调 8999 schema 拉取并缓存。
// 拉取失败返回 nil + 错误（调用方按无入参降级，纯文字对话不受影响）。
func (h *Handler) getChatflowInputs(ctx context.Context, channelID, workflowUUID, orgID, userID string) ([]wanwu.ChatflowInput, error) {
	if inputs, ok := h.chatflowInputCache.Get(channelID, workflowUUID); ok {
		return inputs, nil
	}
	wanwuClient := wanwu.NewClient(h.cfg.BFF.ApiBaseUrl)
	inputs, err := wanwuClient.GetChatflowInputs(ctx, h.cfg.Workflow.Endpoint, workflowUUID, orgID, userID)
	if err != nil {
		log.Warnf("[Chatflow] get inputs failed channel=%s uuid=%s: %v", channelID, workflowUUID, err)
		return nil, err
	}
	h.chatflowInputCache.Set(channelID, workflowUUID, inputs)
	log.Infof("[Chatflow] fetched inputs channel=%s uuid=%s count=%d", channelID, workflowUUID, len(inputs))
	return inputs, nil
}
