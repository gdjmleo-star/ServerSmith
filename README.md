# ServerSmith — 轻量流量额度管理与服务器运维面板

## 项目简介

替代哪吒探针，管理多台服务器的流量额度、计费周期、健康监控与成本核算。

## 构建

```bash
# 一键编译所有平台
make build-all

# 或单独编译目标平台
make build-linux-amd64
make build-linux-arm64
make build-darwin
```

编译产物在 `dist/` 目录：

```
dist/
├── serversmith-darwin-arm64
├── serversmith-darwin-amd64
├── serversmith-linux-amd64
└── serversmith-linux-arm64
```

## 部署

### 单二进制部署（推荐）

```bash
# 上传到管理服务器
scp dist/serversmith-linux-arm64 root@your-server:/usr/local/bin/serversmith

# 创建数据目录
mkdir -p /data/serversmith

# 运行
DB_PATH=/data/serversmith/serversmith.db PORT=8080 /usr/local/bin/serversmith
```

### 探针部署

```bash
# 在每台节点上执行
wget -O /usr/local/bin/serversmith-agent https://管理端地址/download/agent
chmod +x /usr/local/bin/serversmith-agent

# 注册 systemd 服务
cat > /etc/systemd/system/serversmith-agent.service << EOF
[Unit]
Description=ServerSmith Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/serversmith-agent --server http://管理端IP:8080 --server-id IP地址
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl enable --now serversmith-agent
```

### Docker 开发环境

```bash
make dev
# 后端: http://localhost:18080
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `18080` | 管理端监听端口 |
| `DB_PATH` | `/data/serversmith.db` | SQLite 数据库路径 |
| `TZ` | `Asia/Shanghai` | 时区 |
