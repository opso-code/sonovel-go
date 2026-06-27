.PHONY: build test release clean build-all verify

# 版本
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 构建单平台
build:
	go build -ldflags="-s -w" -o ./bin/go-sonovel ./cmd/sonovel

# 构建所有平台（本地）
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" \
	  -o ./bin/linux-amd64/go-sonovel ./cmd/sonovel

	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" \
	  -o ./bin/windows-amd64/go-sonovel.exe ./cmd/sonovel

	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" \
	  -o ./bin/darwin-amd64/go-sonovel ./cmd/sonovel

	@echo "All platforms built successfully!"
	ls -lh bin/*/

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

# 跨平台打包（用于 CI/CD）
dist:
	@echo "Creating release archives..."
	mkdir -p ./release/linux-amd64 \
	          ./release/windows-amd64 ./release/darwin-amd64

	# Linux amd64
	cp bin/linux-amd64/go-sonovel ./release/linux-amd64/
	cp -r rules ./release/linux-amd64/
	cp config.toml ./release/linux-amd64/ || true
	cp README.md ./release/linux-amd64/ || true
	cp LICENSE ./release/linux-amd64/ || true
	tar czf go-sonovel-linux-amd64.tar.gz -C ./release/linux-amd64 .

	# Windows amd64
	cp bin/windows-amd64/go-sonovel.exe ./release/windows-amd64/
	cp -r rules ./release/windows-amd64/
	cp config.toml ./release/windows-amd64/ || true
	cp README.md ./release/windows-amd64/ || true
	cp LICENSE ./release/windows-amd64/ || true
	cd ./release/windows-amd64 && zip -r ../../go-sonovel-windows-amd64.zip .
	cd ../..

	# macOS amd64
	cp bin/darwin-amd64/go-sonovel ./release/darwin-amd64/
	cp -r rules ./release/darwin-amd64/
	cp config.toml ./release/darwin-amd64/ || true
	cp README.md ./release/darwin-amd64/ || true
	cp LICENSE ./release/darwin-amd64/ || true
	tar czf go-sonovel-darwin-amd64.tar.gz -C ./release/darwin-amd64 .

	@echo "Release archives created:"
	ls -lh ./release/*.tar.gz ./release/*.zip
