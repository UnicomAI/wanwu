package request

type CreateSensitiveWordTableReq struct {
	TableName string `json:"tableName" validate:"required"` // 敏感词表名 [EN] Sensitive word list name
	Remark    string `json:"remark"`                        // 备注 [EN] Remark
	CommonCheck
}

type UpdateSensitiveWordTableReq struct {
	TableId   string `json:"tableId" validate:"required"`   // 敏感词表id [EN] Sensitive word list id
	TableName string `json:"tableName" validate:"required"` // 敏感词表名 [EN] Sensitive word list name
	Remark    string `json:"remark"`                        // 备注 [EN] Remark
	CommonCheck
}

type DeleteSensitiveWordTableReq struct {
	TableId string `json:"tableId" validate:"required"` // 敏感词表id [EN] Sensitive word list id
	CommonCheck
}

type GetSensitiveVocabularyReq struct {
	TableId string `json:"tableId" form:"tableId" validate:"required"` // 敏感词表id [EN] Sensitive word list id
	CommonCheck
}

type DeleteSensitiveVocabularyReq struct {
	TableId string `json:"tableId" validate:"required"` // 敏感词表id [EN] Sensitive word list id
	WordId  string `json:"wordId" validate:"required"`  // 敏感词id [EN] Sensitive word id
	CommonCheck
}

type UploadSensitiveVocabularyReq struct {
	TableId       string `json:"tableId" validate:"required"`    // 敏感词表id [EN] Sensitive word list id
	ImportType    string `json:"importType" validate:"required"` // 上传敏感词方式，single：单条添加，file：批量上传 [EN] How to upload sensitive words, single: add a single item, file: upload in batches
	Word          string `json:"word"`                           // 敏感词 [EN] Sensitive words
	SensitiveType string `json:"sensitiveType"`                  // 敏感词类型 (涉政:Political, 辱骂:Revile, 涉黄:Pornography, 暴恐:ViolentTerror, 违禁:Illegal, 信息安全:InformationSecurity, 其他:Other) [EN] Sensitive word types (Political: Political, Insult: Revile, Pornography: Pornography, Violent Terror: ViolentTerror, Prohibited: Illegal, Information Security: InformationSecurity, Others: Other)
	FileName      string `json:"fileName"`                       // 文件名 [EN] file name
	CommonCheck
}

type UpdateSensitiveWordTableReplyReq struct {
	TableId string `json:"tableId" validate:"required"` // 敏感词表id [EN] Sensitive word list id
	Reply   string `json:"reply"`                       // 回复设置 [EN] Reply settings
	CommonCheck
}
