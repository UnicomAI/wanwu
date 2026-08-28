package orm

import (
	"context"
	"errors"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

const textKeyCustomSkillPublishNotFound = "mcp_custom_skill_publish_not_found"

type CustomSkillPublishSnapshot struct {
	Markdown   string
	ObjectPath string
}

func (c *Client) PublishCustomSkill(ctx context.Context, publish *model.CustomSkillPublish, snapshot *CustomSkillPublishSnapshot) *errs.Status {
	if publish.SkillID == "" {
		return toErrStatus("mcp_custom_skill_publish_skill_id_required")
	}
	if publish.Version == "" {
		return toErrStatus("mcp_custom_skill_publish_version_required")
	}

	customSkill, errStatus := c.GetCustomSkill(ctx, publish.SkillID)
	if errStatus != nil {
		return errStatus
	}

	markdown := customSkill.Markdown
	objectPath := customSkill.ObjectPath
	if snapshot != nil {
		if snapshot.Markdown != "" {
			markdown = snapshot.Markdown
		}
		if snapshot.ObjectPath != "" {
			objectPath = snapshot.ObjectPath
		}
	}

	publish.Markdown = markdown
	publish.ObjectPath = objectPath
	publish.UserId = customSkill.UserID
	publish.OrgId = customSkill.OrgID
	if err := c.db.WithContext(ctx).Create(publish).Error; err != nil {
		return toErrStatus("mcp_custom_skill_publish", err.Error())
	}
	return nil
}

func (c *Client) UpdatePublishCustomSkill(ctx context.Context, skillId, versionDesc string) *errs.Status {
	if skillId == "" {
		return toErrStatus("mcp_skill_config_invalid_arg")
	}
	latest, st := c.getLatestCustomSkillPublish(ctx, skillId)
	if st != nil {
		return st
	}
	if latest == nil {
		return toErrStatus(textKeyCustomSkillPublishNotFound)
	}
	if err := c.db.WithContext(ctx).Model(latest).Update("version_description", versionDesc).Error; err != nil {
		return toErrStatus("mcp_custom_skill_publish_update", err.Error())
	}
	return nil
}

// GetPublishCustomSkillHistoryList 返回的 total：无分页，表示该 skill 发布历史全量条数，与 len(list) 一致。
func (c *Client) GetPublishCustomSkillHistoryList(ctx context.Context, skillId string) ([]*model.CustomSkillPublish, int64, *errs.Status) {
	if skillId == "" {
		return nil, 0, toErrStatus("mcp_skill_config_invalid_arg")
	}
	var list []*model.CustomSkillPublish
	if err := sqlopt.SQLOptions(
		sqlopt.WithSkillID(skillId),
	).Apply(c.db).WithContext(ctx).Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, 0, toErrStatus("mcp_custom_skill_publish_history", err.Error())
	}
	return list, int64(len(list)), nil
}

// GetPublishCustomSkillByLatest 返回该 skill 最新发布记录；尚未发布时返回 (nil, nil)。
func (c *Client) GetPublishCustomSkillByLatest(ctx context.Context, skillId string) (*model.CustomSkillPublish, *errs.Status) {
	return c.getLatestCustomSkillPublish(ctx, skillId)
}

func (c *Client) GetPublishCustomSkillByVersion(ctx context.Context, skillId, version string) (*model.CustomSkillPublish, *errs.Status) {
	if skillId == "" || version == "" {
		return nil, toErrStatus("mcp_skill_config_invalid_arg")
	}
	var publish model.CustomSkillPublish
	if err := sqlopt.SQLOptions(
		sqlopt.WithSkillID(skillId),
		sqlopt.WithVersion(version),
	).Apply(c.db).WithContext(ctx).First(&publish).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus(textKeyCustomSkillPublishNotFound, skillId)
		}
		return nil, toErrStatus("mcp_custom_skill_publish_get", skillId, err.Error())
	}
	return &publish, nil
}

func (c *Client) GetPublishCustomSkillByIDList(ctx context.Context, skillIdList []string) ([]*model.CustomSkillPublish, *errs.Status) {
	if len(skillIdList) == 0 {
		return []*model.CustomSkillPublish{}, nil
	}
	skills, st := c.GetCustomSkillBySkillIds(ctx, skillIdList)
	if st != nil {
		return nil, st
	}
	return c.latestPublishesForSkills(ctx, skills)
}

func (c *Client) latestPublishesForSkills(ctx context.Context, skills []*model.CustomSkill) ([]*model.CustomSkillPublish, *errs.Status) {
	if len(skills) == 0 {
		return []*model.CustomSkillPublish{}, nil
	}
	skillIDs := make([]string, 0, len(skills))
	for _, cs := range skills {
		skillIDs = append(skillIDs, util.Int2Str(cs.ID))
	}
	var rows []*model.CustomSkillPublish
	if err := c.db.WithContext(ctx).
		Where("skill_id IN ?", skillIDs).
		Order("skill_id ASC, created_at DESC, id DESC").
		Find(&rows).Error; err != nil {
		return nil, toErrStatus("mcp_custom_skill_publish_get", err.Error())
	}
	latestBySkill := make(map[string]*model.CustomSkillPublish, len(skillIDs))
	for _, p := range rows {
		if p == nil {
			continue
		}
		if _, ok := latestBySkill[p.SkillID]; ok {
			continue
		}
		latestBySkill[p.SkillID] = p
	}
	list := make([]*model.CustomSkillPublish, 0, len(skills))
	for _, cs := range skills {
		sid := util.Int2Str(cs.ID)
		if p, ok := latestBySkill[sid]; ok {
			list = append(list, p)
		}
	}
	return list, nil
}

func (c *Client) getLatestCustomSkillPublish(ctx context.Context, skillId string) (*model.CustomSkillPublish, *errs.Status) {
	if skillId == "" {
		return nil, toErrStatus("mcp_skill_config_invalid_arg")
	}
	var publish model.CustomSkillPublish
	if err := sqlopt.SQLOptions(
		sqlopt.WithSkillID(skillId),
	).Apply(c.db).WithContext(ctx).Order("created_at DESC").
		First(&publish).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, toErrStatus("mcp_custom_skill_publish_get", skillId, err.Error())
	}
	return &publish, nil
}
