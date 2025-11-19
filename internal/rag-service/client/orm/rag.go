package orm

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"gorm.io/gorm"
)

func (c *Client) DeleteRag(ctx context.Context, req *rag_service.RagDeleteReq) *err_code.Status {
	err := sqlopt.WithRagID(req.RagId).Apply(c.db.WithContext(ctx)).First(&model.RagInfo{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return toErrStatus("rag_delete_err", "rag not found: "+req.RagId)
	} else if err != nil {
		return toErrStatus("rag_delete_err", err.Error())
	}

	err = sqlopt.WithRagID(req.RagId).Apply(c.db.WithContext(ctx)).Delete(&model.RagInfo{}).Error
	if err != nil {
		return toErrStatus("rag_delete_err", err.Error())
	}

	return nil
}
func (c *Client) GetRag(ctx context.Context, req *rag_service.RagDetailReq) (*rag_service.RagInfo, *err_code.Status) {
	info := &model.RagInfo{}

	// 获取 rag 信息 [EN] Get rag information
	err := sqlopt.WithRagID(req.RagId).Apply(c.db.WithContext(ctx)).First(info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, toErrStatus("rag_get_err", "rag not found: "+req.RagId)
	} else if err != nil {
		return nil, toErrStatus("rag_get_err", err.Error())
	}

	// 反序列化敏感词表 [EN] Deserialize sensitive word list
	var sensitiveIds []string
	if info.SensitiveConfig.TableIds != "" {
		err = json.Unmarshal([]byte(info.SensitiveConfig.TableIds), &sensitiveIds)
		if err != nil {
			return nil, toErrStatus("rag_get_err", "sensitive "+err.Error())
		}
	}
	knowledgeConfig := info.KnowledgeBaseConfig
	kbGlobalConfig := &rag_service.RagGlobalConfig{
		MaxHistory:        int32(knowledgeConfig.MaxHistory),
		Threshold:         float32(knowledgeConfig.Threshold),
		TopK:              int32(knowledgeConfig.TopK),
		MatchType:         knowledgeConfig.MatchType,
		PriorityMatch:     knowledgeConfig.PriorityMatch,
		SemanticsPriority: float32(knowledgeConfig.SemanticsPriority),
		KeywordPriority:   float32(knowledgeConfig.KeywordPriority),
		TermWeightEnable:  knowledgeConfig.TermWeightEnable,
		TermWeight:        float32(knowledgeConfig.TermWeight),
		UseGraph:          knowledgeConfig.UseGraph,
	}

	var perKbConfig []*rag_service.RagPerKnowledgeConfig
	if info.KnowledgeBaseConfig.MetaParams != "" {
		err = json.Unmarshal([]byte(info.KnowledgeBaseConfig.MetaParams), &perKbConfig)
		if err != nil {
			return nil, toErrStatus("rag_get_err", "kb_meta "+err.Error())
		}
	}

	// 填充 rag 的信息 [EN] Fill in rag information
	resp := &rag_service.RagInfo{
		RagId: info.RagID,
		BriefConfig: &common.AppBriefConfig{
			Name:       info.BriefConfig.Name,
			Desc:       info.BriefConfig.Desc,
			AvatarPath: info.BriefConfig.AvatarPath,
		},
		ModelConfig: &common.AppModelConfig{
			Model:     info.ModelConfig.Model,
			ModelId:   info.ModelConfig.ModelId,
			Provider:  info.ModelConfig.Provider,
			ModelType: info.ModelConfig.ModelType,
			Config:    info.ModelConfig.Config,
		},
		RerankConfig: &common.AppModelConfig{
			Model:     info.RerankConfig.Model,
			ModelId:   info.RerankConfig.ModelId,
			Provider:  info.RerankConfig.Provider,
			ModelType: info.RerankConfig.ModelType,
			Config:    info.RerankConfig.Config,
		},
		KnowledgeBaseConfig: &rag_service.RagKnowledgeBaseConfig{
			PerKnowledgeConfigs: perKbConfig,
			GlobalConfig:        kbGlobalConfig,
		},
		SensitiveConfig: &rag_service.RagSensitiveConfig{
			Enable:   info.SensitiveConfig.Enable,
			TableIds: sensitiveIds,
		},
	}

	return resp, nil
}

func (c *Client) GetRagList(ctx context.Context, req *rag_service.RagListReq) (*rag_service.RagListResp, *err_code.Status) {
	info := make([]*model.RagInfo, 0)
	err := sqlopt.SQLOptions(
		sqlopt.WithUserID(req.Identity.UserId),
		sqlopt.WithOrgID(req.Identity.OrgId),
		sqlopt.LikeBriefName(req.Name),
	).Apply(c.db.WithContext(ctx)).Find(&info).Error

	if err != nil {
		return nil, toErrStatus("rag_list_err", err.Error())
	}

	var list []*common.AppBrief
	for _, v := range info {
		list = append(list, &common.AppBrief{
			AppId:      v.RagID,
			AppType:    constant.AppTypeRag,
			AvatarPath: v.BriefConfig.AvatarPath,
			Name:       v.BriefConfig.Name,
			Desc:       v.BriefConfig.Desc,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
			OrgId:      v.OrgID,
			UserId:     v.UserID,
		})
	}

	return &rag_service.RagListResp{
		RagInfos: list,
		Total:    int64(len(info)),
	}, nil
}

func (c *Client) GetRagByIds(ctx context.Context, req *rag_service.GetRagByIdsReq) (*rag_service.AppBriefList, *err_code.Status) {
	if len(req.RagIdList) == 0 {
		return nil, toErrStatus("rag_list_err", "ragIdList cannot be empty")
	}

	info := make([]*model.RagInfo, 0)
	err := sqlopt.InRagIds(req.RagIdList).Apply(c.db.WithContext(ctx)).Order("updated_at DESC").Find(&info).Error
	if err != nil {
		return nil, toErrStatus("rag_list_err", err.Error())
	}

	var list []*common.AppBrief
	for _, v := range info {
		list = append(list, &common.AppBrief{
			AppId:      v.RagID,
			AppType:    constant.AppTypeRag,
			AvatarPath: v.BriefConfig.AvatarPath,
			Name:       v.BriefConfig.Name,
			Desc:       v.BriefConfig.Desc,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
			OrgId:      v.OrgID,
			UserId:     v.UserID,
		})
	}

	return &rag_service.AppBriefList{
		RagInfos: list,
	}, nil
}

func (c *Client) CreateRag(ctx context.Context, rag *model.RagInfo) *err_code.Status {
	if rag.ID != 0 {
		return toErrStatus("rag_create_err", "create rag but id err") // todo
	}
	if rag.RagID == "" {
		return toErrStatus("rag_create_err", "ragID cannot be empty")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// 检查是否有重复ragID [EN] Check if there are duplicate ragIDs
		if err := sqlopt.WithRagID(rag.RagID).Apply(tx).First(&model.RagInfo{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("rag_create_err", "failed to check ragID: "+err.Error())
			}
		} else {
			return toErrStatus("rag_create_err", "repeated ragID: "+rag.RagID)
		}

		// 默认开关开启 + 默认值 [EN] Default switch on + default value
		rag.KnowledgeBaseConfig.MaxHistory = 0
		rag.KnowledgeBaseConfig.Threshold = 0.4
		rag.KnowledgeBaseConfig.TopK = 5

		if err := tx.Create(rag).Error; err != nil {
			return toErrStatus("rag_create_err", err.Error()) // todo
		}
		return nil
	})
}

