package response

type CreateSensitiveWordTableResp struct {
	TableId string `json:"tableId"` //敏感词表id
}

type SensitiveWordTableDetail struct {
	TableId   string `json:"tableId"`   // 敏感词表id
	TableName string `json:"tableName"` // 敏感词表名
	Remark    string `json:"remark"`    // 备注
	Reply     string `json:"reply"`     // 回复设置
	CreatedAt string `json:"createdAt"` // 敏感词表创建时间
	UpdatedAt string `json:"updatedAt"` // 敏感词表更新时间
	Type      string `json:"type"`      // 敏感词表类型
}

type SensitiveWordVocabularyDetail struct {
	WordId        string `json:"wordId"`        // 敏感词id
	Word          string `json:"word"`          // 敏感词
	SensitiveType string `json:"sensitiveType"` // 敏感词类型
}
