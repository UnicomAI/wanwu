package service

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	git_util "github.com/UnicomAI/wanwu/pkg/git-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	customSkillFileType = ".zip"
)

var skillBiz = &SkillBiz{}

type SkillBiz struct {
}

func init() {
	InitBizService(skillBiz)
}
func (*SkillBiz) BizType() string {
	return constant.BizModuleResourceSkill
}
func (*SkillBiz) SearchBizOwner(ctx *gin.Context, bizId string) (userId, orgId string, err error) {
	_, exist := config.Cfg().AgentSkill(bizId)
	if exist {
		return getUserID(ctx), getOrgID(ctx), nil
	}
	skill, _ := GetAcquiredSkill(ctx, "", "", bizId)
	if skill != nil {
		userId, orgId = skill.UserId, skill.OrgId
		return
	}
	publish, err := mcp.CustomSkillGet(ctx, &mcp_service.CustomSkillGetReq{
		SkillId: bizId,
	})
	if err != nil {
		return "", "", err
	}
	userId, orgId = publish.Skill.UserId, publish.Skill.OrgId
	return
}

func GetCustomSkill(ctx *gin.Context, userId, orgId, skillId string) (*response.PublishedSkillDetail, error) {
	publish, err := mcp.CustomSkillGet(ctx.Request.Context(), &mcp_service.CustomSkillGetReq{
		SkillId: skillId,
	})
	if err != nil {
		return nil, err
	}
	skill := publish.GetSkill()
	version := publish.GetVersion()
	if skill == nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_custom_not_found")
	}
	variables, err := getCustomSkillVariables(ctx, skill.SkillId)
	if err != nil {
		return nil, err
	}

	publishType := ""
	appResp, err := app.GetAppListByIds(ctx.Request.Context(), &app_service.GetAppListByIdsReq{
		AppIdsList: []string{skillId},
		AppType:    constant.AppTypeSkill,
	})
	if err != nil {
		log.Errorf("GetCustomSkill app list failed, skillId=%s err=%v", skillId, err)
	}
	if appResp != nil && len(appResp.Infos) > 0 {
		publishType = appResp.Infos[0].PublishType
	}

	// 判断是否已发布
	isPublished := publishType != "" || version != ""
	if publishType == "" {
		publishType = constant.AppPublishPrivate
	}

	detail := toPublishedSkillDetail(ctx, skill, variables, publish, isPublished, publishType, version)

	return detail, nil
}

func ensureLegacyCustomSkillWorkspace(ctx *gin.Context, skill *mcp_service.CustomSkill) error {
	if skill == nil {
		return nil
	}
	zipURL := strings.TrimSpace(skill.ObjectPath)
	if zipURL == "" {
		return nil
	}

	workspaceExists, workspaceDir, err := customSkillOverwriteWorkspaceExists(skill.SkillId)
	if err != nil {
		log.Errorf("[wga-skill-legacy] skill %v check overwrite workspace err: %v", skill.SkillId, err)
		return err
	}
	if workspaceExists {
		log.Infof("[wga-skill-legacy] skill %v overwrite workspace exists, workspaceDir=%s, skip import", skill.SkillId, workspaceDir)
		return nil
	}

	log.Infof("[wga-skill-legacy] skill %v overwrite workspace not found, expectedWorkspaceDir=%s, start import from objectPath", skill.SkillId, workspaceDir)
	if err := importLegacyCustomSkillWorkspace(ctx, skill.SkillId, zipURL); err != nil {
		log.Errorf("[wga-skill-legacy] skill %v import overwrite workspace err: %v", skill.SkillId, err)
		return err
	}
	workspaceExists, workspaceDir, err = customSkillOverwriteWorkspaceExists(skill.SkillId)
	if err != nil {
		log.Errorf("[wga-skill-legacy] skill %v recheck overwrite workspace err: %v", skill.SkillId, err)
		return err
	}
	if !workspaceExists {
		log.Errorf("[wga-skill-legacy] skill %v import finished but overwrite workspace still not found, expectedWorkspaceDir=%s", skill.SkillId, workspaceDir)
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill workspace import finished but workspace not found: %s", skill.SkillId))
	}
	log.Infof("[wga-skill-legacy] skill %v import overwrite workspace success, workspaceDir=%s", skill.SkillId, workspaceDir)
	return nil
}

