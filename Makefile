# Makefile for dbweb

# 变量定义
BINARY_NAME=dbweb
MAIN_PATH=./main.go
BUILD_DIR=./build
GO_VERSION=1.23.10

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "可用的命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# 构建
.PHONY: build
build: ## 构建项目
	@echo "构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建（开发模式，包含调试信息）
.PHONY: build-dev
build-dev: ## 构建开发版本（包含调试信息）
	@echo "构建开发版本 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建（生产模式，优化）
.PHONY: build-release
build-release: ## 构建发布版本（优化）
	@echo "构建发布版本 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 运行
.PHONY: run
run: ## 运行项目
	@echo "运行 $(BINARY_NAME)..."
	@go run $(MAIN_PATH)

# 安装依赖
.PHONY: deps
deps: ## 下载依赖
	@echo "下载依赖..."
	@go mod download
	@go mod tidy
	@echo "依赖下载完成"

# 更新依赖
.PHONY: deps-update
deps-update: ## 更新所有依赖到最新版本
	@echo "更新依赖..."
	@go get -u ./...
	@go mod tidy
	@echo "依赖更新完成"

# 格式化代码
.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...
	@echo "格式化完成"

# 代码检查
.PHONY: vet
vet: ## 运行 go vet 检查代码
	@echo "运行 go vet..."
	@go vet ./...
	@echo "检查完成"

# 运行测试
.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	@go test -v ./...

# 运行测试（带覆盖率）
.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 清理
.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean
	@echo "清理完成"

# 安装到系统路径
.PHONY: install
install: build-release ## 安装到系统路径
	@echo "安装 $(BINARY_NAME) 到系统路径..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "安装完成"

# 卸载
.PHONY: uninstall
uninstall: ## 从系统路径卸载
	@echo "卸载 $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "卸载完成"

# 交叉编译 - Linux
.PHONY: build-linux
build-linux: ## 构建 Linux 版本
	@echo "构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# 交叉编译 - Windows
.PHONY: build-windows
build-windows: ## 构建 Windows 版本
	@echo "构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# 交叉编译 - macOS
.PHONY: build-darwin
build-darwin: ## 构建 macOS 版本
	@echo "构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64, $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

# 构建所有平台
.PHONY: build-all
build-all: build-linux build-windows build-darwin ## 构建所有平台版本
	@echo "所有平台构建完成"

# 检查 Go 版本
.PHONY: check-go
check-go: ## 检查 Go 版本
	@echo "检查 Go 版本..."
	@go version
	@echo "要求的 Go 版本: $(GO_VERSION)"

# 代码检查（包括格式化、vet、测试）
.PHONY: check
check: fmt vet test ## 运行所有代码检查

# 开发模式（格式化、检查、构建、运行）
.PHONY: dev
dev: fmt vet build run ## 开发模式：格式化、检查、构建并运行

# 发布前检查
.PHONY: pre-release
pre-release: clean deps fmt vet test build-release ## 发布前检查：清理、更新依赖、格式化、检查、测试、构建

