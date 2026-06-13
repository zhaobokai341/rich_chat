-- Chat database schema
-- Table for chat sessions (direct and group chats)
CREATE TABLE IF NOT EXISTS chat_sessions (
    id SERIAL PRIMARY KEY,
    session_type VARCHAR(20) NOT NULL CHECK (session_type IN ('direct', 'group')), -- direct or group chat
    name VARCHAR(100), -- for group chats only, can be null for direct chats
    created_by INTEGER REFERENCES users(id), -- user who created the session (for groups)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- Junction table for users in chat sessions (many-to-many relationship)
-- For direct chats, there will be exactly 2 users
-- For group chats, there can be multiple users
CREATE TABLE IF NOT EXISTS chat_session_participants (
    session_id INTEGER REFERENCES chat_sessions(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    left_at TIMESTAMPTZ, -- NULL means still participating
    PRIMARY KEY (session_id, user_id)
);

-- Index table for message metadata (unencrypted, for fast querying)
CREATE TABLE IF NOT EXISTS message_index (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    sender_id INTEGER NOT NULL REFERENCES users(id),
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file', 'audio', 'video', 'system')),
    is_read BOOLEAN DEFAULT FALSE, -- for read receipts
    is_deleted BOOLEAN DEFAULT FALSE, -- soft delete flag
    reply_to_message_id INTEGER REFERENCES message_index(id) -- for reply threads
);

-- Additional indexes for message_index table
CREATE INDEX IF NOT EXISTS idx_message_index_session_sent_at_col ON message_index(session_id, sent_at);
CREATE INDEX IF NOT EXISTS idx_message_index_sender_sent_at_col ON message_index(sender_id, sent_at); 
CREATE INDEX IF NOT EXISTS idx_message_index_sent_at_desc_col ON message_index(sent_at DESC);

-- Table for encrypted message content (separate from index for security)
-- This table contains the actual encrypted message content
CREATE TABLE IF NOT EXISTS encrypted_messages (
    message_id INTEGER PRIMARY KEY REFERENCES message_index(id) ON DELETE CASCADE,
    encrypted_content TEXT NOT NULL, -- AES-256-GCM encrypted content
    encrypted_session_key BYTEA, -- Session key encrypted with recipient's public key (for offline retrieval)
    encryption_key_id VARCHAR(100), -- reference to which key was used for encryption
    iv BYTEA NOT NULL, -- initialization vector for AES decryption
    auth_tag BYTEA NOT NULL, -- authentication tag for AES-GCM
    content_type VARCHAR(50) DEFAULT 'text/plain', -- MIME type of original content
    file_size INTEGER, -- for file attachments
    checksum VARCHAR(64), -- SHA-256 checksum of original content before encryption
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table for message read receipts (who has read which messages)
CREATE TABLE IF NOT EXISTS message_read_receipts (
    id SERIAL PRIMARY KEY,
    message_id INTEGER NOT NULL REFERENCES message_index(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(message_id, user_id) -- each user can read a message only once
);

-- Table for group chat metadata (extends chat_sessions for group-specific data)
CREATE TABLE IF NOT EXISTS group_chats (
    session_id INTEGER PRIMARY KEY REFERENCES chat_sessions(id) ON DELETE CASCADE,
    description TEXT,
    avatar_url VARCHAR(255),
    max_members INTEGER DEFAULT 100,
    privacy_level VARCHAR(20) DEFAULT 'private' CHECK (privacy_level IN ('public', 'private', 'invite_only'))
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_message_index_session_sent_at ON message_index(session_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_message_index_sender_sent_at ON message_index(sender_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_message_index_sent_at_desc ON message_index(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_message_read_receipts_user_msg ON message_read_receipts(user_id, message_id);
CREATE INDEX IF NOT EXISTS idx_chat_session_participants_user ON chat_session_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_session_participants_session_left ON chat_session_participants(session_id, left_at);

-- Function to update the updated_at timestamp for chat sessions
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at for chat_sessions
CREATE TRIGGER update_chat_sessions_updated_at 
    BEFORE UPDATE ON chat_sessions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();