#!/bin/bash
# ServerSmith Agent 一键安装脚本 (Linux / macOS)
set -euo pipefail

MANAGER_URL="${MANAGER_URL:-}"
SERVER_ID="${SERVER_ID:-}"

if [ -z "$MANAGER_URL" ] || [ -z "$SERVER_ID" ]; then
  echo "用法: curl -sSL $MANAGER_URL/agent/install.sh | MANAGER_URL=http://your-server:18080 SERVER_ID=1 bash"
  exit 1
fi

# 检测架构和系统
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  GOARCH="amd64" ;;
  aarch64|arm64) GOARCH="arm64" ;;
  *) echo "[ERROR] 不支持的架构: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux)  PLATFORM="linux" ;;
  darwin) PLATFORM="darwin" ;;
  *) echo "[ERROR] 不支持的系统: $OS"; exit 1 ;;
esac

BINARY="serversmith-agent-${PLATFORM}-${GOARCH}"
INSTALL_PATH="/usr/local/bin/serversmith-agent"
SERVICE_NAME="serversmith-agent"

echo "[INFO] 平台: ${PLATFORM}/${GOARCH}"
echo "[INFO] 从 ${MANAGER_URL}/agent/${BINARY} 下载探针..."

curl -sSL --max-time 60 "${MANAGER_URL}/agent/${BINARY}" -o /tmp/serversmith-agent
chmod +x /tmp/serversmith-agent
mv /tmp/serversmith-agent "$INSTALL_PATH"
echo "[OK] 探针已安装到 $INSTALL_PATH"

# 写 systemd (Linux) 或 launchd (macOS)
if [ "$PLATFORM" = "linux" ]; then
  cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=ServerSmith Agent
After=network.target

[Service]
ExecStart=${INSTALL_PATH} --server ${MANAGER_URL} --server-id ${SERVER_ID}
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable $SERVICE_NAME
  systemctl restart $SERVICE_NAME
  echo "[OK] systemd 服务已启动"

elif [ "$PLATFORM" = "darwin" ]; then
  PLIST_PATH="/Library/LaunchDaemons/com.serversmith.agent.plist"
  cat > "$PLIST_PATH" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.serversmith.agent</string>
  <key>ProgramArguments</key>
  <array>
    <string>${INSTALL_PATH}</string>
    <string>--server</string><string>${MANAGER_URL}</string>
    <string>--server-id</string><string>${SERVER_ID}</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
</dict>
</plist>
EOF
  launchctl load "$PLIST_PATH"
  echo "[OK] launchd 服务已启动"
fi

echo ""
echo "✅ ServerSmith 探针安装完成！约 30 秒后面板将显示此服务器在线。"
