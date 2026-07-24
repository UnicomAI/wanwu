package model

// WgaArtifact WGA thread 级产物文件累积清单
// 记录某 thread 下智能体用 write 工具写过的产物文件名（basename），跨 run 累积、去重。
// 用于"用户跨 run 说把报告发来"时回发——workspace 无修改时间字段无法按 run 过滤历史，
// write 清单是"智能体主动写过的文件"的可靠信号，比正文子串匹配精确。
// 触发点：handleWGASSEResponse 末尾 extractProducedFiles 提取本次 write 产物后落库累积。
type WgaArtifact struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	ChannelID string `gorm:"uniqueIndex:idx_wa_ch_thread_name;size:64"`
	UserID    string `gorm:"uniqueIndex:idx_wa_ch_thread_name;size:64"`
	ThreadID  string `gorm:"uniqueIndex:idx_wa_ch_thread_name;size:128"`
	FileName  string `gorm:"uniqueIndex:idx_wa_ch_thread_name;size:255"` // basename（去目录前缀的完整文件名）
}

// TableName 表名
func (WgaArtifact) TableName() string {
	return "channel_wga_artifacts"
}
