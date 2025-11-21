package model

const (
	PermissionTypeNone   = -1 //No operation permission
	PermissionTypeView   = 0  //View permissions
	PermissionTypeEdit   = 10 //Edit permissions
	PermissionTypeGrant  = 20 //The reasons for the discontinuous values ​​of authorization permissions prevent the subsequent intermediate permissions. The current logic is authorization permissions>editing permissions>viewing permissions.
	PermissionTypeSystem = 30 //The reason why the system management authorization permissions are discontinuous in value prevents intermediate permissions in the future. The current logic is system management authorization permissions>authorization permissions>editing permissions>viewing permissions.
)

// The KnowledgePermission business is unique. Therefore, a knowledge base, a user, and an organization can only have one
type KnowledgePermission struct {
	Id             uint32 `json:"id" gorm:"primary_key;type:bigint(20) auto_increment;not null;comment:'id';"` // Primary Key
	PermissionId   string `gorm:"column:permission_id;uniqueIndex:idx_unique_permission_id;type:varchar(64);not null;default:''" json:"permissionId"`
	KnowledgeId    string `gorm:"column:knowledge_id;uniqueIndex:idx_knowledge_id_org_user,priority:1;type:varchar(64);not null;default:''" json:"knowledgeId"`
	GrantUserId    string `gorm:"column:grant_user_id;type:varchar(64);not null;default:'';comment:'有权限的用户id';" json:"permissionUserId"`
	GrantOrgId     string `gorm:"column:grant_org_id;type:varchar(64);not null;default:''comment:'有权限的组织id';" json:"permissionOrgId"`
	PermissionType int    `gorm:"column:permission_type;type:tinyint(1);not null;default:0;comment:'权限类型0：读权限，10：编辑权限 20：授权权限，一个知识库只有一个人有授权权限'" json:"permissionType"`
	CreatedAt      int64  `gorm:"column:create_at;type:bigint(20);not null;" json:"createAt"` // Create Time
	UpdatedAt      int64  `gorm:"column:update_at;type:bigint(20);not null;" json:"updateAt"` // Update Time
	OrgId          string `gorm:"column:org_id;uniqueIndex:idx_knowledge_id_org_user,priority:2;type:varchar(64);not null;default:'';" json:"orgId"`
	UserId         string `gorm:"column:user_id;uniqueIndex:idx_knowledge_id_org_user,priority:3;type:varchar(64);not null;default:'';" json:"userId"`
}

func (KnowledgePermission) TableName() string {
	return "knowledge_permission"
}
