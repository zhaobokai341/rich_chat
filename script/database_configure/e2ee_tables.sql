-- Additional tables required for E2EE (End-to-End Encryption)
-- Execute this AFTER running chat_schema.sql

-- ============================================
-- User Encryption Keys Table for E2EE / 用户加密密钥表(端到端加密)
-- ============================================
CREATE TABLE IF NOT EXISTS user_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    public_key TEXT NOT NULL,           -- RSA/Ed25519 public key in PEM format / RSA/Ed25519公钥PEM格式
    encrypted_private_key BYTEA, -- Encrypted private key with user's password-derived key (nullable for client-only storage) / 用用户密码派生密钥加密的私钥（客户端本地存储时可为空）
    key_algorithm VARCHAR(20) DEFAULT 'RSA-2048', -- Algorithm used (RSA-2048, Ed25519) / 使用的算法(RSA-2048,Ed25519)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

COMMENT ON TABLE user_keys IS 'Stores user encryption keys for E2EE';
COMMENT ON COLUMN user_keys.public_key IS 'Public key for E2EE - can be shared publicly';
COMMENT ON COLUMN user_keys.encrypted_private_key IS 'Private key encrypted with user password - NEVER store plaintext';

-- ============================================
-- Offline Encrypted Messages for E2EE / 端到端加密离线消息表
-- ============================================
CREATE TABLE IF NOT EXISTS offline_encrypted_messages (
    id SERIAL PRIMARY KEY,
    recipient_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_id INTEGER REFERENCES chat_sessions(id) ON DELETE CASCADE, -- Nullable: allows E2EE without formal session
    encrypted_session_key BYTEA NOT NULL,
    encrypted_content BYTEA NOT NULL,
    iv BYTEA NOT NULL,
    auth_tag BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    delivered_at TIMESTAMPTZ,
    is_delivered BOOLEAN DEFAULT FALSE
);

COMMENT ON TABLE offline_encrypted_messages IS 'Stores encrypted messages for offline users';
COMMENT ON COLUMN offline_encrypted_messages.encrypted_session_key IS 'Session key encrypted with recipient''s public key';
COMMENT ON COLUMN offline_encrypted_messages.encrypted_content IS 'Message content encrypted with session key - never store plaintext';
COMMENT ON COLUMN offline_encrypted_messages.session_id IS 'Optional chat session ID - can be NULL for direct E2EE messages without formal session';

-- ============================================
-- User Online Status Table / 用户在线状态表
-- ============================================
CREATE TABLE IF NOT EXISTS user_online_status (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_online BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    connection_count INTEGER DEFAULT 0, -- Number of active WebSocket connections / 活跃WebSocket连接数
    device_info JSONB DEFAULT '{}'::jsonb -- Browser/device info / 浏览器/设备信息
);

COMMENT ON TABLE user_online_status IS 'Backup for redis-based presence system';
COMMENT ON COLUMN user_online_status.connection_count IS 'Supports multiple devices per user';

-- ============================================
-- Performance Indexes / 性能索引
-- ============================================

-- User keys indexes
CREATE INDEX IF NOT EXISTS idx_user_keys_user_id ON user_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_user_keys_active ON user_keys(is_active) WHERE is_active = TRUE;

-- Offline messages indexes
CREATE INDEX IF NOT EXISTS idx_offline_messages_recipient ON offline_encrypted_messages(recipient_id, is_delivered);
CREATE INDEX IF NOT EXISTS idx_offline_messages_session ON offline_encrypted_messages(session_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_offline_messages_sender ON offline_encrypted_messages(sender_id, created_at DESC);

-- Online status index
CREATE INDEX IF NOT EXISTS idx_user_online_status_online ON user_online_status(is_online) WHERE is_online = TRUE;

-- ============================================
-- Trigger for user_keys updated_at
-- ============================================
CREATE TRIGGER update_user_keys_updated_at 
    BEFORE UPDATE ON user_keys 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();