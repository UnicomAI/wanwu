package safety

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/app-service/client"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	safety_service.UnimplementedSafetyServiceServer
	cli client.IClient
}

func NewService(cli client.IClient) *Service {
	return &Service{
		cli: cli,
	}
}

func errStatus(code errs.Code, status *errs.Status) error {
	return grpc_util.ErrorStatusWithKey(code, status.TextKey, status.Args...)
}

func (s *Service) CreateSensitiveWordTable(ctx context.Context, req *safety_service.CreateSensitiveWordTableReq) (*safety_service.CreateSensitiveWordTableResp, error) {
	tableId, err := s.cli.CreateSensitiveWordTable(ctx, req.UserId, req.OrgId, req.TableName, req.Remark)
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	return &safety_service.CreateSensitiveWordTableResp{TableId: tableId}, nil
}

func (s *Service) UpdateSensitiveWordTable(ctx context.Context, req *safety_service.UpdateSensitiveWordTableReq) (*emptypb.Empty, error) {
	err := s.cli.UpdateSensitiveWordTable(ctx, util.MustU32(req.TableId), req.TableName, req.Remark)
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateSensitiveWordTableReply(ctx context.Context, req *safety_service.UpdateSensitiveWordTableReplyReq) (*emptypb.Empty, error) {
	err := s.cli.UpdateSensitiveWordTableReply(ctx, util.MustU32(req.TableId), req.Reply)
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteSensitiveWordTable(ctx context.Context, req *safety_service.DeleteSensitiveWordTableReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteSensitiveWordTable(ctx, util.MustU32(req.TableId))
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetSensitiveWordTableList(ctx context.Context, req *safety_service.GetSensitiveWordTableListReq) (*safety_service.SensitiveWordTables, error) {
	tables, err := s.cli.GetSensitiveWordTableList(ctx, req.UserId, req.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	ret := &safety_service.SensitiveWordTables{
		Total: int64(len(tables)),
	}
	for _, table := range tables {
		ret.List = append(ret.List, toProtoSensitiveWordTable(table))
	}
	return ret, nil
}

func (s *Service) GetSensitiveVocabularyList(ctx context.Context, req *safety_service.GetSensitiveVocabularyListReq) (*safety_service.SensitiveWordVocabularyResp, error) {
	words, count, err := s.cli.GetSensitiveVocabularyList(ctx, util.MustU32(req.TableId), toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	ret := &safety_service.SensitiveWordVocabularyResp{
		Total:    count,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}
	for _, word := range words {
		ret.List = append(ret.List, toProtoSensitiveWordVocabulary(word))
	}
	return ret, nil
}

func (s *Service) UploadSensitiveVocabulary(ctx context.Context, req *safety_service.UploadSensitiveVocabularyReq) (*emptypb.Empty, error) {
	err := s.cli.UploadSensitiveVocabulary(ctx, req.UserId, req.OrgId, req.ImportType, req.Word, req.SensitiveType, req.FilePath, util.MustU32(req.TableId))
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteSensitiveVocabulary(ctx context.Context, req *safety_service.DeleteSensitiveVocabularyReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteSensitiveVocabulary(ctx, util.MustU32(req.TableId), util.MustU32(req.WordId))
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetSensitiveWordTableListWithWordsByIDs(ctx context.Context, req *safety_service.GetSensitiveWordTableListByIDsReq) (*safety_service.SensitiveWordTableListWithWords, error) {
	tables, err := s.cli.GetSensitiveWordTableListWithWordsByIDs(ctx, req.TableIds)
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	ret := &safety_service.SensitiveWordTableListWithWords{}
	for _, table := range tables {
		ret.Details = append(ret.Details, toProtoSensitiveWordTableWithWords(table))
	}
	return ret, nil
}

func (s *Service) GetSensitiveWordTableListByIDs(ctx context.Context, req *safety_service.GetSensitiveWordTableListByIDsReq) (*safety_service.SensitiveWordTables, error) {
	tables, err := s.cli.GetSensitiveWordTableListByIDs(ctx, req.TableIds)
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	ret := &safety_service.SensitiveWordTables{
		Total: int64(len(tables)),
	}
	for _, table := range tables {
		ret.List = append(ret.List, toProtoSensitiveWordTable(table))
	}
	return ret, nil
}

func (s *Service) GetSensitiveWordTableByID(ctx context.Context, req *safety_service.GetSensitiveWordTableByIDReq) (*safety_service.SensitiveWordTable, error) {
	table, err := s.cli.GetSensitiveWordTableByID(ctx, util.MustU32(req.TableId))
	if err != nil {
		return nil, errStatus(errs.Code_AppSafety, err)
	}
	ret := toProtoSensitiveWordTable(table)
	return ret, nil
}

func toProtoSensitiveWordTable(sensitiveWordTable *model.SensitiveWordTable) *safety_service.SensitiveWordTable {
	return &safety_service.SensitiveWordTable{
		TableId:   util.Int2Str(sensitiveWordTable.ID),
		TableName: sensitiveWordTable.Name,
		Remark:    sensitiveWordTable.Remark,
		Reply:     sensitiveWordTable.Reply,
		Version:   sensitiveWordTable.Version,
		CreatedAt: sensitiveWordTable.CreatedAt,
	}
}

func toProtoSensitiveWordVocabulary(sensitiveWordVocabulary *model.SensitiveWordVocabulary) *safety_service.SensitiveWordVocabulary {
	return &safety_service.SensitiveWordVocabulary{
		WordId:        util.Int2Str(sensitiveWordVocabulary.ID),
		Word:          sensitiveWordVocabulary.Content,
		SensitiveType: sensitiveWordVocabulary.SensitiveType,
	}
}

func toProtoSensitiveWordTableWithWords(sensitiveTable *orm.SensitiveWordTableWithWord) *safety_service.SensitiveWordTableWithWords {
	return &safety_service.SensitiveWordTableWithWords{
		Table: &safety_service.SensitiveWordTable{
			TableId: util.Int2Str(sensitiveTable.ID),
			Reply:   sensitiveTable.Reply,
			Version: sensitiveTable.Version,
		},
		SensitiveWords: sensitiveTable.SensitiveWords,
	}
}

func toOffset(req iReq) int32 {
	if req.GetPageNo() < 1 || req.GetPageSize() < 0 {
		return -1
	}
	return (req.GetPageNo() - 1) * req.GetPageSize()
}

type iReq interface {
	GetPageNo() int32 // 从1开始 [EN] Start from 1
	GetPageSize() int32
}
