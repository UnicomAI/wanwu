package response

type CreateSensitiveWordTableResp struct {
	TableId string `json:"tableId"` //Sensitive word list id
}

type SensitiveWordTableDetail struct {
	TableId   string `json:"tableId"`   // Sensitive word list id
	TableName string `json:"tableName"` // Sensitive word list name
	Remark    string `json:"remark"`    // Remark
	Reply     string `json:"reply"`     // Reply settings
	CreatedAt string `json:"createdAt"` // Sensitive word list creation time
}

type SensitiveWordVocabularyDetail struct {
	WordId        string `json:"wordId"`        // Sensitive word id
	Word          string `json:"word"`          // Sensitive words
	SensitiveType string `json:"sensitiveType"` // Sensitive word type
}