func (c *Client) UpdateRag(ctx context.Context, rag *model.RagInfo) *err_code.Status {
	if rag.RagID == "" {
		return toErrStatus("rag_update_err", "update rag but ragID is empty")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// 检查ragID是否存在 [EN] Check if ragID exists
		if err := sqlopt.WithRagID(rag.RagID).Apply(tx).First(&model.RagInfo{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("rag_update_err", "rag not found: "+rag.RagID)
			} else {
				return toErrStatus("rag_update_err", "failed to check rag: "+err.Error())
			}
		} else {
			// update rag
			updateMap := map[string]interface{}{
				"brief_name":        rag.BriefConfig.Name,
				"brief_desc":        rag.BriefConfig.Desc,
				"brief_avatar_path": rag.BriefConfig.AvatarPath,
			}
			// 只更新指定 ragID 的记录 [EN] Only update records with specified ragID
			if err := sqlopt.WithRagID(rag.RagID).Apply(tx).Model(&model.RagInfo{}).Updates(updateMap).Error; err != nil {
				return toErrStatus("rag_update_err", "failed to update basic rag: "+err.Error())
			}
		}
		return nil
	})
}

func (c *Client) UpdateRagConfig(ctx context.Context, rag *model.RagInfo) *err_code.Status {
	if rag.RagID == "" {
		return toErrStatus("rag_update_err", "update rag but ragID is empty")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// 检查ragID是否存在 [EN] Check if ragID exists
		if err := sqlopt.WithRagID(rag.RagID).Apply(tx).First(&model.RagInfo{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("rag_update_err", "rag not found: "+rag.RagID)
			} else {
				return toErrStatus("rag_update_err", "failed to check rag: "+err.Error())
			}
		} else {
			// update rag
			updateMap := map[string]interface{}{
				"model_provider":   rag.ModelConfig.Provider,
				"model_model":      rag.ModelConfig.Model,
				"model_model_id":   rag.ModelConfig.ModelId,
				"model_model_type": rag.ModelConfig.ModelType,
				"model_config":     rag.ModelConfig.Config,

				"rerank_provider":   rag.RerankConfig.Provider,
				"rerank_model":      rag.RerankConfig.Model,
				"rerank_model_id":   rag.RerankConfig.ModelId,
				"rerank_model_type": rag.RerankConfig.ModelType,
				"rerank_config":     rag.RerankConfig.Config,

				"kb_know_id":     rag.KnowledgeBaseConfig.KnowId,
				"kb_max_history": rag.KnowledgeBaseConfig.MaxHistory,
				"kb_threshold":   rag.KnowledgeBaseConfig.Threshold,
				"kb_top_k":       rag.KnowledgeBaseConfig.TopK,

				"kb_match_type":         rag.KnowledgeBaseConfig.MatchType,
				"kb_priority_match":     rag.KnowledgeBaseConfig.PriorityMatch,
				"kb_semantics_priority": rag.KnowledgeBaseConfig.SemanticsPriority,
				"kb_keyword_priority":   rag.KnowledgeBaseConfig.KeywordPriority,
				"kb_meta_params":        rag.KnowledgeBaseConfig.MetaParams,
				"kb_term_weight":        rag.KnowledgeBaseConfig.TermWeight,
				"kb_term_weight_enable": rag.KnowledgeBaseConfig.TermWeightEnable,
				"kb_use_graph":          rag.KnowledgeBaseConfig.UseGraph,

				"sensitive_enable":    rag.SensitiveConfig.Enable,
				"sensitive_table_ids": rag.SensitiveConfig.TableIds,
			}

			// 只更新指定 ragID 的记录 [EN] Only update records with specified ragID
			if err := sqlopt.WithRagID(rag.RagID).Apply(tx).Model(&model.RagInfo{}).Updates(updateMap).Error; err != nil {
				return toErrStatus("rag_update_err", "failed to update basic rag config: "+err.Error())
			}
		}
		return nil
	})
}

