package request

type KnowledgeOrgSelectReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId" validate:"required"` //知识库id [EN] knowledge base id
	CommonCheck
}

type KnowledgeUserSelectReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId" validate:"required"` //知识库id [EN] knowledge base id
	CommonCheck
}

type KnowledgeUserNoPermitSelectReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId" validate:"required"` //知识库id [EN] knowledge base id
	OrgId       string `json:"orgId" form:"orgId" validate:"required"`             //选择组织id [EN] Select organization id
	Transfer    bool   `json:"transfer" form:"transfer"`                           //是否是转让列表 [EN] Is it a transfer list?
	CommonCheck
}

type KnowledgeUserAddReq struct {
	KnowledgeId       string               `json:"knowledgeId" validate:"required"`       // 知识库id [EN] knowledge base id
	PermissionType    int                  `json:"permissionType"`                        // 权限类型:0: 查看权限; 10: 编辑权限; 20: 授权权限,数值不连续的原因防止后续有中间权限，目前逻辑 授权权限>编辑权限>查看权限 [EN] Permission type: 0: View permission; 10: Edit permission; 20: Authorization permission. The reason for discontinuous values ​​prevents intermediate permissions in the future. The current logic is Authorization permission > Edit permission > View permission.
	KnowledgeUserList []*KnowledgeUserInfo `json:"knowledgeUserList" validate:"required"` // 知识库用户信息 [EN] Knowledge base user information
	CommonCheck
}

type KnowledgeUserEditReq struct {
	KnowledgeId   string             `json:"knowledgeId" validate:"required"`   // 知识库id [EN] knowledge base id
	KnowledgeUser *KnowledgeUserInfo `json:"knowledgeUser" validate:"required"` // 知识库用户信息 [EN] Knowledge base user information
	CommonCheck
}

type KnowledgeUserDeleteReq struct {
	KnowledgeId  string `json:"knowledgeId" validate:"required"`  // 知识库id [EN] knowledge base id
	PermissionId string `json:"permissionId" validate:"required"` // 知识库用户信息权限信息 [EN] Knowledge base user information permission information
	CommonCheck
}

type KnowledgeTransferUserAdminReq struct {
	KnowledgeId   string             `json:"knowledgeId" validate:"required"`   // 知识库id [EN] knowledge base id
	PermissionId  string             `json:"permissionId" form:"permissionId"`  // 权限id编辑时传入 [EN] The permission id is passed in when editing
	KnowledgeUser *KnowledgeUserInfo `json:"knowledgeUser" validate:"required"` // 知识库用户信息 [EN] Knowledge base user information
	CommonCheck
}

type KnowledgeUserInfo struct {
	UserId         string `json:"userId" validate:"required"`
	OrgId          string `json:"orgId" validate:"required"`
	PermissionId   string `json:"permissionId" form:"permissionId"` // 权限id编辑时传入 [EN] The permission id is passed in when editing
	PermissionType int    `json:"permissionType"`                   // 权限类型: -1 删除此用户权限；0: 查看权限; 10: 编辑权限; 20: 授权权限,数值不连续的原因防止后续有中间权限，目前逻辑 授权权限>编辑权限>查看权限 [EN] Permission type: -1 Delete this user permission; 0: View permission; 10: Edit permission; 20: Authorization permission, the reason of discontinuous values ​​prevents subsequent intermediate permissions, the current logic is Authorization permission>Edit permission>View permission
}
