package rag

import (
	"context"
	"encoding/json"
	"time"

	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/orm"
	message_builder "github.com/UnicomAI/wanwu/internal/rag-service/service/message-builder"
	rag_conversation "github.com/UnicomAI/wanwu/internal/rag-service/service/rag-conversation"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	QACategory int32 = 1 // 问答库类型
	RAGDRAFT   int32 = 0 // 草稿
	RAGPUBLISH int32 = 1 // 发布
)

type Service struct {
	cli client.IClient
	rag_service.UnimplementedRagServiceServer
}

func NewService(cli client.IClient) *Service {
	return &Service{
		cli: cli,
	}
}

func errStatus(code errs.Code, status *errs.Status) error {
	// 归属过滤未命中（rag 不存在或不属于请求方）统一报 RagInfoNotExist，
	// 不让调用方从错误码上区分"无权限"和"不存在"
	if status.TextKey == orm.RagNotExistKey {
		code = errs.Code_RagInfoNotExist
	}
	return grpc_util.ErrorStatusWithKey(code, status.TextKey, status.Args...)
}

// identityOf 取请求里的归属信息，为空表示不做归属过滤
func identityOf(identity *rag_service.Identity) (userId, orgId string) {
	return identity.GetUserId(), identity.GetOrgId()
}

func (s *Service) ChatRag(req *rag_service.ChatRagReq, stream grpc.ServerStreamingServer[rag_service.ChatRagResp]) error {
	ctx := stream.Context()
	//1.查询rag详情
	// 草稿只能问自己的；已发布服务公开应用（探索/openapi），不做归属过滤
	chatUserId, chatOrgId := "", ""
	if req.Publish != RAGPUBLISH {
		chatUserId, chatOrgId = identityOf(req.Identity)
	}
	rag, err := s.searchRagDetail(ctx, req.RagId, "", req.Publish, chatUserId, chatOrgId)
	if err != nil {
		return err
	}
	log.Infof("get rag: %+v", http_client.Convert2LogString(rag))
	//2.查询知识库列表
	knowledgeInfoList, err := buildKnowledgeList(ctx, rag)
	if err != nil {
		return err
	}
	knowledgeIds, qaIds, knowledgeIDToName := splitKnowledgeIdList(knowledgeInfoList)
	//3.回填历史会话
	s.prepareHistory(ctx, req, rag)
	//4.构造rag流式问答消息
	return message_builder.BuildMessage(ctx, &message_builder.RagContext{
		MessageId:         util.IfElse(req.DetailId != "", req.DetailId, util.NewID()),
		Req:               req,
		Rag:               rag,
		KnowledgeIDToName: knowledgeIDToName,
		KnowledgeIds:      knowledgeIds,
		QAIds:             qaIds,
	}, stream)
}

// prepareHistory 传了 conversationId 时以服务端存的历史为准，覆盖调用方传上来的 history
func (s *Service) prepareHistory(ctx context.Context, req *rag_service.ChatRagReq, rag *model.RagInfo) {
	if req.ConversationId == "" {
		return
	}
	userId, orgId := identityOf(req.Identity)
	conversation, status := s.cli.FetchRagConversation(ctx, req.ConversationId, req.RagId, userId, orgId)
	if status != nil {
		log.Errorf("知识问答会话不存在或不属于本知识问答，ragId: %s, conversationId: %s, err: %v", req.RagId, req.ConversationId, status)
		req.History = nil
		return
	}
	// 刷新 updated_at，使会话列表按最近聊天排序
	if status := s.cli.TouchRagConversation(ctx, conversation.ConversationID); status != nil {
		log.Errorf("刷新知识问答会话时间失败，conversationId: %s, err: %v", req.ConversationId, status)
	}

	req.History = nil
	maxHistory := rag_conversation.BuildMaxHistory(rag)
	if maxHistory <= 0 {
		return
	}
	// 按时间倒序取最近 maxHistory 轮，再反转成正序
	details, _, err := rag_conversation.SearchDetail(ctx, req.ConversationId, userId, orgId, 0, maxHistory, "desc")
	if err != nil {
		log.Errorf("查询知识问答历史失败，conversationId: %s, err: %v", req.ConversationId, err)
		return
	}
	for i := len(details) - 1; i >= 0; i-- {
		detail := details[i]
		if detail.Prompt == "" {
			continue
		}
		req.History = append(req.History, &rag_service.HistoryItem{
			Query:       detail.Prompt,
			Response:    detail.Response,
			NeedHistory: true,
		})
	}
}

