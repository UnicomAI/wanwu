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
	"github.com/UnicomAI/wanwu/pkg/db"
	"gorm.io/gorm"
)

func (c *Client) DeleteRag(ctx context.Context, req *rag_service.RagDeleteReq) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		err := sqlopt.WithRagID(req.RagId).Apply(tx.WithContext(ctx)).Delete(&model.RagInfo{}).Error
		if err != nil {
			return toErrStatus("rag_delete_err", err.Error())
		}
		err = sqlopt.WithRagID(req.RagId).Apply(tx.WithContext(ctx)).Delete(&model.RagPublish{}).Error
		if err != nil {
			return toErrStatus("rag_delete_err", err.Error())
		}
		return nil
	})
}

func (c *Client) GetRag(ctx context.Context, req *rag_service.RagDetailReq) (*rag_service.RagInfo, *err_code.Status) {
	info := &model.RagInfo{}
	// 获取 rag 信息
	err := sqlopt.WithRagID(req.RagId).Apply(c.db.WithContext(ctx)).First(info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, toErrStatus("rag_get_err", "rag not found: "+req.RagId)
	} else if err != nil {
		return nil, toErrStatus("rag_get_err", err.Error())
	}
	// 构建 rag 信息
	resp, errF := BuildRagInfo(info)
	if errF != nil {
		return nil, errF
	}
	return resp, nil
}

func BuildRagInfo(info *model.RagInfo) (*rag_service.RagInfo, *err_code.Status) {
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

	// 反序列化敏感词表
	var sensitiveIds []string
	if info.SensitiveConfig.TableIds != "" {
		err := json.Unmarshal([]byte(info.SensitiveConfig.TableIds), &sensitiveIds)
		if err != nil {
			return nil, toErrStatus("rag_get_err", "sensitive "+err.Error())
		}
	}

	// 反序列化 - 知识库元数据
	var perKbConfig []*rag_service.RagPerKnowledgeConfig
	if info.KnowledgeBaseConfig.MetaParams != "" {
		err := json.Unmarshal([]byte(info.KnowledgeBaseConfig.MetaParams), &perKbConfig)
		if err != nil {
			return nil, toErrStatus("rag_get_err", "kb_meta "+err.Error())
		}
	}

	// 反序列化 - 问答库配置
	qaConfig := &rag_service.RagQAKnowledgeBaseConfig{}
	if info.QAKnowledgebaseConfig != "" {
		if err := json.Unmarshal([]byte(info.QAKnowledgebaseConfig), qaConfig); err != nil {
			return nil, toErrStatus("rag_get_err", "kb_qa "+err.Error())
		}
	}

	// 反序列化推荐问题
	var recommendQuestion []string
	if info.RecommendQuestion != "" {
		if err := json.Unmarshal([]byte(info.RecommendQuestion), &recommendQuestion); err != nil {
			return nil, toErrStatus("rag_get_err", "recommend_question "+err.Error())
		}
	}

	// 填充 rag 的信息
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
			Provider:  info.RerankConfig.Provider,
			Model:     info.RerankConfig.Model,
			ModelId:   info.RerankConfig.ModelId,
			ModelType: info.RerankConfig.ModelType,
			Config:    info.RerankConfig.Config,
		},
		QArerankConfig: &common.AppModelConfig{
			Provider:  info.QARerankConfig.Provider,
			Model:     info.QARerankConfig.Model,
			ModelId:   info.QARerankConfig.ModelId,
			ModelType: info.QARerankConfig.ModelType,
			Config:    info.QARerankConfig.Config,
		},
		KnowledgeBaseConfig: &rag_service.RagKnowledgeBaseConfig{
			PerKnowledgeConfigs: perKbConfig,
			GlobalConfig:        kbGlobalConfig,
		},
		QAknowledgeBaseConfig: qaConfig,
		SensitiveConfig: &rag_service.RagSensitiveConfig{
			Enable:   info.SensitiveConfig.Enable,
			TableIds: sensitiveIds,
		},
		VisionConfig: &rag_service.RagVisionConfig{
			PicNum: info.VisionConfig.PicNum,
		},
		RecommendQuestion: recommendQuestion,
		Identity: &rag_service.Identity{
			UserId: info.UserID,
			OrgId:  info.OrgID,
		},
		CreateTime: info.CreatedAt,
		UpdateTime: info.UpdatedAt,
	}
	return resp, nil
}

