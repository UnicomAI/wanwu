include .env
include .env.image.${WANWU_ARCH}

LDFLAGS := -X main.buildTime=$(shell date +%Y-%m-%d,%H:%M:%S) \
			-X main.buildVersion=${WANWU_VERSION} \
			-X main.gitCommitID=$(shell git --git-dir=./.git rev-parse HEAD) \
			-X main.gitBranch=$(shell git --git-dir=./.git for-each-ref --format='%(refname:short)->%(upstream:short)' $(shell git --git-dir=./.git symbolic-ref -q HEAD)) \
			-X main.builder=$(shell git config user.name)

build-tidb-setup-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/tidb-setup

build-tidb-setup-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/tidb-setup

build-bff-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/bff-service

build-bff-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/bff-service

build-iam-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/iam-service

build-iam-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/iam-service

build-model-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/model-service

build-model-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/model-service

build-mcp-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/mcp-service

build-mcp-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/mcp-service

build-knowledge-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/knowledge-service

build-knowledge-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/knowledge-service

build-rag-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/rag-service

build-rag-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/rag-service

build-app-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/app-service

build-app-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/app-service

build-channel-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/channel-service

build-channel-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/channel-service

build-operate-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/operate-service

build-operate-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/operate-service

build-assistant-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/assistant-service

build-assistant-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/assistant-service

build-agent-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/amd64/ ./cmd/agent-service

build-agent-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -ldflags "$(LDFLAGS)" -o ./bin/arm64/ ./cmd/agent-service

create-docker-net:
	docker network create ${WANWU_DOCKER_NETWORK}

check:
	go vet ./...
	gofmt -w cmd/ internal/ pkg/
	goimports -w -local github.com/wanwulite/wanwu cmd/ internal/ pkg/
	docker run --rm -t -v $(PWD):/app -w /app golangci/golangci-lint:v2.10.1 bash -c 'golangci-lint run -v --timeout 5m'

check-callback:
	docker run --rm -t -v $(PWD)/callback:/callback -w /callback crpi-6pj79y7ddzdpexs8.cn-hangzhou.personal.cr.aliyuncs.com/gromitlee/python:3.12-slim-isort7.0.0 isort --check-only --diff --color .
	docker run --rm -t -v $(PWD)/callback:/callback -w /callback pyfound/black:25.11.0 black -t py312 --check --diff --color .

doc:
	docker run --name golang-swag --privileged=true --rm -v $(PWD):/app -w /app crpi-6pj79y7ddzdpexs8.cn-hangzhou.personal.cr.aliyuncs.com/gromitlee/golang:1.24.13-bookworm-swag1.16.6 bash -c 'make doc-swag'

doc-swag:
	# swag version v1.16.6
	# v1
	swag fmt  -g guest.go -d internal/bff-service/server/http/handler/v1
	swag init -g guest.go -d internal/bff-service/server/http/handler/v1 -o docs/v1 --md docs --pd
	@echo '// nolint' | cat - docs/v1/docs.go > tmp && mv tmp docs/v1/docs.go
	# v2
	swag fmt  -g common.go -d internal/bff-service/server/http/handler/v2
	swag init -g common.go -d internal/bff-service/server/http/handler/v2 -o docs/v2 --pd
	@echo '// nolint' | cat - docs/v2/docs.go > tmp && mv tmp docs/v2/docs.go
	# openapi
	swag fmt  -g openapi.go -d internal/bff-service/server/http/handler/openapi
	swag init -g openapi.go -d internal/bff-service/server/http/handler/openapi -o docs/openapi --pd
	@echo '// nolint' | cat - docs/openapi/docs.go > tmp && mv tmp docs/openapi/docs.go
	# callback
	swag fmt  -g callback.go -d internal/bff-service/server/http/handler/callback
	swag init -g callback.go -d internal/bff-service/server/http/handler/callback -o docs/callback --pd
	@echo '// nolint' | cat - docs/callback/docs.go > tmp && mv tmp docs/callback/docs.go
	# openurl
	swag fmt  -g openurl.go -d internal/bff-service/server/http/handler/openurl
	swag init -g openurl.go -d internal/bff-service/server/http/handler/openurl -o docs/openurl --pd
	@echo '// nolint' | cat - docs/openurl/docs.go > tmp && mv tmp docs/openurl/docs.go

