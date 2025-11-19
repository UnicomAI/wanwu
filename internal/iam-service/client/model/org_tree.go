package model

import (
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/util"
)

type OrgNode struct {
	id       uint32
	parentID uint32
	roleID   uint32 // The current org built-in administrator role
	name     string
	status   bool
	parent   *OrgNode
	subs     []*OrgNode
}

type idName struct {
	ID   uint32
	Name string
}

func NewOrgTree(orgs []*Org, orgRoles []*OrgRole) (*OrgNode, error) {
	var nodes []*OrgNode
	for _, org := range orgs {
		// check exist
		for _, node := range nodes {
			if node.id == org.ID {
				return nil, fmt.Errorf("build org tree but org %v already exist", org.ID)
			}
		}
		// current
		currNode := &OrgNode{
			id:       org.ID,
			parentID: org.ParentID,
			name:     org.Name,
			status:   org.Status,
		}
		// parent
		for _, node := range nodes {
			if node.id == org.ParentID {
				currNode.parent = node
				node.subs = append(node.subs, currNode)
				break
			}
		}
		// subs
		for _, node := range nodes {
			if node.parentID == org.ID {
				node.parent = currNode
				currNode.subs = append(currNode.subs, node)
			}
		}
		nodes = append(nodes, currNode)
	}
	for _, orgRole := range orgRoles {
		if !orgRole.IsAdmin {
			continue
		}
		var exist bool
		for _, node := range nodes {
			if node.id == orgRole.OrgID {
				node.roleID = orgRole.RoleID
				exist = true
				break
			}
		}
		if !exist {
			return nil, fmt.Errorf("build org tree but org %v not exist", orgRole.OrgID)
		}
	}
	var head *OrgNode
	for _, node := range nodes {
		if node.parent == nil {
			if head != nil {
				return nil, fmt.Errorf("build org tree but top org %v and %v duplicate", node.id, head.id)
			}
			head = node
		}
	}
	return head, nil
}

func (n *OrgNode) GetOrg(orgID uint32) *OrgNode {
	return n.getOrg(orgID)
}

func (n *OrgNode) Select(userOrgs []*OrgUser, userRoles []*UserRole) []idName {
	var roleIDs []uint32
	for _, userRole := range userRoles {
		if userRole.IsAdmin {
			roleIDs = append(roleIDs, userRole.RoleID)
		}
	}
	var orgIDs []uint32
	for _, userOrg := range userOrgs {
		orgIDs = append(orgIDs, userOrg.OrgID)
	}
	var ret []idName
	n.sel(orgIDs, roleIDs, &ret)
	return ret
}

func (n *OrgNode) IsAdmin(orgID uint32, userRoles []*UserRole) bool {
	var roleIDs []uint32
	for _, userRole := range userRoles {
		if userRole.IsAdmin {
			roleIDs = append(roleIDs, userRole.RoleID)
		}
	}
	return n.getOrg(orgID).isAdmin(roleIDs)
}

func (n *OrgNode) GetFullName(orgID uint32) string {
	return n.getOrg(orgID).getFullName()
}

func (n *OrgNode) GetSubs(orgID uint32) []*OrgNode {
	org := n.getOrg(orgID)
	if org == nil {
		return nil
	}
	return org.subs
}

func (n *OrgNode) GetOrgID() uint32 {
	return n.id
}

// --- internal ---

func (n *OrgNode) sel(orgIDs, roleIDs []uint32, list *[]idName) {
	if n == nil || !n.status {
		return
	}
	if n.isAdmin(roleIDs) || util.Exist(orgIDs, n.id) {
		*list = append(*list, idName{ID: n.id, Name: n.getFullName()})
	}
	for _, org := range n.subs {
		org.sel(orgIDs, roleIDs, list)
	}
}

func (n *OrgNode) isAdmin(roleIDs []uint32) bool {
	if n == nil || !n.status {
		return false
	}
	if util.Exist(roleIDs, n.roleID) {
		return true
	}
	for n.parent != nil {
		if util.Exist(roleIDs, n.parent.roleID) {
			return true
		}
		n = n.parent
	}
	return false
}

func (n *OrgNode) getFullName() string {
	if n == nil {
		return ""
	}
	fullName := n.name
	for n.parent != nil && n.parent.parent != nil {
		fullName = n.parent.name + " - " + fullName
		n = n.parent
	}
	return fullName
}

func (n *OrgNode) getOrg(orgID uint32) *OrgNode {
	if n == nil {
		return nil
	}
	if n.id == orgID {
		return n
	}
	for _, node := range n.subs {
		if ret := node.getOrg(orgID); ret != nil {
			return ret
		}
	}
	return nil
}
