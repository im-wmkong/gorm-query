.PHONY: all tidy generate test lint clean

all: tidy generate test

# 整理依赖
tidy:
	go mod tidy

# 执行代码生成 (测试 example 目录下的生成逻辑)
generate:
	go generate ./example/...

# 运行所有单元测试，并开启竞态检测
test:
	go test -v -race ./...

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
	rm -f example/model/*_gen.go
