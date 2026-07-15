package chat

import (
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
)

// pendingAttachmentTTL 待用附件最长存活时间。
// 超时后视为用户放弃：从缓存丢弃（用户后续发文字时不再带回这些附件）。
// 取 10 分钟：足够用户组织指令，又不至于长期占内存。
const pendingAttachmentTTL = 10 * time.Minute

// PendingAttachment 已上传到万悟 minio 的待用附件。
// 存 minio URL（而非原始字节），避免用户连续发多个文件时重复上传、占内存。
// 由 doWGAChat 在"纯附件消息"分支写入，在"有文字指令"分支 Drain 取出拼进 WGA content。
type PendingAttachment struct {
	URL      string // minio 文件路径，作为 WGA 多模态 binary.url
	FileName string
	MimeType string

	// CreatedAt 本批附件的最后追加时间（滑动过期：每次 Append 刷新），
	// 供 cleanupLoop 判断整批是否过期。
	CreatedAt time.Time
}

// AttachmentCache 待用附件内存管理器。
// key = channelID + ":" + platformUserID（复用 question_manager.go 的 keyOf）。
// 进程重启后丢失——丢失时用户发文字会把当条附件单独发，降级为现状，不致命（与 QuestionManager 一致取舍）。
// 用 mutex + map 而非 sync.Map：列表追加是"读-改-写"，mutex 比 CAS 直接（切片不可比较，CAS 会 panic）。
type AttachmentCache struct {
	mu    sync.Mutex
	store map[string][]*PendingAttachment
}

// NewAttachmentCache 创建待用附件管理器并启动超时清理 goroutine。
func NewAttachmentCache() *AttachmentCache {
	c := &AttachmentCache{store: make(map[string][]*PendingAttachment)}
	go c.cleanupLoop()
	return c
}

// Append 追加一个待用附件到 (channelID, userID) 的列表。
// 刷新整批的 CreatedAt 为当前时间（滑动过期：用户持续发文件不会中途过期）。
func (c *AttachmentCache) Append(channelID, userID string, att *PendingAttachment) {
	if att == nil {
		return
	}
	att.CreatedAt = time.Now()
	key := keyOf(channelID, userID)
	c.mu.Lock()
	c.store[key] = append(c.store[key], att)
	c.mu.Unlock()
}

// Drain 取出并清空 (channelID, userID) 的待用附件列表。
// 返回 nil 表示无暂存附件。
func (c *AttachmentCache) Drain(channelID, userID string) []*PendingAttachment {
	key := keyOf(channelID, userID)
	c.mu.Lock()
	list := c.store[key]
	delete(c.store, key)
	c.mu.Unlock()
	return list
}

// cleanupLoop 每 1 分钟扫描一次，丢弃超时批次（整批最后一条 CreatedAt 距今 >= TTL）。
// 附件无需通知上游（区别于 QuestionManager 要调 reject），直接丢弃即可。
func (c *AttachmentCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, list := range c.store {
			if len(list) == 0 {
				delete(c.store, k)
				continue
			}
			if now.Sub(list[len(list)-1].CreatedAt) >= pendingAttachmentTTL {
				count := len(list)
				delete(c.store, k)
				log.Infof("[AttachmentCache] expired pending attachments dropped: key=%s, count=%d", k, count)
			}
		}
		c.mu.Unlock()
	}
}
