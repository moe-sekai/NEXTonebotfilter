# NEXTonebotfilter

独立运行的 OneBot v11 反向 WebSocket 过滤网关,以及一个基于 Next.js + Claude.ai 风格的现代控制台。

源自 [OneBotFilter](https://github.com/ProtoDeath2333/OneBotFilter) 思路,但代码基线来自 Moebot-NEXT-Go 内置的 filter 模块——已加入消息去重、模板系统、运行时事件流、YAML 进出口等增强。

```
NEXTonebotfilter/
├── backend/    Go 后端(filter 网关 + 控制台 REST/SSE API + SQLite 持久化)
└── console/    Next.js 控制台(App Router + Tailwind + SWR)
```

## 后端

```bash
cd backend
go run ./cmd/nextonebotfilter -db data/filter.db -console :8787
```

启动后:
- 反向 WS 网关默认监听 `0.0.0.0:3939/ws`(可在控制台「网关设置」修改)
- 控制台 REST/SSE API 监听 `:8787/api/*`

也可以直接 build:

```bash
cd backend
go build -o nextonebotfilter.exe ./cmd/nextonebotfilter
```

加上 `-web console/out` 让 Go 进程顺便托管 Next.js 静态导出产物,生产环境就只用一个端口。

### 主要 API

- `GET  /api/health`
- `GET  /api/status` 网关 + 上下游连接快照
- `POST /api/gateway/restart`
- `GET/PUT  /api/gateway`
- `GET/POST /api/apps` · `PUT/DELETE /api/apps/{id}`
- `GET/POST /api/templates` · `GET/PUT/DELETE /api/templates/{id}`
- `GET  /api/events?limit=200` · `GET /api/events/stream` (SSE)
- `GET  /api/yaml/export` · `POST /api/yaml/import`
- `POST /api/regex/test`

## 控制台

```bash
cd console
npm install
npm run dev    # http://localhost:3939
```

`/api/*` 默认通过 Next.js rewrite 反代到 `http://127.0.0.1:8787`,可用环境变量
`NEXT_PUBLIC_BACKEND_URL` 覆盖。

控制台页面:

- **概览** — 网关 / 上游 / 下游连接卡片
- **下游 App** — 配置每个下游 ws 客户端,可选模板或自定义规则
- **规则模板** — 复用规则;`default` 模板兼任全局 ID 兜底
- **实时事件** — SSE 订阅 allow / block / prefix_pass / up / down
- **网关设置** — Host/Port/Suffix/Token/去重等
- **YAML 进出** — 与原版 OneBotFilter 配置兼容

## 与 Moebot 的关系

后端逻辑直接源自 `Moebot-NEXT-Go/internal/filter`,把对 Moebot 主仓数据库 / 模型的依赖替换成内置的 `internal/store`(GORM + SQLite),不再依赖任何 Moebot 主仓的代码。

## 启动脚本

构建脚本和启动脚本是分开的:**build 一次,start 任意多次**。`start.*` 不依赖 Node、不会跑 npm,只启动那个二进制文件。

| 场景 | Linux / macOS | Windows |
| --- | --- | --- |
| 一次性构建(产出 `backend/nextonebotfilter.exe`,内嵌 Next.js 控制台) | `./build.sh` | `build.cmd` |
| 启动(单进程,日志同步进终端 + `data/nextonebotfilter.log`) | `./start.sh` | `start.cmd` |

`start.*` 会检查二进制是否存在,缺了会自动调一次 `build.*`。**之后再启动就完全不碰 Node**,前端已经被 `go:embed` 进去了,不需要 `console/out` 也不需要 `node_modules`。

默认参数:
- 控制台 + API: `http://localhost:8787`(`PORT=9000 ./start.sh` 改端口)
- OneBot 反向 WS 网关: `0.0.0.0:3939/ws`(在控制台「网关设置」中调整)
- 数据库: `data/nextonebotfilter.db`,日志: `data/nextonebotfilter.log`

## Docker

```bash
docker compose up -d --build
# 控制台 + API: http://localhost:8787
# OneBot 反向 WS:  ws://localhost:3939/ws  (可在控制台改端口)
```

镜像采用三阶段构建:`node:22-alpine` 出 Next.js 静态产物 → `golang:1.23-alpine` 把它们 `embed` 进 Go 二进制(纯 Go SQLite,无 CGO)→ `alpine:3.20` 运行。SQLite 数据持久化到 `./data`。
