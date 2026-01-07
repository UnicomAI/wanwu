package model

const (
	PermissionTypeNone   = -1 //未操作权限
	PermissionTypeView   = 0  //查看权限
	PermissionTypeEdit   = 10 //编辑权限
	PermissionTypeGrant  = 20 //授权权限 数值不连续的原因防止后续有中间权限，目前逻辑 授权权限>编辑权限>查看权限
	PermissionTypeSystem = 30 //系统管理授权权限 数值不连续的原因防止后续有中间权限，目前逻辑 系统管理授权权限>授权权限>编辑权限>查看权限
)

// KnowledgePermission 业务唯一所以，一个知识库，一个用户，一个组织 只能有一条
type KnowledgePermission struct {
	Id             uint32 `json:"id" gorm:"primary_key;type:bigint auto_increment;not null;comment:'id';"` // Primary Key
	PermissionId   string `gorm:"column:permission_id;uniqueIndex:idx_unique_permission_id;type:varchar(64);not null;default:''" json:"permissionId"`
	KnowledgeId    string `gorm:"column:knowledge_id;uniqueIndex:idx_knowledge_id_org_user,priority:1;type:varchar(64);not null;default:''" json:"knowledgeId"`
	GrantUserId    string `gorm:"column:grant_user_id;type:varchar(64);not null;default:'';comment:'有权限的用户id';" json:"permissionUserId"`
	GrantOrgId     string `gorm:"column:grant_org_id;type:varchar(64);not null;default:''comment:'有权限的组织id';" json:"permissionOrgId"`
	PermissionType int    `gorm:"column:permission_type;not null;default:0;comment:'权限类型0：读权限，10：编辑权限 20：授权权限，一个知识库只有一个人有授权权限'" json:"permissionType"`
	CreatedAt      int64  `gorm:"column:create_at;type:bigint;autoCreateTime:milli;not null;" json:"createAt"` // Create Time
	UpdatedAt      int64  `gorm:"column:update_at;type:bigint;autoUpdateTime:milli;not null;" json:"updateAt"` // Update Time
	OrgId          string `gorm:"column:org_id;uniqueIndex:idx_knowledge_id_org_user,priority:2;type:varchar(64);not null;default:'';" json:"orgId"`
	UserId         string `gorm:"column:user_id;uniqueIndex:idx_knowledge_id_org_user,priority:3;type:varchar(64);not null;default:'';" json:"userId"`
}

func (KnowledgePermission) TableName() string {
	return "knowledge_permission"
}
