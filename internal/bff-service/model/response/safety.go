package response

type CreateSensitiveWordTableResp struct {
	TableId string `json:"tableId"` //敏感词表id [EN] Sensitive word list id
}

type SensitiveWordTableDetail struct {
	TableId   string `json:"tableId"`   // 敏感词表id [EN] Sensitive word list id
	TableName string `json:"tableName"` // 敏感词表名 [EN] Sensitive word list name
	Remark    string `json:"remark"`    // 备注 [EN] Remark
	Reply     string `json:"reply"`     // 回复设置 [EN] Reply settings
	CreatedAt string `json:"createdAt"` // 敏感词表创建时间 [EN] Sensitive word list creation time
}

type SensitiveWordVocabularyDetail struct {
	WordId        string `json:"wordId"`        // 敏感词id [EN] Sensitive word id
	Word          string `json:"word"`          // 敏感词 [EN] Sensitive words
	SensitiveType string `json:"sensitiveType"` // 敏感词类型 [EN] Sensitive word type
}
