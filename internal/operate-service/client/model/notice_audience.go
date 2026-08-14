package model

// MessageAudience 正向收件名单（AudienceType IN (1,3)——特定用户名单 + 点对点单行）：
// 每个收件人一行，是"不该收到这条消息的人"的反向剔除后的正向结果集。
// 事件时冻结（不受 joinedAt 屏蔽）：名单在消息写入那一刻定格，之后加入组织的新成员
// 不会补投历史消息。
// 查询侧必须 EXISTS/IN 半连接（禁 INNER JOIN）：收件名单按 (user_id, org_id) 前缀
// 驱动索引反查 message_id，若改为 INNER JOIN 会让 notice_messages 成为驱动表、
// 候选集退化为全表——读扩散会被击穿成写扩散（每用户扫全库）。
type MessageAudience struct {
	ID uint32 `gorm:"primary_key"`
	// 引用 notice_messages.id（无外键）；本体后写，孤儿行不可见
	MessageID int64  `gorm:"not null;uniqueIndex:uk_notice_msg_audience_user_org_msg,priority:3"`
	UserID    string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_notice_msg_audience_user_org_msg,priority:1"`
	OrgID     string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_notice_msg_audience_user_org_msg,priority:2"`
}

func (MessageAudience) TableName() string { return "notice_message_audience" }

// MessageOrg 按组织发送（AudienceType=2 组织内）：每个目标组织一行。
// 读侧分支 A 由 (org_id, created_at) 前缀驱动范围扫（read-diffusion 直读组织消息），
// created_at 与本体同写同时刻（冗余存储，供 org 索引上的水位范围截断）。
// 排除语义已从产品设计中整体移除（不再有"全局但排除某组织"场景），故无 is_exclude 列。
type MessageOrg struct {
	ID uint32 `gorm:"primary_key"`
	// 引用 notice_messages.id（无外键）；本体后写，孤儿行不可见
	MessageID int64  `gorm:"not null;uniqueIndex:uk_notice_msg_org_message_org,priority:1;index:idx_notice_msg_org_org_created,priority:3"`
	OrgID     string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_notice_msg_org_message_org,priority:2;index:idx_notice_msg_org_org_created,priority:1"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;not null;index:idx_notice_msg_org_org_created,priority:2"`
}

func (MessageOrg) TableName() string { return "notice_message_orgs" }
