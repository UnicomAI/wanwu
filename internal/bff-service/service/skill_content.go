package service

import (
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	git_util "github.com/UnicomAI/wanwu/pkg/git-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

// skillContentSource 描述一次内容查询的数据来源。
type skillContentSource struct {
	source   string // response.SkillContentSourcePublished | SkillContentSourceDraft
	version  string // published 时的版本号（同时也是 git tag 名）
	skillDir string // published 时的 git 仓库根目录
}

func (s *skillContentSource) isPublished() bool {
	return s.source == response.SkillContentSourcePublished
}

func (s *skillContentSource) meta() response.SkillContentMeta {
	return response.SkillContentMeta{Source: s.source, Version: s.version}
}

// resolveSkillContentSource 判定本次查询应该读已发布版本还是草稿工作区。
//
// 判定以发布记录为准：发布过就读该版本的 git tag。存量 skill（git 工作区能力上线前
// 发布、之后才懒迁移建仓）没有历史 tag，此时回退读草稿——这类 skill 必定有工作区。
// 注意 tag 存在但读取失败不会回退：那属于异常，静默展示草稿会把草稿冒充成已发布内容。
func resolveSkillContentSource(ctx *gin.Context, customSkillID string) (*skillContentSource, error) {
	draft := &skillContentSource{source: response.SkillContentSourceDraft}

	publish, err := mcp.GetPublishCustomSkillByLatest(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillByLatestReq{
		SkillId: customSkillID,
	})
	if err != nil {
		return nil, err
	}
	version := publish.GetVersion()
	if version == "" {
		return draft, nil // 从未发布
	}

	ws, err := resolveSkillWorkspace(customSkillID) // 只读场景，不用 resolveInitializedSkillWorkspace 触发 git 初始化
	if err != nil {
		return nil, err
	}
	if !ws.repo.IsInitialized() {
		log.Warnf("[SkillContent] skill %s published %s but git repo not initialized, fallback to draft", customSkillID, version)
		return draft, nil
	}
	if err := git_util.ValidateTagName(version); err != nil {
		log.Warnf("[SkillContent] skill %s published version %s is not a valid tag name, fallback to draft: %v", customSkillID, version, err)
		return draft, nil
	}
	exists, err := ws.repo.TagExists(version)
	if err != nil {
		log.Errorf("[SkillContent] skill %s check tag %s err: %v", customSkillID, version, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_content_read_tree_failed")
	}
	if !exists {
		log.Warnf("[SkillContent] skill %s published %s has no git tag, fallback to draft", customSkillID, version)
		return draft, nil
	}
	return &skillContentSource{
		source:   response.SkillContentSourcePublished,
		version:  version,
		skillDir: ws.skillDir,
	}, nil
}

// GetSkillContentFiles 获取 Skill 内容文件树：已发布返回最新发布版本，未发布返回草稿工作区。
func GetSkillContentFiles(ctx *gin.Context, userId, orgId string, req request.GetSkillContentFilesReq) (*response.SkillContentFilesResp, error) {
	src, err := resolveSkillContentSource(ctx, req.CustomSkillID)
	if err != nil {
		return nil, err
	}
	if !src.isPublished() {
		draftResp, err := GetSkillWorkspaceFiles(ctx, userId, orgId, request.GetSkillWorkspaceFilesReq(req))
		if err != nil {
			return nil, err
		}
		return &response.SkillContentFilesResp{SkillContentMeta: src.meta(), Files: draftResp.Files}, nil
	}

	entries, err := listSkillTreeEntries(src.skillDir, src.version)
	if err != nil {
		return nil, err
	}
	commitTimeMs, err := git_util.GetTreeishCommitTimeMs(src.skillDir, src.version)
	if err != nil {
		// 时间戳拿不到不影响文件树可用性，降级为 0 并记录
		log.Errorf("[SkillContent] get commit time skill=%s version=%s err: %v", req.CustomSkillID, src.version, err)
		commitTimeMs = 0
	}
	return &response.SkillContentFilesResp{
		SkillContentMeta: src.meta(),
		Files:            buildTreeFromEntries(entries, commitTimeMs),
	}, nil
}

// GetSkillContentFile 读取 Skill 内容中的单个文件。
func GetSkillContentFile(ctx *gin.Context, userId, orgId string, req request.GetSkillContentFileReq) (*response.SkillContentFileResp, error) {
	src, err := resolveSkillContentSource(ctx, req.CustomSkillID)
	if err != nil {
		return nil, err
	}
	if !src.isPublished() {
		draftResp, err := GetSkillWorkspaceFile(ctx, userId, orgId, request.GetSkillWorkspaceFileReq(req))
		if err != nil {
			return nil, err
		}
		return &response.SkillContentFileResp{SkillContentMeta: src.meta(), SkillWorkspaceFileResp: *draftResp}, nil
	}

	blob, err := readSkillTreeBlob(src.skillDir, src.version, req.Path)
	if err != nil {
		return nil, err
	}
	if !blob.Exists {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_file_not_found")
	}
	if blob.IsTree {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_path_is_directory")
	}
	if blob.Content == nil { // ReadBlobAtTreeish 在超过上限时只返回 Size
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_file_too_large")
	}
	commitTimeMs, err := git_util.GetTreeishCommitTimeMs(src.skillDir, src.version)
	if err != nil {
		log.Errorf("[SkillContent] get commit time skill=%s version=%s err: %v", req.CustomSkillID, src.version, err)
		commitTimeMs = 0
	}
	return &response.SkillContentFileResp{
		SkillContentMeta: src.meta(),
		SkillWorkspaceFileResp: response.SkillWorkspaceFileResp{
			Content: string(blob.Content),
			Size:    blob.Size,
			ModTime: commitTimeMs,
		},
	}, nil
}


