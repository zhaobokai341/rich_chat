DROP DATABASE IF EXISTS rich_chat;
CREATE DATABASE rich_chat;

\c rich_chat;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    bio TEXT NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    registered_time TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP NOT NULL,
    lock_until TIMESTAMP -- NULL means unlocked, timestamp means locked until this time
);

CREATE TABLE IF NOT EXISTS ip_block_list (
    ip VARCHAR(50) NOT NULL,
    block_until TIMESTAMP NOT NULL,
    reason TEXT,
    created_time TIMESTAMP DEFAULT NOW()
);