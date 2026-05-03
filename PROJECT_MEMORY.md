# ServerSmith 项目核心记忆库

**项目名称**：ServerSmith — 轻量流量额度管理与服务器运维面板
**架构负责**：穆凝雪 (13-OTC)
**开发负责**：玲玲 (14-Software)
**最高指挥官**：Boss（叶语）
**创建日期**：2026-04-28
**版本**：V1.5（P5 完成）

---

## 🟢 当前状态（2026-05-03 18:40 CST）

| 阶段 | 状态 | 验收标准 |
|------|------|----------|
| P0 骨架+API+探针+仪表盘 | ✅ 完成 | 18080:200 / 13000:200 |
| P1 服务器管理 UI | ✅ 完成 | 同上 |
| P2 告警+成本+TopEmby | ✅ 完成 | 同上 |
| P3 详情页+批量重启+探针管理+告警配置接入 | ✅ 完成 | 18080:200 / 13000:200 |
| P4 Makefile + 数据迁移 + 探针版本 | ✅ 完成（凝雪审查通过） | 双服务架构 |
| P5 Go embed 二周目 + 生产加固 | ✅ 完成 | 单二进制内嵌前端 + TD 修补 |

---

## 📦 P3 核心交付清单（2026-05-03）

### 后端新增/修改

| 文件 | 改动 |
|------|------|
| `db/sqlite.go` | `servers` 表新增 `agent_version TEXT` 字段（含迁移） |
| `api/sql_queries.go` | `listServersSQL` / `getServerByIDSQL` 均加入 `s.agent_version` |
| `api/plans.go` | `scanServerRow` 新增 `agentVersion sql.NullString` 扫描字段 |
| `api/servers.go` | `ServerResponse` 新增 `AgentVersion *string` |
| `api/report.go` | 从 `X-Agent-Version` header 读取并写入 DB（仅 UPDATE，非 INSERT 时更新）|
| `api/reboot_tasks.go` | 新建：`GET /api/servers/:id/reboot-tasks?limit=5` |
| `api/batch_reboot.go` | 新建：`POST /api/servers/batch-reboot`（最多 50 台，并发执行，含 reboot_tasks 记录）|
| `main.go` | 注册 `/api/servers/batch-reboot`（必须在 `/api/servers/` 通配前注册）+ `/api/servers/:id/reboot-tasks` |

### 前端新增/修改

| 文件 | 改动 |
|------|------|
| `hooks/useServers.ts` | `Server` 类型加 `agent_version: string \| null` |
| `hooks/useRebootTasks.ts` | 新建：拉取指定服务器最近 N 条重启记录 |
| `components/ServerDetail.tsx` | 新建：服务器详情面板（基础信息+流量+重启记录+手动重启按钮）|
| `components/ProbeTable.tsx` | 新建：探针状态总览表（在线/离线/未安装判定，agent_version 展示）|
| `components/ServerList.tsx` | 全量重写：加 checkbox 多选 + 批量重启按钮；名称点击跳详情页 |
| `app/servers/[id]/page.tsx` | 新建：服务器详情页（≤15行 shell）|
| `app/servers/[id]/edit/page.tsx` | 加「🔔 告警配置」按钮，弹 `AlertConfigForm` |
| `app/probes/page.tsx` | 新建：探针管理页（≤11行 shell）|
| `app/layout.tsx` | 导航栏加「探针管理」入口 |

---

## 🩸 P3 踩坑记录（不可再犯）

### 坑1：`/api/servers/batch-reboot` 路由必须在通配 `/api/servers/` 之前注册
- **现象**：如果 `mux.HandleFunc("/api/servers/", ...)` 先注册，`/api/servers/batch-reboot` 会被通配拦截，匹配不到正确 handler。
- **铁律**：精确路由必须在通配路由之前注册。`batch-reboot` 是固定路径，必须单独先注册。

### 坑2：韩文乱码路径（修改了错误路径的文件）
- **现象**：上下文 summary 里有两处文件路径含韩文字符（`系统개발`），导致写入了错误路径。
- **铁律**：文件路径一律用 `系统开发`（简体中文），发现路径异常必须先确认后再写。

### 坑3：`scanServerRow` 列数不对称会导致 scan 静默失败返回 nil
- **现象**：SQL 加了 `agent_version` 列但 `Scan()` 没加对应变量，所有 server 查询返回 nil，前端空白。
- **铁律**：SELECT 列数必须和 Scan 变量数量严格一一对应，改 SQL 必须同步改 Scan。

### 坑4：agentVersion 变量声明位置不能在 else 块内
- **现象**：`report.go` 中一开始把 `agentVersion := ...` 放在 `} else {` 块里，导致上面的 if 块引用时编译报错。
- **铁律**：跨 if/else 共用的变量必须提前到两个分支之前声明。

---

## 🔑 核心架构不变量

- **Go 管理端端口**：18080
- **Next.js 前端端口**：13000
- **SQLite 数据路径**：`/data/serversmith.db`（容器内），宿主机挂载到项目 `data/` 目录
- **时区**：UTC 存储，Asia/Shanghai 显示；`calcNextCycle` 用 `db.TZ` 解析 Asia/Shanghai；`expire_at` 用 23:59:59 落 UTC 防止跨天
- **NowUTC() 唯一来源**：`db/timezone.go`，其他地方不得重定义
- **page.tsx 红线**：≤ 20 行，业务逻辑全在组件和 hook 里
- **单文件行数红线**：< 300 行，超限必须拆
- **事务内禁止网络请求**：Emby/SSH 调用必须在事务提交后

---

## ⏭️ 下一步（待 P3 凝雪审查）