func customSkillOverwriteWorkspaceExists(skillId string) (bool, string, error) {
	store, err := NewGeneralAgentSkillWorkspaceStore(skillId)
	if err != nil {
		return false, "", err
	}
	ok, info, err := store.GetLastRunDir()
	if err != nil {
		return false, info.Dir, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if !ok {
		return false, info.Dir, nil
	}
	stat, err := os.Stat(info.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, info.Dir, nil
		}
		return false, info.Dir, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("stat custom skill workspace %s err: %v", info.Dir, err))
	}
	if !stat.IsDir() {
		return false, info.Dir, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill workspace is not a directory: %s", info.Dir))
	}
	return true, info.Dir, nil
}

func importLegacyCustomSkillWorkspace(ctx *gin.Context, skillId, zipURL string) error {
	store, err := NewGeneralAgentSkillWorkspaceStore(skillId)
	if err != nil {
		return err
	}
	dirs, err := PrepareWgaWorkspaceDirs(store, util.GenUUID(), true)
	if err != nil {
		return err
	}
	log.Infof("[wga-skill-legacy] skill %v prepared overwrite workspace, outputDir=%s", skillId, dirs.OutputDir)
	skillDir := filepath.Join(dirs.OutputDir, generalAgentWorkspaceSkillDirName)
	data, err := downloadCustomSkillZip(ctx, zipURL)
	if err != nil {
		_ = CleanupWgaWorkspace(store)
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("download skill zip err: %v", err))
	}
	log.Infof("[wga-skill-legacy] skill %v downloaded legacy skill zip, size=%d, skillDir=%s", skillId, len(data), skillDir)
	if _, err := importSkillDataIntoWorkspace(data, skillDir); err != nil {
		_ = CleanupWgaWorkspace(store)
		return err
	}
	return nil
}

func rollbackCustomSkillWorkspace(ctx *gin.Context, skillId, version string) error {
	objectPath, err := getCustomSkillVersionObjectPath(ctx, skillId, version)
	if err != nil {
		return err
	}
	data, err := downloadCustomSkillZip(ctx, objectPath)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("download skill version zip err: %v", err))
	}
	return overwriteCustomSkillWorkspaceFromZip(ctx, skillId, data)
}

func getCustomSkillVersionObjectPath(ctx *gin.Context, skillId, version string) (string, error) {
	publish, err := mcp.GetPublishCustomSkillByVersion(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillByVersionReq{
		SkillId: skillId,
		Version: version,
	})
	if err != nil {
		return "", err
	}
	if publish.GetObjectPath() == "" {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "skill_version_package_not_available", "skill version package objectPath is empty")
	}
	return publish.GetObjectPath(), nil
}

func overwriteCustomSkillWorkspaceFromZip(ctx *gin.Context, skillId string, data []byte) error {
	store, err := NewGeneralAgentSkillWorkspaceStore(skillId)
	if err != nil {
		return err
	}
	dirs, err := PrepareWgaWorkspaceDirs(store, util.GenUUID(), true)
	if err != nil {
		return err
	}
	skillRoot := dirs.OutputDir
	skillDir := filepath.Join(skillRoot, generalAgentWorkspaceSkillDirName)
	stagingDir := filepath.Join(skillRoot, ".skill-rollback-"+util.GenUUID())
	defer func() { _ = os.RemoveAll(stagingDir) }()

	if _, err := importSkillDataIntoWorkspace(data, stagingDir); err != nil {
		return err
	}
	return replaceCustomSkillDirWithGitGuard(skillRoot, skillDir, stagingDir)
}

