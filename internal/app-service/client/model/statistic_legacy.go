package model

// 只读旧统计表模型（不参与 AutoMigrate）。TableName 钉死现网表名。

// LegacyModelStatistic 对应 model_statistics
type LegacyModelStatistic struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`

	OrgID    string `gorm:"size:64"`
	UserID   string `gorm:"size:64"`
	ModelID  string `gorm:"size:64"`
	Provider string `gorm:"size:64"`
	Date     string `gorm:"size:16"`

	Model             string `gorm:"size:128"`
	ModelType         string `gorm:"size:32"`
	PromptTokens      int64
	CompletionTokens  int64
	TotalTokens       int64
	FirstTokenLatency int64
	Costs             int64
	CallCount         int32
	StreamCount       int32
	NonStreamCount    int32
	CallFailure       int32
	StreamFailure     int32
	NonStreamFailure  int32
}

func (LegacyModelStatistic) TableName() string { return "model_statistics" }

// LegacyAppStatistic 对应 app_statistics
type LegacyAppStatistic struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`

	OrgID   string `gorm:"size:64"`
	UserID  string `gorm:"size:64"`
	AppID   string `gorm:"size:64"`
	AppType string `gorm:"size:64"`
	Date    string `gorm:"size:16"`

	CallCount   int32
	CallFailure int32

	StreamCount   int32
	StreamFailure int32
	StreamCosts   int64

	NonStreamCount   int32
	NonStreamFailure int32
	NonStreamCosts   int64

	WebCallCount       int32
	WebCallFailure     int32
	OpenapiCallCount   int32
	OpenapiCallFailure int32
	WebUrlCallCount    int32
	WebUrlCallFailure  int32
}

func (LegacyAppStatistic) TableName() string { return "app_statistics" }

// LegacyAPIKeyStatistic 对应 api_key_statistics
type LegacyAPIKeyStatistic struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`

	OrgID      string `gorm:"size:64"`
	UserID     string `gorm:"size:64"`
	APIKeyID   string `gorm:"size:64"`
	MethodPath string `gorm:"size:128"`
	Date       string `gorm:"size:16"`

	CallCount        int32
	CallFailure      int32
	StreamCount      int32
	NonStreamCount   int32
	StreamFailure    int32
	NonStreamFailure int32
	StreamCosts      int64
	NonStreamCosts   int64
}

func (LegacyAPIKeyStatistic) TableName() string { return "api_key_statistics" }

// LegacyAPIKeyRecord 对应 api_key_records
type LegacyAPIKeyRecord struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`

	OrgID          string
	UserID         string
	APIKeyID       string
	MethodPath     string
	CallTime       int64
	ResponseStatus string
	IsStream       bool
	StreamCosts    int64
	NonStreamCosts int64
	RequestBody    string `gorm:"type:longtext"`
	ResponseBody   string `gorm:"type:longtext"`
	Date           string // 不迁入 V2
}

func (LegacyAPIKeyRecord) TableName() string { return "api_key_records" }
