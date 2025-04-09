.PHONY: all build test clean fmt lint release

# 包名
PACKAGE := github.com/ntshibin/core

# Go 相关命令
GO := go
GOFMT := gofmt

# 默认目标
all: build test

# 构建
build:
	@echo "Building $(PACKAGE)..."
	$(GO) build ./...

# 运行测试
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# 清理构建文件
clean:
	@echo "Cleaning..."
	$(GO) clean -i ./...
	rm -rf dist/ build/ bin/ *.log
	find . -name "*.test" -delete
	find . -name "*.out" -delete
	find . -name "*.prof" -delete
	find . -name "*.trace" -delete
	find . -name "*.cover" -delete

# 格式化代码
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .


# 发布新版本
# 使用方式: make release VERSION=v1.0.0
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "请指定版本号，例如: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "发布版本 $(VERSION)..."
	@echo "清理不必要的文件..."
	$(MAKE) clean
	@echo "检查代码格式..."
	$(MAKE) fmt
	@echo "运行测试..."
	$(MAKE) test
	@echo "创建并推送标签..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# 安装依赖
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# 显示帮助信息
help:
	@echo "可用的命令:"
	@echo "  make all        - 构建并测试"
	@echo "  make build      - 构建包"
	@echo "  make test       - 运行测试"
	@echo "  make clean      - 清理构建文件"
	@echo "  make fmt        - 格式化代码"
	@echo "  make lint       - 运行代码检查"
	@echo "  make release    - 发布新版本 (需要指定 VERSION 参数)"
	@echo "  make deps       - 安装依赖"
	@echo "  make help       - 显示此帮助信息" 