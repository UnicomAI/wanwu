package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

const (
	finSkillsHubOwnerUser = "finskills-hub"
	finSkillsHubOwnerOrg  = "finskills-hub"
)

func finSkillsHubBaseURL() string {
	if config.Cfg() == nil {
		return ""
	}
	return strings.TrimRight(strings.TrimSpace(config.Cfg().FinSkillsHub.BaseURL), "/")
}

func isLocalSquareSkillID(skillID string) bool {
	if skillID == "" {
		return false
	}
	for _, r := range skillID {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func hubThreadID(skillID string) string {
	return "finskills-hub:" + skillID
}

type hubCatalogSkill struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Version      string `json:"version"`
	Author       string `json:"author"`
	Content      string `json:"content"`
	InstallCount int32  `json:"install_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func hubGetJSON(ctx context.Context, rawURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("finskills-hub %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, out)
}

func listHubCatalogSkills(ctx context.Context, name string) ([]hubCatalogSkill, error) {
	base := finSkillsHubBaseURL()
	if base == "" {
		return nil, nil
	}
	q := url.Values{}
	q.Set("limit", "100")
	q.Set("page", "1")
	if name != "" {
		q.Set("search", name)
	}
	var payload struct {
		Skills []hubCatalogSkill `json:"skills"`
	}
	if err := hubGetJSON(ctx, base+"/skills?"+q.Encode(), &payload); err != nil {
		return nil, err
	}
	return payload.Skills, nil
}

func getHubCatalogSkill(ctx context.Context, skillID string) (*hubCatalogSkill, error) {
	base := finSkillsHubBaseURL()
	if base == "" {
		return nil, fmt.Errorf("finskills-hub base_url is empty")
	}
	var skill hubCatalogSkill
	if err := hubGetJSON(ctx, base+"/skills/"+url.PathEscape(skillID), &skill); err != nil {
		return nil, err
	}
	if skill.ID == "" {
		skill.ID = skillID
	}
	return &skill, nil
}

func hubDownloadURL(slug string) string {
	return finSkillsHubBaseURL() + "/api/v1/download?slug=" + url.QueryEscape(slug)
}

func appendHubShareSkills(ctx *gin.Context, userId, orgId, name string, list []*response.SharedSkillInfo) []*response.SharedSkillInfo {
	skills, err := listHubCatalogSkills(ctx.Request.Context(), name)
	if err != nil {
		log.Warnf("finskills-hub catalog list failed: %v", err)
		return list
	}
	if len(skills) == 0 {
		return list
	}
	acquired := hubAcquiredByCatalogID(ctx, userId, orgId, skills)
	for _, skill := range skills {
		if skill.ID == "" {
			continue
		}
		list = append(list, &response.SharedSkillInfo{
			SkillBasicInfo: response.SkillBasicInfo{
				SkillId:   skill.ID,
				Name:      skill.Name,
				Author:    skill.Author,
				Desc:      skill.Description,
				CreatedAt: skill.CreatedAt,
				UpdatedAt: skill.UpdatedAt,
			},
			IsShared:      acquired[skill.ID],
			DownloadCount: skill.InstallCount,
			AcquiredCount: skill.InstallCount,
		})
	}
	return list
}

func hubAcquiredByCatalogID(ctx *gin.Context, userId, orgId string, skills []hubCatalogSkill) map[string]bool {
	out := map[string]bool{}
	threads := make([]string, 0, len(skills))
	for _, skill := range skills {
		if skill.ID != "" {
			threads = append(threads, hubThreadID(skill.ID))
		}
	}
	if len(threads) == 0 {
		return out
	}
	resp, err := mcp.GetCustomSkillListByThreadIDList(ctx.Request.Context(), &mcp_service.GetCustomSkillListByThreadIDListReq{
		Identity:        &mcp_service.Identity{UserId: finSkillsHubOwnerUser, OrgId: finSkillsHubOwnerOrg},
		WgaThreadIdList: threads,
	})
	if err != nil || resp == nil {
		return out
	}
	localByHub := map[string]string{}
	var localIDs []string
	for _, item := range resp.GetList() {
		skill := item.GetSkill()
		if skill == nil {
			continue
		}
		thread := skill.GetWgaThreadId()
		hubID := strings.TrimPrefix(thread, "finskills-hub:")
		if hubID == "" || hubID == thread || skill.GetSkillId() == "" {
			continue
		}
		localByHub[hubID] = skill.GetSkillId()
		localIDs = append(localIDs, skill.GetSkillId())
	}
	if len(localIDs) == 0 {
		return out
	}
	acquiredResp, err := mcp.CheckAcquiredSkill(ctx.Request.Context(), &mcp_service.CheckAcquiredSkillReq{
		SkillIds: localIDs,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil || acquiredResp == nil {
		return out
	}
	for hubID, localID := range localByHub {
		out[hubID] = acquiredResp.GetAcquiredMap()[localID]
	}
	return out
}

func hubSkillDetail(ctx *gin.Context, userId, orgId string, skillID string) (*response.SharedSkillDetail, error) {
	skill, err := getHubCatalogSkill(ctx.Request.Context(), skillID)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	acquired := hubAcquiredByCatalogID(ctx, userId, orgId, []hubCatalogSkill{*skill})
	return &response.SharedSkillDetail{
		SharedSkillInfo: response.SharedSkillInfo{
			SkillBasicInfo: response.SkillBasicInfo{
				SkillId:   skill.ID,
				Name:      skill.Name,
				Author:    skill.Author,
				Desc:      skill.Description,
				CreatedAt: skill.CreatedAt,
				UpdatedAt: skill.UpdatedAt,
			},
			IsShared:      acquired[skill.ID],
			DownloadCount: skill.InstallCount,
			AcquiredCount: skill.InstallCount,
		},
		SkillMarkdown: skill.Content,
	}, nil
}

func shareHubSkill(ctx *gin.Context, userId, orgId, skillID string) error {
	skill, err := getHubCatalogSkill(ctx.Request.Context(), skillID)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if strings.TrimSpace(skill.Slug) == "" {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, "finskills-hub skill slug is empty")
	}
	localID, err := ensureHubSkillPublished(ctx, skill)
	if err != nil {
		return err
	}
	acquiredResp, err := mcp.CheckAcquiredSkill(ctx.Request.Context(), &mcp_service.CheckAcquiredSkillReq{
		SkillIds: []string{localID},
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err == nil && acquiredResp.GetAcquiredMap()[localID] {
		return nil
	}
	_, err = mcp.AcquiredSkillCreate(ctx.Request.Context(), &mcp_service.AcquiredSkillCreateReq{
		Identity:      &mcp_service.Identity{UserId: userId, OrgId: orgId},
		CustomSkillId: localID,
	})
	if err != nil {
		return err
	}
	incrementCustomSkillAcquired(ctx, localID)
	return nil
}

func ensureHubSkillPublished(ctx *gin.Context, skill *hubCatalogSkill) (string, error) {
	owner := &mcp_service.Identity{UserId: finSkillsHubOwnerUser, OrgId: finSkillsHubOwnerOrg}
	existing, err := mcp.GetCustomSkillByThreadID(ctx.Request.Context(), &mcp_service.GetCustomSkillByThreadIDReq{
		WgaThreadId: hubThreadID(skill.ID),
	})
	localID := ""
	if err == nil && existing.GetSkill().GetSkillId() != "" {
		localID = existing.GetSkill().GetSkillId()
	}
	if localID == "" {
		created, err := mcp.CustomSkillCreate(ctx.Request.Context(), &mcp_service.CustomSkillCreateReq{
			Name:        skill.Name,
			Author:      skill.Author,
			Desc:        skill.Description,
			WgaThreadId: hubThreadID(skill.ID),
			Identity:    owner,
		})
		if err != nil {
			return "", err
		}
		localID = created.GetSkillId()
	}
	version := skill.Version
	if version == "" {
		version = "1.0.0"
	}
	latest, err := mcp.GetPublishCustomSkillByLatest(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillByLatestReq{SkillId: localID})
	if err == nil && latest.GetVersion() == version && latest.GetObjectPath() != "" {
		unpublishHubSkill(ctx, localID)
		return localID, nil
	}
	objectPath := hubDownloadURL(skill.Slug)
	_, err = mcp.CreatePublishCustomSkill(ctx.Request.Context(), &mcp_service.PublishCustomSkillReq{
		SkillId:     localID,
		Version:     version,
		VersionDesc: "finskills-hub",
		Identity:    owner,
		ObjectPath:  objectPath,
		Markdown:    skill.Content,
	})
	if err != nil {
		return "", err
	}
	unpublishHubSkill(ctx, localID)
	return localID, nil
}

// 源 skill 只给「我添加的」引用，不进入广场「共享」。
func unpublishHubSkill(ctx *gin.Context, skillID string) {
	if _, err := app.UnPublishApp(ctx.Request.Context(), &app_service.UnPublishAppReq{
		AppId:   skillID,
		AppType: constant.AppTypeSkill,
		UserId:  finSkillsHubOwnerUser,
	}); err != nil {
		log.Warnf("unpublish finskills-hub skill failed, skillId=%s err=%v", skillID, err)
	}
}

func downloadHubSkill(ctx context.Context, skillID string) ([]byte, error) {
	skill, err := getHubCatalogSkill(ctx, skillID)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if strings.TrimSpace(skill.Slug) == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "finskills-hub skill slug is empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hubDownloadURL(skill.Slug), nil)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("finskills-hub download %s", resp.Status))
	}
	return body, nil
}
