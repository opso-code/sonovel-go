.PHONY: build release clean dist

# 版本
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 构建单平台
build:
	go build -ldflags="-s -w" -o ./bin/go-sonovel ./cmd/sonovel

# 清理
clean:
	rm -rf bin/*
	rm -rf dist/*

# 发布
release:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

# 跨平台打包（用于 CI/CD）
dist: clean
	test -d dist || mkdir dist
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" \
	  -o ./dist/go-sonovel-linux-amd64/go-sonovel ./cmd/sonovel

	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" \
	  -o ./dist/go-sonovel-windows-amd64/go-sonovel.exe ./cmd/sonovel

	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" \
	  -o ./dist/go-sonovel-darwin-amd64/go-sonovel ./cmd/sonovel

	@echo "All platforms built successfully!"

	@echo "Creating release archives..."

	# Linux amd64
	cp -r rules ./dist/go-sonovel-linux-amd64/
	cp config.toml ./dist/go-sonovel-linux-amd64/
	cp README.md ./dist/go-sonovel-linux-amd64/
	tar czf ./dist/go-sonovel-linux-amd64.tar.gz -C ./dist go-sonovel-linux-amd64 --remove-files

	# Windows amd64
	cp -r rules ./dist/go-sonovel-windows-amd64/
	cp config.toml ./dist/go-sonovel-windows-amd64/
	cp README.md ./dist/go-sonovel-windows-amd64/
	(cd ./dist && zip -qmr go-sonovel-windows-amd64.zip ./go-sonovel-windows-amd64)

	# macOS amd64
	cp -r rules ./dist/go-sonovel-darwin-amd64/
	cp config.toml ./dist/go-sonovel-darwin-amd64/
	cp README.md ./dist/go-sonovel-darwin-amd64/
	tar czf ./dist/go-sonovel-darwin-amd64.tar.gz -C ./dist go-sonovel-darwin-amd64 --remove-files

	@echo "Release archives created:"
	ls -lh ./dist
