package chat

import (
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
)

// pendingInputTTL 对话流追问状态最长存活时间。
// 超时视为用户放弃：从缓存丢弃状态（用户后续重发会重新触发追问）。
// 取 5 分钟（与 QuestionManager 一致）：足够用户组织回复，又不至于长期占内存。
const pendingInputTTL = 5 * time.Minute

// ChatflowPendingInput 对话流待补的单个必填入参。
type ChatflowPendingInput struct {
	Name string // 入参变量名（如 num）
	Type string // schema 类型：integer/number/boolean/string
}

// ChatflowPendingInputState 一个 (channel,user) 的对话流追问会话状态。
// 由 handleChatflowMessage 检测到必填文字入参未填时写入，
// 由 handleChatflowInputReply 在用户回复时读取并逐步补全。
type ChatflowPendingInputState struct {
	UUID           string                 // 对话流 uuid（ch.AppID）
	ConversationID string                 // 已创建的会话 id（追问前已创建，补全后复用）
	OriginalQuery  string                 // 用户最初发的文字，作为最终 chat 的 query（USER_INPUT）
	PendingInputs  []ChatflowPendingInput // 待补的必填入参（逐个追问，首个是当前要问的）
	FilledParams   map[string]any         // 已补的入参值
	Attachments    []*PendingAttachment   // 追问前已上传的附件（补完后一起填进 parameters）
	CreatedAt      time.Time
}

// ChatflowPendingInputCache 对话流追问状态内存管理器。
// key = channelID + ":" + userID（复用 keyOf）。进程重启后丢失——丢失时用户重发重新触发追问，不致命
// （与 QuestionManager / AttachmentCache 一致取舍）。用 sync.Map 存指针（*State），
// 回复时直接改指针指向的对象字段，无需重新 Set。
type ChatflowPendingInputCache struct {
	store sync.Map
}

// NewChatflowPendingInputCache 创建追问状态管理器并启动超时清理 goroutine。
func NewChatflowPendingInputCache() *ChatflowPendingInputCache {
	c := &ChatflowPendingInputCache{}
	go c.cleanupLoop()
	return c
}

// Set 记录一条追问状态（覆盖同 channel+user 的旧状态）。
func (c *ChatflowPendingInputCache) Set(channelID, userID string, s *ChatflowPendingInputState) {
	if s == nil {
		return
	}
	s.CreatedAt = time.Now()
	c.store.Store(keyOf(channelID, userID), s)
}

// Get 读取追问状态。
func (c *ChatflowPendingInputCache) Get(channelID, userID string) (*ChatflowPendingInputState, bool) {
	v, ok := c.store.Load(keyOf(channelID, userID))
	if !ok {
		return nil, false
	}
	return v.(*ChatflowPendingInputState), true
}

// Delete 删除追问状态，返回被删的状态（用于补全完成后清理）。
func (c *ChatflowPendingInputCache) Delete(channelID, userID string) *ChatflowPendingInputState {
	v, ok := c.store.LoadAndDelete(keyOf(channelID, userID))
	if !ok {
		return nil
	}
	return v.(*ChatflowPendingInputState)
}

// cleanupLoop 每 1 分钟扫描一次，丢弃超时状态（CreatedAt 距今 >= TTL）。
// 追问状态无需通知上游（区别于 QuestionManager 要调 reject），直接丢弃即可。
func (c *ChatflowPendingInputCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.store.Range(func(k, v any) bool {
			s, ok := v.(*ChatflowPendingInputState)
			if !ok {
				c.store.Delete(k)
				return true
			}
			if now.Sub(s.CreatedAt) >= pendingInputTTL {
				c.store.Delete(k)
				log.Infof("[ChatflowPendingInput] expired pending input state dropped: key=%v", k)
			}
			return true
		})
	}
}