- P3 等待 13-OTC (凝雪) 审查
- 审查通过后视情况是否推进 P4（生产部署 / Go embed 打包前端 / 探针 musl 编译）
- SSH 实际 reboot 需要管理服务器配好 SSH 密钥才能对节点生效（当前 SSH 超时失败是预期行为）

- 项目路径：`/Users/omnileo/.openclaw/Data_Source/系统开发/ServerSmith/`
- Go 管理端：`server/` 目录
- Agent 探针：`agent/` 目录
- Next.js 前端：`web/` 目录
- SQLite 数据：`data/` 目录
- Docker 开发沙盒：`docker-compose.yml`

### ✅ P0-2: Go 管理端 API
- API 端口：18080
- SQLite 初始化含 WAL 模式、6 张核心表自动建表
- 路由注册：`/api/report`, `/api/servers`, `/api/dashboard`, `/api/alerts`
- 探针上报接口完整实现（自动注册、增量计算、阈值告警）
- 服务器 CRUD
- 仪表盘汇总 & 按线路汇总
- 告警列表
- 注意：本地无 Go 环境，编译需在 Docker 容器内执行

### ✅ P0-3: 探针 Agent
- 采集 CPU/内存/磁盘/网卡流量/uptime
- 每 30 秒上报，失败重试 3 次（间隔 5 秒）
- 零本地存储
- 部署传 `--server` 和 `--server-id` 参数

### ✅ P0-4: 流量累计逻辑
- 管理端收到上报 → 查上一条快照 → 算网卡增量 → 累加 used_bytes
- 自动检查 100%/90% 流量警戒线 → 写 alerts
- CPU >85% / 内存 >90% / 磁盘 >90% 自动告警
- 计费周期复位 Cron：每 5 分钟扫描，归零 used_bytes + SSH 重启
- 离线检测 Cron：每 60 秒扫描 5 分钟无上报的服务器
- 到期提醒 Cron：每天 08:00 扫描

### ✅ P0-5: 前端仪表盘
- Next.js + Tailwind CSS + shadcn/ui
- 端口：13000
- 深色运维驾驶舱
- 首屏：总数/在线/离线/月支出/剩余流量/未读告警
- 线路卡片：线路类型/节点数/总额度/已用/剩余/进度条/月租/每GB成本/异常数
- 最近告警列表
- Next build 通过 ✅

### 🔄 凝雪审查打回后修复记录（2026-05-03）

| # | 问题 | 修复方式 | 状态 |
|---|------|----------|------|
| 1 | Docker 构建失败（无 go.sum） | Dockerfile 改为先 COPY go.mod → go mod download → 再 COPY 源码；已生成 go.sum | ✅ |
| 2 | NowUTC 重复定义 | sqlite.go 删除 NowUTC，统一由 timezone.go 导出 | ✅ |
| 3 | servers.go 410 行超红线 | 拆为 servers.go(CRUD, 239行) / plans.go(118行) / reboot.go(124行) / sql_queries.go(42行) | ✅ |
| 4 | SSH 端口被更新为 0 | updateServer 改为先读当前值，仅覆盖明确传入的字段 | ✅ |
| 5 | 手动重启假装成功 | reboot.go 实现真实 SSH（10s 超时 + context + 失败记录） | ✅ |
| 6 | 周期自动重启占位符 | cron/cycle_reset.go sshReboot 改为真实 SSH 命令（10s 超时） | ✅ |
| 7 | 前端 page.tsx 近 300 行 | 拆为 StatCard / CarrierCard / AlertsList 组件 + useDashboard hook | ✅ |
| 8 | Docker Compose version 过时 | 移除 `version: '3.8'` | ✅ |

### 🔄 P0 复审四尾巴修复（2026-05-03 15:46 CST）

| # | 问题 | 修复方式 | 验收 |
|---|------|----------|------|
| 1 | 前端容器不常驻 | web Dockerfile.dev 改为全量 COPY + npm install，compose 移除 command 覆盖，添加 restart: unless-stopped，指定 networks 与 server 同网段 | ✅ 13000:200 |
| 2 | Dockerfile 吞错误 | 移除 `2>/dev/null; exit 0`，改为直接 `RUN go mod download`，失败即失败 | ✅ 构建失败不再被伪装 |
| 3 | SSH 用户写死 root | 新增 `ssh_user` 列到 servers 表（含迁移），ServerRequest/ServerResponse 含 ssh_user，CRUD 支持读写，reboot.go/cron 均使用 db 中的 ssh_user | ✅ 编辑/创建可设用户，重启不再写死 root |
| 4 | 缺核心测试 | 3 个测试覆盖：流量增量累计 / 北京时间复位计算（5 组 case） / 到期日不偏移（3 组 case） + 2 个 db 时区测试 | ✅ 5/5 PASS |

**全量验收（2026-05-03 15:57 CST）**：
- ✅ Docker 容器列表：server 和 web 均 running
- ✅ Go 生产构建：BUILD OK
- ✅ Go vet：零警告
- ✅ Go 测试：5/5 PASS（流量增量、BJ 复位、到期日偏移、NowUTC、时区偏移）
- ✅ 前端 13000：HTTP 200
- ✅ 后端 18080：HTTP 200

**验收结果（2026-05-03 15:40 CST）**：
- ✅ Docker build server（dev + prod 镜像均通过）
- ✅ Go build 零错误
- ✅ Go vet 零警告
- ✅ ESLint 零错误
- ✅ Next build 通过

### ✅ P1: 服务器管理 UI（2026-05-03 16:30 CST）

