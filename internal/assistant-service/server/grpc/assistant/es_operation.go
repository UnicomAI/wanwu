package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	minio_service "github.com/UnicomAI/wanwu/internal/assistant-service/service/minio-service"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/es"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SaveToES saves a document to ES.
func (s *Service) SaveToES(ctx context.Context, req *assistant_service.SaveToESReq) (*emptypb.Empty, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(req.DocJson), &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal doc json: %v", err)
	}

	if err := es.Assistant().IndexDocument(ctx, req.IndexName, doc); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteFromES deletes documents from ES by conditions.
func (s *Service) DeleteFromES(ctx context.Context, req *assistant_service.DeleteFromESReq) (*emptypb.Empty, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	conditions := make(map[string]interface{})
	for k, v := range req.Conditions {
		conditions[k] = v
	}

	//查询数据
	fields, _, _ := es.Assistant().SearchByFields(ctx, req.IndexName, conditions, 0, 1000, "desc")
	detailList := buildConversationDetails(fields)

	if err := es.Assistant().DeleteByFields(ctx, req.IndexName, conditions); err != nil {
		return nil, err
	}

	//删除历史记录中的minio数据
	asyncDeleteMinio(detailList)

	return &emptypb.Empty{}, nil
}

// SearchFromES searches documents in ES by conditions.
func (s *Service) SearchFromES(ctx context.Context, req *assistant_service.SearchFromESReq) (*assistant_service.SearchFromESResp, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	conditions := make(map[string]interface{})
	for k, v := range req.Conditions {
		conditions[k] = v
	}

	from := int((req.PageNo - 1) * req.PageSize)
	size := int(req.PageSize)

	docs, total, err := es.Assistant().SearchByFields(ctx, req.IndexName, conditions, from, size, req.SortOrder)
	if err != nil {
		return nil, err
	}

	docJsonList := make([]string, 0, len(docs))
	for _, doc := range docs {
		// 替换minio地址为用户访问的服务器地址
		docStr := strings.ReplaceAll(string(doc), "http://"+config.Cfg().Minio.EndPoint, os.Getenv("MINIO_DOWNLOAD_URL"))
		docJsonList = append(docJsonList, docStr)
	}

	return &assistant_service.SearchFromESResp{
		DocJsonList: docJsonList,
		Total:       total,
	}, nil
}

func asyncDeleteMinio(detailList []*model.ConversationDetails) {
	marshal, _ := json.Marshal(detailList)
	log.Infof("开始异步删除文件数据, detailList %s", string(marshal))
	if len(detailList) > 0 {
		safe_go_util.SafeGo(func() {
			for _, detail := range detailList {
				if len(detail.FileInfo) > 0 {
					for _, info := range detail.FileInfo {
						log.Infof("异步删除输入文件：%s", info.FileUrl)
						_ = minio_service.DeleteFile(context.Background(), info.FileUrl)
					}
				}
				if len(detail.ResponseFiles) > 0 {
					for _, file := range detail.ResponseFiles {
						log.Infof("异步删除输出文件：%s", file.FileUrl)
						_ = minio_service.DeleteFile(context.Background(), file.FileUrl)
					}
				}
			}
		})
	}
}

func buildConversationDetails(fields []json.RawMessage) []*model.ConversationDetails {
	var detailList []*model.ConversationDetails
	if len(fields) > 0 {
		for _, field := range fields {
			var detail model.ConversationDetails
			if err := json.Unmarshal(field, &detail); err != nil {
				log.Warnf("解析ES文档失败: %v", err)
				continue
			}
			detailList = append(detailList, &detail)
		}
	}
	return detailList
}
