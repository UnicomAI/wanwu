package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type importedSkillRoot struct {
	zipPath string
	name    string
}

func ImportGeneralAgentSkillConversation(ctx *gin.Context, userId, orgId string, req request.ImportGeneralAgentSkillConversationReq) (*response.ImportGeneralAgentSkillConversationResp, error) {
	if err := checkModelConfig(ctx, req.ModelConfig); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.ZipUrl) == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "zipUrl is required")
	}
	modelConfigString, err := req.ModelConfig.ConfigString()
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, err.Error())
	}

	previewID := util.GenUUID()

	threadResp, err := assistant.WgaConversationCreate(ctx.Request.Context(), &assistant_service.WgaConversationCreateReq{
		Prompt: gin_util.I18nKey(ctx, "wga_skill_import_title"),
		ModelConfig: &common.AppModelConfig{
			ModelId:   req.ModelConfig.ModelId,
			Provider:  req.ModelConfig.Provider,
			Model:     req.ModelConfig.Model,
			ModelType: req.ModelConfig.ModelType,
			Config:    modelConfigString,
		},
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	customSkillResp, err := mcp.CustomSkillCreate(ctx.Request.Context(), &mcp_service.CustomSkillCreateReq{
		Name:            gin_util.I18nKey(ctx, "wga_skill_import_title"),
		Avatar:          req.Avatar.Key,
		Author:          getUserNameById(ctx, userId),
		WgaThreadId:     threadResp.ThreadId,
		PreviewThreadId: previewID,
		Identity:        &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		_, _ = assistant.WgaConversationDelete(ctx.Request.Context(), &assistant_service.WgaConversationDeleteReq{
			ThreadId: threadResp.ThreadId,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		return nil, err
	}
	customSkillID := customSkillResp.SkillId

	store, err := NewGeneralAgentSkillWorkspaceStore(customSkillID)
	if err != nil {
		rollbackImportedSkillConversation(ctx, userId, orgId, threadResp.ThreadId, customSkillID)
		return nil, err
	}
	skillDir := filepath.Join(GetWgaWorkspaceThreadDir(store), generalAgentWorkspaceSkillDirName)
	fm, err := importSkillIntoWorkspace(ctx, req.ZipUrl, skillDir)
	if err != nil {
		rollbackImportedSkillConversation(ctx, userId, orgId, threadResp.ThreadId, customSkillID)
		return nil, err
	}
	_, err = mcp.UpdateCustomSkillBasicMeta(ctx.Request.Context(), &mcp_service.UpdateCustomSkillBasicMetaReq{
		SkillId: customSkillID,
		Name:    fm.Name,
		Desc:    fm.Description,
	})
	if err != nil {
		rollbackImportedSkillConversation(ctx, userId, orgId, threadResp.ThreadId, customSkillID)
		return nil, err
	}

	return &response.ImportGeneralAgentSkillConversationResp{
		CustomSkillID: customSkillID,
		ThreadID:      threadResp.ThreadId,
		PreviewID:     previewID,
	}, nil
}

func rollbackImportedSkillConversation(ctx *gin.Context, userId, orgId, threadId, customSkillId string) {
	if customSkillId != "" {
		_, _ = mcp.CustomSkillDelete(ctx.Request.Context(), &mcp_service.CustomSkillDeleteReq{SkillId: customSkillId})
		_ = cleanupCustomSkillWorkspace(customSkillId)
	}
	if threadId != "" {
		_, _ = assistant.WgaConversationDelete(ctx.Request.Context(), &assistant_service.WgaConversationDeleteReq{
			ThreadId: threadId,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
	}
}

func importSkillIntoWorkspace(ctx *gin.Context, zipURL, skillDir string) (*util.SkillFrontMatter, error) {
	data, err := minio_util.DownloadFileDirect(ctx.Request.Context(), zipURL)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("download skill zip err: %v", err))
	}
	return importSkillDataIntoWorkspace(data, skillDir)
}

func importSkillDataIntoWorkspace(data []byte, skillDir string) (*util.SkillFrontMatter, error) {
	if err := util.RecreateDir(skillDir); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("prepare skill dir err: %v", err))
	}
	if err := unzipSkillDir(data, skillDir); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("unzip skill err: %v", err))
	}
	fm, err := readImportedSkillFrontMatter(skillDir)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFSkillParse, err.Error())
	}
	return fm, nil
}

func unzipSkillDir(data []byte, destDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	skillRoot, err := findSkillRootInZip(reader)
	if err != nil {
		return err
	}
	return util.UnzipSubDir(reader, skillRoot.zipPath, filepath.Join(destDir, skillRoot.name))
}

func findSkillRootInZip(reader *zip.Reader) (*importedSkillRoot, error) {
	var skillRoots []*importedSkillRoot
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		// EFS(bit 11)未置位 → 本地编码，先按 GBK/GB18030 解码成 UTF-8，再做路径清理。
		// 必须与 UnzipSubDir 内的解码保持一致：此处得到的 zipPath 会作为 subDir 传给
		// UnzipSubDir，若编码不一致（GBK vs UTF-8），IsZipEntryInSubDir 的 HasPrefix 匹配会失败。
		entryName := file.Name
		if file.Flags&util.ZipEFSFlag == 0 {
			entryName = util.DecodeGBKToUTF8(entryName)
		}
		cleanName, err := util.CleanZipEntryName(entryName)
		if err != nil {
			return nil, err
		}
		if path.Base(cleanName) == "SKILL.md" {
			fm, err := parseZipSkillFrontMatter(file)
			if err != nil {
				return nil, err
			}
			skillDirName, err := cleanSkillDirName(fm.Name)
			if err != nil {
				return nil, err
			}
			skillRoots = append(skillRoots, &importedSkillRoot{
				zipPath: path.Dir(cleanName),
				name:    skillDirName,
			})
		}
	}
	switch len(skillRoots) {
	case 0:
		return nil, fmt.Errorf("SKILL.md file not found in the zip archive")
	case 1:
		return skillRoots[0], nil
	default:
		return nil, fmt.Errorf("expected exactly one SKILL.md in the zip archive, got %d", len(skillRoots))
	}
}

func parseZipSkillFrontMatter(file *zip.File) (*util.SkillFrontMatter, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open SKILL.md file: %v", err)
	}
	defer func() { _ = rc.Close() }()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read SKILL.md file: %v", err)
	}
	return util.ParseSkillFrontMatter(string(content))
}

func cleanSkillDirName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("skill name is required")
	}
	if name == "." || name == ".." || path.IsAbs(name) || filepath.IsAbs(name) ||
		strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, ":") {
		return "", fmt.Errorf("invalid skill name: %s", name)
	}
	return name, nil
}

func readImportedSkillFrontMatter(skillDir string) (*util.SkillFrontMatter, error) {
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil, err
	}
	var skillDirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillMdPath := filepath.Join(skillDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMdPath); err == nil {
			skillDirs = append(skillDirs, filepath.Join(skillDir, entry.Name()))
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	switch len(skillDirs) {
	case 0:
		return nil, fmt.Errorf("SKILL.md file not found in imported skill directory")
	case 1:
		data, err := os.ReadFile(filepath.Join(skillDirs[0], "SKILL.md"))
		if err != nil {
			return nil, err
		}
		return util.ParseSkillFrontMatter(string(data))
	default:
		return nil, fmt.Errorf("expected exactly one imported skill directory, got %d", len(skillDirs))
	}
}
