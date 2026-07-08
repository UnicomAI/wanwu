package chat

import (
	"context"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/internal/channel-service/wanwu"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// pendingQuestionTTL pending question 最长存活时间。
// 超时后视为用户放弃：close CancelCh（通知 SSE goroutine 退出）并尝试调 WGA question/reject。
const pendingQuestionTTL = 5 * time.Minute

// PendingQuestion 待回答的 WGA question。
// 由 SSE goroutine 在收到 ACTIVITY_SNAPSHOT(question, pending) 时写入，
// 由用户后续消息触发的 handleQuestionReply 读取并回答。
type PendingQuestion struct {
	QuestionID string
	RunID      string
	ThreadID   string
	ApiKey     string
	Questions  []wanwu.WGAQuestion
	CreatedAt  time.Time

	// CancelCh 在超时或放弃时被 close，通知阻塞中的 SSE goroutine 退出读取循环。
	// 由 QuestionManager 在删除条目时 close（用 once 保证幂等）。
	CancelCh chan struct{}
	once     sync.Once
}

// cancel 幂等地 close CancelCh。
func (p *PendingQuestion) cancel() {
	p.once.Do(func() { close(p.CancelCh) })
}

// QuestionManager pending question 内存管理器。
// key = channelID + ":" + platformUserID。进程重启后丢失——丢失时用户回复会被当作普通对话发给 WGA（降级，不致命）。
type QuestionManager struct {
	store   sync.Map
	baseURL string // 用于超时清理时 new wanwu.Client 调 reject
}

// NewQuestionManager 创建 pending question 管理器并启动超时清理 goroutine。
func NewQuestionManager(baseURL string) *QuestionManager {
	m := &QuestionManager{baseURL: baseURL}
	go m.cleanupLoop()
	return m
}

// keyOf 构造存储 key。
func keyOf(channelID, userID string) string {
	return channelID + ":" + userID
}

// Set 记录一条 pending question（覆盖同 channel+user 的旧条目，先 cancel 旧的）。
func (m *QuestionManager) Set(channelID, userID string, pq *PendingQuestion) {
	key := keyOf(channelID, userID)
	if old, ok := m.store.LoadAndDelete(key); ok {
		old.(*PendingQuestion).cancel() // 旧的 SSE goroutine 退出（理论上同 user 不会并发两条）
	}
	pq.CreatedAt = time.Now()
	if pq.CancelCh == nil {
		pq.CancelCh = make(chan struct{})
	}
	m.store.Store(key, pq)
}

// Get 读取 pending question。
func (m *QuestionManager) Get(channelID, userID string) (*PendingQuestion, bool) {
	v, ok := m.store.Load(keyOf(channelID, userID))
	if !ok {
		return nil, false
	}
	return v.(*PendingQuestion), true
}

// Delete 删除 pending question 并 close 其 CancelCh。
// 用于放弃/超时场景：close CancelCh 通知 SSE goroutine 退出（WGA 不再推后续事件）。
func (m *QuestionManager) Delete(channelID, userID string) *PendingQuestion {
	v, ok := m.store.LoadAndDelete(keyOf(channelID, userID))
	if !ok {
		return nil
	}
	pq := v.(*PendingQuestion)
	pq.cancel()
	return pq
}

// Complete 删除 pending question 但不 close CancelCh。
// 用于回答成功场景：从 store 移除条目（避免拦截用户下一条消息 / 超时清理误调 reject），
// 但保留 CancelCh 开启——SSE goroutine 需继续读 WGA 推来的后续事件（工具调用/产物/RUN_FINISHED）。
// 若 WGA 长时间不推事件，5 分钟超时清理兜底 close CancelCh（此时条目已不在 store，清理不会调 reject）。
func (m *QuestionManager) Complete(channelID, userID string) *PendingQuestion {
	v, ok := m.store.LoadAndDelete(keyOf(channelID, userID))
	if !ok {
		return nil
	}
	return v.(*PendingQuestion)
}

// cleanupLoop 每 1 分钟扫描一次，清理超时条目：close CancelCh + 异步调 reject 放弃。
func (m *QuestionManager) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		m.store.Range(func(k, v any) bool {
			pq := v.(*PendingQuestion)
			if now.Sub(pq.CreatedAt) >= pendingQuestionTTL {
				m.store.Delete(k)
				pq.cancel()
				// 异步通知 WGA 放弃该 question，失败仅记日志
				go func(p *PendingQuestion) {
					cli := wanwu.NewClient(m.baseURL)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := cli.RejectQuestion(ctx, p.ApiKey, p.RunID, p.QuestionID); err != nil {
						log.Warnf("[Question] timeout reject failed: runId=%s, questionId=%s, err=%v",
							p.RunID, p.QuestionID, err)
					}
				}(pq)
			}
			return true
		})
	}
}
