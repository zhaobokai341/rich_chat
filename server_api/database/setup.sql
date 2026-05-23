DROP DATABASE IF EXISTS rich_chat;
CREATE DATABASE rich_chat;

\c rich_chat;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL
);
