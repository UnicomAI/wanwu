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

// GetKnowledgeKeywordsList 返回关键词列表 [EN] GetKnowledgeKeywordsList returns the keyword list
func (s *Service) GetKnowledgeKeywordsList(ctx context.Context, req *knowledgebase_keywords_service.GetKnowledgeKeywordsListReq) (*knowledgebase_keywords_service.GetKnowledgeKeywordsListResp, error) {
	// 查询关键词列表 [EN] Query keyword list
	keywordsList, total, err := orm.GetKeywordsList(ctx, req)
	if err != nil {
		log.Errorf(fmt.Sprintf("GetKnowledgeKeywordsList 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsListFailed)
	}
	// 构造返回体 [EN] Construct return body
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

func buildKeywordsInfoList(ctx context.Context, keywordsList []*model.KnowledgeKeywords) ([]*knowledgebase_keywords_service.KeywordsInfo, error) {
	var keywordsInfoList []*knowledgebase_keywords_service.KeywordsInfo
	for _, k := range keywordsList {
		// 获取知识库信息 [EN] Get knowledge base information
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
		keywordsInfoList = append(keywordsInfoList, keywordsInfo)
	}
	return keywordsInfoList, nil
}

func buildKeywordsInfo(ctx context.Context, keywords *model.KnowledgeKeywords) (*knowledgebase_keywords_service.KeywordsInfo, error) {
	// 获取知识库信息 [EN] Get knowledge base information
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
	// 反序列化id列表 [EN] Deserialize id list
	var knowledgeIds []string
	err := json.Unmarshal([]byte(k.KnowledgeBaseIds), &knowledgeIds)
	if err != nil {
		log.Errorf("反序列化错误")
		return nil, nil, err
	}
	// 根据id获取知识库列表 [EN] Get the knowledge base list based on id
	knowledgeList, _, errk := orm.SelectKnowledgeByIdList(ctx, knowledgeIds, k.UserId, k.OrgId)
	if errk != nil {
		log.Errorf("查询知识库名称失败")
		return nil, nil, errk
	}
	// 构造知识库名字列表 [EN] Construct a list of knowledge base names
	var knowledgeNames []string
	for _, v := range knowledgeList {
		knowledgeNames = append(knowledgeNames, v.Name)
	}
	return knowledgeIds, knowledgeNames, nil
}

// GetKnowledgeKeywordsDetail 返回关键词信息 [EN] GetKnowledgeKeywordsDetail returns keyword information
func (s *Service) GetKnowledgeKeywordsDetail(ctx context.Context, req *knowledgebase_keywords_service.GetKnowledgeKeywordsDetailReq) (*knowledgebase_keywords_service.GetKnowledgeKeywordsDetailResp, error) {
	// 查询关键词 [EN] Query keywords
	keywords, err := orm.GetKeywordsById(ctx, req.Id)
	if err != nil {
		log.Errorf(fmt.Sprintf("GetKnowledgeKeywords 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsInfoFailed)
	}
	// 构造返回体 [EN] Construct return body
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
	// 序列化字符串列表 [EN] Serialized list of strings
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

// CreateKnowledgeKeywords 新增关键词 [EN] CreateKnowledgeKeywords New keywords
func (s *Service) CreateKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.CreateKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	// 检查有无同名关键词 [EN] Check whether there are keywords with the same name
	err := orm.CheckRepeatedKeywords(ctx, &knowledgebase_keywords_service.UpdateKnowledgeKeywordsReq{
		Id:     0,
		Detail: req,
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsRepeated)
	} else if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	// 构建关键词结构 [EN] Build keyword structure
	knowledgeKeywords, err := buildKeywordsModel(req, 0)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	// 创建关键词 [EN] Create keywords
	err = orm.CreateKeywords(ctx, knowledgeKeywords)
	if err != nil {
		log.Errorf("创建关键词失败")
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	return nil, nil
}

// DeleteKnowledgeKeywords 删除关键词 [EN] DeleteKnowledgeKeywords Delete keywords
func (s *Service) DeleteKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.DeleteKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	err := orm.DeleteKeywords(ctx, req.Id)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsDeleteFailed)
	}
	return nil, nil
}

// UpdateKnowledgeKeywords 更新关键词 [EN] UpdateKnowledgeKeywords update keywords
func (s *Service) UpdateKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.UpdateKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	// 检查有无同名关键词 [EN] Check whether there are keywords with the same name
	err := orm.CheckRepeatedKeywords(ctx, req)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsRepeated)
	} else if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsUpdateFailed)
	}
	// 更新关键词 [EN] Update keywords
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
