# sql配置

本项目用的是postgresql数据库，以下是我配置时的步骤
以Linux系统为例

1. 下载postgresql数据库
```bash
sudo apt-get update && sudo apt-get install postgresql
```

2. 设置密码
```bash
sudo -i -u postgres // 切换用户
psql // 进入数据库命令行
```
随后输入`\password postgres`输入两次密码，随后输入`\q`再输入`exit`退出。

3. 设置开机自启动并启动服务
```bash
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

4.修改配置文件
进入`/etc/postgresql/18/main/pg_hba.conf`文件，找到以下内容
```
local   all             postgres                                peer
```
修改为
```
local   all             postgres                                md5
```
保存退出，随后执行
```bash
sudo systemctl restart postgresql
```

5. 创建数据库和数据表
直接输入
```bash
psql -U postgres -f server_api/database/setup.sql 
```
即可自动配置
