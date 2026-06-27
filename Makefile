.PHONY: build test release clean build-all verify

# 版本
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 构建单平台
build:
	go build -ldflags="-s -w" -o ./bin/go-sonovel ./cmd/sonovel

# 构建所有平台（本地）
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/go-sonovel-linux-amd64 ./cmd/sonovel
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/go-sonovel-windows-amd64.exe ./cmd/sonovel
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/go-sonovel-darwin-amd64 ./cmd/sonovel
	@echo "All platforms built successfully!"
	@ls -lh bin/

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

# 验证
verify:
	go mod download
	go test ./...
	go build ./...