// replaceCustomSkillDirWithGitGuard 在持有仓库锁时替换 skill 目录并保护 Git 元数据。
func replaceCustomSkillDirWithGitGuard(skillRoot, skillDir, stagingDir string) error {
	repo := git_util.Open(skillRoot)
	mu := repo.GetMutex()
	mu.Lock()
	defer mu.Unlock()

	gitInitialized := repo.IsInitialized()
	if err := replaceCustomSkillDir(skillDir, stagingDir); err != nil {
		return err
	}

	if gitInitialized {
		if ok, err := git_util.HasGitMetadata(skillRoot); err != nil {
			return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("skill workspace git metadata missing after rollback: %v", err))
		} else if !ok {
			return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("%s not found", filepath.Join(skillRoot, ".git")))
		}
	}
	return nil
}

// replaceCustomSkillDir 使用备份目录替换当前 skill 目录。
func replaceCustomSkillDir(skillDir, stagingDir string) error {
	if ok, err := git_util.HasGitMetadata(skillDir); err != nil {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("check skill workspace git metadata err: %v", err))
	} else if ok {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("refuse to replace git repository directory: %s", skillDir))
	}

	backupDir := filepath.Join(filepath.Dir(skillDir), ".skill-rollback-backup-"+util.GenUUID())
	backupCreated := false
	defer func() { _ = os.RemoveAll(backupDir) }()

	if _, err := os.Stat(skillDir); err == nil {
		if err := os.Rename(skillDir, backupDir); err != nil {
			return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("backup current skill workspace err: %v", err))
		}
		backupCreated = true
	} else if !os.IsNotExist(err) {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("stat current skill workspace err: %v", err))
	}

	if err := os.Rename(stagingDir, skillDir); err != nil {
		if backupCreated {
			_ = os.Rename(backupDir, skillDir)
		}
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("replace skill workspace err: %v", err))
	}
	return nil
}

func downloadCustomSkillZip(ctx *gin.Context, zipURL string) ([]byte, error) {
	zipURL = strings.TrimSpace(zipURL)
	lowerURL := strings.ToLower(zipURL)
	if strings.HasPrefix(lowerURL, "http://") || strings.HasPrefix(lowerURL, "https://") {
		if data, err := minio_util.DownloadFileDirect(ctx.Request.Context(), zipURL); err == nil {
			return data, nil
		}
		data, _, err := minio_util.DownloadFile(ctx.Request.Context(), zipURL)
		return data, err
	}
	endpoint := strings.TrimSpace(config.Cfg().Minio.Endpoint)
	if endpoint == "" {
		data, _, err := minio_util.DownloadFile(ctx.Request.Context(), zipURL)
		return data, err
	}
	if strings.HasPrefix(strings.ToLower(endpoint), "http://") || strings.HasPrefix(strings.ToLower(endpoint), "https://") {
		endpoint = strings.TrimRight(endpoint, "/") + "/" + strings.TrimLeft(zipURL, "/")
	} else {
		endpoint = "http://" + strings.TrimRight(endpoint, "/") + "/" + strings.TrimLeft(zipURL, "/")
	}
	data, _, err := minio_util.DownloadFile(ctx.Request.Context(), endpoint)
	return data, err
}

func DeleteCustomSkill(ctx *gin.Context, userId, orgId, skillId string) error {
	publish, err := mcp.CustomSkillGet(ctx.Request.Context(), &mcp_service.CustomSkillGetReq{
		SkillId: skillId,
	})
	if err != nil {
		return err
	}
	skill := publish.GetSkill()
	if skill == nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_custom_not_found")
	}

	if _, err := app.DeleteApp(ctx.Request.Context(), &app_service.DeleteAppReq{
		AppId:   skillId,
		AppType: constant.AppTypeSkill,
		UserId:  userId,
		OrgId:   orgId,
	}); err != nil {
		return err
	}
	if skill.WgaThreadId != "" {
		if err := deleteGeneralAgentConversationForSkillDelete(ctx, userId, orgId, request.DeleteGeneralAgentConversationReq{ThreadID: skill.WgaThreadId}); err != nil {
			return err
		}
	}
	if skill.PreviewThreadId != "" {
		if err := deleteWgaConversationHistory(ctx, userId, orgId, skill.PreviewThreadId); err != nil {
			return err
		}
	}

	if _, err := mcp.CustomSkillDelete(ctx.Request.Context(), &mcp_service.CustomSkillDeleteReq{
		SkillId: skillId,
		Identity: &mcp_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	}); err != nil {
		return err
	}
	return cleanupCustomSkillWorkspace(skill.SkillId)
}

