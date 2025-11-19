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
	PermissionType int    `json:"permissionType"` // 权限类型: -1 删除此用户权限；0: 查看权限; 10: 编辑权限; 20: 授权权限,数值不连续的原因防止后续有中间权限，目前逻辑 授权权限>编辑权限>查看权限 [EN] Permission type: -1 Delete this user permission; 0: View permission; 10: Edit permission; 20: Authorization permission, the reason of discontinuous values ​​prevents subsequent intermediate permissions, the current logic is Authorization permission>Edit permission>View permission
	PermissionId   string `json:"permissionId"`
	Transfer       bool   `json:"transfer"` //是否显示转让按钮 [EN] Whether to display the transfer button
}