func (c *Client) GetRagList(ctx context.Context, req *rag_service.RagListReq) (*rag_service.RagListResp, *err_code.Status) {
	info := make([]*model.RagInfo, 0)
	err := sqlopt.SQLOptions(
		sqlopt.WithUserID(req.Identity.UserId),
		sqlopt.WithOrgID(req.Identity.OrgId),
		sqlopt.LikeBriefName(req.Name),
	).Apply(c.db.WithContext(ctx)).Order("updated_at DESC").Find(&info).Error

	if err != nil {
		return nil, toErrStatus("rag_list_err", err.Error())
	}

	return &rag_service.RagListResp{
		RagInfos: toAppBriefs(info),
		Total:    int64(len(info)),
	}, nil
}

// AdminRagPageList 管理员中心知识问答列表，按 updated_at 倒序分页；total 恒为匹配总数。
func (c *Client) AdminRagPageList(ctx context.Context, req *rag_service.AdminRagPageListReq) (*rag_service.AdminRagPageListResp, *err_code.Status) {
	query := sqlopt.SQLOptions(
		sqlopt.LikeBriefName(db.EscapeLike(req.Name)),
		sqlopt.WithUserIDs(req.UserId),
		sqlopt.WithOrgIDs(req.OrgId),
	).Apply(c.db.WithContext(ctx)).Model(&model.RagInfo{})

	var info []*model.RagInfo
	var total int64
	if err := query.
		Count(&total).
		Order("updated_at DESC").
		Offset(int(req.PageSize * (req.PageNum - 1))).
		Limit(int(req.PageSize)).
		Find(&info).Error; err != nil {
		return nil, toErrStatus("rag_admin_page_list_err", err.Error())
	}

	return &rag_service.AdminRagPageListResp{
		RagInfos: toAppBriefs(info),
		Total:    total,
	}, nil
}

// toAppBriefs 将知识问答转为通用应用简要信息
func toAppBriefs(info []*model.RagInfo) []*common.AppBrief {
	list := make([]*common.AppBrief, 0, len(info))
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
	return list
}

func (c *Client) GetRagByIds(ctx context.Context, req *rag_service.GetRagByIdsReq) (*rag_service.AppBriefList, *err_code.Status) {
	// InRagIds 在 len==0 时不加条件，此处必须短路，否则会查全表。
	if len(req.RagIdList) == 0 {
		return &rag_service.AppBriefList{}, nil
	}
	info := make([]*model.RagInfo, 0)
	err := sqlopt.InRagIds(req.RagIdList).Apply(c.db.WithContext(ctx)).Order("updated_at DESC").Find(&info).Error
	if err != nil {
		return nil, toErrStatus("rag_list_err", err.Error())
	}

	return &rag_service.AppBriefList{
		RagInfos: toAppBriefs(info),
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
		// 检查是否有重复ragID
		if err := sqlopt.WithRagID(rag.RagID).Apply(tx).First(&model.RagInfo{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("rag_create_err", "failed to check ragID: "+err.Error())
			}
		} else {
			return toErrStatus("rag_create_err", "repeated ragID: "+rag.RagID)
		}

		// 默认开关开启 + 默认值
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
		// 检查ragID是否存在
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
			// 只更新指定 ragID 的记录
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
		// 检查ragID是否存在
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

				"qa_rerank_provider":   rag.QARerankConfig.Provider,
				"qa_rerank_model":      rag.QARerankConfig.Model,
				"qa_rerank_model_id":   rag.QARerankConfig.ModelId,
				"qa_rerank_model_type": rag.QARerankConfig.ModelType,
				"qa_rerank_config":     rag.QARerankConfig.Config,

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

				"qa_knowledgebase_config": rag.QAKnowledgebaseConfig,

				"sensitive_enable":    rag.SensitiveConfig.Enable,
				"sensitive_table_ids": rag.SensitiveConfig.TableIds,

				"vision_pic_num": rag.VisionConfig.PicNum,

				"recommend_question": rag.RecommendQuestion,
			}

			// 只更新指定 ragID 的记录
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
	// 找出所有符合条件记录中最大的索引号
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
	// 没有匹配的记录
	if maxIndex == nil {
		return 1, nil
	}
	// 返回最大索引 + 1
	return *maxIndex + 1, nil
}

