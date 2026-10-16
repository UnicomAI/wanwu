package knowledge_keywords

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_keywords_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-keywords-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// GetKnowledgeKeywordsList 返回关键词列表
func (s *Service) GetKnowledgeKeywordsList(ctx context.Context, req *knowledgebase_keywords_service.GetKnowledgeKeywordsListReq) (*knowledgebase_keywords_service.GetKnowledgeKeywordsListResp, error) {
	// 查询关键词列表
	keywordsList, total, err := orm.GetKeywordsList(ctx, req)
	if err != nil {
		log.Errorf(fmt.Sprintf("GetKnowledgeKeywordsList 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsListFailed)
	}
	// 构造返回体
	keywordsInfoList, err := buildKeywordsInfoList(ctx, keywordsList)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsListFailed)
	}
	resp := &knowledgebase_keywords_service.GetKnowledgeKeywordsListResp{
		Keywords: keywordsInfoList,
		Total:    total,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	}
	return resp, nil
}

// GetKnowledgeKeywordsListByKnowledgeId 返回指定知识库下的关键词列表
func (s *Service) GetKnowledgeKeywordsListByKnowledgeId(ctx context.Context, req *knowledgebase_keywords_service.GetKnowledgeKeywordsListByKnowledgeIdReq) (*knowledgebase_keywords_service.GetKnowledgeKeywordsListByKnowledgeIdResp, error) {
	keywordsInfoList, err := orm.GetKeywordsListByKnowledgeId(ctx, req.KnowledgeId, req.UserId, req.OrgId)
	if err != nil {
		log.Errorf(fmt.Sprintf("GetKnowledgeKeywordsListByKnowledgeId 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsListFailed)
	}
	resp := &knowledgebase_keywords_service.GetKnowledgeKeywordsListByKnowledgeIdResp{
		Keywords: keywordsInfoList,
	}
	return resp, nil
}

func buildKeywordsInfoList(ctx context.Context, keywordsList []*model.KnowledgeKeywords) ([]*knowledgebase_keywords_service.KeywordsInfo, error) {
	var keywordsInfoList []*knowledgebase_keywords_service.KeywordsInfo
	for _, k := range keywordsList {
		// 获取知识库信息
		knowledgeIds, knowledgeNames, err := GetKnowledgeInfo(ctx, k)
		if err != nil {
			return nil, err
		}
		keywordsInfo := &knowledgebase_keywords_service.KeywordsInfo{
			Id:                 k.Id,
			Name:               k.Name,
			Alias:              k.Alias,
			KnowledgeBaseIds:   knowledgeIds,
			KnowledgeBaseNames: knowledgeNames,
			UpdatedAt:          wanwu_util.Time2Str(k.UpdatedAt),
		}
		if keywordsInfo.KnowledgeBaseNames == nil {
			continue
		}
		keywordsInfoList = append(keywordsInfoList, keywordsInfo)
	}
	return keywordsInfoList, nil
}

func buildKeywordsInfo(ctx context.Context, keywords *model.KnowledgeKeywords) (*knowledgebase_keywords_service.KeywordsInfo, error) {
	// 获取知识库信息
	knowledgeIds, knowledgeNames, err := GetKnowledgeInfo(ctx, keywords)
	if err != nil {
		return nil, err
	}
	keywordsInfo := &knowledgebase_keywords_service.KeywordsInfo{
		Id:                 keywords.Id,
		Name:               keywords.Name,
		Alias:              keywords.Alias,
		KnowledgeBaseIds:   knowledgeIds,
		KnowledgeBaseNames: knowledgeNames,
		UpdatedAt:          strconv.FormatInt(keywords.UpdatedAt, 10)}
	return keywordsInfo, nil
}

func GetKnowledgeInfo(ctx context.Context, k *model.KnowledgeKeywords) ([]string, []string, error) {
	// 反序列化id列表
	var knowledgeIds []string
	err := json.Unmarshal([]byte(k.KnowledgeBaseIds), &knowledgeIds)
	if err != nil {
		log.Errorf("反序列化错误")
		return nil, nil, err
	}
	// 根据id获取知识库列表
	knowledgeList, _, errk := orm.SelectKnowledgeByIdList(ctx, knowledgeIds, k.UserId, k.OrgId)
	if errk != nil {
		log.Errorf("查询知识库名称失败")
		return nil, nil, errk
	}
	// 构造知识库名字列表
	var knowledgeNames []string
	for _, v := range knowledgeList {
		knowledgeNames = append(knowledgeNames, v.Name)
	}
	return knowledgeIds, knowledgeNames, nil
}

// GetKnowledgeKeywordsDetail 返回关键词信息
func (s *Service) GetKnowledgeKeywordsDetail(ctx context.Context, req *knowledgebase_keywords_service.GetKnowledgeKeywordsDetailReq) (*knowledgebase_keywords_service.GetKnowledgeKeywordsDetailResp, error) {
	// 查询关键词
	keywords, err := orm.GetKeywordsById(ctx, req.Id)
	if err != nil {
		log.Errorf(fmt.Sprintf("GetKnowledgeKeywords 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsInfoFailed)
	}
	// 校验关键词归属：仅允许访问本租户(org+user)的关键词，避免按 id 跨租户越权读取。
	// 列表/重复校验接口已通过 WithPermit(org,user) 做此限制，此 detail 接口此前遗漏。
	if keywords.OrgId != req.Identity.OrgId || keywords.UserId != req.Identity.UserId {
		log.Errorf(fmt.Sprintf("GetKnowledgeKeywordsDetail 跨租户越权: keyword(org=%v,user=%v) caller(org=%v,user=%v)",
			keywords.OrgId, keywords.UserId, req.Identity.OrgId, req.Identity.UserId))
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsInfoFailed)
	}
	// 构造返回体
	keywordsInfo, err := buildKeywordsInfo(ctx, keywords)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsInfoFailed)
	}
	resp := &knowledgebase_keywords_service.GetKnowledgeKeywordsDetailResp{
		Detail: keywordsInfo,
	}
	return resp, nil
}

func buildKeywordsModel(req *knowledgebase_keywords_service.CreateKnowledgeKeywordsReq, id uint32) (*model.KnowledgeKeywords, error) {
	// 序列化字符串列表
	idJsonBytes, err := json.Marshal(req.KnowledgeBaseIds)
	if err != nil {
		return nil, err
	}
	knowledgeBaseIds := string(idJsonBytes)
	knowledgeKeywords := &model.KnowledgeKeywords{
		Name:             req.Name,
		Alias:            req.Alias,
		KnowledgeBaseIds: knowledgeBaseIds,
		UserId:           req.Identity.UserId,
		OrgId:            req.Identity.OrgId,
	}
	if id != 0 {
		knowledgeKeywords.Id = id
	}
	return knowledgeKeywords, nil
}

// CreateKnowledgeKeywords 新增关键词
func (s *Service) CreateKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.CreateKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	// 检查有无同名关键词
	err := orm.CheckRepeatedKeywords(ctx, &knowledgebase_keywords_service.UpdateKnowledgeKeywordsReq{
		Id:     0,
		Detail: req,
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsRepeated)
	} else if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	// 构建关键词结构
	knowledgeKeywords, err := buildKeywordsModel(req, 0)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	// 创建关键词
	err = orm.CreateKeywords(ctx, knowledgeKeywords)
	if err != nil {
		log.Errorf("创建关键词失败")
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	return nil, nil
}

// DeleteKnowledgeKeywords 删除关键词
func (s *Service) DeleteKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.DeleteKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	err := orm.DeleteKeywords(ctx, req.Id)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsDeleteFailed)
	}
	return nil, nil
}

// UpdateKnowledgeKeywords 更新关键词
func (s *Service) UpdateKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.UpdateKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	// 检查有无同名关键词
	err := orm.CheckRepeatedKeywords(ctx, req)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsRepeated)
	} else if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsUpdateFailed)
	}
	// 更新关键词
	knowledgeKeywords, err := buildKeywordsModel(req.Detail, req.Id)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsUpdateFailed)
	}
	err = orm.UpdateKeywords(ctx, knowledgeKeywords)
	if err != nil {
		log.Errorf("更新关键词失败")
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsUpdateFailed)
	}
	return nil, nil
}
