.PHONY: all build build-gateway build-seaking build-relay clean tidy run-gateway run-seaking run-relay migrate docker-up docker-down

# 默认目标
all: build

# 构建所有服务
build: build-gateway build-seaking build-relay

build-gateway:
	@echo "Building gateway..."
	cd gateway && go build -o ../bin/gateway ./cmd

build-seaking:
	@echo "Building seaking..."
	cd seaking && go build -o ../bin/seaking ./cmd

build-relay:
	@echo "Building relay..."
	cd relay && go build -o ../bin/relay ./cmd

# 清理
clean:
	rm -rf bin/
	rm -rf */logs/

# 依赖管理
tidy:
	cd common && go mod tidy
	cd gateway && go mod tidy
	cd seaking && go mod tidy
	cd relay && go mod tidy

# 运行服务
run-gateway:
	cd gateway && go run ./cmd -c config -cPath "./,./configs/"

run-seaking:
	cd seaking && go run ./cmd -c config -cPath "./,./configs/"

run-relay:
	cd relay && go run ./cmd -c config -cPath "./,./configs/"

# 数据库迁移
migrate-seaking:
	cd seaking && go run ./cmd -c config -cPath "./,./configs/" -migrate

migrate-relay:
	cd relay && go run ./cmd -c config -cPath "./,./configs/" -migrate

migrate: migrate-seaking migrate-relay

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

# 测试
test:
	cd common && go test ./...
	cd gateway && go test ./...
	cd seaking && go test ./...
	cd relay && go test ./...

# 格式化
fmt:
	cd common && go fmt ./...
	cd gateway && go fmt ./...
	cd seaking && go fmt ./...
	cd relay && go fmt ./...

# 帮助
help:
	@echo "Available targets:"
	@echo "  build          - Build all services"
	@echo "  build-gateway  - Build gateway service"
	@echo "  build-seaking  - Build seaking service"
	@echo "  build-relay    - Build relay service"
	@echo "  clean          - Clean build artifacts"
	@echo "  tidy           - Run go mod tidy for all modules"
	@echo "  run-gateway    - Run gateway service"
	@echo "  run-seaking    - Run seaking service"
	@echo "  run-relay      - Run relay service"
	@echo "  migrate        - Run database migrations"
	@echo "  docker-up      - Start docker containers"
	@echo "  docker-down    - Stop docker containers"
	@echo "  test           - Run tests"
	@echo "  fmt            - Format code"