func (s *Service) CreateRag(ctx context.Context, in *rag_service.CreateRagReq) (*rag_service.CreateRagResp, error) {
	// 检查是否有重名应用
	rag, _ := s.cli.FetchRagFirstByName(ctx, in.AppBrief.Name, in.Identity.UserId, in.Identity.OrgId)
	if rag != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_RagDuplicateName)
	}
	ragId := util.NewID()
	err := s.cli.CreateRag(ctx, &model.RagInfo{
		RagID: ragId,
		BriefConfig: model.AppBriefConfig{
			Name:       in.AppBrief.Name,
			Desc:       in.AppBrief.Desc,
			AvatarPath: in.AppBrief.AvatarPath,
		},
		PublicModel: model.PublicModel{
			OrgID:  in.Identity.OrgId,
			UserID: in.Identity.UserId,
		},
	})
	if err != nil {
		return nil, errStatus(errs.Code_RagCreateErr, err) // todo
	}
	return &rag_service.CreateRagResp{RagId: ragId}, nil
}

func (s *Service) UpdateRag(ctx context.Context, in *rag_service.UpdateRagReq) (*emptypb.Empty, error) {
	userId, orgId := identityOf(in.Identity)
	originalRag, err := s.cli.FetchRagFirst(ctx, in.RagId, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	if originalRag.BriefConfig.Name != in.AppBrief.Name {
		// 检查是否有重名应用
		rag, _ := s.cli.FetchRagFirstByName(ctx, in.AppBrief.Name, in.Identity.UserId, in.Identity.OrgId)
		if rag != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_RagDuplicateName)
		}
	}
	if err = s.cli.UpdateRag(ctx, &model.RagInfo{
		RagID: in.RagId,
		BriefConfig: model.AppBriefConfig{
			Name:       in.AppBrief.Name,
			Desc:       in.AppBrief.Desc,
			AvatarPath: in.AppBrief.AvatarPath,
		},
	}, userId, orgId); err != nil {
		return nil, errStatus(errs.Code_RagUpdateErr, err)
	}
	return &emptypb.Empty{}, nil
}

// normalizeRagConfigReq 把请求里缺省的配置段补成空结构体。
// 下面的组装逻辑按"字段缺省即取默认值"写，全是裸解引用，调用方少传一段就会 panic
func normalizeRagConfigReq(in *rag_service.UpdateRagConfigReq) {
	if in.ModelConfig == nil {
		in.ModelConfig = &common.AppModelConfig{}
	}
	if in.RerankConfig == nil {
		in.RerankConfig = &common.AppModelConfig{}
	}
	if in.QArerankConfig == nil {
		in.QArerankConfig = &common.AppModelConfig{}
	}
	if in.KnowledgeBaseConfig == nil {
		in.KnowledgeBaseConfig = &rag_service.RagKnowledgeBaseConfig{}
	}
	if in.KnowledgeBaseConfig.GlobalConfig == nil {
		in.KnowledgeBaseConfig.GlobalConfig = &rag_service.RagGlobalConfig{}
	}
	if in.QAknowledgeBaseConfig == nil {
		in.QAknowledgeBaseConfig = &rag_service.RagQAKnowledgeBaseConfig{}
	}
	if in.QAknowledgeBaseConfig.GlobalConfig == nil {
		in.QAknowledgeBaseConfig.GlobalConfig = &rag_service.RagQAGlobalConfig{}
	}
	if in.SensitiveConfig == nil {
		in.SensitiveConfig = &rag_service.RagSensitiveConfig{}
	}
	if in.VisionConfig == nil {
		in.VisionConfig = &rag_service.RagVisionConfig{}
	}
}

