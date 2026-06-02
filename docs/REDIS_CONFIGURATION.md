# Redis 配置

以下是我配置redis的步骤，以Linux系统为例

1. 安装redis
```bash
sudo apt-get update && sudo apt-get install redis-server
```

2. 设置redis为开机自启动并启动服务
```bash
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

3.修改配置文件
```bash
sudo nano /etc/redis/redis.conf
```
我新增以下内容
```
port 5350
requirepass 123456
maxmemory 256mb
maxmemory-policy allkeys-lru
```

4. 重启服务
```bash
sudo systemctl restart redis-server
```