| 模块 | 路径 | 行数 |
|------|------|------|
| 配置常量 | `lib/config.ts` | 4 |
| CRUD Hook | `hooks/useServers.ts` | 112 |
| 历史 Hook | `hooks/useHistory.ts` | 40 |
| 服务器列表表格 | `components/ServerList.tsx` | 216 |
| 表单组件 | `components/ServerForm.tsx` | 255 |
| 流量折线图 | `components/TrafficChart.tsx` | 103 |
| Toast通知 | `components/Toast.tsx` | 30 |
| 列表页 | `app/servers/page.tsx` | 48 |
| 新增页 | `app/servers/new/page.tsx` | 32 |
| 编辑页 | `app/servers/[id]/edit/page.tsx` | 80 |
| 历史图页 | `app/servers/[id]/history/page.tsx` | 48 |

后端修改：`updateServer` 增加 plan_type 变更处理（DELETE 旧 plan + INSERT 新 plan，全事务）

支侟库：`recharts@3.8.1`（已加入 package.json）

**P1 全量验收**：
- ✅ Go build + vet 零警告
- ✅ Next.js build 成功（6 个路由 all OK）
- ✅ 容器 18080:200 / 13000:200 / 13000/servers:200
- ✅ 单文件均 < 300 行

---

## 📦 P4 核心交付清单（2026-05-03 18:10 CST）

| 任务 | 改动 | 说明 |
|------|------|------|
| Go embed | `server/web.go` 新建 + `main.go` 注册 `/` 路由 | Next.js 静态产物嵌入 Go 二进制，不再需要独立 Node 服务 |
| Makefile | `Makefile` 新建 | `make build-all` 一键编译 darwin+linux amd64+linux arm64 |
| Dockerfile 重构 | `server/Dockerfile` 三阶段构建 | Node build → Go build → 最小 alpine 运行时 |
| docker-compose 精简 | `docker-compose.yml` 移除 web 服务 | 开发环境只跑 Go 后端 |
| 数据迁移 | `server/db/sqlite.go` 补 `expire_notify_days` | ✅ 通过 |
| 探针版本上报 | `agent/main.go` + `server/api/report.go` | ✅ 通过 |
| API 地址统一 | `web/src/lib/config.ts` / `useDashboard.ts` | ✅ 通过 |
| `.gitignore` / `README.md` | 新建 | ✅ 通过 |
| Go embed 内嵌前端 | `server/web.go` + `next.config.ts` `output:'export'` | ❌ **被否决** |

### 🩸 P4 要点记录

1. **Go embed 方案被否决**：Next.js 16 的 `output: 'export'` 与 `'use client'` + 动态路由不兼容。
2. **当前架构**：回到双服务部署（Go API 18080 + Node 前端 13000）。
3. `API_BASE` 默认值改为空字符串，`useDashboard.ts` 的硬编码已消除。

---

## 📦 P5 核心交付清单（2026-05-03 18:40 CST）

### 子任务 A：Go embed 二周目

| 文件 | 操作 | 说明 |
|------|------|------|
| `web/src/app/servers/layout.tsx` | 新建 | layout 层提供 `generateStaticParams`，构建时请求 API 生成所有服务器 ID |
| `web/next.config.ts` | 修改 | 加回 `output: 'export'` + `trailingSlash: true` |
| `server/web.go` | 重建 | Go embed + SPA fallback（文件不存在时回退到 index.html） |
| `server/main.go` | 修改 | 注册 embed handler + 版本变量 `Version` |
| `Makefile` | 改写 | 两阶段构建 `build-web` → `build-server`，顺序依赖 |
| `server/Dockerfile` | 修改 | 三阶段构建传递 ldflags 版本号 |

**构建流程**：`make build-web` → web/out/ 静态导出 → `make build-server` → web-out/ 拷贝到 server/ → Go embed 编译。API 不可达时 `generateStaticParams` 返回空数组，不阻塞构建。

**SPA fallback 机制**：`webHandler()` 先尝试精确路径匹配，404 时回退到 `/`（index.html），由客户端路由处理未预生成的动态路径。

### 子任务 B：生产加固

| 文件 | 操作 | 说明 | 关联 TD |
|------|------|------|---------|
| `server/cron/expiry_notify.go` | 重写 | 读 `server_plans.expire_notify_days` 字段，每台服务器独立通知天数（fallback 7） | TD-01 ✅ |
| `server/api/servers.go` | 审阅 | plan_type 变更清理逻辑在 P1 已实现（DELETE旧+INSERT新+全事务），无需新增 | TD-03 ✅ 已修 |
| `server/api/backup.go` | 新建 | `POST /api/backup` 执行 VACUUM INTO 生成时间戳备份文件 | TD-05 ✅ |
| `server/main.go` | 修改 | 注册 `/api/backup` 路由 | |

### P5 红线确认

1. ✅ `generateStaticParams` 在 layout 层，不透传 `"use client"` 页面
2. ✅ 构建时 API 调用有 5 秒超时 + try/catch，不可达时返回空数组
3. ✅ Go embed 有 SPA fallback，404 路由回退 index.html
4. ✅ 到期提醒读库 `expire_notify_days`，不再硬编码

### 🩸 P5 要点记录

1. **Go embed 解决方案**：P4 被否决后，P5 通过 `servers/layout.tsx` 服务端 layout + `generateStaticParams` 打通静态导出路径。三个 `'use client'` 页面无需改动渲染逻辑。
2. **TD-03 已不存在**：`updateServer` 中 plan_type 变更的清理逻辑在 P1 已经实现，是记忆库中遗留的过期债务条目。
3. **备份可靠性**：`VACUUM INTO` 是 SQLite 官方推荐的热备份方式，无需锁表，不影响运行中的服务。

