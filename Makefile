APP := sonovel-go
CMD := ./cmd/sonovel
DIST := dist
BIN := sonovel-go
VERSION ?= 0.0.1

GO ?= go
GOFLAGS ?=
LDFLAGS ?= -s -w
GOPROXY ?= https://proxy.golang.org,direct
GOCACHE ?= $(CURDIR)/.gocache
GOMODCACHE ?= $(CURDIR)/.gomodcache

.PHONY: help tidy test run run-tui run-web init build build-all package-all clean

help:
	@echo "Targets:"
	@echo "  make run            # 本地运行 CLI (显示 help)"
	@echo "  make init           # 生成 config.toml 并初始化目录"
	@echo "  make run-tui        # 本地运行 TUI"
	@echo "  make run-web        # 本地运行 Web UI (:7765)"
	@echo "  make test           # 运行 go test"
	@echo "  make build          # 构建当前平台二进制到项目根目录(sonovel-go/sonovel-go.exe)"
	@echo "  make build-all      # 构建并打包 x64 多平台压缩包(仅保留压缩包)"
	@echo "  make clean          # 清理 dist"

tidy:
	$(GO) mod tidy

test:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) test ./...

run:
	$(GO) run $(CMD)

init:
	$(GO) run $(CMD) init --config ./config.toml --rules-dir ./rules --out ./downloads

run-tui:
	$(GO) run $(CMD) tui --config ./config.toml --rules ./rules/main.json --out ./downloads

run-web:
	$(GO) run $(CMD) web --config ./config.toml --port 7765 --rules ./rules/main.json --out ./downloads

build:
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o ./$(BIN)$$( [ "$$(go env GOOS)" = "windows" ] && echo ".exe" ) $(CMD)

build-all:
	@mkdir -p $(DIST)/packages
	rm -rf $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle
	rm -rf $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle
	rm -rf $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle
	mkdir -p $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/rules
	mkdir -p $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/rules
	mkdir -p $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/rules
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_linux_amd64 $(CMD)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_darwin_amd64 $(CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_windows_amd64.exe $(CMD)
	cp $(DIST)/$(APP)_$(VERSION)_linux_amd64 $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/$(BIN)
	cp $(DIST)/$(APP)_$(VERSION)_darwin_amd64 $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/$(BIN)
	cp $(DIST)/$(APP)_$(VERSION)_windows_amd64.exe $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/$(BIN).exe
	cp config.toml $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/
	cp config.toml $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/
	cp config.toml $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/
	cp -R rules/* $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/rules/
	cp -R rules/* $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/rules/
	cp -R rules/* $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/rules/
	tar -czf $(DIST)/packages/$(APP)_$(VERSION)_linux_amd64.tar.gz -C $(DIST) $(APP)_$(VERSION)_linux_amd64_bundle
	tar -czf $(DIST)/packages/$(APP)_$(VERSION)_darwin_amd64.tar.gz -C $(DIST) $(APP)_$(VERSION)_darwin_amd64_bundle
	cd $(DIST) && zip -rq packages/$(APP)_$(VERSION)_windows_amd64.zip $(APP)_$(VERSION)_windows_amd64_bundle

clean:
	rm -rf $(DIST)
