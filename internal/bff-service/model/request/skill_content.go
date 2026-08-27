package request

import (
	"fmt"
)

// --- Skill 内容查询请求结构体 ---
// 与 workspace 系列的区别：这些接口只读，且已发布 skill 返回最新发布版本内容，
// 从未发布的 skill 回退返回草稿工作区内容。
// GET 接口使用 form tag（query parameter），POST 接口使用 json tag（request body）。

type GetSkillContentFilesReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
}

type GetSkillContentFileReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
	Path          string `form:"path"`                              // 文件路径（相对 skill 内容根目录）
}

// Check 校验获取内容文件树请求。
func (r *GetSkillContentFilesReq) Check() error {
	return nil
}

// Check 校验读取内容文件请求。
func (r *GetSkillContentFileReq) Check() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	return checkRelPath(r.Path)
}