---

## 🎯 1. 核心目标

替代哪吒探针，管理老板的 100+ 台服务器，核心解决四个问题：

1. **分线路流量看板** — 按线路类型（4837/CN2/CMI等）汇总剩余流量，点进去看每台节点明细
2. **独立计费周期+自动复位** — 每台节点独立设置复位日，到点自动重启并归零流量
3. **成本管理** — 记录每台服务器的月租，流量节点自动算每GB成本
4. **健康监控 + 到期提醒** — 所有服务器（不分类型）监控健康状态和续费到期日

---

## 🏗️ 2. 架构总览

```
┌──────────────────────────────────────────────┐
│  管理服务器 (1C1G 低配，SSD 20G+)              │
│  Go 单二进制 + SQLite (WAL) + 内嵌前端         │
│                                               │
│  对外端口: 8080 (Web) + 端口 (探针上报)         │
│  SSH 免密登录到 100+ 节点                       │
└──────────────┬───────────────────────────────┘
               │
   ┌───────────┼──────────────┐
   ▼           ▼              ▼
┌───────┐ ┌───────┐      ┌───────┐
│节点01  │ │节点02  │ ...  │节点100 │
│GoEdge  │ │GoEdge  │      │GoEdge  │
│+探针   │ │+探针   │      │+探针   │
└───────┘ └───────┘      └───────┘

每台节点：探针二进制 (3-5MB) + systemd 自启
探针 → POST /api/report (每 30 秒上报系统指标)
管理端 → SSH reboot (到计费周期日)
```

---

## 📊 3. 数据模型（ER 设计）

### 3.1 表结构

#### 表一：servers（服务器主表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增主键 |
| name | TEXT NOT NULL | 服务器名称/备注，如"日本-软银-01" |
| host | TEXT NOT NULL | IP 地址或域名 |
| ssh_port | INTEGER DEFAULT 22 | SSH 端口 |
| server_type | TEXT NOT NULL DEFAULT 'node' | 类型: 'node'（流量节点） / 'project'（项目部署） |
| carrier_type | TEXT | 线路类型，节点机填写: '4837' / 'CN2_GIA' / 'CN2_GT' / 'CMI' / 'CUII' / '软银' / '9929' / '其他'. 项目服务器填 NULL |
| monthly_rent | REAL | 月租费用（元/月），所有服务器都填 |
| status | TEXT DEFAULT 'offline' | 当前状态: online / offline / unknown |
| last_report_at | DATETIME | 最近一次探针上报时间 |
| created_at | DATETIME DEFAULT NOW | 创建时间 |
| updated_at | DATETIME DEFAULT NOW | 更新时间 |

#### 表二：server_plans（计费配置）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增主键 |
| server_id | INTEGER FK → servers.id | 关联服务器 |
| plan_type | TEXT NOT NULL | 计费类型: 'traffic'（流量节点） / 'no_limit'（项目服务器） |
| --- 以下是 plan_type='traffic' 时有效 --- | | |
| total_quota | REAL | 月总流量额度（GB） |
| initial_used_gb | REAL DEFAULT 0 | ⭐ **期初已用流量（GB）**——添加服务器时手动输入，表示当前周期已经用了多少 |
| cycle_day | INTEGER | 计费复位日（1-31，如5号则=5） |
| cycle_time | TEXT DEFAULT '02:00' | 复位时间（如 02:00 凌晨） |
| last_cycle_end | DATETIME | 上个计费周期结束时间 |
| next_cycle_at | DATETIME | 下个计费周期开始时间（管理端自动计算） |
| used_bytes | INTEGER DEFAULT 0 | 当前周期已用流量（字节，管理端=期初值+探针增量累计） |
| --- 以下是 plan_type='no_limit' 时有效 --- | | |
| expire_at | DATETIME | 服务器到期时间 |
| expire_notify_days | INTEGER DEFAULT 7 | 到期前 N 天开始提醒 |

**🚨 关键！used_bytes 初始化逻辑：**
```
添加节点时：
  used_bytes = initial_used_gb × 1,000,000,000（转为字节）
从那之后：
  探针每30秒上报一次网卡总流量
  管理端算出增量，累加到 used_bytes 上
```
这样你第一次添加服务器时，输入

**🚨 核心设计说明（Boss 的血泪教训）：**
- 流量复位和服务器重启无关。流量按 **计费周期** 算，不是按重启事件算。
- 哪怕手动重启了节点，流量计数器不清零——等到 `next_cycle_at` 才清零。
- `used_bytes` 不存储在探针上，探针只上报网卡总字节，由管理端计算增量。
- 这样节点临时重启→增量继续计算，完全不受影响。

#### 表三：traffic_snapshots（流量快照——用于计算增量）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增主键 |
| server_id | INTEGER FK → servers.id | |
| report_at | DATETIME | 上报时间 |
| net_in | INTEGER | 本机网卡总入字节 |
| net_out | INTEGER | 本机网卡总出字节 |
| cpu_percent | REAL | CPU 使用率 |
| mem_percent | REAL | 内存使用率 |
| disk_percent | REAL | 磁盘使用率 |

**存储策略**：
- 实时数据：traffic_snapshots 保留最近 7 天
- 每天凌晨 03:00 归档旧数据到 `traffic_history` 汇总表（按天聚合）
- 或直接定时清理 7 天以上数据

#### 表四：traffic_history（历史汇总——趋势图表用）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| server_id | INTEGER FK | |
| date | DATE | 日期 |
| bytes_in | INTEGER | 当天总入 |
| bytes_out | INTEGER | 当天总出 |
| avg_cpu | REAL | 当天平均 CPU |
| avg_mem | REAL | 当天平均内存 |