func GetCustomSkillList(ctx *gin.Context, userId, orgId, name string) (*response.ListResult, error) {
	resp, err := mcp.CustomSkillGetList(ctx.Request.Context(), &mcp_service.CustomSkillGetListReq{
		Name:     name,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}

	customSkillList := make([]*response.PublishedSkillInfo, 0, len(resp.List))
	for _, item := range resp.List {
		customSkillList = append(customSkillList, toCustomSkillListItem(ctx, item))
	}

	fillCustomSkillPublishInfo(ctx, customSkillList)

	return &response.ListResult{
		List:  customSkillList,
		Total: resp.Total,
	}, nil
}

func fillCustomSkillPublishInfo(ctx *gin.Context, skillList []*response.PublishedSkillInfo) {
	if len(skillList) == 0 {
		return
	}
	skillIds := make([]string, 0, len(skillList))
	for _, skill := range skillList {
		if skill != nil && skill.SkillId != "" {
			skillIds = append(skillIds, skill.SkillId)
		}
	}
	appResp, err := app.GetAppListByIds(ctx.Request.Context(), &app_service.GetAppListByIdsReq{
		AppIdsList: skillIds,
		AppType:    constant.AppTypeSkill,
	})
	if err != nil {
		log.Errorf("fillCustomSkillPublishInfo app list failed, err=%v", err)
	}
	publishTypeBySkill := make(map[string]string)
	if appResp != nil {
		for _, appInfo := range appResp.Infos {
			publishTypeBySkill[appInfo.AppId] = appInfo.PublishType
		}
	}
	publishBySkill, err := getCustomSkillPublishMap(ctx, skillIds)
	if err != nil {
		log.Errorf("fillCustomSkillPublishInfo publish map failed, err=%v", err)
	}
	for _, skill := range skillList {
		if skill == nil {
			continue
		}
		publishType := publishTypeBySkill[skill.SkillId]
		version := ""
		if publish := publishBySkill[skill.SkillId]; publish != nil {
			version = publish.GetVersion()
		}
		skill.IsPublished = publishType != "" || version != ""
		if publishType == "" {
			publishType = constant.AppPublishPrivate
		}
		skill.PublishType = publishType
		skill.Version = version
	}
}

// GetCustomSkillListDetail 返回一批 custom skill 的详情（含变量）。
// userId/orgId 是 assistant-service callback 传入的智能体（Assistant）创建者身份，
// 不是发起 HTTP 请求的调用者身份——发布态智能体常被非创建者调用，二者不一定一致。
// 当前 var 查询走 mcp 层按 skill 记录自带的 owner 隐式解析，这两个参数不改变返回内容，
// 仅用于日志/审计，便于排查发布态智能体被非创建者调用时的问题。
func GetCustomSkillListDetail(ctx *gin.Context, userId, orgId string, skillIdList []string) (*response.CustomSkillDetailListResp, error) {
	skillIdList = uniqueSkillIDs(skillIdList)
	log.Debugf("callback custom skill list, assistantCreator=%s/%s, skillIds=%v", orgId, userId, skillIdList)
	publishBySkill, err := getCustomSkillPublishMap(ctx, skillIdList)
	if err != nil {
		return nil, err
	}

	skillDetailList := make([]*response.CustomSkillListDetail, 0, len(skillIdList))
	for _, skillId := range skillIdList {
		publish := publishBySkill[skillId]
		if publish == nil || publish.GetSkill() == nil {
			log.Warnf("callback custom skill list ignored unpublished skill, skillId: %s", skillId)
			continue
		}
		if publish.GetObjectPath() == "" {
			log.Warnf("callback custom skill list ignored skill with empty latest publish objectPath, skillId: %s", skillId)
			continue
		}
		skill := publish.GetSkill()
		detail := &response.CustomSkillListDetail{
			SkillBasicInfo: response.SkillBasicInfo{
				SkillId: skill.GetSkillId(),
				Name:    skill.GetName(),
				Avatar:  cacheSkillAvatar(ctx, skill.GetAvatar()),
				Author:  skill.GetAuthor(),
				Desc:    skill.GetDesc(),
			},
			ObjectPath: publish.GetObjectPath(),
		}
		// 通过 mcp gRPC 拉取该 custom skill 的变量集（per-request，无缓存）。
		// 失败只打日志、不中断 callback：让 LLM 拿到无 var 的 skill 也比直接拒绝整批好。
		// fail-closed 风格：err != nil 时不赋 Variables，与 skill_acquired / skill_builtin 对齐。
		vars, varsErr := getCustomSkillVariables(ctx, skillId)
		if varsErr != nil {
			log.Warnf("callback custom skill list failed to fetch variables, skillId: %s, err: %v", skillId, varsErr)
		} else {
			detail.Variables = toSkillVariables(vars)
		}
		skillDetailList = append(skillDetailList, detail)
	}

	return &response.CustomSkillDetailListResp{SkillList: skillDetailList}, nil
}

func uniqueSkillIDs(skillIdList []string) []string {
	seen := make(map[string]struct{}, len(skillIdList))
	result := make([]string, 0, len(skillIdList))
	for _, skillId := range skillIdList {
		if _, ok := seen[skillId]; ok {
			continue
		}
		seen[skillId] = struct{}{}
		result = append(result, skillId)
	}
	return result
}

// buildCustomSkillPublishPackage 从 HEAD:skill 打包发布 zip。
// 只有包校验和上传都成功后，才创建发布版本 tag。
func buildCustomSkillPublishPackage(ctx *gin.Context, userId, orgId, skillId, version string) (string, string, func(), error) {
	ws, err := resolveSkillWorkspace(skillId)
	if err != nil {
		return "", "", nil, err
	}
	if _, err := ensureGitInitializedAt(skillId, ws.skillDir); err != nil {
		return "", "", nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_init_git_failed")
	}

	if !ws.repo.HasHead() {
		return "", "", nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_uncommitted_changes", "工作区尚未提交，请先提交后再发布")
	}

	hasChanges, err := ws.repo.HasChangesInSubDir(generalAgentWorkspaceSkillDirName)
	if err != nil {
		return "", "", nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_check_git_status_failed")
	}
	if hasChanges {
		return "", "", nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_uncommitted_changes", "工作区有未提交的更改，请先提交后再发布")
	}

	tagExists, err := ws.repo.TagExists(version)
	if err != nil {
		return "", "", nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("check git tag failed for skill %s: %v", skillId, err))
	}
	if tagExists {
		return "", "", nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_version_exists", fmt.Sprintf("版本 %s 已存在，请使用其他版本号", version))
	}

	zipBytes, err := ws.repo.ArchivePath("HEAD", generalAgentWorkspaceSkillDirName)
	if err != nil {
		return "", "", nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("archive git HEAD:%s failed for skill %s: %v", generalAgentWorkspaceSkillDirName, skillId, err))
	}

	markdown, _, err := util.ExtractSkillMarkdownFromZip(zipBytes)
	if err != nil {
		return "", "", nil, grpc_util.ErrorStatus(errs.Code_BFFSkillParse, err.Error())
	}

	fileName, _, err := minio.UploadFileCommon(ctx.Request.Context(), bytes.NewReader(zipBytes), customSkillFileType, int64(len(zipBytes)), true)
	if err != nil {
		return "", "", nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("upload custom skill %s zip err: %v", skillId, err))
	}

	commitHash, err := ws.repo.CreateTag(version)
	if err != nil {
		return "", "", nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("create git tag %s failed for skill %s: %v", version, skillId, err))
	}
	log.Infof("[skill-publish] created tag %s for skill %s at commit %s", version, skillId, commitHash)
	rollback := func() {
		if err := ws.repo.DeleteTag(version); err != nil {
			log.Warnf("[skill-publish] rollback tag %s for skill %s failed: %v", version, skillId, err)
		}
	}
	return path.Join(minio.BucketFileUpload, minio.DirFileNotExpire, fileName), markdown, rollback, nil
}

