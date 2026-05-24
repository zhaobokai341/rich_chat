DROP DATABASE IF EXISTS rich_chat;
CREATE DATABASE rich_chat;

\c rich_chat;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    bio TEXT,
    password_hash VARCHAR(255) NOT NULL,
    registered_time TIMESTAMP,
    last_login TIMESTAMP NOT NULL,
    lock_until TIMESTAMP -- NULL means unlocked, timestamp means locked until this time
);
