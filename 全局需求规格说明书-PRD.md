# ServerSmith 全局需求规格说明书（PRD）

**项目代号**：ServerSmith
**版本**：V4.0（全生命周期覆盖）
**签发人**：穆凝雪 (13-OTC)
**签发时间**：2026-05-03 17:42 CST
**最后更新**：2026-05-03

---

## 1. 项目定位

轻量流量额度管理与服务器运维面板。一剑斩杀 Boss 的 100+ 台服务器的成本、流量、健康、到期四大痛点。

一句话：**"零废话运维"**——打开面板看到所有服务器的状态、成本、流量余量，不再需要登录每一台机器。

---

## 2. 用户故事

| ID | 用户 | 需求 | 优先级 |
|----|------|------|--------|
| US-01 | Boss | 打开驾驶舱就能看到所有服务器的在线/离线、月总支出、剩余流量 | P0 |
| US-02 | Boss | 添加一台新服务器（流量节点或项目机），设好月租和周期 | P0 |
| US-03 | Boss | 探针自动上报系统指标，管理端自动累计流量 | P0 |
| US-04 | Boss | 每月到计费日自动归零流量并重启服务器 | P0 |
| US-05 | Boss | 服务器离线 5 分钟自动标记 + 告警 | P0 |
| US-06 | Boss | 到期前 N 天自动提醒 | P0 |
| US-07 | Boss | 分线路（4837/CN2等）查看流量汇总和剩余 | P0 |
| US-08 | Boss | 服务器列表页看全量表格，分页、编辑、删除 | P1 |
| US-09 | Boss | 新增/编辑服务器表单，区分流量节点和项目机 | P1 |
| US-10 | Boss | 查看单台服务器的流量历史趋势图 | P1 |
| US-11 | Boss | 首页看到月度总支出 + 各线路每GB成本 | P2 |
| US-12 | Boss | 页头有 TopEmby 跳转按钮 | P2 |
| US-13 | Boss | 为每台服务器配置离线告警阈值和流量告警阈值 | P2 |
| US-14 | Boss | 告警列表页，区分未读/已读，可标记 | P2 |
| US-15 | Boss | 页头 NotificationBadge 显示未读告警数 | P2 |
| US-16 | Boss | 点击服务器名称进入详情页，看到完整信息 | P3 |
| US-17 | Boss | 多选服务器批量重启 | P3 |
| US-18 | Boss | 探针管理页，看哪些服务器装了探针、状态如何 | P3 |
| US-19 | Boss | 编辑页直接配置告警（不用跳其他页面） | P3 |
| US-20 | Boss | 生产部署：Go 单二进制 + embed 前端 + 不依赖 Node/Docker | P4 |
| US-21 | Boss | 跨平台编译（Linux ARM64 / Linux AMD64） | P4 |
| US-22 | Boss | OpenClaw Agent 集成（在模联端也能看到服务器状态） | 未来 |

---

## 3. 分阶段交付

### P0 — 基础设施（✅ 已完成）

| 模块 | 文件 |
|------|------|
| Go 项目骨架 + SQLite WAL + 6 张表自动建表 | `server/` |
| Docker 开发沙盒（Go API + Next.js 前端 + SQLite 挂载） | `docker-compose.yml` |
| API：上报/服务器CRUD/仪表盘/告警 | `server/api/` |
| Cron：周期复位/离线检测/到期提醒 | `server/cron/` |
| 探针 Agent（CPU/内存/磁盘/网卡，每30秒上报） | `agent/` |
| 前端仪表盘（暗色驾驶舱+StatCard+CarrierCard+告警列表） | `web/` |
| 时区模块（UTC存，北京显） | `server/db/timezone.go` |

### P1 — 服务器管理 UI（✅ 已完成）

| 模块 | 文件清单 |
|------|----------|
| 服务器列表页 | `app/servers/page.tsx`, `components/ServerList.tsx` |
| 新增服务器表单 | `app/servers/new/page.tsx`, `components/ServerForm.tsx` |
| 编辑服务器表单 | `app/servers/[id]/edit/page.tsx`, `components/ServerForm.tsx` |
| 流量历史趋势图 | `app/servers/[id]/history/page.tsx`, `components/TrafficChart.tsx` |
| CRUD Hook | `hooks/useServers.ts` |
| 历史数据 Hook | `hooks/useHistory.ts` |
| Toast 全局通知 | `components/Toast.tsx` |
| 配置常量 | `lib/config.ts` |

### P2 — 成本与告警管理（⚠️ 已交付，待 Boss 签收）

