package knowledge_parse_template

import (
	knowledgebase_template_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-template-service"
	grpc_provider "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/grpc-provider"
	"google.golang.org/grpc"
)

type Service struct {
	knowledgebase_template_service.UnimplementedKnowledgeBaseTemplateServiceServer
}

var parseTemplateService = Service{}

func init() {
	grpc_provider.AddGrpcContainer(&parseTemplateService)
}

func (s *Service) GrpcType() string {
	return "grpc_knowledge_parse_template_service"
}

func (s *Service) Register(serv *grpc.Server) error {
	knowledgebase_template_service.RegisterKnowledgeBaseTemplateServiceServer(serv, s)
	return nil
}
