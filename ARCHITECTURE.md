# 聊天软件架构文档

## 1. 项目概述

本文档概述了一个支持万级并发用户的高性能聊天软件架构。系统包含三个主要组件：

- **server_api**: 核心业务逻辑处理器，提供统一的 HTTP + WebSocket API
- **server_web**: 网页前端（React 单页应用），消费 server_api 接口
- **client**: 命令行聊天客户端，使用 server_api

### 1.1 通信模型

**统一 API 服务器:**

- **HTTP REST API**: 用户注册、登录、获取历史消息、用户管理等非实时操作
- **WebSocket**: 实时消息推送、在线状态更新、输入指示等实时功能
- **单端口服务**: 两种协议在同一端口上运行（如 :8080）
- **共享认证机制**: HTTP 和 WebSocket 使用相同的 JWT Token 进行认证

---

## 2. 技术栈推荐

### 2.1 后端服务 (server_api)

#### **首选：Go 语言**

**推荐理由:**

- ✅ **高性能**: 编译型语言，网络应用性能接近 C/C++
- ✅ **并发处理**: 内置 goroutine 模型，完美处理万级并发连接
- ✅ **语法简洁**: 清晰、极简的语法，易于学习和维护
- ✅ **丰富的生态**: 成熟的库支持（Gin、Echo、gRPC、WebSocket、Redis 客户端）
- ✅ **单文件部署**: 静态编译，部署简单
- ✅ **HTTP + WebSocket 统一**: Gin/Echo 框架支持在同一端口同时处理两种协议
- ✅ **生产就绪**: Docker、Kubernetes、Discord、Twitch 等都用于类似场景

**推荐框架:**

- **Gin**: 轻量级、快速、支持中间件（推荐）
- **Echo**: 现代化、功能丰富、文档良好
- **Gorilla WebSocket**: Go 最流行的 WebSocket 库
- **GORM/sqlx**: 数据库 ORM/查询构建器

**备选方案: Rust**

- 优点: 内存安全，极致性能
- 缺点: 学习曲线陡峭，语法更复杂

---

### 2.2 网页端 (server_web)

#### **首选：React + TypeScript + Vite**

**推荐理由:**

- ✅ **高性能**: 快速的虚拟 DOM、代码分割、懒加载
- ✅ **生态系统**: 最大的生态系统（antd、Material-UI、Socket.IO 客户端）
- ✅ **类型安全**: TypeScript 防止许多运行时错误
- ✅ **现代工具链**: Vite 提供闪电般的开发体验
- ✅ **HTTP + WebSocket 客户端**: Axios 处理 REST API + 原生 WebSocket API

**备选方案: Vue 3 + TypeScript**

- 优点: 学习曲线更平缓，API 更简单
- 缺点: 生态系统相比 React 较小

---

### 2.3 命令行客户端 (client)

#### **首选：Go + Bubbletea**

**推荐理由:**

- ✅ **跨平台**: 单一代码库可编译为 Windows/Linux/macOS
- ✅ **简单的终端 UI 库**: Bubbletea、tview、gocui 支持良好的终端界面
- ✅ **与后端同语言**: 代码共享，更易维护
- ✅ **小体积二进制**: 无运行时依赖

**备选方案: Python + Rich**

- 优点: 快速原型开发，丰富的 CLI 库（rich、prompt_toolkit）
- 缺点: 性能较慢，需要 Python 运行时

---

### 2.4 数据库

#### **主数据库：PostgreSQL**

**推荐理由:**

- ✅ **可靠性**: ACID 兼容，数据完整性好
- ✅ **JSON 支持**: 存储灵活的消息元数据
- ✅ **高性能**: 优秀的读写性能
- ✅ **扩展性**: 全文搜索、地理空间查询等扩展

**缓存和实时数据：Redis**

**推荐理由:**

