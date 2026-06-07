DROP DATABASE IF EXISTS rich_chat;
CREATE DATABASE rich_chat;

\c rich_chat;
-- Set default timezone to UTC for consistent timestamp handling
SET TIME ZONE 'UTC';