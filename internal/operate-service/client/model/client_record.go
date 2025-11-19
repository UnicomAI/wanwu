package model

// ClientRecord client table
type ClientRecord struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	ClientId  string `gorm:"index:idx_client_id"`
	CreatedAt int64  `gorm:"index:idx_client_created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"index:idx_client_updated_at"`
}
