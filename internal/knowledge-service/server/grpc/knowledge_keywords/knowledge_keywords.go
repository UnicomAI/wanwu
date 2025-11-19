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

// GetKnowledgeKeywordsList returns the keyword list
func (s *Service) GetKnowledgeKeywordsList(ctx context.Context, req *knowledgebase_keywords_service.GetKnowledgeKeywordsListReq) (*knowledgebase_keywords_service.GetKnowledgeKeywordsListResp, error) {
	// Query keyword list
	keywordsList, total, err := orm.GetKeywordsList(ctx, req)
	if err != nil {
		log.Errorf(fmt.Sprintf("GetKnowledgeKeywordsList 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsListFailed)
	}
	// Construct return body
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
		// Get knowledge base information
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
	// Get knowledge base information
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
	// Deserialize id list
	var knowledgeIds []string
	err := json.Unmarshal([]byte(k.KnowledgeBaseIds), &knowledgeIds)
	if err != nil {
		log.Errorf("反序列化错误")
		return nil, nil, err
	}
	// Get the knowledge base list based on id
	knowledgeList, _, errk := orm.SelectKnowledgeByIdList(ctx, knowledgeIds, k.UserId, k.OrgId)
	if errk != nil {
		log.Errorf("查询知识库名称失败")
		return nil, nil, errk
	}
	// Construct a list of knowledge base names
	var knowledgeNames []string
	for _, v := range knowledgeList {
		knowledgeNames = append(knowledgeNames, v.Name)
	}
	return knowledgeIds, knowledgeNames, nil
}

// GetKnowledgeKeywordsDetail returns keyword information
func (s *Service) GetKnowledgeKeywordsDetail(ctx context.Context, req *knowledgebase_keywords_service.GetKnowledgeKeywordsDetailReq) (*knowledgebase_keywords_service.GetKnowledgeKeywordsDetailResp, error) {
	// Query keywords
	keywords, err := orm.GetKeywordsById(ctx, req.Id)
	if err != nil {
		log.Errorf(fmt.Sprintf("GetKnowledgeKeywords 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsInfoFailed)
	}
	// Construct return body
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
	// Serialized list of strings
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

// CreateKnowledgeKeywords New keywords
func (s *Service) CreateKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.CreateKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	// Check whether there are keywords with the same name
	err := orm.CheckRepeatedKeywords(ctx, &knowledgebase_keywords_service.UpdateKnowledgeKeywordsReq{
		Id:     0,
		Detail: req,
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsRepeated)
	} else if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	// Build keyword structure
	knowledgeKeywords, err := buildKeywordsModel(req, 0)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	// Create keywords
	err = orm.CreateKeywords(ctx, knowledgeKeywords)
	if err != nil {
		log.Errorf("创建关键词失败")
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsCreateFailed)
	}
	return nil, nil
}

// DeleteKnowledgeKeywords Delete keywords
func (s *Service) DeleteKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.DeleteKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	err := orm.DeleteKeywords(ctx, req.Id)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsDeleteFailed)
	}
	return nil, nil
}

// UpdateKnowledgeKeywords update keywords
func (s *Service) UpdateKnowledgeKeywords(ctx context.Context, req *knowledgebase_keywords_service.UpdateKnowledgeKeywordsReq) (*emptypb.Empty, error) {
	// Check whether there are keywords with the same name
	err := orm.CheckRepeatedKeywords(ctx, req)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsRepeated)
	} else if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeKeywordsUpdateFailed)
	}
	// Update keywords
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
