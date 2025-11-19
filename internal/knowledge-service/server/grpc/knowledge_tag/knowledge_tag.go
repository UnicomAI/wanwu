package knowledge_tag

import (
	"context"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_tag_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-tag-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/generator"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) SelectKnowledgeTagList(ctx context.Context, req *knowledgebase_tag_service.KnowledgeTagSelectReq) (*knowledgebase_tag_service.KnowledgeTagSelectListResp, error) {
	if len(req.KnowledgeId) > 0 {
		relation := orm.SelectKnowledgeTagListWithRelation(ctx, req.UserId, req.OrgId, req.TagName, []string{req.KnowledgeId})
		if relation.TagErr != nil {
			log.Errorf(fmt.Sprintf("获取知识库标签列表失败(%v)  参数(%v)", relation.TagErr, req))
			return nil, util.ErrCode(errs.Code_KnowledgeTagSelectFailed)
		}
		return buildKnowledgeTagListRespByRelation(relation), nil
	} else {
		tagList, err := orm.SelectKnowledgeTagList(ctx, req.UserId, req.OrgId, req.TagName)
		if err != nil {
			log.Errorf(fmt.Sprintf("获取知识库标签列表失败(%v)  参数(%v)", err, req))
			return nil, util.ErrCode(errs.Code_KnowledgeTagSelectFailed)
		}
		return buildKnowledgeTagListResp(tagList), nil
	}

}

func (s *Service) CreateKnowledgeTag(ctx context.Context, req *knowledgebase_tag_service.CreateKnowledgeTagReq) (*knowledgebase_tag_service.CreateKnowledgeTagResp, error) {
	//1. Duplicate name verification
	err := orm.CheckSameKnowledgeTagName(ctx, req.UserId, req.OrgId, req.TagName)
	if err != nil {
		return nil, err
	}
	//2. Create a knowledge base
	tagModel := buildKnowledgeTagModel(req)
	err = orm.CreateKnowledgeTag(ctx, tagModel)
	if err != nil {
		log.Errorf("CreateKnowledgeTag error %v params %v", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeTagCreateFailed)
	}
	//3.Return results
	return &knowledgebase_tag_service.CreateKnowledgeTagResp{
		TagId: tagModel.TagId,
	}, nil
}

