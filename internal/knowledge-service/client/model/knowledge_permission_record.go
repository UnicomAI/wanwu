package model

const (
	RecordOptionAdd    = 0 //添加权限
	RecordOptionDelete = 1 //删除权限
	RecordOptionEdit   = 2 //修改权限
)

// KnowledgePermissionRecord 业务唯一所以，一个知识库，一个用户，一个组织 只能有一条
type KnowledgePermissionRecord struct {
	Id                 uint32 `json:"id" gorm:"primary_key;type:bigint auto_increment;not null;comment:'id';"` // Primary Key
	RecordId           string `gorm:"column:record_id;uniqueIndex:idx_record_id;type:varchar(64);not null;default:''" json:"recordId"`
	KnowledgeId        string `gorm:"column:knowledge_id;;type:varchar(64);not null;default:''" json:"knowledgeId"`
	Option             int    `gorm:"column:option;not null;default:0;comment:'操作类型0：添加权限，1：删除权限, 2:修改权限'" json:"option"`
	OperatorUserId     string `gorm:"column:operator_user_id;type:varchar(64);not null;default:'';comment:'有权限的用户id';" json:"operatorUserId"`
	OperatorOrgId      string `gorm:"column:operator_org_id;type:varchar(64);not null;default:''comment:'有权限的组织id';" json:"operatorOrgId"`
	FromPermissionType int    `gorm:"column:from_permission_type;not null;default:0;comment:'权限类型-1:无权限，0：读权限，10：编辑权限 20：授权权限，一个知识库只有一个人有授权权限'" json:"fromPermissionType"`
	ToPermissionType   int    `gorm:"column:to_permission_type;not null;default:0;comment:'权限类型-1:无权限，0：读权限，10：编辑权限 20：授权权限，一个知识库只有一个人有授权权限'" json:"toPermissionType"`
	OwnerOrgId         string `gorm:"column:owner_org_id;type:varchar(64);not null;default:'';" json:"ownerOrgId"`
	OwnerUserId        string `gorm:"column:owner_user_id;type:varchar(64);not null;default:'';" json:"ownerUserId"`
	CreatedAt          int64  `gorm:"column:create_at;type:bigint;autoCreateTime:milli;not null;" json:"createAt"` // Create Time
	UpdatedAt          int64  `gorm:"column:update_at;type:bigint;autoUpdateTime:milli;not null;" json:"updateAt"` // Update Time

}

func (KnowledgePermissionRecord) TableName() string {
	return "knowledge_permission_record"
}
