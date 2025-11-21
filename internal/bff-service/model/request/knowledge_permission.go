package request

type KnowledgeOrgSelectReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId" validate:"required"` //knowledge base id
	CommonCheck
}

type KnowledgeUserSelectReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId" validate:"required"` //knowledge base id
	CommonCheck
}

type KnowledgeUserNoPermitSelectReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId" validate:"required"` //knowledge base id
	OrgId       string `json:"orgId" form:"orgId" validate:"required"`             //Select organization id
	Transfer    bool   `json:"transfer" form:"transfer"`                           //Is it a transfer list?
	CommonCheck
}

type KnowledgeUserAddReq struct {
	KnowledgeId       string               `json:"knowledgeId" validate:"required"`       // knowledge base id
	PermissionType    int                  `json:"permissionType"`                        // Permission type: 0: View permission; 10: Edit permission; 20: Authorization permission. The reason for discontinuous values ​​prevents intermediate permissions in the future. The current logic is Authorization permission > Edit permission > View permission.
	KnowledgeUserList []*KnowledgeUserInfo `json:"knowledgeUserList" validate:"required"` // Knowledge base user information
	CommonCheck
}

type KnowledgeUserEditReq struct {
	KnowledgeId   string             `json:"knowledgeId" validate:"required"`   // knowledge base id
	KnowledgeUser *KnowledgeUserInfo `json:"knowledgeUser" validate:"required"` // Knowledge base user information
	CommonCheck
}

type KnowledgeUserDeleteReq struct {
	KnowledgeId  string `json:"knowledgeId" validate:"required"`  // knowledge base id
	PermissionId string `json:"permissionId" validate:"required"` // Knowledge base user information permission information
	CommonCheck
}

type KnowledgeTransferUserAdminReq struct {
	KnowledgeId   string             `json:"knowledgeId" validate:"required"`   // knowledge base id
	PermissionId  string             `json:"permissionId" form:"permissionId"`  // The permission id is passed in when editing
	KnowledgeUser *KnowledgeUserInfo `json:"knowledgeUser" validate:"required"` // Knowledge base user information
	CommonCheck
}

type KnowledgeUserInfo struct {
	UserId         string `json:"userId" validate:"required"`
	OrgId          string `json:"orgId" validate:"required"`
	PermissionId   string `json:"permissionId" form:"permissionId"` // The permission id is passed in when editing
	PermissionType int    `json:"permissionType"`                   // Permission type: -1 Delete this user permission; 0: View permission; 10: Edit permission; 20: Authorization permission, the reason of discontinuous values ​​prevents subsequent intermediate permissions, the current logic is Authorization permission>Edit permission>View permission
}