docker-image-frontend:
	@printf "==========================================\n"
	@printf "前端镜像构建说明：\n"
	@printf "==========================================\n"
	@printf "1. 确保 wanwu 和 wanwu-workflow 仓库在同一目录下，均更新至最新状态\n"
	@printf "2. 在该目录下执行：\n"
	@printf "   cp wanwu/Dockerfile.frontend.bak Dockerfile.frontend\n"
	@printf "   docker build -f Dockerfile.frontend -t wanwulite/wanwu-frontend:${WANWU_VERSION}-$(shell git rev-parse --short HEAD)-${WANWU_ARCH} .\n"
	@printf "==========================================\n"

docker-image-backend:
	docker build -f Dockerfile.backend --build-arg WANWU_ARCH=${WANWU_ARCH} -t wanwulite/wanwu-backend:${WANWU_VERSION}-$(shell git rev-parse --short HEAD)-${WANWU_ARCH} .

docker-image-rag:
	docker build -f Dockerfile.rag --build-arg WANWU_ARCH=${WANWU_ARCH} -t wanwulite/rag:${WANWU_VERSION}-$(shell git rev-parse --short HEAD)-${WANWU_ARCH} .

docker-image-callback:
	docker build -f Dockerfile.callback --build-arg WANWU_ARCH=${WANWU_ARCH} -t wanwulite/callback:${WANWU_VERSION}-$(shell git rev-parse --short HEAD)-${WANWU_ARCH} .

docker-image-callback-base:
	docker build -f Dockerfile.callback-base --build-arg WANWU_ARCH=${WANWU_ARCH} -t wanwulite/callback-base:${WANWU_VERSION}-$(shell git rev-parse --short HEAD)-${WANWU_ARCH} .

docker-image-wga-sandbox:
	docker build -f Dockerfile.wga-sandbox --build-arg WANWU_ARCH=${WANWU_ARCH} -t wanwulite/wga-sandbox:${WANWU_VERSION}-$(shell git rev-parse --short HEAD)-${WANWU_ARCH} .

docker-image-wga-sandbox-base:
	docker build -f Dockerfile.wga-sandbox-base --build-arg WANWU_ARCH=${WANWU_ARCH} -t wanwulite/wga-sandbox-base:1.0.0.152-opencode1.16.2-${WANWU_ARCH} .

grpc-protoc:
	protoc --proto_path=. --go_out=paths=source_relative:api --go-grpc_out=paths=source_relative:api proto/*/*.proto

i18n-jsonl:
	go test ./pkg/i18n -run TestI18nConvertXlsx2Jsonl

init:
	go mod tidy
	go mod vendor

pb:
	docker run --name golang-grpc --privileged=true --rm -v $(PWD):/app -w /app crpi-6pj79y7ddzdpexs8.cn-hangzhou.personal.cr.aliyuncs.com/gromitlee/golang:1.24.13-bookworm-protoc29.4-gengo1.34.1-gengrpc1.5.1-gengw2.20.0-genapi2.20.0 bash -c 'make grpc-protoc'

# --- mysql ---
run-mysql:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up -d mysql

stop-mysql:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down mysql

# --- mysql-setup ---
run-mysql-setup:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up mysql-setup

stop-mysql-setup:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down mysql-setup

# --- tidb ---
run-tidb:
	docker-compose -f docker-compose.tidb.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up -d tidb

stop-tidb:
	docker-compose -f docker-compose.tidb.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down tidb

# --- tidb-setup ---
run-tidb-setup:
	docker-compose -f docker-compose.tidb.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up tidb-setup

stop-tidb-setup:
	docker-compose -f docker-compose.tidb.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down tidb-setup

# --- oceanbase ---
run-oceanbase:
	docker-compose -f docker-compose.oceanbase.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up -d oceanbase

stop-oceanbase:
	docker-compose -f docker-compose.oceanbase.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down oceanbase

# --- redis ---
run-redis:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up -d redis

stop-redis:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down redis

# --- minio ---
run-minio:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up -d minio

stop-minio:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down minio

# --- kafka ---
run-kafka:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up -d kafka

stop-kafka:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down kafka

# --- elastic-setup ---
run-es-setup:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up -d es-setup

stop-es-setup:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down es-setup

# --- elastic ---
run-es:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		up -d es

stop-es:
	docker-compose -f docker-compose.yaml \
		--env-file .env.image.${WANWU_ARCH} \
		--env-file .env \
		down es