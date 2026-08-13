package model

// ReadWatermark 已读水位线（单行二元组水位；一键已读=全部未读不分类别，无 Category 维度）。
//
// 写入协议：同一行取快照边界 → 事务内 SELECT ... FOR UPDATE → 成对字典序单调更新。
// 禁分列 MAX(created_at)/MAX(id) 拼伪游标的原因：created_at 与 id 并不同序
// （同一毫秒可落多条、id 雪花可能乱序），分列取各自最大值拼出的二元组不落在任何真实
// 消息的 (created_at, id) 上，会导致"水位覆盖了尚未读到的消息"——必须取同一行的
// (created_at, id) 成对推进，水位才是真实排序键上的切点。
// 读侧：«未读» = (created_at, id) > (WatermarkAt, WatermarkID) 且无已读/已删状态行。
type ReadWatermark struct {
	ID     uint32 `gorm:"primary_key"`
	UserID string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_notice_watermark_user_org,priority:1"`
	OrgID  string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_notice_watermark_user_org,priority:2"`
	// 水位时间分量；无行等价 (0,0)
	WatermarkAt int64 `gorm:"not null;default:0"`
	// WatermarkID 水位 ID 分量；与 WatermarkAt 构成二元组，与排序键 (created_at,id) 同构。
	// 必要性：created_at 是毫秒级，同毫秒可落多条消息，仅时间分量无法表达"同毫秒内读到哪条"；
	// 配上 ID 分量后二元组与排序键完全同构，读侧 (created_at,id) > (WatermarkAt,WatermarkID)
	// 的字典序比较才能精确定位"比这条新的都未读"。
	WatermarkID int64 `gorm:"not null;default:0"`
	UpdatedAt   int64 `gorm:"autoUpdateTime:milli;not null;default:0"`
}

func (ReadWatermark) TableName() string { return "notice_watermark" }