#### 表五：reboot_tasks（重启任务记录）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| server_id | INTEGER FK | |
| action | TEXT | 'cycle_reboot' / 'manual_reboot' |
| triggered_at | DATETIME | 触发时间 |
| executed_at | DATETIME | 执行时间 |
| result | TEXT | 'success' / 'failed' / 'pending' |
| error_msg | TEXT | 失败原因 |

#### 表六：alerts（告警记录）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| server_id | INTEGER FK | |
| alert_type | TEXT | 'traffic_exceeded' / 'expiring' / 'offline' / 'high_cpu' / 'high_mem' / 'high_disk' |
| message | TEXT | |
| level | TEXT | 'info' / 'warning' / 'critical' |
| created_at | DATETIME | |
| acknowledged | BOOLEAN DEFAULT FALSE | |

---

## 🔄 4. API 接口设计

### 4.1 探针上报接口（由 Agent 调）

```
POST /api/report
Body: {
  "server_id": "agent的内置标识或IP",
  "net_in": 1256789000,
  "net_out": 987654321,
  "cpu_percent": 23.5,
  "mem_percent": 62.1,
  "disk_percent": 45.0,
  "uptime_sec": 360000
}
Response: { "ok": true, "next_cycle_at": "2026-05-05T02:00:00" }
```

管理端收到后：
1. 更新 `servers.status = 'online'`, `last_report_at = NOW`
2. 查上一次快照，计算流量增量 → 累加到 `used_bytes`
3. 存储新快照到 `traffic_snapshots`
4. 检查是否超限 → 写告警

### 4.2 服务器管理 API

```
GET    /api/servers          — 服务器列表（含实时状态、已用/剩余流量）
POST   /api/servers          — 添加服务器（含计费配置 + 线路类型）
GET    /api/servers/:id      — 单台详情（含流量使用率、健康趋势）
PUT    /api/servers/:id      — 编辑服务器信息
DELETE /api/servers/:id      — 删除服务器

GET    /api/servers/:id/history?days=7  — 历史趋势数据
POST   /api/servers/:id/reboot          — 手动重启
GET    /api/dashboard                   — 总览看板（总数/在线/超限/到期）
GET    /api/dashboard/by-carrier        — 按线路类型汇总剩余流量（核心API）
GET    /api/alerts                      — 告警列表
```

#### `/api/dashboard/by-carrier` 返回格式：
```json
[
  {
    "carrier_type": "4837",
    "server_count": 5,
    "total_quota_gb": 20000,
    "used_gb": 7800,
    "remaining_gb": 12200,
    "usage_percent": 39,
    "total_rent": 2500.00,
    "cost_per_gb": 0.125
  },
  {
    "carrier_type": "CN2",
    "server_count": 2,
    "total_quota_gb": 5000,
    "used_gb": 1800,
    "remaining_gb": 3200,
    "usage_percent": 63,
    "total_rent": 1800.00,
    "cost_per_gb": 0.36
  }
]
```

#### `/api/dashboard/costs` 返回格式（成本页面）：
```json
{
  "total_monthly_rent": 12500.00,
  "node_rent": 9800.00,
  "project_rent": 2700.00,
  "by_carrier": [
    { "carrier": "4837", "rent": 2500, "quota_gb": 20000, "used_gb": 7800, "cost_per_gb": 0.125 },
    { "carrier": "CN2", "rent": 1800, "quota_gb": 5000, "used_gb": 1800, "cost_per_gb": 0.36 }
  ]
}
```
`cost_per_gb` 计算方式：该线路所有节点月租总和 ÷ 该线路总流量额度

### 4.3 计费周期复位（Cron 任务）

```
每 5 分钟扫描一次：
  SELECT * FROM server_plans
  WHERE plan_type = 'traffic'
    AND next_cycle_at <= NOW

  对命中的每条：
  ① 记录 used_bytes 到上一个周期的报表（可选）
  ② used_bytes = 0   ← 归零
  ③ 计算 next_cycle_at = 下个月的 cycle_day cycle_time
  ④ 执行 SSH reboot 该节点
  ⑤ 写 reboot_tasks 记录
```

### 4.4 到期提醒（Cron 任务）

```
每天 08:00 执行一次：
  SELECT * FROM servers s
  JOIN server_plans p ON s.id = p.server_id
  WHERE expire_at IS NOT NULL
    AND expire_at BETWEEN NOW AND NOW + expire_notify_days
    AND 今天还没提醒过

  对命中的每条 → 写 alerts (alert_type='expiring')
```

**所有服务器（节点/项目）都参与到期提醒，只要设置了 expire_at 就行。**

### 4.5 离线检测

```
每 1 分钟扫描：
  SELECT * FROM servers WHERE last_report_at < NOW - 5 分钟
  AND status = 'online'

  对命中的每条 → status = 'offline', 写 alerts (alert_type='offline')
```

---

## 🧾 4.6 成本统计逻辑

### 4.6.1 数据来源
- 每台服务器添加时手动输入 `monthly_rent`（月租）
- 流量节点额外输入 `total_quota`（月流量总额度，GB）

### 4.6.2 老板需求（简化为月度固定成本）

Boss 的核心诉求：**不需要按周期算、不需要按天折算，只看每月固定的总成本。**

- 每台服务器按月租算（月租是固定的，不管复位日是5号还是10号）
- 每台节点的流量配额也是按月固定的
- 所有服务器的计费周期不一样，但统一按「一个月」算成本

### 4.6.3 计算公式

