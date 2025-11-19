package response

type KnowledgeUserPermissionResp struct {
	KnowledgeUserInfoList []*KnowledgeUserInfo `json:"knowledgeUserInfoList"`
}

type KnowOrgInfoResp struct {
	KnowOrgInfoList []*KnowOrgInfo `json:"knowOrgInfoList"`
}

type KnowOrgInfo struct {
	OrgId   string `json:"orgId"`
	OrgName string `json:"orgName"`
}

type KnowOrgUserInfoResp struct {
	OrgId        string          `json:"orgId"`
	OrgName      string          `json:"orgName"`
	UserInfoList []*KnowUserInfo `json:"userInfoList"`
}

type KnowUserInfo struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
}

type KnowledgeUserInfo struct {
	UserId         string `json:"userId"`
	UserName       string `json:"userName"`
	OrgId          string `json:"orgId"`
	OrgName        string `json:"orgName"`
	PermissionType int    `json:"permissionType"` // Permission type: -1 Delete this user permission; 0: View permission; 10: Edit permission; 20: Authorization permission, the reason of discontinuous values ​​prevents subsequent intermediate permissions, the current logic is Authorization permission>Edit permission>View permission
	PermissionId   string `json:"permissionId"`
	Transfer       bool   `json:"transfer"` //Whether to display the transfer button
}
