#!/bin/bash
# ============================================================
# ServerSmith — 一键安装脚本 (Linux amd64/arm64)
# 用法:
#   curl -sSL https://raw.githubusercontent.com/gdjmleo-star/ServerSmith/main/install.sh | bash
# 或:
#   bash <(curl -sSL https://raw.githubusercontent.com/gdjmleo-star/ServerSmith/main/install.sh)
# ============================================================
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
info()  { echo -e "${CYAN}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
err()   { echo -e "${RED}[ERR]${NC} $1"; exit 1; }

# ---------- 检测环境 ----------
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  GOARCH="amd64" ;;
  aarch64) GOARCH="arm64" ;;
  *)       err "不支持的架构: $ARCH (仅支持 amd64 / arm64)" ;;
esac

if [ "$(uname -s)" != "Linux" ]; then
  err "仅支持 Linux 系统"
fi

INSTALL_DIR="/opt/serversmith"
DATA_DIR="/var/lib/serversmith"
SERVICE_NAME="serversmith"

# ---------- 安装依赖 ----------
info "检测系统依赖..."

install_pkg() {
  if command -v apt-get &>/dev/null; then
    apt-get update -qq && apt-get install -y -qq "$@" 2>&1 | tail -1
  elif command -v yum &>/dev/null; then
    yum install -y -q "$@" 2>&1 | tail -1
  elif command -v apk &>/dev/null; then
    apk add -q "$@" 2>&1 | tail -1
  else
    err "不支持的包管理器 (仅支持 apt/yum/apk)"
  fi
}

if command -v go &>/dev/null; then
  ok "Go $(go version | awk '{print $3}')"
else
  info "未检测到 Go，正在安装 Go 1.22..."
  GO_TAR="go1.22.0.linux-${GOARCH}.tar.gz"
  curl -sSL "https://go.dev/dl/${GO_TAR}" -o /tmp/go.tar.gz
  tar -C /usr/local -xzf /tmp/go.tar.gz
  export PATH="/usr/local/go/bin:$PATH"
  ok "Go 1.22 安装完成"
fi

# ---------- 克隆仓库 ----------
info "克隆 ServerSmith..."
if [ -d "$INSTALL_DIR" ]; then
  cd "$INSTALL_DIR" && git pull
else
  git clone --depth=1 https://github.com/gdjmleo-star/ServerSmith.git "$INSTALL_DIR"
fi
cd "$INSTALL_DIR"
ok "代码已获取"

# ---------- 构建前端 ----------
info "构建前端 (Next.js 静态导出)..."
install_pkg nodejs npm 2>/dev/null || true
cd web
npm install --silent 2>&1 | tail -1
npm run build 2>&1 | tail -3
cd ..
ok "前端构建完成"

# ---------- 构建后端 (单二进制) ----------
info "构建后端 (Go embed 单二进制)..."
mkdir -p server/web-out
cp -r web/out/* server/web-out/
cd server
CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/bin/serversmith .
cd "$INSTALL_DIR"
ok "二进制: /usr/local/bin/serversmith"

# ---------- 创建数据目录 ----------
mkdir -p "$DATA_DIR"
cat > "$INSTALL_DIR/.env" <<EOF
DB_PATH=$DATA_DIR/serversmith.db
LISTEN=:18080
EOF
ok "数据目录: $DATA_DIR"

# ---------- 创建 systemd 服务 ----------
info "注册 systemd 服务..."
cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=ServerSmith — 轻量流量额度管理与服务器运维面板
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/.env
ExecStart=/usr/local/bin/serversmith
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"
ok "服务已启动 (serversmith)"

# ---------- 完成 ----------
IP=$(curl -s ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ServerSmith 安装完成!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "  管理面板:  ${CYAN}http://$IP:18080${NC}"
echo -e "  健康检查:  ${CYAN}http://$IP:18080/health${NC}"
echo ""
echo -e "  数据目录:  $DATA_DIR"
echo -e "  systemd:   systemctl status ${SERVICE_NAME}"
echo -e "  日志:      journalctl -u ${SERVICE_NAME} -f"
echo ""