```
全局每月总支出 = 全部服务器的 monthly_rent 之和（含节点+项目）

按线路计算每GB成本：
  某线路月总支出 = 该线路下所有节点 monthly_rent 之和
  某线路月总配额 = 该线路下所有节点 total_quota 之和
  某线路每GB成本 = 总月租 / 总配额
```

### 4.6.4 成本数据来源

| 数据 | 来源 | 说明 |
|------|------|------|
| 服务器月租 | 添加服务器时手动输入 `monthly_rent` | 固定值，不改就不变 |
| 节点流量配额 | 添加节点时手动输入 `total_quota` | 固定值 |
| 总支出 | 全部月租相加 | 每月固定这么多 |
| 每GB成本 | 月租÷流量 | 算一次就知道平均值了 |

**🚨 不追踪月度内按天的成本波动**——哪个服务器几号复位不影响成本，成本只看每月固定支出。

---

## 🧩 5. 探针 Agent 设计

### 5.1 功能

- 采集：CPU%、内存%、磁盘%、网卡总流量（in/out）、uptime
- 每 30 秒 POST 到管理端
- 管理端收到上报后比较上次值算流量增量
- 如果 POST 失败（管理端挂了），重试 3 次，每次间隔 5 秒，之后原地等待下次采集
- 本地不存任何数据，不缓存

**🚨 重点：探针不上报周期内的历史已用流量**——探针只上报当前网卡总字节数，初始已用流量由管理端 `server_plans.initial_used_gb` 承担。

**例：你添加一台节点时输入「已用 500GB」，之后探针每30秒上报网卡字节增量，管理系统：已用 = 500G + 增量，正确。**

### 5.2 部署方式

```bash
# 节点机上一次命令搞定
wget -O /usr/local/bin/serversmith-agent https://管理端地址/download/agent
chmod +x /usr/local/bin/serversmith-agent

# 创建 systemd 服务
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

systemctl enable serversmith-agent
systemctl start serversmith-agent
```

### 5.3 探针上报时服务器标识

100 台机器的探针，管理端怎么知道数据是谁报的？两种方案：

- **方案 A（推荐）**：部署时传入 `--server-id`（用 IP 或主机名），管理端用 IP 做唯一标识
- **方案 B**：部署时传一个管理端生成的 token 令牌

推荐方案 A，简单粗暴，IP 唯一。

---

## 🔗 6. TopEmby 对接

TopEmby 管理后台侧边栏加一行按钮：

```tsx
// 位置: src/app/(admin)/admin/layout.tsx 或侧边栏配置文件
{
  key: "server-smith",
  label: "服务器管理",
  icon: <ServerIcon />,
  href: "https://管理端域名:8080",
  target: "_blank"
}
```

**仅此而已。** 不共享数据库，不共用密码，各自独立部署。

---

## 🧱 7.5 本地开发环境架构（13-OTC 补充定稿）

**定位**：Docker 只作为本地开发沙盒，用来保证 Go 管理端、Next.js 前端、SQLite 数据文件的开发环境可复现。生产目标仍然是 **Go 单二进制 + SQLite + embed 内嵌前端**，不允许演变成复杂容器编排项目。

### 7.5.0 当前宿主机 Docker/端口基线（2026-05-03 15:07 CST）

已运行容器：

| 容器 | 用途/归属 | 宿主机端口 |
|------|-----------|------------|
| topemby-app-1 | TopEmby 应用 | 3000 |
| filebrowser-data-source | Data_Source 文件浏览 | 8080 |
| topemby-redis-1 | TopEmby 内部 Redis | 未暴露宿主机端口 |
| topemby-postgres-1 | TopEmby 内部 Postgres | 未暴露宿主机端口 |
| emby-redis | 旧/独立 Redis | 6379 |
| emby-pg | 旧/独立 Postgres | 5432 |

宿主机已占用关键端口：

| 端口 | 占用方 | 结论 |
|------|--------|------|
| 3000 | TopEmby | ServerSmith 前端开发禁止使用 |
| 8080 | filebrowser-data-source | ServerSmith 管理端禁止使用 |
| 5432 | emby-pg | 禁止碰，且本项目不需要 Postgres |
| 6379 | emby-redis | 禁止碰，且本项目不需要 Redis |
| 11434 | Ollama | 禁止碰 |

**ServerSmith 本地端口分配定稿**：

| 模块 | 宿主机端口 | 说明 |
|------|------------|------|
| Go 管理端 API | 18080 | 避开现有 8080 |
| Next.js 前端开发服务 | 13000 | 避开现有 3000 |
| 探针模拟上报 | 不单独暴露 | 直接上报到 18080 |
| SQLite | 不占端口 | 数据文件挂载到项目目录 |

**端口红线**：玲玲不准占用 3000、8080、5432、6379、11434。开发前必须先确认端口未被占用。若未来端口冲突，优先调整 ServerSmith，不准动 TopEmby、FileBrowser、Postgres、Redis、Ollama。

### 7.5.1 开发环境组成

| 模块 | 要求 | 红线 |
|------|------|------|
| Go 管理端 | 独立服务进程，负责 API、Cron、SQLite、SSH 重启任务 | 不准拆成微服务 |
| SQLite | 使用本地挂载的数据目录保存数据库文件，开启 WAL | 不准把数据藏在容器临时层 |
| Next.js 前端 | 使用 Next.js + Tailwind CSS + shadcn/ui 开发 | 禁止 Ant Design、Element Plus、外部 CDN |
| 探针 Agent | 独立 Go 程序，可在本地模拟上报 | 不准让探针承担历史流量存储 |
| 反向代理 | 本地开发阶段不需要 | 不准引入 Nginx/Caddy 增加复杂度 |
| 缓存/队列 | 不需要 | 禁止 Redis、RabbitMQ、Kafka 等无意义组件 |

