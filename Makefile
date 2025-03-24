# Go参数
GO = go
GOPATH = $(shell $(GO) env GOPATH)
GOBIN = $(GOPATH)/bin
GOFMT = gofmt -s -w
GOTEST = $(GO) test

# 项目参数
MODULE_NAME = github.com/ntshibin/core
VERSION = v0.1.0
TARGET = ./bin/

# 模块列表
MODULES = gerror glog gconf gcache ghttp

# 颜色设置
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

# 构建目标
.PHONY: all
all: clean tidy fmt vet test build

# 清理构建目录
.PHONY: clean
clean:
	@echo "$(BLUE)清理构建环境...$(NC)"
	@rm -rf $(TARGET)
	@mkdir -p $(TARGET)
	@echo "$(GREEN)清理完成.$(NC)"

# 依赖管理
.PHONY: tidy
tidy:
	@echo "$(BLUE)整理依赖关系...$(NC)"
	@$(GO) mod tidy
	@echo "$(GREEN)整理完成.$(NC)"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "$(BLUE)格式化代码...$(NC)"
	@$(GOFMT) .
	@echo "$(GREEN)格式化完成.$(NC)"

# 代码检查
.PHONY: vet
vet:
	@echo "$(BLUE)检查代码...$(NC)"
	@$(GO) vet ./...
	@echo "$(GREEN)检查完成.$(NC)"

# 运行测试
.PHONY: test
test:
	@echo "$(BLUE)运行测试...$(NC)"
	@$(GOTEST) -v ./...
	@echo "$(GREEN)测试完成.$(NC)"

# 构建所有模块
.PHONY: build
build:
	@echo "$(BLUE)构建所有模块...$(NC)"
	@for module in $(MODULES); do \
		echo "$(YELLOW)构建 $$module...$(NC)"; \
		$(GO) build -v -o $(TARGET)$$module ./$$module; \
	done
	@echo "$(GREEN)构建完成.$(NC)"

# 为每个模块创建独立的构建目标
define make-module-target
.PHONY: $(1)
$(1):
	@echo "$(BLUE)构建 $(1)...$(NC)"
	@$(GO) build -v -o $(TARGET)$(1) ./$(1)
	@echo "$(GREEN)构建完成.$(NC)"
endef

$(foreach module,$(MODULES),$(eval $(call make-module-target,$(module))))

# Git提交
.PHONY: commit
commit:
	@echo "$(BLUE)提交变更到Git...$(NC)"
	@read -p "提交信息: " message; \
	git add .; \
	git commit -m "$$message"; \
	echo "$(GREEN)提交完成.$(NC)"

# Git推送
.PHONY: push
push:
	@echo "$(BLUE)推送到远程仓库...$(NC)"
	@git push origin HEAD
	@echo "$(GREEN)推送完成.$(NC)"

# 创建新的Git标签
.PHONY: tag
tag:
	@echo "$(BLUE)创建Git标签...$(NC)"
	@read -p "标签版本(例如 v0.1.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	echo "$(GREEN)标签创建完成.$(NC)"

# 推送Git标签
.PHONY: push-tag
push-tag:
	@echo "$(BLUE)推送标签到远程仓库...$(NC)"
	@git push --tags
	@echo "$(GREEN)标签推送完成.$(NC)"

# 完整发布流程
.PHONY: release
release: all commit tag push push-tag
	@echo "$(GREEN)发布流程完成.$(NC)"

# 帮助信息
.PHONY: help
help:
	@echo "$(YELLOW)可用目标:$(NC)"
	@echo "  $(GREEN)all$(NC) - 执行完整构建流程(清理、整理依赖、格式化、检查、测试、构建)"
	@echo "  $(GREEN)clean$(NC) - 清理构建目录"
	@echo "  $(GREEN)tidy$(NC) - 整理依赖关系"
	@echo "  $(GREEN)fmt$(NC) - 格式化代码"
	@echo "  $(GREEN)vet$(NC) - 检查代码"
	@echo "  $(GREEN)test$(NC) - 运行测试"
	@echo "  $(GREEN)build$(NC) - 构建所有模块"
	@echo "  $(GREEN)commit$(NC) - 提交变更到Git"
	@echo "  $(GREEN)push$(NC) - 推送到远程仓库"
	@echo "  $(GREEN)tag$(NC) - 创建Git标签"
	@echo "  $(GREEN)push-tag$(NC) - 推送Git标签"
	@echo "  $(GREEN)release$(NC) - 执行完整发布流程"
	@echo "单独构建模块: $(GREEN)$(MODULES)$(NC)" 