func (s *Service) UpdateKnowledgeTag(ctx context.Context, req *knowledgebase_tag_service.UpdateKnowledgeTagReq) (*emptypb.Empty, error) {
	//1. Query knowledge base tag details
	knowledgeTag, err := orm.SelectKnowledgeTagDetail(ctx, req.UserId, req.OrgId, req.TagId)
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该标签的权限 参数(%v)", req))
		return nil, util.ErrCode(errs.Code_KnowledgeTagAccessDenied)
	}
	//2. Duplicate name verification
	if knowledgeTag.Name == req.TagName {
		//How to modify the name so that it is the same as the original name without modification
		return &emptypb.Empty{}, nil
	}
	err = orm.CheckSameKnowledgeTagName(ctx, req.UserId, req.OrgId, req.TagName)
	if err != nil {
		return nil, err
	}
	//3. Update knowledge base
	err = orm.UpdateKnowledgeTag(ctx, req.TagName, knowledgeTag.Id)
	if err != nil {
		log.Errorf("知识库标签更新失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeTagUpdateFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteKnowledgeTag(ctx context.Context, req *knowledgebase_tag_service.DeleteKnowledgeTagReq) (*emptypb.Empty, error) {
	//1. Query knowledge base tag details
	knowledgeTag, err := orm.SelectKnowledgeTagDetail(ctx, req.UserId, req.OrgId, req.TagId)
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该标签的权限 参数(%v)", req))
		return nil, err
	}
	//2. Delete knowledge base tags
	err = orm.DeleteKnowledgeTag(ctx, req.TagId, knowledgeTag.Id)
	if err != nil {
		log.Errorf("知识库标签删除失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeTagDeleteFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) BindKnowledgeTag(ctx context.Context, req *knowledgebase_tag_service.BindKnowledgeTagReq) (*emptypb.Empty, error) {
	err := orm.BindKnowledgeTag(ctx, buildKnowledgeTagRelationModelList(req), req.KnowledgeId)
	if err != nil {
		log.Errorf("BindKnowledgeTag error %v params %v", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeTagBindFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) TagBindCount(ctx context.Context, req *knowledgebase_tag_service.TagBindCountReq) (*knowledgebase_tag_service.TagBindCountResp, error) {
	//1. Query knowledge base tag details
	knowledgeTag, err := orm.SelectKnowledgeTagDetail(ctx, req.UserId, req.OrgId, req.TagId)
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该标签的权限 参数(%v)", req))
		return nil, util.ErrCode(errs.Code_KnowledgeTagAccessDenied)
	}
	count, _ := orm.SelectKnowledgeCountByTagId(ctx, knowledgeTag.TagId)
	return &knowledgebase_tag_service.TagBindCountResp{BindCount: count}, nil
}

// buildKnowledgeTagListResp constructs a knowledge base tag list and returns the results
func buildKnowledgeTagListResp(knowledgeTagList []*model.KnowledgeTag) *knowledgebase_tag_service.KnowledgeTagSelectListResp {
	if len(knowledgeTagList) == 0 {
		return &knowledgebase_tag_service.KnowledgeTagSelectListResp{}
	}
	var retList []*knowledgebase_tag_service.KnowledgeTagInfo
	for _, knowledgeTag := range knowledgeTagList {
		retList = append(retList, buildKnowledgeTag(knowledgeTag, false))
	}
	return &knowledgebase_tag_service.KnowledgeTagSelectListResp{
		KnowledgeTagList: retList,
	}
}

// buildKnowledgeTagListRespByRelation constructs a knowledge base tag list and returns the results
func buildKnowledgeTagListRespByRelation(tagRelation *orm.TagRelation) *knowledgebase_tag_service.KnowledgeTagSelectListResp {
	if len(tagRelation.TagList) == 0 {
		return &knowledgebase_tag_service.KnowledgeTagSelectListResp{}
	}
	var retList []*knowledgebase_tag_service.KnowledgeTagInfo
	//Only the specified knowledge base ID can be entered here, so the map that directly combines tagId and knowledgeId will not be overwritten.
	var relationMap = make(map[string]string)
	list := tagRelation.RelationList
	if len(list) > 0 {
		for _, relation := range list {
			relationMap[relation.TagId] = relation.KnowledgeId
		}
	}
	for _, knowledgeTag := range tagRelation.TagList {
		retList = append(retList, buildKnowledgeTag(knowledgeTag, len(relationMap[knowledgeTag.TagId]) > 0))
	}
	return &knowledgebase_tag_service.KnowledgeTagSelectListResp{
		KnowledgeTagList: retList,
	}
}

// buildKnowledgeTag constructs knowledge base tag
func buildKnowledgeTag(knowledgeTag *model.KnowledgeTag, selected bool) *knowledgebase_tag_service.KnowledgeTagInfo {
	return &knowledgebase_tag_service.KnowledgeTagInfo{
		TagId:    knowledgeTag.TagId,
		TagName:  knowledgeTag.Name,
		Selected: selected,
	}
}

// buildKnowledgeTagModel constructs the knowledge base tag model
func buildKnowledgeTagModel(req *knowledgebase_tag_service.CreateKnowledgeTagReq) *model.KnowledgeTag {
	return &model.KnowledgeTag{
		TagId:     generator.GetGenerator().NewID(),
		Name:      req.TagName,
		OrgId:     req.OrgId,
		UserId:    req.UserId,
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}
}

// buildKnowledgeTagRelationModel constructs a knowledge base tag relationship model
func buildKnowledgeTagRelationModelList(req *knowledgebase_tag_service.BindKnowledgeTagReq) []*model.KnowledgeTagRelation {
	var retList []*model.KnowledgeTagRelation
	for _, tagId := range req.TagIdList {
		retList = append(retList, &model.KnowledgeTagRelation{
			TagId:       tagId,
			KnowledgeId: req.KnowledgeId,
			OrgId:       req.OrgId,
			UserId:      req.UserId,
			CreatedAt:   time.Now().UnixMilli(),
			UpdatedAt:   time.Now().UnixMilli(),
		})
	}
	return retList
}