- ✅ **速度**: 内存操作，亚毫秒级响应
- ✅ **发布订阅**: 完美的实时消息分发
- ✅ **数据结构**: 列表、集合用于消息队列
- ✅ **会话管理**: 用户会话、在线状态

**备选方案: MongoDB**

- 优点: 无 schema，水平扩展能力强
- 缺点: 事务操作一致性较差

---

### 2.5 通信协议

**统一 server_api 方案:**

| 组件 | 协议 | 技术 | 用途 |
|------|------|------|------|
| 网页 ↔ server_api | HTTP REST + WebSocket | Axios + 原生 WS | 单一端点处理所有操作 |
| CLI 客户端 ↔ server_api | HTTP REST + WebSocket | Go net/http + Gorilla WS | 与网页相同 |
| 内部服务 | gRPC（可选） | Protocol Buffers | 未来微服务扩展 |

**为何选择此架构:**

1. **简化部署**
   - 单个服务器二进制文件同时处理 HTTP 和 WebSocket
   - 只需管理、监控一个端口
   - 负载均衡配置更简单

2. **更好性能**
   - Web 服务器和 API 服务器之间无额外跳转
   - 浏览器直连 WebSocket 到 API
   - 实时功能延迟更低

3. **更易开发**
   - 认证、验证、业务逻辑单一代码库
   - 共享中间件（CORS、限流、日志）
   - HTTP 和 WebSocket 错误处理一致

4. **可扩展性**
   - 通过 sticky sessions 或 Redis Pub/Sub 水平扩展
   - 如需要可稍后拆分为微服务
   - CDN 友好，适用于静态资源

---

## 3. 系统架构

### 3.1 高层架构

**推荐架构（简化版）:**

```
┌──────────────────┐
│   网页浏览器      │
│   (React SPA)    │
└────────┬─────────┘
         │
         ├─────────────────────────────┐
         │ HTTP REST + WebSocket       │
         ▼                             │
┌──────────────────┐                   │
│   server_api     │◄──────────────────┤
│   (Go: Gin + WS) │                   │
│   端口 :8080      │                   │
└────────┬─────────┘                   │
         │                             │
         ├─────────────────────────────┤
         │ HTTP REST + WebSocket       │
         ▼                             │
┌──────────────────┐                   │
│   CLI 客户端      │                   │
│   (Go 二进制)     │                   │
└──────────────────┘                   │
         │
         ▼
┌──────────────────┐
│   PostgreSQL     │
│   + Redis        │
└──────────────────┘
```

**与原设计的核心区别:**

1. **移除 server_web 后端**: 前端现在是静态单页应用，由 Nginx/CDN 提供服务
2. **客户端直连 server_api**: 网页和 CLI 都直接连接到同一个 API
3. **单一通信层**: HTTP + WebSocket 在同一端口
4. **更简单的数据流**: 无中间 server_web 层

### 3.2 server_api 结构

**统一服务器组件:**

```
server_api/
├── HTTP 路由（REST API）
│   ├── POST /api/v1/auth/register      # 用户注册
│   ├── POST /api/v1/auth/login         # 登录并获取 JWT Token
│   ├── GET  /api/v1/users/me           # 获取当前用户信息
│   ├── GET  /api/v1/conversations      # 列出会话列表
│   ├── POST /api/v1/conversations      # 创建新会话
│   ├── GET  /api/v1/messages           # 获取历史消息
│   └── PUT  /api/v1/users/profile      # 更新个人资料
│
├── WebSocket 端点
│   └── GET /ws                          # WebSocket 升级
│       ├── 通过 Header 中的 JWT Token 认证
│       ├── 处理实时消息
│       ├── 在线状态更新
│       ├── 输入指示
│       └── 实时通知
│
└── 共享中间件
    ├── 认证（JWT 验证）
    ├── CORS 处理
    ├── 限流
    ├── 日志记录
    └── 错误处理
```

### 3.3 连接流程

**网页客户端连接流程:**

