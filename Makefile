.PHONY: build-all build-server build-web build-linux-amd64 build-linux-arm64 build-darwin dev clean

# ============================================================
# ServerSmith — 跨平台构建工具
# ============================================================
# P5 两阶段构建:
#   1. build-web:   Next.js 静态导出 (output: 'export')
#   2. build-server: Go embed 编译单二进制
# ============================================================

VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-X main.Version=$(VERSION)"

# ---------- 全量构建 ----------

build-all: build-web build-server build-linux build-darwin

# ---------- 前后端分别构建 ----------

build-web:
	@echo "========================"
	@echo " Building Next.js static export..."
	@echo "========================"
	cd web && npm run build 2>&1 | tail -5
	@echo "✅ Web build complete (out/)"

build-server:
	@echo "========================"
	@echo " Copying static files & building Go binary..."
	@echo "========================"
	mkdir -p server/web-out && cp -r web/out/* server/web-out/
	cd server && go build $(LDFLAGS) -o ../dist/serversmith .
	@echo "✅ Server binary: dist/serversmith"

# ---------- 跨平台编译 ----------

build-linux: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	@echo ">>> linux/amd64 ..."
	mkdir -p server/web-out && cp -r web/out/* server/web-out/
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ../dist/serversmith-linux-amd64 .
	@echo "✅ dist/serversmith-linux-amd64"

build-linux-arm64:
	@echo ">>> linux/arm64 ..."
	mkdir -p server/web-out && cp -r web/out/* server/web-out/
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o ../dist/serversmith-linux-arm64 .
	@echo "✅ dist/serversmith-linux-arm64"

build-darwin:
	@echo ">>> darwin/arm64 ..."
	mkdir -p server/web-out && cp -r web/out/* server/web-out/
	cd server && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o ../dist/serversmith-darwin-arm64 .
	@echo "✅ dist/serversmith-darwin-arm64"

# ---------- 开发环境 ----------

dev:
	docker compose up --build -d

# ---------- 清理 ----------

clean:
	rm -rf dist/ server/web-out/ web/out/ web/.next/
	@echo "✅ cleaned"