| 模块 | 文件清单 |
|------|----------|
| 成本展示（月总支出 + 线路单价） | `components/CostSummary.tsx`, `hooks/useCostData.ts` |
| TopEmby 导航 | `app/layout.tsx` |
| 告警配置对话框 | `components/AlertConfigForm.tsx` |
| 告警列表页 | `app/alerts/page.tsx`, `components/AlertsList.tsx` |
| 未读告警角标 | `components/NotificationBadge.tsx` |
| 告警 CRUD Hook | `hooks/useAlerts.ts` |
| 后端 alert_configs 表 + 迁移 + API | `server/api/alert_configs.go` |
| 后端告警已读 + 筛选 | `server/api/alerts.go` |
| 后端仪表盘扩展成本字段 | `server/api/dashboard.go` |

### P3 — 详情、批量、探针管理（💡 已签发蓝图）

| 模块 | 计划文件 |
|------|----------|
| 服务器详情只读页 | `app/servers/[id]/page.tsx`, `components/ServerDetail.tsx` |
| 编辑页接入告警配置 | 修改 `app/servers/[id]/edit/page.tsx` |
| 批量重启 + checkbox | 修改 `components/ServerList.tsx` |
| 后端批量重启 API | 修改 `server/api/servers.go` |
| 后端重启记录查询 | 新建 `server/api/reboot_tasks.go` |
| 探针管理页 | `app/probes/page.tsx`, `components/ProbeTable.tsx` |
| 探针版本号 migration | 修改 `server/db/sqlite.go` |

### P4 — 生产就绪（💡 蓝图待写）

| 模块 | 内容 |
|------|------|
| Go embed 内嵌前端 | `server/main.go` 使用 `embed.FS` |
| 跨平台编译 | `Makefile` + `GOOS`/`GOARCH` 交叉编译 |
| 单二进制交付 | 前端静态资源内嵌，不再需要 Node |
| 生产运行指令 | 文档：`nohup ./serversmith &` 即可运行 |

### 未来功能（💡 未排期）

| 模块 | 说明 |
|------|------|
| OpenClaw Agent 集成 | 在模联端看到 ServerSmith 数据 |
| 机柜图 | 可视化 100+ 台全局布局 |
| 流量对战 | 各线路流量占比可视化 |
| 成本趋势 | 月度成本变化曲线 |
| 服务器分组 | 按业务/线路/数据中心分组 |
| 用户权限 | 多用户登录，不同可见度 |

---

## 4. 技术债务与已知限制

| 编号 | 描述 | 严重度 | 计划修复 |
|------|------|--------|----------|
| TD-01 | 后端没有 `expire_notify_days` 字段参与到期扫描（P0写死的7天） | 中 | P4 |
| TD-02 | 前端 `useServer` 的 `avg_cpu`/`avg_mem` 定义在 TS 里但后端未返回 | 低 | 允许，不阻塞 |
| TD-03 | 编辑时 `plan_type` 变更后端没处理自动删/建 server_plans | 中 | P4 |
| TD-04 | 没有全量 End-to-End 测试 | 中 | P4 |
| TD-05 | 没有备份/恢复 SQLite 的功能 | 低 | P4 |

---

## 5. 架构红线（永久生效）

每次审查时使用第二条零妥协的标准：

1. **单文件不超 300 行**。超了就是烂。
2. **禁止硬编码 IP、Token、配置开关**。全部抽离到 `lib/config.ts`。
3. **禁止 Ant Design / Element Plus / 外部 CDN**。只允许 shadcn/ui。
4. **禁止 fetch 不包 try/catch**。API 不可达就红色 toast，不准白屏。
5. **禁止删除不二次确认**。
6. **禁止不加时区的无格式时间落库**。
7. **禁止 Docker Compose 成为生产架构**。Docker 只给开发用。
8. **禁止在业务组件中直接写死字符串和数字**。全部走 config 或 props。
9. **禁止页面组件超 100 行**（page.tsx 只做壳，业务逻辑下沉）。
10. **禁止 Schema 中用执行层资源（Emby账号/服务器）作为业务数据（订单/余额）的外键**。

---

## 6. 文件索引

| 文档 | 位置 |
|------|------|
| 项目核心记忆库（含ER图/数据流转） | `PROJECT_MEMORY.md` |
| P1 架构蓝图 | `P1架构蓝图-2026-05-03.md` |
| P2 架构蓝图 | `P2架构蓝图-2026-05-03.md` |
| P3 架构蓝图 | `P3架构蓝图-2026-05-03.md` |
| 全局需求规格（本文） | `全局需求规格说明书-PRD.md` |

---

## 7. 交付签收记录

| 阶段 | 状态 | 验收时间 | 验收人 |
|------|------|----------|--------|
| P0 | ✅ 通过 | 2026-05-03 15:57 | 13-OTC |
| P1 | ✅ 通过 | 2026-05-03 16:37 | 13-OTC |
| P2 | ⏳ 待 Boss 签收 | — | — |
| P3 | 💡 蓝图已签发，待开发 | — | — |
| P4 | 💡 蓝图待写 | — | — |

---

**文档版本**：V4.0
**签发**：穆凝雪
**时间**：2026-05-03 17:42 CST
