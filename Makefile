WANWU_VERSION := v0.3.3

LDFLAGS := -X main.buildTime=$(shell date +%Y-%m-%d,%H:%M:%S) \
			-X main.buildVersion=${WANWU_VERSION} \
			-X main.gitCommitID=$(shell git --git-dir=./.git rev-parse HEAD) \
			-X main.gitBranch=$(shell git --git-dir=./.git for-each-ref --format='%(refname:short)->%(upstream:short)' $(shell git --git-dir=./.git symbolic-ref -q HEAD)) \
			-X main.builder=$(shell git config user.name)

build: build-tidb-setup build-bff build-iam build-model build-mcp build-knowledge build-rag build-app build-operate build-assistant build-agent

build-tidb-setup:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/tidb-setup

build-bff:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/bff-service

build-iam:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/iam-service

build-model:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/model-service

build-mcp:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/mcp-service

build-knowledge:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/knowledge-service

build-rag:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/rag-service

build-app:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/app-service

build-operate:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/operate-service

build-assistant:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/assistant-service

build-agent:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/ ./cmd/agent-service

check:
	go vet ./...
	go fmt ./...
	docker run --rm -t -v $(PWD):/app -w /app golangci/golangci-lint:v1.64.8 bash -c 'golangci-lint run -v --timeout 3m'

check-callback:
	docker run --rm -t -v $(PWD)/callback:/callback -w /callback crpi-6pj79y7ddzdpexs8.cn-hangzhou.personal.cr.aliyuncs.com/gromitlee/python:3.12-slim-isort7.0.0 isort --check-only --diff --color .
	docker run --rm -t -v $(PWD)/callback:/callback -w /callback pyfound/black:25.11.0 black -t py312 --check --diff --color .

doc-swag:
	# swag version v1.16.4
	# v1
	swag fmt  -g guest.go -d internal/bff-service/server/http/handler/v1
	swag init -g guest.go -d internal/bff-service/server/http/handler/v1 -o docs/v1 --md docs --pd
	# openapi
	swag fmt  -g openapi.go -d internal/bff-service/server/http/handler/openapi
	swag init -g openapi.go -d internal/bff-service/server/http/handler/openapi -o docs/openapi --pd
	# callback
	swag fmt  -g callback.go -d internal/bff-service/server/http/handler/callback
	swag init -g callback.go -d internal/bff-service/server/http/handler/callback -o docs/callback --pd
	# openurl
	swag fmt  -g openurl.go -d internal/bff-service/server/http/handler/openurl
	swag init -g openurl.go -d internal/bff-service/server/http/handler/openurl -o docs/openurl --pd

docker: docker-image-backend docker-image-frontend docker-image-rag docker-image-agent docker-image-callback

docker-base: docker-image-agent-base docker-image-callback-base

docker-image-backend:
	docker build -f Dockerfile.backend -t wanwulite/wanwu-backend:${WANWU_VERSION}-$(shell git rev-parse --short HEAD) .

docker-image-frontend:
	docker build -f Dockerfile.frontend -t wanwulite/wanwu-frontend:${WANWU_VERSION}-$(shell git rev-parse --short HEAD) .

docker-image-rag:
	docker build -f Dockerfile.rag -t wanwulite/rag:${WANWU_VERSION}-$(shell git rev-parse --short HEAD) .

docker-image-agent:
	docker build -f Dockerfile.agent -t wanwulite/agent:${WANWU_VERSION}-$(shell git rev-parse --short HEAD) .

docker-image-agent-base:
	docker build -f Dockerfile.agent-base -t wanwulite/agent-base:${WANWU_VERSION}-$(shell git rev-parse --short HEAD) .

docker-image-callback:
	docker build -f Dockerfile.callback -t wanwulite/callback:${WANWU_VERSION}-$(shell git rev-parse --short HEAD) .

docker-image-callback-base:
	docker build -f Dockerfile.callback-base -t wanwulite/callback-base:${WANWU_VERSION}-$(shell git rev-parse --short HEAD) .

grpc-protoc:
	protoc --proto_path=. --go_out=paths=source_relative:api --go-grpc_out=paths=source_relative:api proto/*/*.proto

i18n-jsonl:
	go test ./pkg/i18n -run TestI18nConvertXlsx2Jsonl