func (c *Client) PublishRag(ctx context.Context, rag *model.RagPublish) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// 查找是否已存在相同 ragID 和版本号的记录
		if err := sqlopt.SQLOptions(sqlopt.WithRagID(rag.RagID), sqlopt.WithVersion(rag.Version)).
			Apply(tx).First(&model.RagPublish{}).Error; err == nil {
			return toErrStatus("rag_publish_err", "repeated version: "+rag.Version)
		}
		if err := tx.Create(rag).Error; err != nil {
			return toErrStatus("rag_publish_err", err.Error()) // todo
		}
		return nil
	})
}

func (c *Client) FetchPublishRagFirst(ctx context.Context, ragId, version string) (*model.RagPublish, *err_code.Status) {
	if ragId == "" {
		return nil, toErrStatus("rag_get_err", "get rag but ragID is empty")
	}
	rag := &model.RagPublish{}
	// 如果version为空，则返回最新版本
	if err := sqlopt.SQLOptions(sqlopt.WithRagID(ragId), sqlopt.WithVersion(version)).
		Apply(c.db.WithContext(ctx)).
		Order("created_at DESC").
		Limit(1).
		First(rag).Error; err != nil {
		return nil, toErrStatus("rag_get_err", "failed to fetch rag: "+err.Error())
	}
	return rag, nil
}

func (c *Client) UpdatePublishRag(ctx context.Context, rag *model.RagPublish) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := sqlopt.WithID(rag.ID).Apply(tx).Model(&model.RagPublish{}).
			Update("description", rag.Description).Error; err != nil {
			return toErrStatus("rag_update_err", "failed to update rag: "+err.Error())
		}
		return nil
	})
}

func (c *Client) FetchPublishRagList(ctx context.Context, ragId string) ([]*model.RagPublish, *err_code.Status) {
	if ragId == "" {
		return nil, toErrStatus("rag_get_err", "get rag but ragID is empty")
	}
	ragList := make([]*model.RagPublish, 0)
	err := sqlopt.SQLOptions(sqlopt.WithRagID(ragId)).Apply(c.db.WithContext(ctx)).Order("created_at DESC").Find(&ragList).Error
	if err != nil {
		return nil, toErrStatus("rag_get_err", "failed to fetch rag: "+err.Error())
	}
	return ragList, nil
}

// FetchPublishRagListByRagIds 按 rag_id IN 返回全部匹配的发布记录（含历史版本）；每个 rag 取最新由 GetPublishRagDescBatch 按 CreatedAt 聚合。
func (c *Client) FetchPublishRagListByRagIds(ctx context.Context, ragIds []string) ([]*model.RagPublish, *err_code.Status) {
	if len(ragIds) == 0 {
		return nil, toErrStatus("rag_get_err", "ragIds cannot be empty")
	}
	ragList := make([]*model.RagPublish, 0)
	err := sqlopt.SQLOptions(sqlopt.InRagIds(ragIds)).Apply(c.db.WithContext(ctx)).Order("created_at DESC").Find(&ragList).Error
	if err != nil {
		return nil, toErrStatus("rag_get_err", "failed to fetch rag: "+err.Error())
	}
	return ragList, nil
}
