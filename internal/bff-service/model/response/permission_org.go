package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type OrgID struct {
	OrgID string `json:"orgId"`
}

type OrgInfo struct {
	OrgID     string         `json:"orgId"`
	Name      string         `json:"name"`
	Remark    string         `json:"remark"`
	CreatedAt string         `json:"createdAt"`
	Creator   IDName         `json:"creator"`
	Status    bool           `json:"status"`
	UserCount int64          `json:"userCount"`
	Admins    []string       `json:"admins"`
	Avatar    request.Avatar `json:"avatar"`
}

type AdminOrgTreeNode struct {
	OrgID    string              `json:"orgId"`
	Name     string              `json:"name"`
	Children []*AdminOrgTreeNode `json:"children"`
	HasPerm  bool                `json:"hasPerm"`
	IsSystem bool                `json:"isSystem"` // 是否系统（顶级）组织
}