```
1. 从 CDN/Nginx 加载单页应用
   ↓
2. HTTP POST /api/v1/auth/login → 获取 JWT Token
   ↓
3. 建立 WebSocket 连接到 ws://server_api/ws?token={jwt}
   ↓
4. 通过 WebSocket 进行实时通信
   ↓
5. HTTP GET /api/v1/messages 获取历史消息
```

**认证流程:**

```
HTTP 登录请求:
POST /api/v1/auth/login
{
  "username": "john",
  "password": "secret"
}

响应:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",  // JWT token
  "user_id": 123,
  "username": "john"
}

WebSocket 连接:
GET ws://localhost:8080/ws?token=eyJhbGciOiJIUzI1NiIs...
升级为 WebSocket → 开始实时通信
```

---

## 4. 数据库设计

### 4.1 PostgreSQL 表结构

```sql
-- 用户表
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    status VARCHAR(20) DEFAULT 'offline',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 会话表
CREATE TABLE conversations (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL, -- 'direct'（私聊）或 'group'（群聊）
    name VARCHAR(100), -- 群聊名称
    created_at TIMESTAMP DEFAULT NOW()
);

-- 会话参与者表
CREATE TABLE conversation_participants (
    conversation_id BIGINT REFERENCES conversations(id),
    user_id BIGINT REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

-- 消息表
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT REFERENCES conversations(id),
    sender_id BIGINT REFERENCES users(id),
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text', -- text, image, file
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 性能索引
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_conversation_participants_user ON conversation_participants(user_id);
```

### 4.2 Redis 数据结构

```redis
# 在线用户
SET online_users:{user_id} {timestamp}

# 用户会话
HSET session:{user_id} token {token} device {device_info}

# 消息投递队列
LPUSH message_queue:{user_id} {message_json}

# 房间发布订阅
SUBSCRIBE chat_room:{conversation_id}

# 正在输入指示
SET typing:{conversation_id}:{user_id} {timestamp} EX 5
```

---

## 5. API 设计

### 5.1 HTTP REST API 端点

#### **认证与用户管理**

**用户注册:**
```http
POST /api/v1/auth/register
Content-Type: application/json

请求:
{
  "username": "john",
  "email": "john@example.com",
  "password": "secretpass123"
}

响应 (201 Created):
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com"
  }
}
```

**用户登录:**
```http
POST /api/v1/auth/login
Content-Type: application/json

请求:
{
  "username": "john",
  "password": "secretpass123"
}

响应 (200 OK):
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": 1,
  "username": "john",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

#### **会话管理**

**获取会话列表:**
```http
GET /api/v1/conversations
Authorization: Bearer {token}

响应 (200 OK):
{
  "conversations": [
    {
      "id": 123,
      "type": "direct",
      "name": null,
      "participants": [
        {"id": 1, "username": "john"},
        {"id": 2, "username": "jane"}
      ],
      "last_message": {
        "content": "Hello!",
        "sender_id": 2,
        "created_at": "2024-01-01T12:00:00Z"
      },
      "unread_count": 3
    }
  ]
}
```

**创建会话:**
```http
POST /api/v1/conversations
Authorization: Bearer {token}
Content-Type: application/json

请求（私聊）:
{
  "type": "direct",
  "participant_ids": [1, 2]
}

请求（群聊）:
{
  "type": "group",
  "name": "项目团队",
  "participant_ids": [1, 2, 3, 4]
}

响应 (201 Created):
{
  "success": true,
  "conversation_id": 456
}
```

#### **消息历史**

```http
GET /api/v1/messages?conversation_id=123&page=1&limit=50
Authorization: Bearer {token}

