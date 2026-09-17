.PHONY: all tidy generate check-generated test test-contract lint clean

all: tidy generate test

# 整理依赖
tidy:
	go mod tidy

# 执行代码生成 (测试 example 目录下的生成逻辑)
generate:
	go generate ./example/...

# 在干净 checkout 中检查已跟踪和未跟踪的生成文件
check-generated: generate
	@test -z "$$(git status --porcelain -- example/model/schema)" || { git status --short -- example/model/schema; exit 1; }

# 运行单元测试和公开能力组合测试，并开启竞态检测
test:
	go test -race -count=1 ./...

test-contract:
	go test -race -count=2 -shuffle=on ./test

# 静态检查（与 CI 一致）
lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed; skipping. Install: https://golangci-lint.run"; \
	fi

# 清理可能生成的临时文件或缓存
clean:
	go clean -testcache
