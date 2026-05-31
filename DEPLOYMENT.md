# 部署指南 / Deployment Guide

## 架构说明

本项目采用**前后端分离**架构，为避免跨域问题（CORS），推荐使用以下部署方案：

### 开发环境

前端通过 Vite 代理访问后端 API，无需 CORS 配置。

**Vite 代理配置** (`server_web/vite.config.ts`):
```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:2316',
      changeOrigin: true,
    },
  },
}
```

**启动开发环境**:
```bash
# Terminal 1: 启动后端
cd server_api
export $(grep -v '^#' .env | xargs)
go run .

# Terminal 2: 启动前端
cd server_web
npm run dev
```

前端访问 `http://localhost:5173/api/...` 会被自动代理到 `http://localhost:2316/api/...`

---

### 生产环境

#### 方案 1: Nginx 反向代理（推荐）⭐

使用 Nginx 同时提供前端静态文件和代理 API 请求，**完全避免 CORS 问题**。

**Nginx 配置文件** (`/etc/nginx/sites-available/rich_chat`):

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        root /var/www/rich_chat/dist;
        index index.html;
        try_files $uri $uri/ /index.html;
        
        # 缓存静态资源
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # API 反向代理（对浏览器来说是同源）
    location /api/ {
        proxy_pass http://localhost:2316/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持（如果未来需要）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
    gzip_min_length 1000;
}
```

**部署步骤**:

```bash
# 1. 构建前端
cd server_web
npm run build
# 生成 dist/ 目录

# 2. 复制前端文件到 Nginx
sudo cp -r dist/* /var/www/rich_chat/

# 3. 配置 Nginx
sudo ln -s /etc/nginx/sites-available/rich_chat /etc/nginx/sites-enabled/
sudo nginx -t  # 测试配置
sudo systemctl reload nginx

# 4. 启动后端（使用 systemd 或 supervisor）
cd server_api
export $(grep -v '^#' .env | xargs)
./server_api &
```

**优势**:
- ✅ 无需 CORS 配置
- ✅ 性能最优（Nginx 处理静态文件）
- ✅ 支持 HTTPS（配合 Let's Encrypt）
- ✅ 可以添加负载均衡、限流等

---

#### 方案 2: Node.js BFF 层（备选）

如果 server_web 需要有自己的后端（Node.js/Express），可以让它代理 API 请求。

**Express 代理示例**:

```javascript
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();

// 提供静态文件
app.use(express.static('dist'));

// 代理 API 请求
app.use('/api', createProxyMiddleware({
  target: 'http://localhost:2316',
  changeOrigin: true,
}));

app.listen(3000);
```

**优势**:
- ✅ 可以在 BFF 层做额外的业务逻辑
- ✅ 可以聚合多个 API
- ✅ 可以做服务端渲染（SSR）

**劣势**:
- ❌ 增加了一层网络跳转
- ❌ 性能略低于 Nginx 方案

---

#### 方案 3: 直接暴露 API + CORS（不推荐）

如果必须让前端直接访问 API，才需要配置 CORS。

**Gin CORS 中间件**:

```go
import "github.com/gin-contrib/cors"

func main() {
    r := gin.Default()
    
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:5173", "https://your-domain.com"}
    config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "user_token", "user_id"}
    config.AllowCredentials = true
    
    r.Use(cors.New(config))
    
    // ... routes
}
```

**为什么不推荐**:
- ❌ 需要处理 OPTIONS 预检请求
- ❌ 增加服务器负担
- ❌ 安全性较低（API 直接暴露）
- ❌ 配置复杂

---

## 总结

| 方案 | 复杂度 | 性能 | 安全性 | 推荐场景 |
|------|--------|------|--------|----------|
| **Nginx 反向代理** | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 生产环境首选 |
| **Node.js BFF** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 需要 SSR 或复杂逻辑 |
| **CORS** | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | 快速原型开发 |

**当前项目已配置**:
- ✅ 开发环境：Vite 代理（无需 CORS）
- ✅ 生产环境：推荐使用 Nginx 反向代理

**不需要在 server_api 中添加 CORS 中间件！**