func findFirstCustomSkillDir(skillId string) (string, error) {
	store, err := NewGeneralAgentSkillWorkspaceStore(skillId)
	if err != nil {
		return "", err
	}
	workspaceDir := GetWgaWorkspaceThreadDir(store)
	if workspaceDir == "" {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill workspace not found: %s", skillId))
	}
	stat, err := os.Stat(workspaceDir)
	if err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill workspace not found: %s", skillId))
	}
	if !stat.IsDir() {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill workspace is not a directory: %s", workspaceDir))
	}
	return findFirstCustomSkillDirInWorkspace(workspaceDir, skillId)
}

func findFirstCustomSkillDirInWorkspace(workspaceDir, skillId string) (string, error) {
	var skillDir string
	rootSkillFound := false
	err := filepath.WalkDir(workspaceDir, func(currentPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if currentPath == workspaceDir {
				return nil
			}
			name := entry.Name()
			if strings.HasPrefix(name, ".") || name == "tmp" || name == "input" || name == "output" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != "SKILL.md" {
			return nil
		}
		parentDir := filepath.Dir(currentPath)
		if parentDir == workspaceDir {
			rootSkillFound = true
			return filepath.SkipAll
		}
		skillDir = parentDir
		return filepath.SkipAll
	})
	if err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to scan custom skill workspace %s: %v", workspaceDir, err))
	}
	if skillDir != "" {
		return skillDir, nil
	}
	if rootSkillFound {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill SKILL.md is at workspace root, refuse to zip workspace: %s", skillId))
	}
	return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill SKILL.md not found in workspace: %s", skillId))
}

