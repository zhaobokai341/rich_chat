# WebSocket Real-time Chat and Encrypted Message Storage Implementation Guide
# WebSocket实时聊天与加密消息存储实现指南

**Document Version / 文档版本:** 1.0  
**Last Updated / 最后更新:** 2026-06-02  
**Project / 项目:** Rich Chat  
**Status / 状态:** Design Document - Ready for Implementation / 设计文档 - 准备实施

---

## Table of Contents / 目录

1. [Overview / 概述](#1-overview--概述)
2. [Architecture Design / 架构设计](#2-architecture-design--架构设计)
3. [Database Schema / 数据库模式](#3-database-schema--数据库模式)
4. [Encryption Strategy / 加密策略](#4-encryption-strategy--加密策略)
5. [Implementation Structure / 实现结构](#5-implementation-structure--实现结构)
6. [WebSocket Message Protocol / WebSocket消息协议](#6-websocket-message-protocol--websocket消息协议)
7. [RESTful API Endpoints / RESTful API端点](#7-restful-api-endpoints--restful-api端点)
8. [Security Considerations / 安全考虑](#8-security-considerations--安全考虑)
9. [Configuration / 配置](#9-configuration--配置)
10. [Implementation Roadmap / 实施路线图](#10-implementation-roadmap--实施路线图)
11. [Monitoring & Operations / 监控与运维](#11-monitoring--operations--监控与运维)
12. [Testing Strategy / 测试策略](#12-testing-strategy--测试策略)
13. [Troubleshooting Guide / 故障排除指南](#13-troubleshooting-guide--故障排除指南)
14. [Performance Optimization / 性能优化](#14-performance-optimization-tips--性能优化建议)
15. [Summary / 总结](#15-summary--总结)

---

## 1. Overview / 概述

### Current State / 当前状态

✅ **Completed / 已完成:**
- User authentication with JWT / 用户JWT认证
- RESTful API for user management / 用户管理RESTful API
- PostgreSQL + Redis dual-layer architecture / PostgreSQL+Redis双层架构
- Multi-language support (Chinese/English) / 多语言支持(中文/英文)
- Security features (rate limiting, account lockout) / 安全功能(速率限制、账户锁定)

❌ **Not Implemented / 未实现:**
- WebSocket real-time messaging / WebSocket实时消息
- Chat message storage / 聊天消息存储
- Message encryption / 消息加密
- Online status tracking / 在线状态跟踪
- Group chat support / 群聊支持

### Goals / 目标

This document provides a comprehensive implementation guide for adding:

本文档提供以下功能的完整实现指南：

1. **Real-time messaging via WebSocket** - Instant message delivery between users / **通过WebSocket实现实时消息** - 用户间即时消息传递
2. **Encrypted message storage** - AES-256-GCM encryption for database persistence / **加密消息存储** - 使用AES-256-GCM加密进行数据库持久化
3. **Session management** - Direct and group chat sessions / **会话管理** - 私聊和群聊会话
4. **Online presence** - Real-time online/offline status / **在线状态** - 实时在线/离线状态
5. **Message history** - Paginated retrieval of encrypted messages / **消息历史** - 加密消息的分页检索

---

## 2. Architecture Design / 架构设计

### 2.1 Technology Stack / 技术栈

**Backend / 后端:**
- **Language:** Go 1.26.2
- **Web Framework:** Gin (`github.com/gin-gonic/gin v1.12.0`)
- **WebSocket Library:** Gorilla WebSocket (`github.com/gorilla/websocket`) - *NEW*
- **Encryption:** AES-256-GCM (`crypto/aes` standard library) - *NEW*
- **Database:** PostgreSQL via `sqlx` + `lib/pq`
- **Cache:** Redis via `go-redis/redis/v8`
- **Logging:** Logrus (`github.com/sirupsen/logrus`)

**Frontend Web / 网页前端:**
- **Framework:** React + TypeScript + Vite
- **WebSocket:** Native WebSocket API - *NEW*

**CLI Client / CLI客户端:**
- **TUI Framework:** Bubble Tea + Lip Gloss
- **WebSocket Client:** Gorilla WebSocket - *NEW*

### 2.2 Communication Architecture / 通信架构

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│  Web Client │         │   CLI Client │         │ Mobile App  │
│  (React)    │         │   (Go TUI)   │         │ (Future)    │
└──────┬──────┘         └──────┬───────┘         └──────┬──────┘
       │ WebSocket             │ WebSocket              │ WebSocket
       │ wss://                │ wss://                 │ wss://
       └───────────────────────┼────────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   Go Server (Gin)   │
                    │   Port: 2316        │
                    │                     │
                    │  HTTP Routes:       │
                    │  /api/auth/*        │
                    │  /api/users/*       │
                    │  /api/chat/*        │
                    │                     │
                    │  WebSocket:         │
                    │  /ws?token={jwt}    │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │   Message Hub       │
                    │  - Connection Mgmt  │
                    │  - Broadcast        │
                    │  - Rate Limiting    │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
     ┌────────▼──────┐ ┌──────▼──────┐ ┌──────▼────────┐
     │  PostgreSQL   │ │    Redis    │ │ Encryption    │
     │               │ │             │ │ Service       │
     │ - Users       │ │ - Sessions  │ │               │
     │ - Messages    │ │ - Presence  │ │ - AES-256-GCM │
     │ - Sessions    │ │ - Rate      │ │ - Key Mgmt    │
     │ - Keys        │ │   Limiting  │ │               │
     └───────────────┘ └─────────────┘ └───────────────┘
```

### 2.3 Unified Port Strategy / 统一端口策略

**Single Port Approach / 单端口方案:**
- HTTP and WebSocket share port **2316** / HTTP和WebSocket共用**2316**端口
- HTTP endpoints use `/api/*` prefix / HTTP端点使用`/api/*`前缀
- WebSocket endpoint uses `/ws` path / WebSocket端点使用`/ws`路径
- Simplifies deployment and firewall configuration / 简化部署和防火墙配置

**Authentication Flow / 认证流程:**

1. Client authenticates via `/api/auth/login` → receives JWT token / 客户端通过登录接口认证→获取JWT令牌
2. Client connects to WebSocket: `wss://server:2316/ws?token={jwt_token}` / 客户端连接WebSocket
3. Server validates JWT token before upgrading connection / 服务器在升级连接前验证JWT令牌
4. Invalid/expired tokens result in **401 Unauthorized** / 无效/过期令牌返回**401未授权**
5. Valid connections are registered in Message Hub / 有效连接在消息中心注册

---

## 3. Database Schema / 数据库模式

### 3.1 New Tables Required / 需要新增的数据表

Execute this SQL **after** the existing `script/setup.sql` / 在现有`script/setup.sql`**之后**执行此SQL:

```sql
-- ============================================
-- Chat Sessions Table / 聊天会话表
-- ============================================
CREATE TABLE IF NOT EXISTS chat_sessions (
    id SERIAL PRIMARY KEY,
    session_uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    session_type VARCHAR(20) DEFAULT 'direct', -- 'direct' (1-on-1) or 'group' / '私聊'或'群聊'
    session_name VARCHAR(100), -- Optional name for group chats / 群聊可选名称
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE chat_sessions IS 'Stores chat session metadata';
COMMENT ON COLUMN chat_sessions.session_type IS 'Type of chat: direct (1-on-1) or group';

-- ============================================
-- Session Participants Table / 会话参与者表
-- ============================================
CREATE TABLE IF NOT EXISTS chat_session_participants (
    id SERIAL PRIMARY KEY,
    session_id INTEGER REFERENCES chat_sessions(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    last_read_message_id INTEGER, -- Last message ID read by this user / 该用户最后阅读的消息ID
    role VARCHAR(20) DEFAULT 'member', -- 'admin' or 'member' for group chats / 群聊中的'管理员'或'成员'
    UNIQUE(session_id, user_id)
);

COMMENT ON TABLE chat_session_participants IS 'Maps users to chat sessions they participate in';
COMMENT ON COLUMN chat_session_participants.last_read_message_id IS 'Tracks read status per user';

-- ============================================
-- Encrypted Messages Table / 加密消息表
-- ============================================
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    session_id INTEGER REFERENCES chat_sessions(id) ON DELETE CASCADE,
    sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    
    -- Encryption fields / 加密字段
    encrypted_content BYTEA NOT NULL, -- AES-256-GCM encrypted message content / AES-256-GCM加密的消息内容
    iv BYTEA NOT NULL, -- Initialization Vector (12 bytes for GCM) / 初始化向量(GCM为12字节)
    encryption_algorithm VARCHAR(20) DEFAULT 'AES-256-GCM',
    
    -- Message metadata / 消息元数据
    message_type VARCHAR(20) DEFAULT 'text', -- 'text', 'image', 'file', 'system' / '文本','图片','文件','系统'
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    delivered BOOLEAN DEFAULT FALSE, -- True if delivered to all online participants / 是否已交付给所有在线参与者
    read_by JSONB DEFAULT '[]'::jsonb, -- Array of user IDs who have read this message / 已读此消息的用户ID数组
    
    -- Optional fields for future extensions / 未来扩展的可选字段
    parent_message_id INTEGER REFERENCES chat_messages(id), -- For threaded replies / 用于线程回复
    edited_at TIMESTAMPTZ, -- Timestamp if message was edited / 如果消息被编辑的时间戳
    is_deleted BOOLEAN DEFAULT FALSE -- Soft delete flag / 软删除标志
);

COMMENT ON TABLE chat_messages IS 'Stores encrypted chat messages';
COMMENT ON COLUMN chat_messages.encrypted_content IS 'Encrypted message body - never store plaintext';
COMMENT ON COLUMN chat_messages.iv IS 'Initialization vector for AES-GCM decryption';
COMMENT ON COLUMN chat_messages.read_by IS 'JSON array of user IDs: [1, 2, 3]';

-- ============================================
-- Message Index Table / 消息索引表
-- ============================================
CREATE TABLE IF NOT EXISTS chat_message_index (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES chat_messages(id) ON DELETE CASCADE,
    session_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL,
    message_type VARCHAR(20) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    
    -- Denormalized for fast queries (no sensitive data) / 反规范化以加快查询速度(无敏感数据)
    INDEX created_at_idx ON chat_message_index(sent_at DESC),
    INDEX session_time_idx ON chat_message_index(session_id, sent_at DESC)
);

COMMENT ON TABLE chat_message_index IS 'Fast query index without encrypted content';
COMMENT ON COLUMN chat_message_index.is_deleted IS 'Allows filtering deleted messages in queries';

-- ============================================
-- User Encryption Keys Table / 用户加密密钥表
-- ============================================
CREATE TABLE IF NOT EXISTS user_encryption_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- For Phase 2: End-to-End Encryption / 第二阶段：端到端加密
    public_key TEXT NOT NULL, -- RSA/Ed25519 public key (PEM format) / RSA/Ed25519公钥(PEM格式)
    private_key_encrypted BYTEA NOT NULL, -- Private key encrypted with user password / 用用户密码加密的私钥
    key_version INTEGER DEFAULT 1, -- Supports key rotation / 支持密钥轮换
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE user_encryption_keys IS 'Stores user encryption keys for E2EE (Phase 2)';
COMMENT ON COLUMN user_encryption_keys.private_key_encrypted IS 'NEVER store plaintext private keys';

-- ============================================
-- Online Status Backup Table / 在线状态备份表
-- ============================================
CREATE TABLE IF NOT EXISTS user_online_status (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_online BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    connection_count INTEGER DEFAULT 0, -- Number of active WebSocket connections / 活跃WebSocket连接数
    device_info JSONB DEFAULT '{}'::jsonb -- Browser/device info / 浏览器/设备信息
);

COMMENT ON TABLE user_online_status IS 'Backup for Redis-based presence system';
COMMENT ON COLUMN user_online_status.connection_count IS 'Supports multiple devices per user';
```

### 3.2 Performance Indexes / 性能索引

```sql
-- ============================================
-- Indexes for Query Optimization / 查询优化索引
-- ============================================

-- Messages table indexes / 消息表索引
CREATE INDEX idx_chat_messages_session_id ON chat_messages(session_id);
CREATE INDEX idx_chat_messages_sender_id ON chat_messages(sender_id);
CREATE INDEX idx_chat_messages_sent_at ON chat_messages(sent_at DESC);
CREATE INDEX idx_chat_messages_delivered ON chat_messages(delivered) WHERE delivered = FALSE;
CREATE INDEX idx_chat_messages_parent ON chat_messages(parent_message_id) WHERE parent_message_id IS NOT NULL;

-- Session participants indexes / 会话参与者索引
CREATE INDEX idx_chat_session_participants_user_id ON chat_session_participants(user_id);
CREATE INDEX idx_chat_session_participants_session ON chat_session_participants(session_id);

-- Message index table indexes / 消息索引表索引
CREATE INDEX idx_chat_message_index_session_time ON chat_message_index(session_id, sent_at DESC);
CREATE INDEX idx_chat_message_index_sender ON chat_message_index(sender_id, sent_at DESC);
CREATE INDEX idx_chat_message_index_deleted ON chat_message_index(is_deleted) WHERE is_deleted = TRUE;

-- Online status index / 在线状态索引
CREATE INDEX idx_user_online_status_online ON user_online_status(is_online) WHERE is_online = TRUE;

-- Add partial index for unread messages / 为未读消息添加部分索引
CREATE INDEX idx_chat_messages_unread ON chat_messages(session_id, sent_at DESC) 
WHERE delivered = TRUE AND NOT (read_by @> '[current_user_id]'::jsonb);
```

### 3.3 Data Retention Policy / 数据保留策略

```sql
-- ============================================
-- Automated Cleanup Functions / 自动清理函数
-- ============================================

-- Function to archive old messages / 归档旧消息的函数
CREATE OR REPLACE FUNCTION archive_old_messages(days_to_keep INTEGER DEFAULT 365)
RETURNS VOID AS $$
BEGIN
    -- Move messages older than specified days to archive table / 将超过指定天数的消息移至归档表
    INSERT INTO chat_messages_archive
    SELECT * FROM chat_messages
    WHERE sent_at < NOW() - (days_to_keep || ' days')::INTERVAL;
    
    -- Delete archived messages from main table / 从主表中删除已归档消息
    DELETE FROM chat_messages
    WHERE sent_at < NOW() - (days_to_keep || ' days')::INTERVAL;
    
    RAISE NOTICE 'Archived messages older than % days', days_to_keep;
END;
$$ LANGUAGE plpgsql;

-- Schedule daily cleanup (requires pg_cron extension) / 安排每日清理(需要pg_cron扩展)
-- SELECT cron.schedule('daily-message-archive', '0 2 * * *', 'SELECT archive_old_messages(365)');
```

---

## 4. Encryption Strategy / 加密策略

### 4.1 Two-Phase Approach / 两阶段方法

#### Phase 1: Transport-Layer Encryption (Recommended for Initial Release) / 第一阶段：传输层加密(初始版本推荐)

**How It Works / 工作原理:**
1. Server holds a master encryption key (from environment variable) / 服务器持有主加密密钥(来自环境变量)
2. Messages are encrypted immediately upon arrival at server / 消息到达服务器后立即加密
3. Encrypted messages stored in PostgreSQL / 加密消息存储在PostgreSQL中
4. Messages decrypted only when sending to authorized users / 仅在发送给授权用户时解密
5. **Server can see plaintext** during processing / **服务器在处理过程中可以看到明文**

**Pros / 优点:**
- ✅ Simple implementation / 实现简单
- ✅ Fast performance / 性能快速
- ✅ Easy key management / 密钥管理简单
- ✅ Good security for most use cases / 大多数用例的安全性良好

**Cons / 缺点:**
- ❌ Server has access to plaintext / 服务器可访问明文
- ❌ Not suitable for highly sensitive communications / 不适合高度敏感的通信

**Security Level / 安全级别:** ⭐⭐⭐ (Good for general chat apps) / ⭐⭐⭐(适用于一般聊天应用)

---

#### Phase 2: End-to-End Encryption (E2EE) - Future Enhancement / 第二阶段：端到端加密(E2EE) - 未来增强

**How It Works / 工作原理:**
1. Each user generates RSA-2048 or Ed25519 key pair on registration / 每个用户注册时生成RSA-2048或Ed25519密钥对
2. Public keys stored in database, freely accessible / 公钥存储在数据库中，可公开访问
3. Private keys encrypted with user's password-derived key / 私钥使用用户密码派生的密钥加密
4. **Client-side encryption:** / **客户端加密:**
   - Sender generates random symmetric key (session key) / 发送方生成随机对称密钥(会话密钥)
   - Encrypt message with session key (AES-256-GCM) / 用会话密钥加密消息(AES-256-GCM)
   - Encrypt session key with recipient's public key / 用接收方公钥加密会话密钥
   - Send: `[encrypted_session_key + encrypted_message + IV]` / 发送
5. **Client-side decryption:** / **客户端解密:**
   - Recipient decrypts session key with their private key / 接收方用自己的私钥解密会话密钥
   - Decrypt message with session key / 用会话密钥解密消息
6. **Server never sees plaintext** / **服务器永远看不到明文**

**Pros / 优点:**
- ✅ Maximum privacy / 最大隐私保护
- ✅ Server cannot read messages / 服务器无法读取消息
- ✅ Resistant to server compromise / 抵抗服务器入侵

**Cons / 缺点:**
- ❌ Complex implementation / 实现复杂
- ❌ Slower performance / 性能较慢
- ❌ Key management challenges / 密钥管理挑战
- ❌ No server-side search / 无法服务器端搜索

**Security Level / 安全级别:** ⭐⭐⭐⭐⭐ (Maximum privacy, like Signal/WhatsApp) / ⭐⭐⭐⭐⭐(最大隐私，如Signal/WhatsApp)

---

### 4.2 Recommended Encryption Algorithm: AES-256-GCM / 推荐加密算法：AES-256-GCM

**Why AES-256-GCM? / 为什么选择AES-256-GCM?**

| Feature / 特性 | AES-256-GCM | AES-256-CBC | ChaCha20-Poly1305 |
|----------------|-------------|-------------|-------------------|
| Authenticated Encryption / 认证加密 | ✅ Yes / 是 | ❌ No / 否 | ✅ Yes / 是 |
| Hardware Acceleration / 硬件加速 | ✅ Excellent / 优秀 | ✅ Good / 良好 | ❌ Limited / 有限 |
| Go Standard Library / Go标准库 | ✅ Built-in / 内置 | ✅ Built-in / 内置 | ✅ Built-in / 内置 |
| Performance / 性能 | ⚡⚡⚡ Fast / 快 | ⚡⚡ Medium / 中 | ⚡⚡ Medium / 中 |
| Industry Adoption / 行业采用 | 🌟 Widespread / 广泛 | 🌟 Legacy / 传统 | 🌟 Growing / 增长中 |
| IV Size / IV大小 | 12 bytes / 字节 | 16 bytes / 字节 | 12 bytes / 字节 |

**Key Advantages / 关键优势:**
1. **Authenticated Encryption / 认证加密:** Provides both confidentiality AND integrity / 同时提供机密性和完整性
2. **Tamper Detection / 篡改检测:** Automatically detects if ciphertext was modified / 自动检测密文是否被修改
3. **No Padding Required / 无需填充:** More efficient than CBC mode / 比CBC模式更高效
4. **Parallel Processing / 并行处理:** Can encrypt/decrypt in parallel / 可以并行加密/解密
5. **Go Support / Go支持:** Native support in `crypto/aes` package / `crypto/aes`包原生支持

---

### 4.3 Key Management / 密钥管理

#### Master Encryption Key Configuration / 主加密密钥配置

Add to your `.env` file / 添加到`.env`文件:

```bash
# ============================================
# Encryption Configuration / 加密配置
# ============================================

# Master encryption key (minimum 32 bytes for AES-256) / 主加密密钥(AES-256至少32字节)
# Generate with: openssl rand -base64 32 / 生成命令
MASTER_ENCRYPTION_KEY=your-random-32-byte-key-here-change-in-production

# Encryption algorithm (currently only AES-256-GCM supported) / 加密算法(目前仅支持AES-256-GCM)
ENCRYPTION_ALGORITHM=AES-256-GCM

# Key rotation interval in days (0 = disabled) / 密钥轮换间隔天数(0=禁用)
ENCRYPTION_KEY_ROTATION_DAYS=90
```

**Generate a Secure Key / 生成安全密钥:**

```bash
# Option 1: Using OpenSSL / 选项1：使用OpenSSL
openssl rand -base64 32

# Option 2: Using Go / 选项2：使用Go
go run -e 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { k := make([]byte, 32); rand.Read(k); fmt.Println(base64.StdEncoding.EncodeToString(k)) }'

# Option 3: Using Python / 选项3：使用Python
python3 -c "import secrets, base64; print(base64.b64encode(secrets.token_bytes(32)).decode())"
```

**Key Security Best Practices / 密钥安全最佳实践:**

1. ✅ **Never hardcode keys** / 绝不硬编码密钥
2. ✅ **Use environment variables** / 使用环境变量
3. ✅ **Restrict .env file permissions:** `chmod 600 .env` / 限制.env文件权限
4. ✅ **Never commit .env to Git** / 绝不将.env提交到Git
5. ✅ **Rotate keys periodically** (every 90 days) / 定期轮换密钥(每90天)
6. ✅ **Use different keys per environment** (dev/staging/prod) / 每个环境使用不同密钥
7. ❌ **Don't share keys via email/chat** / 不要通过邮件/聊天分享密钥
8. ❌ **Don't log keys accidentally** / 不要意外记录密钥

---

### 4.4 Go Encryption Implementation Example / Go加密实现示例

Create file: `server_api/service/encryption_service.go` / 创建文件：

```go
package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// EncryptionService handles message encryption and decryption / 加密服务处理消息加密和解密
type EncryptionService interface {
	// EncryptMessage encrypts plaintext using AES-256-GCM / 使用AES-256-GCM加密明文
	EncryptMessage(plaintext []byte) (encrypted []byte, iv []byte, err error)
	
	// DecryptMessage decrypts ciphertext using AES-256-GCM / 使用AES-256-GCM解密密文
	DecryptMessage(encrypted []byte, iv []byte) ([]byte, error)
}

// aesGCMEncryptionService implements EncryptionService using AES-256-GCM / 使用AES-256-GCM实现EncryptionService
type aesGCMEncryptionService struct {
	masterKey []byte // 32 bytes for AES-256 / AES-256为32字节
}

// NewEncryptionService creates a new encryption service from master key string / 从主密钥字符串创建新的加密服务
func NewEncryptionService(masterKey string) (EncryptionService, error) {
	if len(masterKey) < 32 {
		return nil, errors.New("master key must be at least 32 bytes for AES-256")
	}
	
	// Hash the key to ensure exactly 32 bytes / 哈希密钥以确保正好32字节
	hash := sha256.Sum256([]byte(masterKey))
	
	return &aesGCMEncryptionService{
		masterKey: hash[:],
	}, nil
}

// EncryptMessage encrypts plaintext message / 加密明文消息
func (e *aesGCMEncryptionService) EncryptMessage(plaintext []byte) ([]byte, []byte, error) {
	// Create AES cipher block / 创建AES密码块
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return nil, nil, errors.New("failed to create cipher: " + err.Error())
	}
	
	// Create GCM mode / 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, errors.New("failed to create GCM: " + err.Error())
	}
	
	// Generate random IV (12 bytes for GCM) / 生成随机IV(GCM为12字节)
	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, errors.New("failed to generate IV: " + err.Error())
	}
	
	// Encrypt and authenticate / 加密和认证
	// Seal appends ciphertext and tag to dst / Seal将密文和标签附加到dst
	encrypted := gcm.Seal(nil, iv, plaintext, nil)
	
	return encrypted, iv, nil
}

// DecryptMessage decrypts ciphertext message / 解密密文消息
func (e *aesGCMEncryptionService) DecryptMessage(encrypted []byte, iv []byte) ([]byte, error) {
	// Create AES cipher block / 创建AES密码块
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return nil, errors.New("failed to create cipher: " + err.Error())
	}
	
	// Create GCM mode / 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("failed to create GCM: " + err.Error())
	}
	
	// Verify IV size / 验证IV大小
	if len(iv) != gcm.NonceSize() {
		return nil, errors.New("invalid IV size")
	}
	
	// Decrypt and verify authenticity / 解密并验证真实性
	// Open verifies the tag and decrypts / Open验证标签并解密
	plaintext, err := gcm.Open(nil, iv, encrypted, nil)
	if err != nil {
		// This could mean wrong key, tampered data, or corrupted IV / 这可能意味着密钥错误、数据篡改或IV损坏
		return nil, errors.New("decryption failed: invalid key or tampered data")
	}
	
	return plaintext, nil
}
```

**Unit Tests / 单元测试:**

Create file: `server_api/service/encryption_service_test.go` / 创建文件：

```go
package service

import (
	"testing"
)

func TestEncryptionDecryption(t *testing.T) {
	// Create encryption service / 创建加密服务
	service, err := NewEncryptionService("test-master-key-must-be-at-least-32-bytes-long!")
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}
	
	// Test message / 测试消息
	originalMessage := []byte("Hello, this is a secret message!")
	
	// Encrypt / 加密
	encrypted, iv, err := service.EncryptMessage(originalMessage)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	
	// Verify encrypted data is different from original / 验证加密数据与原始不同
	if string(encrypted) == string(originalMessage) {
		t.Error("Encrypted data should be different from original")
	}
	
	// Decrypt / 解密
	decrypted, err := service.DecryptMessage(encrypted, iv)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}
	
	// Verify decrypted message matches original / 验证解密消息与原始匹配
	if string(decrypted) != string(originalMessage) {
		t.Errorf("Decrypted message doesn't match original.\nExpected: %s\nGot: %s", 
			string(originalMessage), string(decrypted))
	}
}

func TestDecryptionWithWrongKey(t *testing.T) {
	// Create two services with different keys / 创建两个具有不同密钥的服务
	service1, _ := NewEncryptionService("key-one-must-be-at-least-32-bytes-long!!")
	service2, _ := NewEncryptionService("key-two-must-be-at-least-32-bytes-long!!")
	
	message := []byte("Secret message")
	
	// Encrypt with service1 / 用service1加密
	encrypted, iv, _ := service1.EncryptMessage(message)
	
	// Try to decrypt with service2 (wrong key) / 尝试用service2解密(错误的密钥)
	_, err := service2.DecryptMessage(encrypted, iv)
	if err == nil {
		t.Error("Decryption should fail with wrong key")
	}
}

func TestDecryptionWithTamperedData(t *testing.T) {
	service, _ := NewEncryptionService("test-master-key-must-be-at-least-32-bytes-long!")
	
	message := []byte("Secret message")
	encrypted, iv, _ := service.EncryptMessage(message)
	
	// Tamper with encrypted data / 篡改加密数据
	encrypted[0] ^= 0xFF
	
	// Try to decrypt tampered data / 尝试解密篡改的数据
	_, err := service.DecryptMessage(encrypted, iv)
	if err == nil {
		t.Error("Decryption should fail with tampered data")
	}
}
```

Run tests / 运行测试:
```bash
cd server_api/service
go test -v -run TestEncryption
```

---

## 5. Implementation Structure / 实现结构

### 5.1 Directory Layout / 目录布局

```
rich_chat/
├── server_api/
│   ├── websocket/                    # NEW: WebSocket module / 新增：WebSocket模块
│   │   ├── connection.go            # WebSocket connection management / WebSocket连接管理
│   │   ├── hub.go                   # Message hub (pub/sub pattern) / 消息中心(发布/订阅模式)
│   │   ├── handler.go               # WebSocket route handler / WebSocket路由处理器
│   │   ├── message.go               # Message type definitions / 消息类型定义
│   │   └── middleware.go            # Authentication & rate limiting / 认证和速率限制
│   │
│   ├── service/
│   │   ├── chat_service.go          # NEW: Chat business logic / 新增：聊天业务逻辑
│   │   ├── encryption_service.go    # NEW: AES-256-GCM encryption / 新增：AES-256-GCM加密
│   │   ├── session_service.go       # NEW: Session management / 新增：会话管理
│   │   └── ...existing files...
│   │
│   ├── database/
│   │   ├── message_repository.go    # NEW: Message CRUD operations / 新增：消息增删改查
│   │   ├── session_repository.go    # NEW: Session CRUD operations / 新增：会话增删改查
│   │   ├── encryption_repository.go # NEW: Key management / 新增：密钥管理
│   │   └── ...existing files...
│   │
│   └── main.go                      # MODIFY: Add WebSocket routes / 修改：添加WebSocket路由
│
├── client/
│   ├── websocket_client.go          # NEW: WebSocket client implementation / 新增：WebSocket客户端实现
│   ├── chat_handler.go              # NEW: Chat UI handler (Bubble Tea) / 新增：聊天UI处理器
│   └── ...existing files...
│
└── server_web/src/
    ├── hooks/
    │   └── useWebSocket.ts          # NEW: React WebSocket hook / 新增：React WebSocket钩子
    ├── components/
    │   ├── ChatWindow.tsx           # NEW: Main chat interface / 新增：主聊天界面
    │   ├── MessageList.tsx          # NEW: Message display component / 新增：消息显示组件
    │   └── MessageInput.tsx         # NEW: Message input component / 新增：消息输入组件
    └── ...existing files...
```

### 5.2 Core Components Detailed / 核心组件详解

#### Component 1: WebSocket Connection / 组件1：WebSocket连接

**File:** `server_api/websocket/connection.go`

**Purpose / 目的:** Represents a single WebSocket connection to a client / 表示到客户端的单个WebSocket连接

**Key Features / 关键特性:**
- Thread-safe write operations / 线程安全的写操作
- Activity tracking for idle timeout / 活动跟踪以实现空闲超时
- Buffered send channel to prevent blocking / 缓冲发送通道以防止阻塞

```go
package websocket

import (
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
)

// Connection represents a single WebSocket connection / 表示单个WebSocket连接
type Connection struct {
	UserID     int               // Associated user ID / 关联的用户ID
	Conn       *websocket.Conn   // underlying WebSocket connection / 底层WebSocket连接
	SendChan   chan []byte       // Buffered channel for outgoing messages / 出站消息的缓冲通道
	LastActive time.Time         // Last activity timestamp / 最后活动时间戳
	mu         sync.RWMutex      // Mutex for thread safety / 线程安全的互斥锁
}

// NewConnection creates a new WebSocket connection / 创建新的WebSocket连接
func NewConnection(userID int, conn *websocket.Conn) *Connection {
	return &Connection{
		UserID:     userID,
		Conn:       conn,
		SendChan:   make(chan []byte, 256), // Buffer 256 messages / 缓冲256条消息
		LastActive: time.Now(),
	}
}

// Write sends a message to the client with timeout / 带超时向客户端发送消息
func (c *Connection) Write(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Set write deadline / 设置写入截止时间
	c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	
	// Write message / 写入消息
	return c.Conn.WriteMessage(messageType, data)
}

// UpdateActivity updates the last active timestamp / 更新最后活动时间戳
func (c *Connection) UpdateActivity() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastActive = time.Now()
}

// IsActive checks if connection is still active (within 5 minutes) / 检查连接是否仍然活跃(5分钟内)
func (c *Connection) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.LastActive) < 5*time.Minute
}

// Close gracefully closes the connection / 优雅地关闭连接
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	close(c.SendChan)
	c.Conn.Close()
}
```

---

#### Component 2: Message Hub / 组件2：消息中心

**File:** `server_api/websocket/hub.go`

**Purpose / 目的:** Central message broker that manages all connections and broadcasts messages / 管理所有连接并广播消息的中央消息代理

**Key Features / 关键特性:**
- Connection registry by user ID / 按用户ID注册连接
- Session-based message broadcasting / 基于会话的消息广播
- Automatic cleanup of disconnected clients / 自动清理断开的客户端
- Thread-safe operations with RWMutex / 使用RWMutex的线程安全操作

```go
package websocket

import (
	"encoding/json"
	"sync"
	
	log "github.com/sirupsen/logrus"
)

// Hub maintains active connections and broadcasts messages / 维护活动连接并广播消息
type Hub struct {
	connections map[int]*Connection  // UserID -> Connection mapping / 用户ID到连接的映射
	
	register   chan *Connection      // Register new connections / 注册新连接
	unregister chan *Connection      // Unregister disconnected clients / 注销断开的客户端
	broadcast  chan BroadcastMessage // Broadcast messages to sessions / 向会话广播消息
	
	mu sync.RWMutex // Read-write mutex for thread safety / 线程安全的读写互斥锁
}

// BroadcastMessage represents a message to broadcast to a session / 表示要广播到会话的消息
type BroadcastMessage struct {
	SessionID       int    // Target session ID / 目标会话ID
	Message         []byte // Serialized WSMessage / 序列化的WSMessage
	ExcludeUserID   int    // Optional: exclude sender from broadcast / 可选：从广播中排除发送者
}

// NewHub creates a new message hub / 创建新的消息中心
func NewHub() *Hub {
	return &Hub{
		connections: make(map[int]*Connection),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		broadcast:   make(chan BroadcastMessage, 256), // Buffer broadcasts / 缓冲广播
	}
}

// Run starts the hub's main loop (call in goroutine) / 启动中心的主循环(在goroutine中调用)
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.handleRegister(conn)
			
		case conn := <-h.unregister:
			h.handleUnregister(conn)
			
		case broadcast := <-h.broadcast:
			h.handleBroadcast(broadcast)
		}
	}
}

// handleRegister registers a new connection / 注册新连接
func (h *Hub) handleRegister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// If user already has a connection, close the old one / 如果用户已有连接，关闭旧连接
	if existing, exists := h.connections[conn.UserID]; exists {
		existing.Close()
		log.WithFields(log.Fields{
			"user_id": conn.UserID,
		}).Warn("Closed existing WebSocket connection for user")
	}
	
	// Register new connection / 注册新连接
	h.connections[conn.UserID] = conn
	
	log.WithFields(log.Fields{
		"user_id": conn.UserID,
		"total_connections": len(h.connections),
	}).Info("WebSocket connection registered")
}

// handleUnregister removes a connection / 移除连接
func (h *Hub) handleUnregister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if existing, exists := h.connections[conn.UserID]; exists {
		if existing == conn { // Only remove if it's the same connection / 仅在是同一连接时移除
			delete(h.connections, conn.UserID)
			conn.Close()
			
			log.WithFields(log.Fields{
				"user_id": conn.UserID,
				"total_connections": len(h.connections),
			}).Info("WebSocket connection unregistered")
		}
	}
}

// handleBroadcast sends message to all participants in a session / 向会话中的所有参与者发送消息
func (h *Hub) handleBroadcast(broadcast BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// TODO: Query database for session participants / 查询数据库获取会话参与者
	// For now, broadcast to all connected users / 目前广播给所有连接的用户
	
	sentCount := 0
	for userID, conn := range h.connections {
		// Skip excluded user (usually the sender) / 跳过排除的用户(通常是发送者)
		if broadcast.ExcludeUserID > 0 && userID == broadcast.ExcludeUserID {
			continue
		}
		
		// TODO: Check if user is participant in session / 检查用户是否是会话参与者
		
		// Send message (non-blocking) / 发送消息(非阻塞)
		select {
		case conn.SendChan <- broadcast.Message:
			sentCount++
		default:
			// Channel full, disconnect user / 通道已满，断开用户
			log.WithFields(log.Fields{
				"user_id": userID,
			}).Warn("User message channel full, disconnecting")
			go h.unregister <- conn
		}
	}
	
	log.WithFields(log.Fields{
		"session_id": broadcast.SessionID,
		"sent_to": sentCount,
	}).Debug("Message broadcast completed")
}

// SendToUser sends message to specific user / 向特定用户发送消息
func (h *Hub) SendToUser(userID int, message []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if conn, ok := h.connections[userID]; ok {
		select {
		case conn.SendChan <- message:
			return true
		default:
			return false // Channel full / 通道已满
		}
	}
	return false // User not connected / 用户未连接
}

// GetUserCount returns number of active connections / 返回活动连接数
func (h *Hub) GetUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// IsUserOnline checks if user has active WebSocket connection / 检查用户是否有活跃的WebSocket连接
func (h *Hub) IsUserOnline(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.connections[userID]
	return exists
}
```

---

#### Component 3: WebSocket Handler / 组件3：WebSocket处理器

**File:** `server_api/websocket/handler.go`

**Purpose / 目的:** Handles WebSocket upgrade, authentication, and message routing / 处理WebSocket升级、认证和消息路由

```go
package websocket

import (
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	EnableCompression: true, // Enable Per-message Deflate / 启用逐消息压缩
	
	// IMPORTANT: Restrict origins in production / 重要：在生产环境中限制来源
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking / 实现适当的来源检查
		// For development, allow all origins / 开发期间允许所有来源
		return true
		
		// Production example / 生产示例:
		// origin := r.Header.Get("Origin")
		// return origin == "https://yourdomain.com"
	},
}

// HandleWebSocket upgrades HTTP connection to WebSocket / 将HTTP连接升级为WebSocket
func HandleWebSocket(c *gin.Context, hub *Hub, jwtService JWTService) {
	// Get JWT token from query parameter / 从查询参数获取JWT令牌
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication token"})
		return
	}
	
	// Validate JWT token and extract user ID / 验证JWT令牌并提取用户ID
	userID, err := jwtService.ValidateToken(token)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"ip": c.ClientIP(),
		}).Warn("Invalid JWT token for WebSocket connection")
		
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}
	
	// Upgrade HTTP connection to WebSocket / 将HTTP连接升级为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to upgrade to WebSocket")
		return
	}
	
	// Create new connection / 创建新连接
	wsConn := NewConnection(userID, conn)
	
	// Register connection with hub / 向中心注册连接
	hub.register <- wsConn
	
	// Start reader and writer goroutines / 启动读写goroutine
	go wsConn.writePump()
	go wsConn.readPump(hub, jwtService)
	
	log.WithFields(log.Fields{
		"user_id": userID,
		"ip": c.ClientIP(),
	}).Info("WebSocket connection established")
}

// writePump pumps messages from hub to WebSocket connection / 将消息从中心泵送到WebSocket连接
func (c *Connection) writePump() {
	defer func() {
		c.Conn.Close()
		log.WithFields(log.Fields{
			"user_id": c.UserID,
		}).Info("WebSocket write pump stopped")
	}()
	
	for message := range c.SendChan {
		c.UpdateActivity()
		
		// Set write deadline / 设置写入截止时间
		c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		
		// Write message / 写入消息
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.WithFields(log.Fields{
				"user_id": c.UserID,
				"error": err.Error(),
			}).Error("WebSocket write error")
			break
		}
	}
}

// readPump pumps messages from WebSocket to hub / 将消息从WebSocket泵送到中心
func (c *Connection) readPump(hub *Hub, jwtService JWTService) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
		log.WithFields(log.Fields{
			"user_id": c.UserID,
		}).Info("WebSocket read pump stopped")
	}()
	
	// Set read deadline / 设置读取截止时间
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	
	// Configure pong handler for heartbeat / 配置心跳的pong处理程序
	c.Conn.SetPongHandler(func(appData string) error {
		c.UpdateActivity()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, 
				websocket.CloseGoingAway, 
				websocket.CloseNormalClosure) {
				log.WithFields(log.Fields{
					"user_id": c.UserID,
					"error": err.Error(),
				}).Error("WebSocket unexpected close")
			}
			break
		}
		
		c.UpdateActivity()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		
		// Parse message / 解析消息
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.WithFields(log.Fields{
				"user_id": c.UserID,
				"error": err.Error(),
			}).Warn("Failed to parse WebSocket message")
			continue
		}
		
		// Handle different message types / 处理不同类型的消息
		switch wsMsg.Type {
		case MessageTypeText:
			handleTextMessage(hub, c.UserID, wsMsg)
			
		case MessageTypeHeartbeat:
			// Respond with pong / 响应pong
			pong := WSResponse{Success: true, Message: "pong"}
			pongData, _ := json.Marshal(pong)
			c.SendChan <- pongData
			
		case MessageTypeReadReceipt:
			handleReadReceipt(hub, c.UserID, wsMsg)
			
		default:
			log.WithFields(log.Fields{
				"user_id": c.UserID,
				"type": wsMsg.Type,
			}).Warn("Unknown message type")
		}
	}
}

// handleTextMessage processes text messages / 处理文本消息
func handleTextMessage(hub *Hub, senderID int, msg WSMessage) {
	// TODO: Implement message validation, encryption, storage, and broadcast / 实现消息验证、加密、存储和广播
	log.WithFields(log.Fields{
		"sender_id": senderID,
		"session_id": msg.SessionID,
		"content_length": len(msg.Content),
	}).Info("Text message received")
}

// handleReadReceipt processes read receipts / 处理已读回执
func handleReadReceipt(hub *Hub, userID int, msg WSMessage) {
	// TODO: Update message read status in database / 更新数据库中的消息已读状态
	log.WithFields(log.Fields{
		"user_id": userID,
		"message_id": msg.MessageID,
	}).Info("Read receipt received")
}
```

---

*(Due to length constraints, I'll continue with the remaining sections in the next part)*

**Continue reading Sections 6-15 in the complete document...**

---

## Document Status / 文档状态

✅ **Sections Completed / 已完成章节:**
1. Overview / 概述
2. Architecture Design / 架构设计
3. Database Schema / 数据库模式
4. Encryption Strategy / 加密策略
5. Implementation Structure (partial) / 实现结构(部分)

📝 **Sections Remaining / 剩余章节:**
6. WebSocket Message Protocol / WebSocket消息协议
7. RESTful API Endpoints / RESTful API端点
8. Security Considerations / 安全考虑
9. Configuration / 配置
10. Implementation Roadmap / 实施路线图
11. Monitoring & Operations / 监控与运维
12. Testing Strategy / 测试策略
13. Troubleshooting Guide / 故障排除指南
14. Performance Optimization / 性能优化
15. Summary / 总结

---

**Note / 注意:** This is a comprehensive design document. The complete implementation would require approximately 10 weeks of development following the roadmap in Section 10.

**注意:** 这是一份综合设计文档。按照第10节的路线图，完整实现大约需要10周的开发时间。

For the complete implementation code examples and remaining sections, please refer to the memory knowledge base or request specific sections to be expanded.

有关完整的实现代码示例和剩余章节，请参考记忆知识库或请求展开特定章节。
