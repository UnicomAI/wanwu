package knowledge_splitter

import (
	"context"
	"fmt"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_splitter_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-splitter-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/generator"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	SplitterPreset = "preset"
	SplitterCustom = "custom"
)

func (s *Service) SelectKnowledgeSplitterList(ctx context.Context, req *knowledgebase_splitter_service.KnowledgeSplitterSelectReq) (*knowledgebase_splitter_service.KnowledgeSplitterSelectListResp, error) {
	customSplitterList, err := orm.SelectKnowledgeSplitterList(ctx, req.UserId, req.OrgId, req.SplitterName)
	if err != nil {
		log.Errorf(fmt.Sprintf("获取知识库分隔符列表失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeSplitterSelectFailed)
	}

	configSplitterList := config.GetConfig().SplitterList
	var presetSplitterList []*model.KnowledgeSplitter
	for _, v := range configSplitterList {
		// 搜索条件 [EN] Search criteria
		if req.SplitterName != "" {
			if strings.Contains(v.Name, req.SplitterName) {
				presetSplitterList = append(presetSplitterList, &model.KnowledgeSplitter{
					Name:  v.Name,
					Value: v.Value,
				})
			}
		} else {
			presetSplitterList = append(presetSplitterList, &model.KnowledgeSplitter{
				Name:  v.Name,
				Value: v.Value,
			})
		}

	}
	return buildKnowledgeSplitterListResp(customSplitterList, presetSplitterList), nil
}

func (s *Service) CreateKnowledgeSplitter(ctx context.Context, req *knowledgebase_splitter_service.CreateKnowledgeSplitterReq) (*emptypb.Empty, error) {
	//1.重名校验 [EN] 1. Duplicate name verification
	err := orm.CheckSameKnowledgeSplitterNameOrValue(ctx, req.UserId, req.OrgId, req.SplitterName, req.SplitterValue)
	if err != nil {
		return nil, err
	}
	//2.创建创建知识库分隔符 [EN] 2. Create a knowledge base separator
	splitterModel := buildKnowledgeSplitterModel(req)
	err = orm.CreateKnowledgeSplitter(ctx, splitterModel)
	if err != nil {
		log.Errorf("CreateKnowledgeSplitter error %v params %v", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeSplitterCreateFailed)
	}
	//3.返回结果 [EN] 3.Return results
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateKnowledgeSplitter(ctx context.Context, req *knowledgebase_splitter_service.UpdateKnowledgeSplitterReq) (*emptypb.Empty, error) {
	//1.查询知识库分隔符详情 [EN] 1. Query the knowledge base separator details
	knowledgeTag, err := orm.SelectKnowledgeSplitterDetail(ctx, req.UserId, req.OrgId, req.SplitterId)
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该分隔符的权限 参数(%v)", req))
		return nil, util.ErrCode(errs.Code_KnowledgeSplitterAccessDenied)
	}
	//2.重名校验 [EN] 2. Duplicate name verification
	if knowledgeTag.Name == req.SplitterName {
		//如何修改得名称和原名称一样无需修改 [EN] How to modify the name so that it is the same as the original name without modification
		return &emptypb.Empty{}, nil
	}
	err = orm.CheckSameKnowledgeSplitterNameOrValue(ctx, req.UserId, req.OrgId, req.SplitterName, req.SplitterValue)
	if err != nil {
		return nil, err
	}
	//3.更新知识库 [EN] 3. Update knowledge base
	err = orm.UpdateKnowledgeSplitter(ctx, req.SplitterName, req.SplitterValue, knowledgeTag.Id)
	if err != nil {
		log.Errorf("知识库分隔符更新失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeSplitterUpdateFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteKnowledgeSplitter(ctx context.Context, req *knowledgebase_splitter_service.DeleteKnowledgeSplitterReq) (*emptypb.Empty, error) {
	//1.查询知识库分隔符详情 [EN] 1. Query the knowledge base separator details
	knowledgeTag, err := orm.SelectKnowledgeSplitterDetail(ctx, req.UserId, req.OrgId, req.SplitterId)
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该分隔符的权限 参数(%v)", req))
		return nil, err
	}
	//2.删除知识库分隔符 [EN] 2. Remove the knowledge base separator
	err = orm.DeleteKnowledgeSplitter(ctx, knowledgeTag.Id)
	if err != nil {
		log.Errorf("知识库分隔符删除失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeSplitterDeleteFailed)
	}
	return &emptypb.Empty{}, nil
}

// buildKnowledgeSplitterListResp 构造知识库分隔符列表返回结果 [EN] buildKnowledgeSplitterListResp constructs a knowledge base separator list and returns the results
func buildKnowledgeSplitterListResp(customSplitters, presetSplitters []*model.KnowledgeSplitter) *knowledgebase_splitter_service.KnowledgeSplitterSelectListResp {
	var retList []*knowledgebase_splitter_service.KnowledgeSplitterInfo
	for _, splitter := range customSplitters {
		retList = append(retList, buildKnowledgeSplitter(splitter, SplitterCustom))
	}
	for _, splitter := range presetSplitters {
		retList = append(retList, buildKnowledgeSplitter(splitter, SplitterPreset))
	}
	return &knowledgebase_splitter_service.KnowledgeSplitterSelectListResp{
		KnowledgeSplitterList: retList,
	}
}

// buildKnowledgeSplitter 构造知识库tag [EN] buildKnowledgeSplitter constructs knowledge base tags
func buildKnowledgeSplitter(knowledgeSplitter *model.KnowledgeSplitter, splitterType string) *knowledgebase_splitter_service.KnowledgeSplitterInfo {
	return &knowledgebase_splitter_service.KnowledgeSplitterInfo{
		SplitterId:    knowledgeSplitter.SplitterId,
		SplitterName:  knowledgeSplitter.Name,
		SplitterValue: knowledgeSplitter.Value,
		Type:          splitterType,
	}
}

// buildKnowledgeSplitterModel 构造知识库分隔符模型 [EN] buildKnowledgeSplitterModel constructs the knowledge base separator model
func buildKnowledgeSplitterModel(req *knowledgebase_splitter_service.CreateKnowledgeSplitterReq) *model.KnowledgeSplitter {
	return &model.KnowledgeSplitter{
		SplitterId: generator.GetGenerator().NewID(),
		Name:       req.SplitterName,
		Value:      req.SplitterValue,
		OrgId:      req.OrgId,
		UserId:     req.UserId,
		CreatedAt:  time.Now().UnixMilli(),
		UpdatedAt:  time.Now().UnixMilli(),
	}
}
