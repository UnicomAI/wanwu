package model

// 消息类别
const (
	NoticeCategoryAnnouncement   int8 = 1 // 公告
	NoticeCategoryProductService int8 = 2 // 产品服务
	NoticeCategoryTicket         int8 = 3 // 工单
)

// 受众类型（AudienceType）
const (
	// AudienceTypePrivate 私密（点对点=audience 单行，读侧与 3 同支 AT IN (1,3)，本期无生产者）
	AudienceTypePrivate int8 = 1
	// AudienceTypeOrg 组织内（notice_message_orgs 指定组织，不级联子组织）
	AudienceTypeOrg int8 = 2
	// AudienceTypeSpecific 特定用户（notice_message_audience 正向名单）
	AudienceTypeSpecific int8 = 3
	// AudienceTypeGlobal 全局（全员可见）
	AudienceTypeGlobal int8 = 4
)

// 文案分支（EventVariant）：同一事件的多条差分本体靠它与 EventID 组成幂等键。
// 一段式：variant 直接承载文案语义，不再有独立的文案细分字段。
// 规则型（应用/模型，走 scope）：应用统一 online「已上线」；模型 online「上线」/offline「下线」。
// 名单操作语义（当前名单型使用，未来名单型消息复用）：shared/unshared/deleted/perm_changed。
const (
	EventVariantOnline  = "online"
	EventVariantOffline = "offline"
	// 名单操作语义（当前名单型使用，未来名单型消息可复用）
	EventVariantShared      = "shared"
	EventVariantUnshared    = "unshared"
	EventVariantDeleted     = "deleted"
	EventVariantPermChanged = "perm_changed"
)

// NoticeMessage 消息本体（读扩散：广播只存 1 份 + 受众规则），一条事件×文案版本一行、多账号共享。
//
// ID 由应用层预生成（雪花），不用自增——写入协议要求"名单先写、本体后写"，写名单时必须已知 MessageID。
// 表名靠 GORM 复数化恰好得到 notice_messages。
//
// 索引名相对 DDL 加了 notice_msg 前缀：postgres 的索引名是 schema 全局唯一的，
// 而 DDL 里 uk_user_org_msg / idx_message 在三张表重复出现，不加前缀在 postgres 下 AutoMigrate 会冲突。
// utf8mb4_bin collate 不写在字段 tag 里——库级已由 configs/middleware/mysql/initdb.d/init.sql
// 的 `COLLATE utf8mb4_bin` 提供，写进 tag 会让 postgres 方言建表失败。
type NoticeMessage struct {
	// --- 幂等键区 ---
	// 应用层预生成（雪花/序列），写入协议要求先于名单行确定
	ID int64 `gorm:"primaryKey;autoIncrement:false"`
	// 幂等键：源业务事件唯一标识（资源+目标状态+版本+时间戳）。
	// 时间戳刻画「这一次动作」，避免反复操作撞键被当重试吞掉；
	// appId/userId 为 UUID 时 key 可达百字符级，故列宽放宽到 128。
	EventID string `gorm:"type:varchar(128);not null;default:'';uniqueIndex:uk_notice_msg_event,priority:1"`
	// 同一事件的文案分支：online/offline（规则型 scope）/shared/unshared/deleted/perm_changed（名单操作语义）/ann_N（公告）
	EventVariant string `gorm:"type:varchar(16);not null;default:'';uniqueIndex:uk_notice_msg_event,priority:2"`

	// --- 时间区 ---
	// 毫秒时间戳；与 ID 共同构成排序键与水位切点
	CreatedAt int64 `gorm:"autoCreateTime:milli;not null;index:idx_notice_msg_at_created,priority:2"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli;not null;default:0"`

	// --- 内容区 ---
	// 1 公告 / 2 产品服务 / 3 工单
	Category int8 `gorm:"not null;index:idx_notice_msg_at_created,priority:3"`
	// SubjectType 消息主体类型（关联对象）：agent/workflow/chatflow/rag/skill/model/knowledge/ticket。
	// 用途：读侧按"主体类型"做应用内定向（如工单/知识库各自的收件箱角标、消息中心按主体筛选），
	// 也是后续扩展"同主体聚合、按主体查全部消息"等能力的事实依据。
	// 它只刻画"这条消息关于什么"，不承载文案语义（文案由 EventVariant + Title/Content 固化）；
	// 子类（如工单类型）按需扩展 variant 或加列，不并入本字段。
	SubjectType string `gorm:"type:varchar(32);not null;default:''"`
	Title       string `gorm:"type:varchar(256);not null;default:''"`
	// 正文（唯一正文模型）；$${超链文字}$$ 占位符按序对应 actions[i]，前端解析
	Content string `gorm:"type:text"`
	// 跳转数组 [{msgType,actionType,actionParams}]；写入侧按各 msgType 的必填字段 schema 校验后落库
	Actions string `gorm:"type:json"`

	// --- 受众区 ---
	// 1 私密 / 2 组织内 / 3 特定用户 / 4 全局；组织维度受众落在 notice_message_orgs / notice_message_audience
	AudienceType int8 `gorm:"not null;index:idx_notice_msg_at_created,priority:1"`

	// --- 操作者区 ---
	// 操作者（真实发送者语义，不承载收件人）；«vis» 组织内/全局分支用于排除本人
	SenderID    string `gorm:"type:varchar(64);not null;default:'';index:idx_notice_msg_at_created,priority:4"`
	SenderOrgID string `gorm:"type:varchar(64);not null;default:'';index:idx_notice_msg_at_created,priority:5"`
}
