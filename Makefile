APP := sonovel-go
CMD := ./cmd/sonovel
DIST := dist
VERSION ?= 0.0.1

GO ?= go
GOFLAGS ?=
LDFLAGS ?= -s -w
GOPROXY ?= https://proxy.golang.org,direct
GOCACHE ?= /tmp/$(APP)-gocache
GOMODCACHE ?= $(CURDIR)/.gomodcache

RULES_SRC ?= ../so-novel/bundle/rules
RULES_DST ?= ./rules

.PHONY: help tidy test run run-tui run-web init build copy-rules build-all package-all clean

help:
	@echo "Targets:"
	@echo "  make run            # 本地运行 CLI (显示 help)"
	@echo "  make init           # 生成 config.toml 并初始化目录"
	@echo "  make run-tui        # 本地运行 TUI"
	@echo "  make run-web        # 本地运行 Web UI (:7765)"
	@echo "  make test           # 运行 go test"
	@echo "  make build          # 构建当前平台二进制到 ./dist"
	@echo "  make build-all      # 构建 x64 多平台二进制到 ./dist (linux/windows/darwin)"
	@echo "  make package-all    # 生成可分发压缩包(二进制+config+rules+README)"
	@echo "  make copy-rules     # 复制规则目录到 ./rules"
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
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_$$(go env GOOS)_$$(go env GOARCH) $(CMD)

copy-rules:
	@mkdir -p $(RULES_DST)
	rm -rf $(RULES_DST)
	cp -R $(RULES_SRC) $(RULES_DST)

build-all:
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_linux_amd64 $(CMD)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_darwin_amd64 $(CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPROXY=$(GOPROXY) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_windows_amd64.exe $(CMD)

package-all: build-all
	@mkdir -p $(DIST)/packages
	rm -rf $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle
	rm -rf $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle
	rm -rf $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle
	mkdir -p $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/rules
	mkdir -p $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/rules
	mkdir -p $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/rules
	cp $(DIST)/$(APP)_$(VERSION)_linux_amd64 $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/$(APP)
	cp $(DIST)/$(APP)_$(VERSION)_darwin_amd64 $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/$(APP)
	cp $(DIST)/$(APP)_$(VERSION)_windows_amd64.exe $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/$(APP).exe
	cp config.toml $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/
	cp config.toml $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/
	cp config.toml $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/
	cp README.md $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/
	cp README.md $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/
	cp README.md $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/
	cp -R rules/* $(DIST)/$(APP)_$(VERSION)_linux_amd64_bundle/rules/
	cp -R rules/* $(DIST)/$(APP)_$(VERSION)_darwin_amd64_bundle/rules/
	cp -R rules/* $(DIST)/$(APP)_$(VERSION)_windows_amd64_bundle/rules/
	tar -czf $(DIST)/packages/$(APP)_$(VERSION)_linux_amd64.tar.gz -C $(DIST) $(APP)_$(VERSION)_linux_amd64_bundle
	tar -czf $(DIST)/packages/$(APP)_$(VERSION)_darwin_amd64.tar.gz -C $(DIST) $(APP)_$(VERSION)_darwin_amd64_bundle
	cd $(DIST) && zip -rq packages/$(APP)_$(VERSION)_windows_amd64.zip $(APP)_$(VERSION)_windows_amd64_bundle

clean:
	rm -rf $(DIST)
