package grpc

import (
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/grpc-provider"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_doc"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_keywords"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_parse_template"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_permission"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_qa"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_report"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_splitter"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_tag"
)