func ensureNoSymlink(dir string) error {
	return filepath.WalkDir(dir, func(currentPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink is not supported: %s", currentPath)
		}
		return nil
	})
}

func toCustomSkillListItem(ctx *gin.Context, publish *mcp_service.PublishCustomSkill) *response.PublishedSkillInfo {
	skill := publish.GetSkill()
	if skill == nil {
		return nil
	}

	return &response.PublishedSkillInfo{
		SkillBasicInfo: response.SkillBasicInfo{
			SkillId:   skill.SkillId,
			Name:      skill.Name,
			Avatar:    cacheSkillAvatar(ctx, skill.Avatar),
			Author:    skill.Author,
			Desc:      skill.Desc,
			CreatedAt: util.Time2Str(skill.CreatedAt),
			UpdatedAt: util.Time2Str(skill.UpdatedAt),
			UserId:    skill.UserId,
			OrgId:     skill.OrgId,
		},
		ThreadID:  skill.WgaThreadId,
		PreviewID: skill.PreviewThreadId,
	}
}

func toPublishedSkillDetail(ctx *gin.Context, skill *mcp_service.CustomSkill, variables []*mcp_service.Variable, publish *mcp_service.PublishCustomSkill, isPublished bool, publishType, version string) *response.PublishedSkillDetail {
	return &response.PublishedSkillDetail{
		PublishedSkillInfo: response.PublishedSkillInfo{
			SkillBasicInfo: response.SkillBasicInfo{
				SkillId:   skill.SkillId,
				Name:      skill.Name,
				Avatar:    cacheSkillAvatar(ctx, skill.Avatar),
				Author:    skill.Author,
				Desc:      skill.Desc,
				CreatedAt: util.Time2Str(skill.CreatedAt),
				UpdatedAt: util.Time2Str(skill.UpdatedAt),
				OrgId:     skill.OrgId,
				UserId:    skill.UserId,
			},
			ThreadID:    skill.WgaThreadId,
			PreviewID:   skill.PreviewThreadId,
			IsPublished: isPublished,
			PublishType: publishType,
			Version:     version,
		},
		Variables:     toSkillVariables(variables),
		SkillMarkdown: config.FixFrontMatterFormat(publish.GetMarkdown()),
	}
}