### 7.5.2 本地启动边界

- 本地开发允许用 Docker 一键启动开发环境。
- Docker 里只允许承载 Go 管理端、Next.js 前端开发服务、SQLite 数据目录。
- SQLite 数据目录必须挂载到项目目录下，确保容器删除后数据不丢。
- 前端开发阶段可以独立热更新；正式交付时必须构建静态产物并由 Go embed 打进管理端。
- 探针上报可用本地模拟 Agent 验证，不需要真实 100 台机器。

### 7.5.3 生产交付边界

最终交付必须满足：

1. 管理端是一个 Go 单二进制。
2. 前端静态资源已内嵌进 Go 二进制。
3. SQLite 数据文件外置保存，可备份、可迁移。
4. 探针 Agent 是独立静态二进制，可丢到 Linux 机器直接运行。
5. 不依赖 Node.js 运行时、不依赖 Docker、不依赖外部 CDN。

### 7.5.4 玲玲执行顺序

1. 先搭本地 Docker 开发沙盒。
2. 再创建 Go 管理端骨架与 SQLite 初始化。
3. 再创建 Next.js + Tailwind CSS + shadcn/ui 前端骨架。
4. 再实现探针本地模拟上报。
5. 再做仪表盘第一版。
6. 最后验证前端可构建并被 Go embed 内嵌。

**裁决**：本地 Docker 是开发脚手架，不是生产架构。谁把 ServerSmith 写成 Docker Compose 堆料项目，直接打回。

---

## 🕒 7.6 时间与时区架构红线（13-OTC 补充定稿）

**定位**：ServerSmith 涉及计费周期复位、到期提醒、离线检测、流量快照、重启任务记录。时间语义必须统一，否则会再次出现「看起来差 8 小时」「复位日提前/延后」「到期提醒错天」这种低级灾难。

### 7.6.1 存储规则

| 场景 | 存储标准 | 显示标准 |
|------|----------|----------|
| traffic_snapshots.report_at | UTC 时间 | 前端按 Asia/Shanghai 显示 |
| servers.last_report_at | UTC 时间 | 前端按 Asia/Shanghai 显示 |
| reboot_tasks.triggered_at/executed_at | UTC 时间 | 前端按 Asia/Shanghai 显示 |
| alerts.created_at | UTC 时间 | 前端按 Asia/Shanghai 显示 |
| server_plans.next_cycle_at | UTC 时间，由本地业务时区换算后落库 | 前端按 Asia/Shanghai 显示 |
| server_plans.expire_at | 业务日期按 Asia/Shanghai 解释，落库时转 UTC | 前端按 Asia/Shanghai 显示 |

### 7.6.2 业务时区定稿

- ServerSmith 的唯一业务时区：**Asia/Shanghai**。
- Boss 设置的「每月 5 号 02:00 复位」，含义固定为 **Asia/Shanghai 的 5 号 02:00**。
- 后端计算 `next_cycle_at` 时，必须先按 Asia/Shanghai 解释 cycle_day + cycle_time，再转换成 UTC 存储。
- 前端展示任何时间时，必须明确按 Asia/Shanghai 格式化，不准让浏览器自由猜。

### 7.6.3 禁止行为

1. 不准把本地时间字符串无时区落库。
2. 不准前端直接传 `new Date()` 生成的模糊时间给后端当业务时间。
3. 不准用服务器系统时区作为业务时区。
4. 不准在 SQLite 里混存 UTC、北京时间、无时区字符串。
5. 不准把 `expire_at` 当纯 UTC 日期直接比较，否则会出现到期日错一天。
6. 不准用前端浏览器所在时区决定复位时间。

### 7.6.4 三类时间语义

| 类型 | 例子 | 规则 |
|------|------|------|
| 事件时间 | 上报时间、告警创建、重启执行 | 后端生成 UTC，前端转 Asia/Shanghai 展示 |
| 业务计划时间 | next_cycle_at、复位日+复位时间 | 用户输入按 Asia/Shanghai 解释，后端转 UTC 存储 |
| 业务日期 | 到期日 expire_at | 用户选择的是 Asia/Shanghai 日期，不是浏览器随便解析的 UTC 日期 |

### 7.6.5 审查验收要求

玲玲交付时必须说明：

1. SQLite 里时间字段是否统一 UTC。
2. 前端展示是否统一 Asia/Shanghai。
3. cycle_day + cycle_time 如何计算 next_cycle_at。
4. 到期日 expire_at 如何避免错一天。
5. 离线判断是否基于 UTC 时间差，而不是字符串比较。

**裁决**：时间不是 UI 装饰，是业务主轴。ServerSmith 如果出现时区混存，直接打回。全是漏洞。

### 7.6.6 ✅ 时区实现方案（2026-05-03 已修复）

**时间库**：`db/timezone.go` 提供统一时区工具函数。

| 需求 | 实现 | 代码位置 |
|------|------|----------|
| UTC 落库 | `db.NowUTC()` 返回 UTC ISO8601 | `db/timezone.go` |
| 北京时区定义 | `db.TZ = Asia/Shanghai` | `db/timezone.go`（init 中加载） |
| 前端展示 | 前端 JS 将 UTC 字符串转 `Asia/Shanghai` 显示 | 前端 `lib/timezone.ts` |
| 复位日+时间 | 用户输入解释为 `db.TZ` → `UTC` 存 `next_cycle_at` | `calcNextCycle()` 用 `time.Now().In(db.TZ)` 构 |
| 到期日 | 用户输入 `2026-06-01` → `Asia/Shanghai 23:59:59` → `UTC` | `createServer()` ÿ 用 `time.ParseInLocation` |
| Cron 触发 | 到期检查按 Asia/Shanghai 08:00 触发 | `expiry_notify.go` 用 `bjNow.Hour()==8` |
| 离线检测 | 基于 UTC 时间差（`last_report_at` UTC vs NOW UTC - 5min）| `offline_check.go` ✅ 已有 |

