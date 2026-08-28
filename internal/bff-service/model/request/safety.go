package request

import "github.com/UnicomAI/wanwu/pkg/util"

type CreateSensitiveWordTableReq struct {
	TableName string `json:"tableName" validate:"required"` // 敏感词表名
	Remark    string `json:"remark"`                        // 备注
	Type      string `json:"type" validate:"required"`      // 敏感词表类型，personal：个人，global：全局
}

func (req *CreateSensitiveWordTableReq) Check() error {
	if err := util.ValidateName(&req.TableName, util.SubjectSensitiveWordTable); err != nil {
		return err
	}
	return util.ValidateDesc(&req.Remark, util.SubjectSensitiveWordTable)
}

type UpdateSensitiveWordTableReq struct {
	TableId   string `json:"tableId" validate:"required"`   // 敏感词表id
	TableName string `json:"tableName" validate:"required"` // 敏感词表名
	Remark    string `json:"remark"`                        // 备注
	CommonCheck
}

type DeleteSensitiveWordTableReq struct {
	TableId string `json:"tableId" validate:"required"` // 敏感词表id
	CommonCheck
}

type GetSensitiveVocabularyReq struct {
	TableId string `json:"tableId" form:"tableId" validate:"required"` // 敏感词表id
	CommonCheck
}

type DeleteSensitiveVocabularyReq struct {
	TableId string `json:"tableId" validate:"required"` // 敏感词表id
	WordId  string `json:"wordId" validate:"required"`  // 敏感词id
	CommonCheck
}

type UploadSensitiveVocabularyReq struct {
	TableId       string `json:"tableId" validate:"required"`    // 敏感词表id
	ImportType    string `json:"importType" validate:"required"` // 上传敏感词方式，single：单条添加，file：批量上传
	Word          string `json:"word"`                           // 敏感词
	SensitiveType string `json:"sensitiveType"`                  // 敏感词类型 (涉政:Political, 辱骂:Revile, 涉黄:Pornography, 暴恐:ViolentTerror, 违禁:Illegal, 信息安全:InformationSecurity, 其他:Other)
	FileName      string `json:"fileName"`                       // 文件名
	CommonCheck
}

type UpdateSensitiveWordTableReplyReq struct {
	TableId string `json:"tableId" validate:"required"` // 敏感词表id
	Reply   string `json:"reply"`                       // 回复设置
	CommonCheck
}
