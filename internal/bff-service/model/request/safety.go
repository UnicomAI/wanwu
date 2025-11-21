package request

type CreateSensitiveWordTableReq struct {
	TableName string `json:"tableName" validate:"required"` // Sensitive word list name
	Remark    string `json:"remark"`                        // Remark
	CommonCheck
}

type UpdateSensitiveWordTableReq struct {
	TableId   string `json:"tableId" validate:"required"`   // Sensitive word list id
	TableName string `json:"tableName" validate:"required"` // Sensitive word list name
	Remark    string `json:"remark"`                        // Remark
	CommonCheck
}

type DeleteSensitiveWordTableReq struct {
	TableId string `json:"tableId" validate:"required"` // Sensitive word list id
	CommonCheck
}

type GetSensitiveVocabularyReq struct {
	TableId string `json:"tableId" form:"tableId" validate:"required"` // Sensitive word list id
	CommonCheck
}

type DeleteSensitiveVocabularyReq struct {
	TableId string `json:"tableId" validate:"required"` // Sensitive word list id
	WordId  string `json:"wordId" validate:"required"`  // Sensitive word id
	CommonCheck
}

type UploadSensitiveVocabularyReq struct {
	TableId       string `json:"tableId" validate:"required"`    // Sensitive word list id
	ImportType    string `json:"importType" validate:"required"` // How to upload sensitive words, single: add a single item, file: upload in batches
	Word          string `json:"word"`                           // Sensitive words
	SensitiveType string `json:"sensitiveType"`                  // Sensitive word types (Political: Political, Insult: Revile, Pornography: Pornography, Violent Terror: ViolentTerror, Prohibited: Illegal, Information Security: InformationSecurity, Others: Other)
	FileName      string `json:"fileName"`                       // file name
	CommonCheck
}

type UpdateSensitiveWordTableReplyReq struct {
	TableId string `json:"tableId" validate:"required"` // Sensitive word list id
	Reply   string `json:"reply"`                       // Reply settings
	CommonCheck
}
