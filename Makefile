# Makefile for smartCmd

# 默认配置（可通过环境变量覆盖）
BASE_URL ?= https://api.openai.com/v1
API_KEY ?= your-key-here
MODEL ?= gpt-4o

# 构建标志
LDFLAGS = -X smartCmd/config.DefaultBaseURL=$(BASE_URL) \
          -X smartCmd/config.DefaultAPIKey=$(API_KEY) \
          -X smartCmd/config.DefaultModel=$(MODEL)

# 输出目录
DIST_DIR = dist

# 目标
.PHONY: all build build-linux build-mac build-win build-all clean test

all: build

# 本地构建（当前平台）
build:
	@echo "Building for current platform..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(DIST_DIR)/smartCmd .
	@echo "Build complete: $(DIST_DIR)/smartCmd"

# Linux 构建
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o $(DIST_DIR)/smartCmd-linux .
	@echo "Build complete: $(DIST_DIR)/smartCmd-linux"

# macOS 构建
build-mac:
	@echo "Building for macOS..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -o $(DIST_DIR)/smartCmd-mac-arm64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o $(DIST_DIR)/smartCmd-mac-amd64 .
	@echo "Build complete: $(DIST_DIR)/smartCmd-mac-*"

# Windows 构建
build-win:
	@echo "Building for Windows..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o $(DIST_DIR)/smartCmd.exe .
	@echo "Build complete: $(DIST_DIR)/smartCmd.exe"

# 构建所有平台
build-all: build-linux build-mac build-win
	@echo "All builds complete!"

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(DIST_DIR)
	@echo "Clean complete!"

# 测试
test:
	@echo "Running tests..."
	go test ./...

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod tidy
	@echo "Dependencies installed!"

# 运行（开发模式）
run:
	go run .

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete!"

# 代码检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/"; \
	fi
