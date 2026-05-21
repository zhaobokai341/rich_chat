# 项目结构说明 / Project Structure

## 简化后的目录结构 / Simplified Directory Structure

```
rich_chat/
├── server_api/              # Go 后端服务
│   ├── cmd/api/            # 应用入口 (main.go)
│   ├── internal/handler/   # HTTP/WebSocket 处理器
│   └── go.mod
│
├── server_web/             # React 网页前端
│   ├── src/
│   │   ├── App.tsx         # 主组件
│   │   └── main.tsx        # 入口文件
│   ├── index.html
│   ├── package.json
│   └── vite.config.ts
│
├── client/                 # Go CLI 客户端
│   ├── cmd/cli/            # 应用入口 (main.go)
│   ├── internal/ui/        # Bubble Tea UI
│   └── go.mod
│
├── docs/                   # 文档
│   └── PROJECT_STRUCTURE.md
│
├── docker-compose.yml      # PostgreSQL + Redis
├── ARCHITECTURE.md         # 架构文档
└── README.md               # 项目说明
```

## 核心文件说明 / Core Files

### server_api/
- `cmd/api/main.go` - 服务器入口，启动 Gin + WebSocket
- `internal/handler/handler.go` - 请求处理器

### server_web/
- `src/main.tsx` - React 应用入口
- `src/App.tsx` - 主应用组件

### client/
- `cmd/cli/main.go` - CLI 客户端入口
- `internal/ui/ui.go` - Bubble Tea TUI 界面

## 下一步 / Next Steps

1. **启动数据库**: `docker-compose up -d`
2. **实现 server_api**: 添加 Gin、WebSocket、数据库连接
3. **实现 client**: 添加 Bubble Tea、Resty、WebSocket 客户端
4. **实现 server_web**: 添加 React 组件、API 调用