**逐个对照审计清单**：
1. ✅ SQLite 所有时间字段统一 UTC（`datetime('now')` = UTC）
2. ✅ 前端展示时由 JS 转为 Asia/Shanghai
3. ✅ `cycle_day` + `cycle_time` 作为 Asia/Shanghai 时间，转 UTC 存 `next_cycle_at`
4. ✅ `expire_at` 防止错一天：Beijing 23:59:59 转 UTC，避免 UTC 零点跨天
5. ✅ 离线判断基于 UTC 时间差，不是字符串比较

---

## 🔴 7. 项目红线（玲玲必须遵守）

1. **carrier_type 在添加服务器时手动选择**——预置常见线路类型：4837、CN2_GIA、CN2_GT、CMI、CUII、软银、9929、其他。支持扩展输入

2. **`/api/dashboard/by-carrier` 只统计 plan_type='traffic' 的节点**——项目服务器不参与线路汇总

3. **used_bytes 初始化 = initial_used_gb + 探针增量**——添加节点时必须输入期初已用量，不能为空。系统不可自己猜

4. **探针不上报 ≠ 节点离线**——管理端超过 5 分钟没收到上报才标记 offline，上报更新 `last_report_at` 就用这个字段判断

5. **used_bytes 按日历周期累加，不以服务器重启为基准**——即使手动重启或意外重启，流量计数器不清零。只有到 `next_cycle_at` 自动复位的时候才清零

6. **SSH 重启必须超时控制**——ssh 命令设置 10 秒超时，单台节点重启失败不影响其他节点

7. **部署脚本只发文件不加密**——Boss 的环境不急加密，先跑通再说。等稳定了再加探针身份认证

8. **前端页面用内嵌静态文件**——Go 编译时用 `embed` 直接打包前端，不依赖 CDN 或外部资源

   **前端基础骨架强制标准**：全面封杀 Ant Design、Element Plus 等老旧企业后台框架。ServerSmith 前端必须使用 **Next.js + Tailwind CSS + shadcn/ui** 作为基础骨架。UI 风格定为「深色运维驾驶舱 + 苹果系统般极简现代质感」。组件可以由玲玲自由组合，但底层颜料盒必须是 shadcn/ui，不准回退到廉价企业后台模板。

9. **不自行实现批量部署脚本**——Boss 有 100 台机器，初次部署让玲玲写个 for 循环脚本就行，不需要在管理端实现批量推送

10. **探针二进制用 musl 交叉编译**——编译为静态链接，扔到任何 Linux 机器上都能跑，不依赖 glibc 版本

---

## 📋 8. 开发优先级

| 优先级 | 模块 | 依赖 | 预估工时 |
|--------|------|------|----------|
| P0 | 探针 agent（采集+上报） | 无 | 0.5 天 |
| P0 | 管理端收数 + SQLite 存储 | 上一步 | 0.5 天 |
| P0 | 流量累计 + 仪表盘展示 | 前两步 | 1 天 |
| P1 | 节点管理 CRUD（添加/编辑） | 无 | 0.5 天 |
| P1 | 计费周期复位 Cron + SSH 重启 | 所有上述 | 1 天 |
| P2 | 到期提醒 Cron | 无 | 0.5 天 |
| P2 | 历史趋势折线图 | 前序 | 0.5 天 |
| P3 | 告警通知（Telegram Bot） | 所有上述 | 0.5 天 |
| P3 | TopEmby 跳转按钮 | 无 | 0.2 天 |

---

## ✅ 13-OTC 签发（技术选型定稿）

### 技术选型理由

**选 Go**，理由如下：

| 维度 | Go | Node.js |
|------|----|---------|
| 探针部署 | 单二进制 3-5MB，scp 过去就能跑 | 需要 Node 运行时 + npm i，每台装依赖 |
| 100 台并发 | goroutine 原生并发，毫秒级处理 100 路上报 | 事件循环 100 路没问题，但单次计算量大时卡主线程 |
| 内存占用 | 管理端跑起来 20-50MB | 空跑 Node 就 30-50MB，加 Express 等翻倍 |
| 跨平台 | 交叉编译 musl，任何 Linux 都能跑 | 依赖 Node 版本，glibc 版本容易踩坑 |
| 长期维护 | 标准库强大，依赖极少 | npm 依赖树，三五年后可能一堆 CVE |
| TopEmby 对接 | 独立服务，语言不相关，HTTP 跳转即可 | — |

**结论**：Go 更适合这个场景。探针二进制扔到 100 台机器上零依赖，管理端部署也只是一个文件，省心。

### 复位时间策略
- 每台服务器在添加时**手动输入**复位周期日和复位时间
- 管理端根据输入自动计算 `next_cycle_at`
- 没有输入复位时间的节点 → 不走流量额度管理，只监控健康度+到期日

### 告警策略
- 面板内仪表盘变色展示（绿/黄/红），不推送 Telegram
- 在 TopEmby 后台加一个按钮跳转过去，随手看一眼就行

---

**架构版本**：V1.0
**签发人**：穆凝雪
**日期**：2026-04-28 21:42 CST
**状态**：✅ 已定稿，等待 Boss 下发玲玲执行
