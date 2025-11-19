package model

// ClientDailyStats 客户端日统计表 [EN] ClientDailyStats client daily statistics table
type ClientDailyStats struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdateAt  int64  `gorm:"autoUpdateTime:milli"`
	Date      string `gorm:"index:idx_client_daily_date"`
	DauCount  int32  `gorm:"index:idx_client_daily_dau_count"`
}
