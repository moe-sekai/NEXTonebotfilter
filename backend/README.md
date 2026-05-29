# NEXTonebotfilter Backend

详见仓库根目录的 README.md。

## 模块结构

```
backend/
├── cmd/nextonebotfilter/     程序入口(filter manager + 控制台 API)
├── internal/filter/          反向 WS 网关 + 规则编译 + 事件总线
├── internal/store/           GORM + SQLite 持久化层与领域模型
└── internal/server/          控制台 REST/SSE API
```

## 与 Moebot 主仓的差异

- 用 `store.Store` 接口替代 `*database.DB`,默认实现 SQLite 自带,不依赖 Moebot
- 移除 `internal_app.go` 中的插件相关 seed 逻辑——独立运行时无插件
- `wsClient.systemTransport` 字段保留,但默认始终为 false;独立网关不需要
  Moebot 内置传输闸门
- 增加 HTTP 控制台 API(REST + SSE),对应控制台前端
