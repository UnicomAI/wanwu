package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/orm"
	message_builder "github.com/UnicomAI/wanwu/internal/rag-service/service/message-builder"
	"github.com/UnicomAI/wanwu/pkg/db"
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
	return grpc_util.ErrorStatusWithKey(code, status.TextKey, status.Args...)
}

func (s *Service) ChatRag(req *rag_service.ChatRagReq, stream grpc.ServerStreamingServer[rag_service.ChatRagResp]) error {
	ctx := stream.Context()
	// 获取rag详情
	rag := &model.RagInfo{}
	switch req.Publish {
	case RAGPUBLISH:
		publishRag, err := s.cli.FetchPublishRagFirst(ctx, req.RagId, "")
		if err != nil {
			return errStatus(errs.Code_RagChatErr, err)
		}
		if publishRag.RagInfo == "" {
			return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", "ragInfo is empty")
		}
		if err := json.Unmarshal([]byte(publishRag.RagInfo), rag); err != nil {
			return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err.Error())
		}
	default:
		ragInfo, err := s.cli.FetchRagFirst(ctx, req.RagId)
		if err != nil {
			return errStatus(errs.Code_RagChatErr, err)
		}
		rag = ragInfo
	}
	log.Infof("get rag: %+v", http_client.Convert2LogString(rag))
	// 反序列化字符串
	knowledgeIds, err1 := buildKnowledgeIdList(rag)
	if err1 != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err1.Error())
	}
	knowledgeInfoList, errk := Knowledge.SelectKnowledgeDetailByIdList(ctx, &knowledgebase_service.KnowledgeDetailSelectListReq{
		UserId:       rag.UserID,
		OrgId:        rag.OrgID,
		KnowledgeIds: knowledgeIds,
	})
	if errk != nil {
		log.Errorf("errk = %s", errk.Error())
		return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", errk.Error())
	}
	if knowledgeInfoList == nil || len(knowledgeInfoList.List) == 0 {
		log.Errorf("knowledgeInfoList = nil")
		return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", "check knowledgeInfoList err: knowledgeInfoList is nil")
	}
	knowledgeIds, qaIds, knowledgeIDToName := splitKnowledgeIdList(knowledgeInfoList)
	return message_builder.BuildMessage(ctx, &message_builder.RagContext{
		MessageId:         util.NewID(),
		Req:               req,
		Rag:               rag,
		KnowledgeIDToName: knowledgeIDToName,
		KnowledgeIds:      knowledgeIds,
		QAIds:             qaIds,
	}, stream)
	//  请求rag
	//buildParams, errk := rag_manage_service.BuildChatConsultParams(req, rag, knowledgeInfoList, knowledgeIds)
	//if errk != nil {
	//	log.Errorf("errk = %s", errk.Error())
	//	return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", errk.Error())
	//}
	//chatChan, errg := rag_manage_service.RagStreamChat(ctx, rag.UserID, buildParams)
	//if errg != nil {
	//	return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", errg.Error())
	//}
	//for text := range chatChan {
	//	resp := &rag_service.ChatRagResp{
	//		Content: text,
	//	}
	//	if err := stream.Send(resp); err != nil {
	//		return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err.Error())
	//	}
	//}
	//return nil
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
	originalRag, err := s.cli.FetchRagFirst(ctx, in.RagId)
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
	}); err != nil {
		return nil, errStatus(errs.Code_RagUpdateErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateRagConfig(ctx context.Context, in *rag_service.UpdateRagConfigReq) (*emptypb.Empty, error) {
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

	if in.QAknowledgeBaseConfig == nil {
		in.QAknowledgeBaseConfig = &rag_service.RagQAKnowledgeBaseConfig{}
	}
	qaConfig := in.QAknowledgeBaseConfig
	if qaConfig.GlobalConfig == nil {
		qaConfig.GlobalConfig = &rag_service.RagQAGlobalConfig{}
	}
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
		QAKnowledgebaseConfig: db.LongText(qaKnowledgeConfig),
		SensitiveConfig: model.SensitiveConfig{
			Enable:   in.SensitiveConfig.Enable,
			TableIds: sensitiveIds,
		},
	}); err != nil {
		return nil, errStatus(errs.Code_RagUpdateErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteRag(ctx context.Context, in *rag_service.RagDeleteReq) (*emptypb.Empty, error) {
	errDelete := s.cli.DeleteRag(ctx, in)
	if errDelete != nil {
		return nil, errStatus(errs.Code_RagDeleteErr, errDelete)
	}
	return nil, nil
}

func (s *Service) GetRagDetail(ctx context.Context, in *rag_service.RagDetailReq) (*rag_service.RagInfo, error) {
	switch in.Publish {
	case RAGDRAFT:
		info, err := s.cli.GetRag(ctx, in)
		if err != nil {
			return nil, errStatus(errs.Code_RagGetErr, err)
		}
		return info, nil
	default:
		rag, err := s.cli.FetchPublishRagFirst(ctx, in.RagId, in.Version)
		if err != nil {
			return nil, errStatus(errs.Code_RagGetErr, err)
		}
		log.Infof("publish_rag = %+v", http_client.Convert2LogString(rag))
		ragInfo := &model.RagInfo{}
		if err := json.Unmarshal([]byte(rag.RagInfo), ragInfo); err != nil {
			return &rag_service.RagInfo{}, grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err.Error())
		}
		log.Infof("ragInfo = %+v", http_client.Convert2LogString(ragInfo))
		ret, err := orm.BuildRagInfo(ragInfo)
		if err != nil {
			return &rag_service.RagInfo{}, errStatus(errs.Code_RagGetErr, err)
		}
		return ret, nil
	}
}

func (s *Service) ListRag(ctx context.Context, in *rag_service.RagListReq) (*rag_service.RagListResp, error) {
	ragList, err := s.cli.GetRagList(ctx, in)
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
	info, err := s.cli.FetchRagFirst(ctx, in.RagId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	index, err := s.cli.FetchRagCopyIndex(ctx, info.BriefConfig.Name, in.Identity.UserId, in.Identity.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_RagGetErr, err)
	}
	replicaName := fmt.Sprintf("%s_%d", info.BriefConfig.Name, index)
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
	draft, err := s.cli.FetchRagFirst(ctx, req.RagId)
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
	rag, err := s.cli.FetchPublishRagFirst(ctx, req.RagId, "")
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
	ragList, err := s.cli.FetchPublishRagList(ctx, req.RagId)
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
	rag, err := s.cli.FetchPublishRagFirst(ctx, req.RagId, req.Version)
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
	//3.覆盖草稿
	err = s.cli.UpdateRagConfig(ctx, ragInfo)
	if err != nil {
		return nil, errStatus(errs.Code_RagUpdateErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetPublishRagDesc(ctx context.Context, req *rag_service.GetPublishRagDescReq) (*rag_service.GetPublishRagDescResp, error) {
	rag, err := s.cli.FetchPublishRagFirst(ctx, req.RagId, "")
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