func (s *Service) UpdateRagConfig(ctx context.Context, in *rag_service.UpdateRagConfigReq) (*emptypb.Empty, error) {
	cfgUserId, cfgOrgId := identityOf(in.Identity)
	// 归属校验前置：越权请求不进入后面的配置组装，也不依赖请求体是否完整
	if _, err := s.cli.FetchRagFirst(ctx, in.RagId, cfgUserId, cfgOrgId); err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	normalizeRagConfigReq(in)
	var sensitiveIds string
	var knowledgeIds string
	if in.SensitiveConfig.TableIds != nil {
		sensitiveIdBytes, err := json.Marshal(in.SensitiveConfig.TableIds)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_update_err", "marshal err:", err.Error())
		}
		sensitiveIds = string(sensitiveIdBytes)
	}
	var knowledgeIdList []string
	for _, perKbConfig := range in.KnowledgeBaseConfig.PerKnowledgeConfigs {
		knowledgeIdList = append(knowledgeIdList, perKbConfig.KnowledgeId)
	}
	if len(knowledgeIdList) > 0 {
		knowledgeIdBytes, err := json.Marshal(knowledgeIdList)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_update_err", "marshal err:", err.Error())
		}
		knowledgeIds = string(knowledgeIdBytes)
	}

	var metaParams string
	perConfig := in.KnowledgeBaseConfig.PerKnowledgeConfigs
	if perConfig != nil {
		kbConfigBytes, err := json.Marshal(perConfig)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_update_err", "marshal err:", err.Error())
		}
		metaParams = string(kbConfigBytes)
	}
	kbGlobalConfig := in.KnowledgeBaseConfig.GlobalConfig

	rerankConfig := model.AppModelConfig{}
	qaRerankConfig := model.AppModelConfig{}

	// 设置检索方式默认值
	if kbGlobalConfig.MatchType == "" || len(knowledgeIdList) == 0 {
		kbGlobalConfig.KeywordPriority = model.KeywordPriorityDefault
		kbGlobalConfig.MatchType = model.MatchTypeDefault
		kbGlobalConfig.PriorityMatch = model.KnowledgePriorityDefault
		kbGlobalConfig.Threshold = model.ThresholdDefault
		kbGlobalConfig.SemanticsPriority = model.SemanticsPriorityDefault
		kbGlobalConfig.TopK = model.TopKDefault
	} else {
		rerankConfig = model.AppModelConfig{
			Provider:  in.RerankConfig.Provider,
			Model:     in.RerankConfig.Model,
			ModelId:   in.RerankConfig.ModelId,
			ModelType: in.RerankConfig.ModelType,
			Config:    in.RerankConfig.Config,
		}
	}

	qaConfig := in.QAknowledgeBaseConfig
	if qaConfig.GlobalConfig.MatchType == "" || len(qaConfig.PerKnowledgeConfigs) == 0 {
		qaConfig.GlobalConfig.KeywordPriority = model.KeywordPriorityDefault
		qaConfig.GlobalConfig.MatchType = model.MatchTypeDefault
		qaConfig.GlobalConfig.PriorityMatch = model.QAPriorityDefault
		qaConfig.GlobalConfig.Threshold = model.ThresholdDefault
		qaConfig.GlobalConfig.SemanticsPriority = model.SemanticsPriorityDefault
		qaConfig.GlobalConfig.TopK = model.TopKDefault
	} else {
		qaRerankConfig = model.AppModelConfig{
			Provider:  in.QArerankConfig.Provider,
			Model:     in.QArerankConfig.Model,
			ModelId:   in.QArerankConfig.ModelId,
			ModelType: in.QArerankConfig.ModelType,
			Config:    in.QArerankConfig.Config,
		}
	}
	in.QAknowledgeBaseConfig.GlobalConfig = qaConfig.GlobalConfig
	// 序列化QAknowledgeBaseConfig
	var qaKnowledgeConfig string
	if in.QAknowledgeBaseConfig != nil {
		knowledgeBaseConfigBytes, err := json.Marshal(in.QAknowledgeBaseConfig)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_update_err", "marshal err:", err.Error())
		}
		qaKnowledgeConfig = string(knowledgeBaseConfigBytes)
		log.Debugf("knowConfig = %s", qaKnowledgeConfig)
	}

	var recommendQuestionStr string
	if len(in.RecommendQuestion) > 0 {
		rqBytes, err := json.Marshal(in.RecommendQuestion)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_update_err", "marshal err:", err.Error())
		}
		recommendQuestionStr = string(rqBytes)
	}

	if err := s.cli.UpdateRagConfig(ctx, &model.RagInfo{
		RagID: in.RagId,
		ModelConfig: model.AppModelConfig{
			Provider:  in.ModelConfig.Provider,
			Model:     in.ModelConfig.Model,
			ModelId:   in.ModelConfig.ModelId,
			ModelType: in.ModelConfig.ModelType,
			Config:    in.ModelConfig.Config,
		},
		RerankConfig:   rerankConfig,
		QARerankConfig: qaRerankConfig,
		KnowledgeBaseConfig: model.KnowledgeBaseConfig{
			KnowId:            knowledgeIds,
			MaxHistory:        int64(kbGlobalConfig.MaxHistory),
			Threshold:         float64(kbGlobalConfig.Threshold),
			TopK:              int64(kbGlobalConfig.TopK),
			MatchType:         kbGlobalConfig.MatchType,
			PriorityMatch:     kbGlobalConfig.PriorityMatch,
			SemanticsPriority: float64(kbGlobalConfig.SemanticsPriority),
			KeywordPriority:   float64(kbGlobalConfig.KeywordPriority),
			TermWeight:        float64(kbGlobalConfig.TermWeight),
			TermWeightEnable:  kbGlobalConfig.TermWeightEnable,
			MetaParams:        metaParams,
			UseGraph:          kbGlobalConfig.UseGraph,
		},
		QAKnowledgebaseConfig: qaKnowledgeConfig,
		SensitiveConfig: model.SensitiveConfig{
			Enable:   in.SensitiveConfig.Enable,
			TableIds: sensitiveIds,
		},
		VisionConfig: model.VisionConfig{
			PicNum: in.VisionConfig.PicNum,
		},
		RecommendQuestion: recommendQuestionStr,
	}, cfgUserId, cfgOrgId); err != nil {
		return nil, errStatus(errs.Code_RagUpdateErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteRag(ctx context.Context, in *rag_service.RagDeleteReq) (*emptypb.Empty, error) {
	errDelete := s.cli.DeleteRag(ctx, in)
	if errDelete != nil {
		return nil, errStatus(errs.Code_RagDeleteErr, errDelete)
	}
	// 应用已删除，会话与明细的清理失败只记日志，不把已成功的删除报成失败
	if status := s.cli.DeleteRagConversationByRagID(ctx, in.RagId); status != nil {
		log.Errorf("删除知识问答会话失败，ragId: %s, err: %v", in.RagId, status)
	}
	if err := rag_conversation.DeleteRagDetail(ctx, in.RagId); err != nil {
		log.Errorf("删除知识问答明细失败，ragId: %s, err: %v", in.RagId, err)
	}
	return nil, nil
}

func (s *Service) GetRagDetail(ctx context.Context, in *rag_service.RagDetailReq) (*rag_service.RagInfo, error) {
	userId, orgId := identityOf(in.Identity)
	detail, err := s.searchRagDetail(ctx, in.RagId, in.Version, in.Publish, userId, orgId)
	if err != nil {
		return nil, err
	}
	ret, errMsg := orm.BuildRagInfo(detail)
	if errMsg != nil {
		return &rag_service.RagInfo{}, errStatus(errs.Code_RagGetErr, errMsg)
	}
	return ret, nil
}

func (s *Service) ListRag(ctx context.Context, in *rag_service.RagListReq) (*rag_service.RagListResp, error) {
	ragList, err := s.cli.GetRagList(ctx, in)
	if err != nil {
		return nil, errStatus(errs.Code_RagListErr, err)
	}
	return ragList, nil
}

// AdminRagPageList 管理员中心知识问答全局分页列表（跨用户/组织）
func (s *Service) AdminRagPageList(ctx context.Context, in *rag_service.AdminRagPageListReq) (*rag_service.AdminRagPageListResp, error) {
	ragList, err := s.cli.AdminRagPageList(ctx, in)
	if err != nil {
		return nil, errStatus(errs.Code_RagListErr, err)
	}
	return ragList, nil
}

func (s *Service) GetRagByIds(ctx context.Context, in *rag_service.GetRagByIdsReq) (*rag_service.AppBriefList, error) {
	ragList, err := s.cli.GetRagByIds(ctx, &rag_service.GetRagByIdsReq{
		RagIdList: in.RagIdList,
	})
	if err != nil {
		return nil, errStatus(errs.Code_RagListErr, err)
	}
	return ragList, nil
}

func (s *Service) CopyRag(ctx context.Context, in *rag_service.CopyRagReq) (*rag_service.CreateRagResp, error) {
	userId, orgId := identityOf(in.Identity)
	info, err := s.cli.FetchRagFirst(ctx, in.RagId, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	// 获取序列号
	index, err := s.cli.FetchRagCopyIndex(ctx, info.BriefConfig.Name, in.Identity.UserId, in.Identity.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	// 生成新名字
	replicaName := util.GenCopyName(info.BriefConfig.Name, index)
	replicaId := util.NewID()
	err = s.cli.CreateRag(ctx, &model.RagInfo{
		RagID: replicaId,
		BriefConfig: model.AppBriefConfig{
			Name:       replicaName,
			Desc:       info.BriefConfig.Desc,
			AvatarPath: info.BriefConfig.AvatarPath,
		},
		ModelConfig:           info.ModelConfig,
		RerankConfig:          info.RerankConfig,
		QARerankConfig:        info.QARerankConfig,
		KnowledgeBaseConfig:   info.KnowledgeBaseConfig,
		QAKnowledgebaseConfig: info.QAKnowledgebaseConfig,
		SensitiveConfig:       info.SensitiveConfig,
		VisionConfig:          info.VisionConfig,
		RecommendQuestion:     info.RecommendQuestion,
		PublicModel:           info.PublicModel,
	})
	if err != nil {
		return nil, errStatus(errs.Code_RagCreateErr, err)
	}
	return &rag_service.CreateRagResp{
		RagId: replicaId,
	}, nil
}

func (s *Service) PublishRag(ctx context.Context, req *rag_service.PublishRagReq) (*emptypb.Empty, error) {
	//1.获取rag草稿信息
	pubUserId, pubOrgId := identityOf(req.Identity)
	draft, err := s.cli.FetchRagFirst(ctx, req.RagId, pubUserId, pubOrgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	//2.序列化信息
	var ragInfo string
	ragInfoBytes, errR := json.Marshal(draft)
	if errR != nil {
		log.Errorf("marshal rag err: %s", errR.Error())
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagPublishErr, "rag_publish_err", "marshal err:", errR.Error())
	}
	ragInfo = string(ragInfoBytes)
	//3.存入发布表
	err = s.cli.PublishRag(ctx, &model.RagPublish{
		RagID:       draft.RagID,
		Version:     req.Version,
		Description: req.Desc,
		RagInfo:     ragInfo,
		UserId:      req.Identity.UserId,
		OrgId:       req.Identity.OrgId,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, errStatus(errs.Code_RagPublishErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdatePublishRag(ctx context.Context, req *rag_service.UpdatePublishRagReq) (*emptypb.Empty, error) {
	userId, orgId := identityOf(req.Identity)
	rag, err := s.cli.FetchPublishRagFirst(ctx, req.RagId, "", userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	rag.Description = req.Desc
	err = s.cli.UpdatePublishRag(ctx, rag)
	if err != nil {
		return nil, errStatus(errs.Code_RagUpdateErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ListPublishRagHistory(ctx context.Context, req *rag_service.ListPublishRagHistoryReq) (*rag_service.ListPublishRagHistoryResp, error) {
	historyList := make([]*rag_service.PublishRagHistory, 0)
	userId, orgId := identityOf(req.Identity)
	ragList, err := s.cli.FetchPublishRagList(ctx, req.RagId, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	for _, rag := range ragList {
		historyList = append(historyList, &rag_service.PublishRagHistory{
			RagId:    rag.RagID,
			Version:  rag.Version,
			Desc:     rag.Description,
			CreateAt: rag.CreatedAt,
		})
	}
	return &rag_service.ListPublishRagHistoryResp{
		HistoryList: historyList,
		Total:       int64(len(historyList)),
	}, nil
}

func (s *Service) OverwriteRagDraft(ctx context.Context, req *rag_service.OverwriteRagDraftReq) (*emptypb.Empty, error) {
	//1.获取该版本rag信息
	userId, orgId := identityOf(req.Identity)
	rag, err := s.cli.FetchPublishRagFirst(ctx, req.RagId, req.Version, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	//2.反序列化配置
	ragInfo := &model.RagInfo{}
	if rag.RagInfo != "" {
		errU := json.Unmarshal([]byte(rag.RagInfo), ragInfo)
		if errU != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagUpdateErr, "rag_update_err", errU.Error())
		}
	}
	//3.覆盖草稿（归属已由上面按 identity 取发布快照保证，这里不再过滤）
	err = s.cli.UpdateRagConfig(ctx, ragInfo, "", "")
	if err != nil {
		return nil, errStatus(errs.Code_RagUpdateErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetPublishRagDesc(ctx context.Context, req *rag_service.GetPublishRagDescReq) (*rag_service.GetPublishRagDescResp, error) {
	userId, orgId := identityOf(req.Identity)
	rag, err := s.cli.FetchPublishRagFirst(ctx, req.RagId, "", userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	return &rag_service.GetPublishRagDescResp{
		RagId:    req.RagId,
		Version:  rag.Version,
		Desc:     rag.Description,
		CreateAt: rag.CreatedAt,
	}, nil
}

func (s *Service) GetPublishRagDescBatch(ctx context.Context, req *rag_service.GetPublishRagDescBatchReq) (*rag_service.GetPublishRagDescBatchResp, error) {
	if len(req.RagIdList) == 0 {
		return &rag_service.GetPublishRagDescBatchResp{List: []*rag_service.GetPublishRagDescResp{}}, nil
	}
	userId, orgId := identityOf(req.Identity)
	ragList, err := s.cli.FetchPublishRagListByRagIds(ctx, req.RagIdList, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	// 按 ragId 分组，取每个 ragId 的最新版本
	ragMap := make(map[string]*model.RagPublish)
	for _, rag := range ragList {
		if existing, ok := ragMap[rag.RagID]; !ok || rag.CreatedAt > existing.CreatedAt {
			ragMap[rag.RagID] = rag
		}
	}
	// 构建响应
	resp := make([]*rag_service.GetPublishRagDescResp, 0, len(ragMap))
	for ragId, rag := range ragMap {
		resp = append(resp, &rag_service.GetPublishRagDescResp{
			RagId:    ragId,
			Version:  rag.Version,
			Desc:     rag.Description,
			CreateAt: rag.CreatedAt,
		})
	}
	return &rag_service.GetPublishRagDescBatchResp{List: resp}, nil
}

func (s *Service) searchRagDetail(ctx context.Context, ragId, version string, publish int32, userId, orgId string) (*model.RagInfo, error) {
	// 获取rag详情
	rag := &model.RagInfo{}
	switch publish {
	case RAGPUBLISH:
		publishRag, err := s.cli.FetchPublishRagFirst(ctx, ragId, version, userId, orgId)
		if err != nil {
			return nil, errStatus(errs.Code_RagChatErr, err)
		}
		if publishRag.RagInfo == "" {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", "ragInfo is empty")
		}
		if err := json.Unmarshal([]byte(publishRag.RagInfo), rag); err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err.Error())
		}
	default:
		ragInfo, err := s.cli.FetchRagFirst(ctx, ragId, userId, orgId)
		if err != nil {
			return nil, errStatus(errs.Code_RagChatErr, err)
		}
		rag = ragInfo
	}
	return rag, nil
}

func buildKnowledgeList(ctx context.Context, rag *model.RagInfo) (*knowledgebase_service.KnowledgeDetailSelectListResp, error) {
	// 反序列化字符串
	knowledgeIds, err := buildKnowledgeIdList(rag)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err.Error())
	}
	knowledgeInfoList, err1 := Knowledge.SelectKnowledgeDetailByIdList(ctx, &knowledgebase_service.KnowledgeDetailSelectListReq{
		UserId:       rag.UserID,
		OrgId:        rag.OrgID,
		KnowledgeIds: knowledgeIds,
	})
	if err1 != nil {
		log.Errorf("SelectKnowledgeDetailByIdList err %s", err1.Error())
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err1.Error())
	}
	if knowledgeInfoList == nil || len(knowledgeInfoList.List) == 0 {
		log.Errorf("knowledgeInfoList is empty")
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", "check knowledgeInfoList err: knowledgeInfoList is nil")
	}
	return knowledgeInfoList, nil
}

func buildKnowledgeIdList(rag *model.RagInfo) ([]string, error) {
	// 反序列化字符串
	var knowledgeIds []string
	if len(rag.KnowledgeBaseConfig.KnowId) > 0 {
		errU := json.Unmarshal([]byte(rag.KnowledgeBaseConfig.KnowId), &knowledgeIds)
		if errU != nil {
			return nil, errU
		}
	}
	if len(rag.QAKnowledgebaseConfig) > 0 {
		// 反序列化qaKnowledgeBaseConfig
		qaKnowledgeBaseConfig := &rag_service.RagQAKnowledgeBaseConfig{}
		err := json.Unmarshal([]byte(rag.QAKnowledgebaseConfig), qaKnowledgeBaseConfig)
		if err != nil {
			return nil, err
		}
		for _, qaConfig := range qaKnowledgeBaseConfig.PerKnowledgeConfigs {
			knowledgeIds = append(knowledgeIds, qaConfig.KnowledgeId)
		}
	}
	return knowledgeIds, nil
}

// 拆分知识库列表
func splitKnowledgeIdList(knowledgeList *knowledgebase_service.KnowledgeDetailSelectListResp) (knowledgeIds []string, qaIds []string, knowledgeIDToName map[string]string) {
	knowledgeIDToName = make(map[string]string)
	for _, info := range knowledgeList.List {
		if info.Category == QACategory {
			qaIds = append(qaIds, info.KnowledgeId)
		} else {
			knowledgeIds = append(knowledgeIds, info.KnowledgeId)
		}
		if _, exists := knowledgeIDToName[info.KnowledgeId]; !exists {
			knowledgeIDToName[info.KnowledgeId] = info.RagName
		}
	}
	return
}