func DownloadCustomSkillVersion(ctx *gin.Context, userId, orgId, skillId, version string) ([]byte, error) {
	objectPath, err := getCustomSkillVersionObjectPath(ctx, skillId, version)
	if err != nil {
		return nil, err
	}
	return downloadCustomSkillZip(ctx, objectPath)
}

func CheckCustomSkill(ctx *gin.Context, userId, orgId, zipUrl string) (*response.CustomSkillCheckResp, error) {
	// 下载文件
	data, err := minio_util.DownloadFileDirect(ctx.Request.Context(), zipUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("download skill zip err: %v", err))
	}

	// 解压并查找SKILL.md文件，验证zip包是否有效
	_, fm, err := util.ExtractSkillMarkdownFromZip(data)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFSkillParse, err.Error())
	}

	return &response.CustomSkillCheckResp{
		Name: fm.Name,
		Desc: fm.Description,
	}, nil
}

func GetSkillSelect(ctx *gin.Context, userId, orgId, name, skillType string, builtInOnlyOntology bool) (*response.ListResult, error) {
	var allSkills []*response.SkillInfo

	// 内建 skills
	if skillType == "" || skillType == constant.SkillTypeBuiltIn {
		for _, skillsCfg := range config.Cfg().AgentSkills {
			if name != "" && !strings.Contains(skillsCfg.Name, name) {
				continue
			}
			if builtInOnlyOntology {
				// 本体智能体平台配置数字员工场景
				if !util.Exist(config.Cfg().Ontology.BuiltinSkills, skillsCfg.SkillId) {
					// 只展示在 ontology 中配置的内建技能
					continue
				}
			} else {
				// 智能体、工作流配置Skills场景
				if util.Exist(config.Cfg().Ontology.BuiltinSkills, skillsCfg.SkillId) {
					// 不展示在 ontology 中配置的内建技能
					continue
				}
			}
			iconUrl := config.Cfg().DefaultIcon.SkillIcon
			if skillsCfg.Avatar != "" {
				iconUrl = skillsCfg.Avatar
			}
			allSkills = append(allSkills, &response.SkillInfo{
				SkillBasicInfo: response.SkillBasicInfo{
					SkillId: skillsCfg.SkillId,
					Desc:    skillsCfg.Desc,
					Author:  skillsCfg.Author,
					Avatar:  request.Avatar{Path: iconUrl},
				},
				SkillName: skillsCfg.Name,
				SkillType: constant.SkillTypeBuiltIn,
			})
		}
	}

	// 自定义 skills（只返回已发布的）
	if skillType == "" || skillType == constant.SkillTypeCustom {
		customResp, err := mcp.GetPublishCustomSkillList(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillListReq{
			Name:     name,
			Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
		})
		if err != nil {
			return nil, err
		}
		for _, item := range customResp.List {
			skill := item.GetSkill()
			allSkills = append(allSkills, &response.SkillInfo{
				SkillBasicInfo: response.SkillBasicInfo{
					SkillId: skill.GetSkillId(),
					Desc:    skill.GetDesc(),
					Author:  skill.GetAuthor(),
					Avatar:  cacheSkillAvatar(ctx, skill.GetAvatar()),
				},
				SkillName: skill.GetName(),
				SkillType: constant.SkillTypeCustom,
			})
		}
	}

	// acquired skills
	if skillType == "" || skillType == constant.SkillTypeAcquired {
		acquiredResp, err := mcp.AcquiredSkillGetList(ctx.Request.Context(), &mcp_service.AcquiredSkillGetListReq{
			Name:     name,
			Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
		})
		if err != nil {
			return nil, err
		}

		// 批量获取源 Skill 的发布信息用于权限校验
		sourceSkillIds := make([]string, 0, len(acquiredResp.List))
		for _, acquired := range acquiredResp.List {
			if acquired != nil && acquired.GetSkill().GetSkill().GetSkillId() != "" {
				sourceSkillIds = append(sourceSkillIds, acquired.GetSkill().GetSkill().GetSkillId())
			}
		}
		appInfoMap, err := getSourceSkillAppInfoMap(ctx, sourceSkillIds)
		if err != nil {
			return nil, err
		}

		for _, acquired := range acquiredResp.List {
			if isAcquiredSkillAccessible(ctx, acquired, appInfoMap) {
				skill := acquired.GetSkill().GetSkill()
				if skill == nil || strings.TrimSpace(skill.GetName()) == "" || strings.TrimSpace(skill.GetDesc()) == "" {
					continue
				}
				allSkills = append(allSkills, &response.SkillInfo{
					SkillBasicInfo: response.SkillBasicInfo{
						SkillId: acquired.AcquiredSkillId,
						Name:    skill.GetName(),
						Desc:    skill.GetDesc(),
						Author:  skill.GetAuthor(),
						Avatar:  cacheSkillAvatar(ctx, skill.GetAvatar()),
					},
					SkillName: skill.GetName(),
					SkillType: constant.SkillTypeAcquired,
				})
			}
		}
	}

	return &response.ListResult{
		List:  allSkills,
		Total: int64(len(allSkills)),
	}, nil
}

