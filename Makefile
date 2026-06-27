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

# 发布
release:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

