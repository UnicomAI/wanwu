package model

import (
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/util"
)

type OrgNode struct {
	id       uint32
	parentID uint32
	roleID   uint32 // 当前org内置管理员角色
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

// GetOrgName 返回组织在数据库中存储的原始名（不含上级前缀）
func (n *OrgNode) GetOrgName(orgID uint32) string {
	if org := n.getOrg(orgID); org != nil {
		return org.name
	}
	return ""
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

func (n *OrgNode) GetFirstClassOrg() *OrgNode {
	if n == nil {
		return nil
	}
	curr := n
	for curr.parent != nil && curr.parent.parent != nil {
		curr = curr.parent
	}
	return curr
}

// GetAncestorIDs 返回从指定组织到根的所有祖先ID（不含自身，不含根节点）
func (n *OrgNode) GetAncestorIDs(orgID uint32) []uint32 {
	node := n.getOrg(orgID)
	if node == nil {
		return nil
	}
	var ancestors []uint32
	for node.parent != nil && node.parent.parent != nil {
		ancestors = append(ancestors, node.parent.id)
		node = node.parent
	}
	return ancestors
}

// CollectDescendants 收集指定组织及其所有后代的ID到集合中
func (n *OrgNode) CollectDescendants(orgID uint32, ids map[uint32]bool) {
	node := n.getOrg(orgID)
	if node == nil {
		return
	}
	collectDescendantsFromNode(node, ids)
}

func collectDescendantsFromNode(node *OrgNode, ids map[uint32]bool) {
	if node == nil {
		return
	}
	ids[node.id] = true
	for _, sub := range node.subs {
		collectDescendantsFromNode(sub, ids)
	}
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