func getCustomSkillVariables(ctx *gin.Context, skillId string) ([]*mcp_service.Variable, error) {
	varResp, err := mcp.GetCustomSkillVars(ctx.Request.Context(), &mcp_service.GetCustomSkillVarsReq{
		SkillId: skillId,
	})
	if err != nil {
		return nil, err
	}
	if varResp == nil {
		return nil, nil
	}
	return varResp.GetVariables(), nil
}

func getAcquiredSkillVariables(ctx *gin.Context, acquiredSkillId string) ([]*mcp_service.Variable, error) {
	varResp, err := mcp.GetAcquiredSkillVars(ctx.Request.Context(), &mcp_service.GetAcquiredSkillVarsReq{
		AcquiredSkillId: acquiredSkillId,
	})
	if err != nil {
		return nil, err
	}
	if varResp == nil {
		return nil, nil
	}
	return varResp.GetVariables(), nil
}

// getCustomSkillPublishMap 批量获取自定义 skill 的发布信息（从 mcp-service）
func getCustomSkillPublishMap(ctx *gin.Context, skillIds []string) (map[string]*mcp_service.PublishCustomSkill, error) {
	ret := make(map[string]*mcp_service.PublishCustomSkill, len(skillIds))
	if len(skillIds) == 0 {
		return ret, nil
	}
	resp, err := mcp.GetPublishCustomSkillByIDList(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillByIDListReq{
		SkillIdList: skillIds,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return ret, nil
	}
	for _, publish := range resp.GetList() {
		if skill := publish.GetSkill(); skill != nil && skill.GetSkillId() != "" {
			ret[skill.GetSkillId()] = publish
		}
	}
	return ret, nil
}

func deleteWgaConversationHistory(ctx *gin.Context, userId, orgId, threadId string) error {
	threadId = strings.TrimSpace(threadId)
	if threadId == "" {
		return nil
	}
	_, err := assistant.DeleteFromES(ctx.Request.Context(), &assistant_service.DeleteFromESReq{
		IndexName: wgaConversationHistoryEventESIndexName,
		Conditions: map[string]string{
			"threadId": threadId,
			"userId":   userId,
			"orgId":    orgId,
		},
	})
	if err != nil && !esIndexNotFound(err, wgaConversationHistoryEventESIndexName) {
		return err
	}
	return nil
}

func cleanupCustomSkillWorkspace(skillId string) error {
	store, err := NewGeneralAgentSkillWorkspaceStore(skillId)
	if err != nil {
		log.Warnf("[wga-skill] skill %v create persistent store for cleanup err: %v", skillId, err)
		return nil
	}
	return CleanupWgaWorkspace(store)
}
