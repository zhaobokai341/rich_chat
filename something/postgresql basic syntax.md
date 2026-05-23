PostgreSQL 的语法非常丰富且高度遵循 SQL 标准。为了方便你快速上手，我将常用语法按功能模块进行了分类整理，并附带了简要说明。

### 1. 数据库操作

```sql
-- 创建数据库
CREATE DATABASE mydb;

-- 切换当前数据库
\c mydb;

-- 修改数据库（例如重命名）
ALTER DATABASE mydb RENAME TO newname;

-- 删除数据库（谨慎使用！）
DROP DATABASE mydb;

-- 查看所有数据库
\l
```

### 2. 表的操作

```sql
-- 创建表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,               -- 自增主键
    username VARCHAR(50) NOT NULL,       -- 不为空
    email VARCHAR(100) UNIQUE,           -- 唯一约束
    created_at TIMESTAMP DEFAULT NOW()   -- 默认当前时间
);

-- 查看表结构
\d users

-- 修改表：添加列
ALTER TABLE users ADD COLUMN age INTEGER;

-- 修改表：删除列
ALTER TABLE users DROP COLUMN age;

-- 修改表：重命名列
ALTER TABLE users RENAME COLUMN username TO name;

-- 删除表（包含数据）
DROP TABLE users;

-- 删除表（仅如果存在）
DROP TABLE IF EXISTS users;
```

### 3. 数据的增删改查 (CRUD)

#### 插入数据
```sql
-- 插入完整行
INSERT INTO users (name, email) VALUES ('ZhangSan', 'zhang@example.com');

-- 批量插入
INSERT INTO users (name, email) VALUES 
    ('LiSi', 'li@example.com'),
    ('WangWu', 'wang@example.com');

-- 插入并返回插入的数据（非常有用的特性）
INSERT INTO users (name, email) VALUES ('ZhaoLiu', 'zhao@example.com') RETURNING *;
```

#### 查询数据
```sql
-- 查询所有
SELECT * FROM users;

-- 条件查询
SELECT * FROM users WHERE name = 'ZhangSan';

-- 分页查询 (LIMIT 和 OFFSET)
SELECT * FROM users ORDER BY id LIMIT 10 OFFSET 20;

-- 排序 (ASC 升序, DESC 降序)
SELECT * FROM users ORDER BY created_at DESC;
```

#### 更新数据
```sql
-- 更新特定行
UPDATE users SET email = 'new_email@example.com' WHERE id = 1;

-- 更新并返回结果
UPDATE users SET name = 'NewName' WHERE id = 1 RETURNING *;
```

#### 删除数据
```sql
-- 删除特定行
DELETE FROM users WHERE id = 1;

-- 清空表（保留表结构，比 DELETE 快，但不可回滚）
TRUNCATE TABLE users;
```

### 4. 模式

PostgreSQL 使用 Schema（模式）来对表进行逻辑分组。默认的 Schema 叫做 `public`。

```sql
-- 创建模式
CREATE SCHEMA my_schema;

-- 在指定模式下创建表
CREATE TABLE my_schema.orders (...);

-- 删除模式
DROP SCHEMA my_schema;

-- 查看当前搜索路径
SHOW search_path;

-- 修改搜索路径（让查询时不用加 schema 前缀）
SET search_path TO my_schema, public;
```

### 5. 索引

索引可以加快查询速度，但会减慢写入速度。

```sql
-- 创建普通 B-Tree 索引
CREATE INDEX idx_users_email ON users(email);

-- 创建唯一索引
CREATE UNIQUE INDEX idx_unique_email ON users(email);

-- 删除索引
DROP INDEX idx_users_email;

-- 查看表上的索引
\d users
```

### 6. 视图

视图是虚拟表，基于 SQL 查询结果。

```sql
-- 创建视图
CREATE VIEW user_summary AS
SELECT id, name, email FROM users;

-- 查询视图
SELECT * FROM user_summary;

-- 删除视图
DROP VIEW user_summary;
```

### 7. 常用函数与聚合

```sql
-- 统计行数
SELECT COUNT(*) FROM users;

-- 求和、平均值、最大值、最小值
SELECT SUM(price), AVG(age), MAX(score), MIN(score) FROM orders;

-- 字符串拼接
SELECT name || ' (' || email || ')' AS user_info FROM users;

-- 数据类型转换
SELECT CAST('123' AS INTEGER);
-- 或者简写
SELECT '123'::INTEGER;
```

### 8. JSON 操作 (PostgreSQL 特色)

PostgreSQL 对 JSON 支持极好，推荐使用 `JSONB` 类型（二进制存储，处理更快）。

```sql
-- 创建包含 JSONB 的表
CREATE TABLE logs (
    id SERIAL PRIMARY KEY,
    data JSONB
);

-- 插入 JSON 数据
INSERT INTO logs (data) VALUES ('{"user": "Alice", "action": "login"}');

-- 查询：JSON 中的某个键
SELECT data->>'user' AS username FROM logs;

-- 查询：过滤 JSON 数据（查找 action 为 login 的记录）
SELECT * FROM logs WHERE data->>'action' = 'login';

-- 更新：修改 JSON 中的某个字段
UPDATE logs SET data = jsonb_set(data, '{action}', '"logout"') WHERE id = 1;
```

### 9. 常用元数据命令

这些命令通常在 `psql` 命令行工具中使用：

```sql
\?           -- 查看所有 psql 命令
\h           -- 查看 SQL 语法帮助（例如 \h SELECT）
\l           -- 列出所有数据库
\dt          -- 列出当前数据库的所有表
\du          -- 列出所有用户（角色）
\conninfo    -- 查看当前连接信息
\q           -- 退出 psql
```

这些语法覆盖了 PostgreSQL 日常开发 90% 以上的场景。如果你对某个特定的语法（如窗口函数、存储过程）感兴趣，可以继续提问。