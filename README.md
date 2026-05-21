# 聊天软件项目 / Chat Application Project

## 项目结构 / Project Structure

- **server_api/**: Go 后端服务（HTTP + WebSocket）
- **server_web/**: React 网页前端
- **client/**: Go CLI 客户端
- **docs/**: 项目文档
- **scripts/**: 构建和部署脚本

## 快速开始 / Quick Start

### 1. 启动数据库 / Start Database
```bash
docker-compose up -d
```

### 2. 运行后端 / Run Backend
```bash
cd server_api
go run cmd/api/main.go
```

### 3. 运行前端 / Run Frontend
```bash
cd server_web
npm install
npm run dev
```

### 4. 运行客户端 / Run Client
```bash
cd client
go run cmd/cli/main.go
```

