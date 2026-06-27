.PHONY: build test release clean

# 版本
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 构建
build:
	go build -ldflags="-s -w" -o ./bin/go-sonovel ./cmd/sonovel

# 测试
test:
	go test ./... -v

# 清理
clean:
	rm -rf bin/
	go clean

# 发布
release:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

# 构建所有平台
build-all:
	@echo "Building for all platforms..."
	@echo "Linux: go build -ldflags=\"-s -w\" -o bin/go-sonovel ./cmd/sonovel"
	@echo "Windows: go build -ldflags=\"-s -w\" -o bin/go-sonovel.exe ./cmd/sonovel"
	@echo "macOS: go build -ldflags=\"-s -w\" -o bin/go-sonovel ./cmd/sonovel"

# 验证
verify:
	go mod download
	go test ./...
	go build ./...