响应 (200 OK):
{
  "messages": [
    {
      "id": 789,
      "conversation_id": 123,
      "sender_id": 2,
      "sender_name": "Jane",
      "content": "你好吗？",
      "message_type": "text",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "has_more": true,
  "next_page": 2
}
```

#### **用户资料**

```http
GET /api/v1/users/me
Authorization: Bearer {token}

响应 (200 OK):
{
  "id": 1,
  "username": "john",
  "email": "john@example.com",
  "avatar_url": "https://example.com/avatar.jpg",
  "status": "online",
  "created_at": "2024-01-01T10:00:00Z"
}
```

---

### 5.2 WebSocket 事件

#### **连接建立**

```javascript
// 客户端使用 JWT Token 连接
const ws = new WebSocket('ws://localhost:8080/ws?token=eyJhbGciOiJIUzI1NiIs...');

ws.onopen = () => {
  console.log('WebSocket 已连接');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleWebSocketEvent(data);
};
```

#### **客户端 → 服务器事件**

**1. 发送消息:**
```json
{
  "event": "send_message",
  "request_id": "req_123456",
  "data": {
    "conversation_id": 123,
    "content": "Hello, World!",
    "type": "text"
  }
}
```

**2. 输入指示:**
```json
{
  "event": "typing_start",
  "data": {
    "conversation_id": 123
  }
}

// 用户停止输入时:
{
  "event": "typing_stop",
  "data": {
    "conversation_id": 123
  }
}
```

**3. 已读回执:**
```json
{
  "event": "mark_read",
  "data": {
    "conversation_id": 123,
    "message_id": 789
  }
}
```

**4. 心跳保持:**
```json
{
  "event": "ping",
  "timestamp": 1704067200
}
```

#### **服务器 → 客户端事件**

**1. 新消息:**
```json
{
  "event": "new_message",
  "data": {
    "message_id": 456,
    "conversation_id": 123,
    "sender_id": 2,
    "sender_name": "Jane",
    "sender_avatar": "https://example.com/jane.jpg",
    "content": "Hello, World!",
    "type": "text",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

**2. 用户在线状态更新:**
```json
{
  "event": "user_online",
  "data": {
    "user_id": 2,
    "status": "online",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

**3. 用户离线:**
```json
{
  "event": "user_offline",
  "data": {
    "user_id": 2,
    "status": "offline",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

**4. 输入指示:**
```json
{
  "event": "user_typing",
  "data": {
    "conversation_id": 123,
    "user_id": 2,
    "user_name": "Jane"
  }
}
```

**5. 已读回执:**
```json
{
  "event": "message_read",
  "data": {
    "conversation_id": 123,
    "user_id": 2,
    "message_id": 456
  }
}
```

**6. Pong（心跳响应）:**
```json
{
  "event": "pong",
  "timestamp": 1704067200
}
```

---

## 6. 扩展策略

### 6.1 支持万级并发

**水平扩展:**

- 在负载均衡器后方部署多个 server_api 实例
- WebSocket 连接使用 sticky sessions
- Redis Cluster 实现分布式缓存

**优化技术:**

- 数据库连接池
- 批量操作的批处理
- 消息历史懒加载
- CDN 用于静态资源（头像、文件）

**性能目标:**

- WebSocket 连接数: 每服务器实例 10K+
- 消息延迟: < 100ms (p95)
- API 响应时间: < 200ms (p95)
- 数据库查询: < 50ms (p95)

---

## 7. 安全考虑

1. **认证**: JWT Token + 刷新机制
2. **加密**: 所有连接使用 TLS，密码使用 bcrypt 哈希
3. **授权**: 基于角色的访问控制（RBAC）
4. **限流**: 防止滥用（每秒消息数限制）
5. **输入验证**: 清理所有用户输入
6. **SQL 注入防护**: 使用预处理语句

---

## 8. 开发路线图

### 第一阶段：基础（第 1-2 周）

- 搭建 Go 项目结构
- 实现基础 WebSocket 服务器
- 设计数据库 schema
- 创建认证系统

### 第二阶段：核心功能（第 3-4 周）

- 实现消息收发
- 构建会话管理
- 添加消息持久化
- 开发 CLI 客户端 MVP

### 第三阶段：网页界面（第 5-6 周）

- 搭建 React + TypeScript 项目
- 实现 WebSocket 客户端
- 构建聊天 UI 组件
- 添加用户资料管理

### 第四阶段：高级功能（第 7-8 周）

- 群聊支持
- 文件/图片分享
- 搜索功能
- 通知系统

### 第五阶段：优化与测试（第 9-10 周）

- 性能优化
- 负载测试（模拟 10K 用户）
- 安全审计
- 文档完善

---

## 9. 技术栈总结

| 层级 | 首选方案 | 备选方案 |
|------|---------|---------|
| **后端语言** | Go 1.21+ | Rust |
| **前端框架** | React + TypeScript | Vue 3 + TypeScript |
| **数据库** | PostgreSQL + Redis | MongoDB |
| **WebSocket 库** | Gorilla WebSocket | Socket.IO |
| **Web 服务器** | Nginx + Go | Caddy |
| **CLI 框架** | Bubbletea (Go) | Rich (Python) |
| **ORM** | GORM / sqlx | Prisma |
| **API 框架** | Gin / Echo | Fiber |
| **消息队列** | Redis Streams | RabbitMQ |
| **部署** | Docker + Docker Compose | Kubernetes |

---

## 10. 项目结构

```
rich_chat/
├── server_api/              # Go 后端
│   ├── cmd/
│   │   └── api/
│   │       └── main.go      # 应用入口
│   ├── internal/
│   │   ├── handler/         # WebSocket 和 HTTP 处理器
│   │   ├── service/         # 业务逻辑
│   │   ├── model/           # 数据模型
│   │   ├── repository/      # 数据库访问
│   │   ├── middleware/      # 认证、日志中间件
│   │   └── config/          # 配置管理
│   ├── pkg/
│   │   ├── websocket/       # WebSocket 工具
│   │   └── jwt/             # JWT 工具
│   ├── migrations/          # 数据库迁移
│   ├── go.mod
│   └── go.sum
│
├── server_web/              # 网页前端
│   ├── src/
│   │   ├── components/      # React 组件
│   │   ├── pages/           # 页面
│   │   ├── hooks/           # 自定义 Hooks
│   │   ├── services/        # API 服务
│   │   ├── store/           # 状态管理
│   │   └── types/           # TypeScript 类型
│   ├── package.json
│   └── vite.config.ts
│
├── client/                  # CLI 客户端
│   ├── cmd/
│   │   └── cli/
│   │       └── main.go
│   ├── internal/
│   │   ├── ui/              # 终端 UI 组件
│   │   ├── websocket/       # WebSocket 客户端
│   │   └── commands/        # CLI 命令
│   └── go.mod
│
└── docs/                    # 文档
    ├── ARCHITECTURE.md      # 架构文档
    ├── API.md               # API 文档
    └── DEPLOYMENT.md        # 部署文档
```

---

## 11. 性能基准参考

**Go WebSocket 性能:**

- 单服务器可处理: 50K+ 并发连接
- 消息吞吐量: 100K+ 消息/秒
- 每个连接内存: ~50KB

**Redis 性能:**

- 读取: ~100K 操作/秒
- 写入: ~80K 操作/秒
- 发布订阅延迟: < 1ms

**PostgreSQL 性能:**

- 简单查询: < 10ms
- 复杂联接: < 50ms
- 适当索引情况下

---

## 12. 监控与可观测性

1. **日志**: Zap（Go 结构化日志）
2. **指标**: Prometheus + Grafana
3. **追踪**: OpenTelemetry
4. **健康检查**: /health 端点
5. **告警**: AlertManager

---

## 结论

此架构为构建支持万级并发用户的可扩展聊天应用提供了坚实基础。推荐的技术栈（Go + React + PostgreSQL + Redis）提供：

- 高性能和并发处理能力
- 丰富的生态系统和社区支持
- 简洁、可维护的代码
- 在生产环境中经过验证的可扩展性

模块化设计支持增量开发，并可根据用户需求增长轻松扩展。