func (c *Client) FetchRagFirst(ctx context.Context, ragId string) (*model.RagInfo, *err_code.Status) {
	if ragId == "" {
		return nil, toErrStatus("rag_get_err", "get rag but ragID is empty")
	}
	rag := &model.RagInfo{}
	if err := sqlopt.WithRagID(ragId).Apply(c.db.WithContext(ctx)).First(rag).Error; err != nil {
		return nil, toErrStatus("rag_get_err", "failed to fetch rag: "+err.Error())
	}
	return rag, nil
}

func (c *Client) FetchRagFirstByName(ctx context.Context, name, userId, orgId string) (*model.RagInfo, *err_code.Status) {
	if name == "" {
		return nil, toErrStatus("rag_get_err", "name is empty")
	}
	rag := &model.RagInfo{}
	if err := sqlopt.SQLOptions(sqlopt.WithName(name), sqlopt.WithUserID(userId), sqlopt.WithOrgID(orgId)).
		Apply(c.db.WithContext(ctx)).First(rag).Error; err != nil {
		return nil, toErrStatus("rag_get_err", "failed to fetch rag: "+err.Error())
	}
	return rag, nil
}

func (c *Client) FetchRagCopyIndex(ctx context.Context, name, userId, orgId string) (int, *err_code.Status) {
	if name == "" {
		return 0, toErrStatus("rag_get_err", "name is empty")
	}
	prefix := name + "_"
	var maxIndex *int
	// 找出所有符合条件记录中最大的索引号 [EN] Find the largest index number among all eligible records
	query := `
        SELECT MAX(CAST(REPLACE(brief_name, ?, '') AS UNSIGNED)) 
		FROM rag_info 
		WHERE brief_name LIKE ? AND brief_name REGEXP ?
    `
	params := []interface{}{prefix, prefix + "%", "^" + prefix + "[0-9]+$"}
	if userId != "" {
		query += " AND user_id = ?"
		params = append(params, userId)
	}
	if orgId != "" {
		query += " AND org_id = ?"
		params = append(params, orgId)
	}
	err := c.db.WithContext(ctx).Raw(query, params...).Scan(&maxIndex).Error
	if err != nil {
		return 0, toErrStatus("rag_get_err", "failed to fetch max rag copy index: "+err.Error())
	}
	// 没有匹配的记录 [EN] No matching records
	if maxIndex == nil {
		return 1, nil
	}
	// 返回最大索引 + 1 [EN] Returns maximum index + 1
	return *maxIndex + 1, nil
}
