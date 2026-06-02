# 项目结构说明 / Project Structure

## 1. 整体架构概览

项目主要分为三个核心部分：
*   **Client (`client/`)**: 基于 Go 的命令行客户端，负责用户交互与服务端通信。
*   **Server API (`server_api/`)**: 基于 Go + Gin 的后端服务，提供 RESTful API，处理业务逻辑与数据持久化。
*   **Server Web (`server_web/`)**: 基于 React + TypeScript + Vite 的前端 Web 应用。

---

## 2. Server API (`server_api/`) - 后端服务层

后端采用了严格的**分层架构**，从上至下依次为 Handler、Service 和 Repository 层。

### 2.1 接入层
*   **职责**: 处理 HTTP 请求，参数解析，调用 Service 层，返回响应。
*   **核心文件**:
    *   `main.go`: 程序入口，初始化路由、数据库、服务和中间件。
    *   `web_server_page.go`: 定义了具体的 HTTP Handler 函数（如 `Login`, `Register`, `GetUserProfile`）。
    *   `safe_policy.go`: 包含安全相关的中间件（如 JWT 校验、IP 限流）。

### 2.2 业务逻辑层
*   **职责**: 封装核心业务逻辑，不直接操作数据库或 HTTP 细节。通过接口定义服务契约。
*   **核心文件**:
    *   `service/interfaces.go`: 定义了 `AuthService`, `UserService`, `TokenService` 等接口及通用错误类型。
    *   `service/auth_service.go`: 实现登录、注册、Token 生成逻辑。
    *   `service/user_service.go`: 实现用户信息查询、修改、删除逻辑。
    *   `service/token_service.go`: 负责 JWT 和验证令牌的生成与校验。
    *   `service/factory.go`: 负责初始化和组装所有服务实例（依赖注入容器）。

### 2.3 数据访问层
*   **职责**: 负责与数据库（PostgreSQL）和缓存交互，提供数据持久化能力。
*   **核心文件**:
    *   `database/interfaces.go`: 定义了 `UserRepository`, `RateLimitRepository`, `TokenRepository` 等数据访问接口。
    *   `database/user_repository.go`: 用户数据的 PostgreSQL 实现。
    *   `database/cached_user_repository.go`: 带有 Redis 缓存装饰器的用户仓储实现。
    *   `database/redis_adapter.go`: Redis 缓存的具体实现。
    *   `database/database_service.go`: 数据库服务的统一入口，管理各 Repository 实例。

### 2.4 基础设施与配置
*   **核心文件**:
    *   `config.go`: 定义配置常量及从环境变量加载配置。
    *   `redis_manager.go`: Redis 连接管理。
    *   `repeatitive_function_provide.go`: 提供日志记录等通用辅助函数。

---

## 3. Client (`client/`) - 命令行客户端

客户端同样采用了分层设计，主要分为 UI、Service 和 API 通信层。

### 3.1 表现层
*   **职责**: 处理终端用户输入输出，控制应用流程。
*   **核心文件**:
    *   `ui_handler.go`: 核心交互逻辑，处理菜单导航、登录/注册流程、密码输入等。
    *   `repeatitive_function_provide.go`: 提供美化打印和简化输入的辅助函数。

### 3.2 业务逻辑层
*   **职责**: 封装客户端的业务操作，如认证状态管理、用户资料管理。
*   **核心文件**:
    *   `auth_service.go`: 处理登录、注册、登出逻辑及凭证存储。
    *   `user_service.go`: 处理获取资料、修改资料、删除账号逻辑。

### 3.3 通信与数据层
*   **职责**: 负责与服务端 API 进行 HTTP 通信。
*   **核心文件**:
    *   `api_client.go`: 定义 `APIClient` 接口及 `RestAPIClient` 实现，封装了所有 API 调用。
    *   `http_client.go`: 封装底层的 HTTP 客户端配置。
    *   `config_manager.go`: 管理本地配置文件（如 Token 和 UserID 的读写）。
    *   `token_extractor.go`: 负责解析 JWT Token。

---

## 4. Server Web (`server_web/`) - Web 前端

基于现代前端技术栈构建的单页应用。

*   **技术栈**: React, TypeScript, Vite。
*   **核心文件**:
    *   `src/main.tsx`: 应用入口组件。
    *   `vite.config.ts`: 构建工具配置，包含 API 代理设置（将 `/api` 请求代理到后端 `localhost:2316`）。
    *   `package.json`: 项目依赖与脚本定义。
    *   `tsconfig.json`: TypeScript 编译配置。

---

## 5. 公共资源与文档

*   **文档 (`docs/`)**:
    *   `PROJECT_STRUCTURE.md`: 项目结构说明。
    *   `server_api/service/`: Service 层重构文档（架构、迁移指南、测试策略）。
    *   `server_api/database/`: Database 层重构文档（Repository 模式、缓存策略）。
    *   `DEPLOYMENT.md`: 部署指南。
*   **国际化 (`lang_pack/`)**:
    *   `client/main.json`: 客户端多语言包。
    *   `server_api/main.json`: 服务端 API 多语言包。
*   **工具 (`lang_pack_load/`)**:
    *   `lang_pack_load.go`: 语言包加载逻辑。