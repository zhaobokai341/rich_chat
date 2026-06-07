#!/bin/bash

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# 切换到脚本所在目录，或者直接使用绝对路径引用文件
# 这里我们使用绝对路径拼接 SQL 文件路径
SQL_DIR="$SCRIPT_DIR/database_configure"

echo "正在执行数据库配置..."

psql -U postgres -f "$SQL_DIR/create_database.sql"
psql -U postgres -d "rich_chat" -f "$SQL_DIR/users_schema.sql"
psql -U postgres -d "rich_chat" -f "$SQL_DIR/ip_blocker_schema.sql"
psql -U postgres -d "rich_chat" -f "$SQL_DIR/chat_schema.sql"

echo "数据库配置完成。"
