# WebSocket Chat Implementation Checklist
# WebSocket聊天实现检查清单

**Project / 项目:** Rich Chat  
**Feature / 功能:** Real-time Chat with Encrypted Messages / 实时聊天与加密消息  
**Status / 状态:** Planning Phase / 规划阶段  

---

## Pre-Implementation Checklist / 实施前检查清单

### Environment Setup / 环境设置

- [ ] Install Gorilla WebSocket dependency / 安装Gorilla WebSocket依赖
  ```bash
  cd server_api
  go get github.com/gorilla/websocket
  ```

- [ ] Generate master encryption key / 生成主加密密钥
  ```bash
  openssl rand -base64 32
  ```

- [ ] Update `.env` file with encryption key / 用加密密钥更新`.env`文件
  ```bash
  MASTER_ENCRYPTION_KEY=your-generated-key-here
  ```

- [ ] Create feature branch / 创建功能分支
  ```bash
  git checkout -b feature/websocket-chat
  ```

- [ ] Review database access credentials / 审查数据库访问凭证
  - [ ] PostgreSQL connection working / PostgreSQL连接正常
  - [ ] Redis connection working / Redis连接正常

---

## Phase 1: Foundation (Week 1-2) / 第一阶段：基础(第1-2周)

### Database Schema / 数据库模式

- [ ] Execute SQL schema from implementation guide / 执行实现指南中的SQL模式
  - [ ] `chat_sessions` table created / 创建chat_sessions表
  - [ ] `chat_session_participants` table created / 创建chat_session_participants表
  - [ ] `chat_messages` table created / 创建chat_messages表
  - [ ] `chat_message_index` table created / 创建chat_message_index表
  - [ ] `user_encryption_keys` table created / 创建user_encryption_keys表
  - [ ] `user_online_status` table created / 创建user_online_status表

- [ ] Create performance indexes / 创建性能索引
  - [ ] All indexes from Section 3.2 created / 创建第3.2节中的所有索引

- [ ] Verify tables in database / 验证数据库中的表
  ```sql
  \dt chat_*
  \di chat_*
  ```

### WebSocket Infrastructure / WebSocket基础设施

- [ ] Create `server_api/websocket/` directory / 创建websocket目录

- [ ] Implement `connection.go` / 实现connection.go
  - [ ] Connection struct defined / 定义Connection结构体
  - [ ] Write() method implemented / 实现Write()方法
  - [ ] Activity tracking implemented / 实现活动跟踪
  - [ ] Thread-safe with mutex / 使用互斥锁保证线程安全

- [ ] Implement `hub.go` / 实现hub.go
  - [ ] Hub struct defined / 定义Hub结构体
  - [ ] Register/unregister logic / 注册/注销逻辑
  - [ ] Broadcast mechanism / 广播机制
  - [ ] Run() goroutine started / 启动Run() goroutine

- [ ] Implement `handler.go` / 实现handler.go
  - [ ] WebSocket upgrader configured / 配置WebSocket升级器
  - [ ] JWT validation integrated / 集成JWT验证
  - [ ] readPump/writePump goroutines / 读写泵goroutine
  - [ ] Heartbeat handling / 心跳处理

- [ ] Update `main.go` / 更新main.go
  - [ ] Import websocket package / 导入websocket包
  - [ ] Initialize Hub in initialize() / 在initialize()中初始化Hub
  - [ ] Add `/ws` route / 添加/ws路由
  - [ ] Start Hub.Run() goroutine / 启动Hub.Run() goroutine

### Testing Phase 1 / 第一阶段测试

- [ ] Test WebSocket connection / 测试WebSocket连接
  ```javascript
  // Browser console test / 浏览器控制台测试
  const ws = new WebSocket('ws://localhost:2316/ws?token=YOUR_JWT_TOKEN');
  ws.onopen = () => console.log('Connected!');
  ws.onclose = () => console.log('Disconnected');
  ```

- [ ] Verify authentication / 验证认证
  - [ ] Valid token connects successfully / 有效令牌成功连接
  - [ ] Invalid token rejected (401) / 无效令牌被拒绝(401)
  - [ ] Missing token rejected (401) / 缺少令牌被拒绝(401)

- [ ] Test heartbeat mechanism / 测试心跳机制
  - [ ] Ping/pong working / Ping/pong正常工作
  - [ ] Idle timeout functioning / 空闲超时功能正常

- [ ] Check connection cleanup / 检查连接清理
  - [ ] Disconnected users removed from Hub / 断开连接的用户从Hub中移除
  - [ ] No memory leaks after multiple connect/disconnect / 多次连接/断开后无内存泄漏

**Phase 1 Completion Criteria / 第一阶段完成标准:**
✅ Users can establish authenticated WebSocket connections / 用户可以建立认证的WebSocket连接  
✅ Connections are properly tracked and cleaned up / 连接被正确跟踪和清理  
✅ Basic heartbeat keeps connections alive / 基本心跳保持连接活跃  

---

## Phase 2: Core Messaging (Week 3-4) / 第二阶段：核心消息(第3-4周)

### Encryption Service / 加密服务

- [ ] Create `service/encryption_service.go` / 创建encryption_service.go
  - [ ] EncryptionService interface defined / 定义EncryptionService接口
  - [ ] AES-256-GCM implementation / AES-256-GCM实现
  - [ ] EncryptMessage() method / EncryptMessage()方法
  - [ ] DecryptMessage() method / DecryptMessage()方法

- [ ] Write unit tests / 编写单元测试
  - [ ] Test encrypt/decrypt roundtrip / 测试加密/解密往返
  - [ ] Test wrong key failure / 测试错误密钥失败
  - [ ] Test tampered data detection / 测试篡改数据检测
  ```bash
  cd server_api/service
  go test -v -run TestEncryption
  ```

### Message Repository / 消息仓库

- [ ] Create `database/message_repository.go` / 创建message_repository.go
  - [ ] MessageRepository interface / MessageRepository接口
  - [ ] CreateMessage() method / CreateMessage()方法
  - [ ] GetMessagesBySession() with pagination / 带分页的GetMessagesBySession()
  - [ ] MarkMessageAsRead() method / MarkMessageAsRead()方法
  - [ ] UpdateDeliveryStatus() method / UpdateDeliveryStatus()方法

- [ ] Integrate with DatabaseService / 与DatabaseService集成
  - [ ] Add messageRepo to DatabaseService / 将messageRepo添加到DatabaseService
  - [ ] Update InitializeDatabaseService() / 更新InitializeDatabaseService()

### Chat Service / 聊天服务

- [ ] Create `service/chat_service.go` / 创建chat_service.go
  - [ ] ChatService interface / ChatService接口
  - [ ] SendMessage() implementation / SendMessage()实现
    - [ ] Encrypt message / 加密消息
    - [ ] Save to database / 保存到数据库
    - [ ] Broadcast via Hub / 通过Hub广播
  - [ ] GetMessageHistory() implementation / GetMessageHistory()实现
    - [ ] Query database / 查询数据库
    - [ ] Decrypt messages / 解密消息
    - [ ] Return paginated results / 返回分页结果

- [ ] Integrate with WebSocket handler / 与WebSocket处理器集成
  - [ ] Call SendMessage() on text message receive / 收到文本消息时调用SendMessage()
  - [ ] Handle broadcast responses / 处理广播响应

### RESTful API Endpoints / RESTful API端点

- [ ] Add routes in `main.go` / 在main.go中添加路由
  - [ ] `POST /api/chat/sessions` - Create session / 创建会话
  - [ ] `GET /api/chat/sessions/:user_id` - Get user sessions / 获取用户会话
  - [ ] `GET /api/chat/sessions/:session_id/messages` - Get message history / 获取消息历史

- [ ] Implement handlers in `web_server_page.go` / 在web_server_page.go中实现处理器
  - [ ] CreateChatSession handler / CreateChatSession处理器
  - [ ] GetUserSessions handler / GetUserSessions处理器
  - [ ] GetSessionMessages handler / GetSessionMessages处理器

### Testing Phase 2 / 第二阶段测试

- [ ] Test message encryption / 测试消息加密
  - [ ] Verify encrypted content in database / 验证数据库中的加密内容
  - [ ] Confirm plaintext not stored / 确认未存储明文

- [ ] Test message delivery / 测试消息传递
  - [ ] Sender sees message sent / 发送者看到消息已发送
  - [ ] Recipient receives message in real-time / 接收者实时接收消息
  - [ ] Message content matches original / 消息内容与原始匹配

- [ ] Test message history / 测试消息历史
  - [ ] Retrieve last 50 messages / 检索最近50条消息
  - [ ] Pagination works correctly / 分页正常工作
  - [ ] Decrypted messages readable / 解密消息可读

- [ ] Load test basic messaging / 基本消息负载测试
  - [ ] Send 100 messages/second / 每秒发送100条消息
  - [ ] Monitor latency (< 100ms P95) / 监控延迟(P95 < 100ms)
  - [ ] Check for message loss / 检查消息丢失

**Phase 2 Completion Criteria / 第二阶段完成标准:**
✅ Messages are encrypted before storage / 消息在存储前加密  
✅ Real-time message delivery working / 实时消息传递正常工作  
✅ Message history retrieval functional / 消息历史检索功能正常  
✅ One-on-one chat fully operational / 一对一聊天完全正常运行  

---

## Phase 3: Advanced Features (Week 5-7) / 第三阶段：高级功能(第5-7周)

### Group Chat Support / 群聊支持

- [ ] Update session creation API / 更新会话创建API
  - [ ] Support multiple participants / 支持多个参与者
  - [ ] Set session_type to 'group' / 设置session_type为'group'
  - [ ] Assign admin role / 分配管理员角色

- [ ] Implement participant management / 实现参与者管理
  - [ ] Add participant to session / 添加参与者到会话
  - [ ] Remove participant from session / 从会话中移除参与者
  - [ ] List session participants / 列出会话参与者

- [ ] Update broadcast logic / 更新广播逻辑
  - [ ] Query session participants from database / 从数据库查询会话参与者
  - [ ] Broadcast only to participants / 仅广播给参与者
  - [ ] Handle participant join/leave events / 处理参与者加入/离开事件

### Read Receipts / 已读回执

- [ ] Implement read receipt tracking / 实现已读回执跟踪
  - [ ] Client sends read receipt via WebSocket / 客户端通过WebSocket发送已读回执
  - [ ] Server updates read_by field in database / 服务器更新数据库中的read_by字段
  - [ ] Broadcast read status to other participants / 向其他参与者广播已读状态

- [ ] Display read status in UI / 在UI中显示已读状态
  - [ ] Show "Delivered" status / 显示"已交付"状态
  - [ ] Show "Read by X users" / 显示"X人已读"
  - [ ] Update in real-time / 实时更新

### Online Status System / 在线状态系统

- [ ] Implement presence tracking with Redis / 用Redis实现在线状态跟踪
  - [ ] Set online status on WebSocket connect / WebSocket连接时设置在线状态
  - [ ] Clear online status on disconnect / 断开连接时清除在线状态
  - [ ] Track last_seen timestamp / 跟踪last_seen时间戳

- [ ] Create online status API / 创建在线状态API
  - [ ] `GET /api/users/:user_id/status` - Get user status / 获取用户状态
  - [ ] `GET /api/chat/sessions/:session_id/online` - Get online participants / 获取在线参与者

- [ ] Display online status in UI / 在UI中显示在线状态
  - [ ] Green dot for online / 在线显示绿点
  - [ ] Gray dot for offline / 离线显示灰点
  - [ ] Show "Last seen X minutes ago" / 显示"X分钟前最后在线"

### Offline Message Handling / 离线消息处理

- [ ] Queue messages for offline users / 为离线用户排队消息
  - [ ] Detect offline recipients / 检测离线接收者
  - [ ] Store undelivered flag / 存储未交付标志
  - [ ] Deliver on reconnection / 重新连接时交付

- [ ] Implement message expiration / 实现消息过期
  - [ ] Set TTL for queued messages / 设置排队消息的TTL
  - [ ] Clean up expired messages / 清理过期消息

### Testing Phase 3 / 第三阶段测试

- [ ] Test group chat / 测试群聊
  - [ ] Create group with 5 users / 创建5人群组
  - [ ] Send message, all receive it / 发送消息，所有人都收到
  - [ ] Remove user, they stop receiving / 移除用户，他们停止接收

- [ ] Test read receipts / 测试已读回执
  - [ ] Mark message as read / 标记消息为已读
  - [ ] Verify read_by updated in database / 验证数据库中read_by已更新
  - [ ] Other users see read status / 其他用户看到已读状态

- [ ] Test online status / 测试在线状态
  - [ ] Connect, status shows online / 连接，状态显示在线
  - [ ] Disconnect, status shows offline / 断开，状态显示离线
  - [ ] Multiple devices, correct count / 多设备，计数正确

- [ ] Test offline messages / 测试离线消息
  - [ ] Send to offline user / 发送给离线用户
  - [ ] User reconnects, receives messages / 用户重新连接，接收消息
  - [ ] Expired messages cleaned up / 过期消息已清理

**Phase 3 Completion Criteria / 第三阶段完成标准:**
✅ Group chat with multiple participants working / 多参与者群聊正常工作  
✅ Read receipts accurately tracked / 已读回执准确跟踪  
✅ Online/offline status real-time accurate / 在线/离线状态实时准确  
✅ Offline messages delivered on reconnection / 离线消息在重新连接时交付  

---

## Phase 4: Frontend Integration (Week 8-9) / 第四阶段：前端集成(第8-9周)

### Web Frontend (React) / Web前端(React)

- [ ] Create `useWebSocket.ts` hook / 创建useWebSocket.ts钩子
  - [ ] WebSocket connection management / WebSocket连接管理
  - [ ] Auto-reconnect with exponential backoff / 指数退避自动重连
  - [ ] Message queue for offline periods / 离线期间消息队列
  - [ ] Event listeners for messages / 消息事件监听器

- [ ] Build chat UI components / 构建聊天UI组件
  - [ ] `ChatWindow.tsx` - Main container / 主容器
  - [ ] `MessageList.tsx` - Scrollable message display / 可滚动消息显示
  - [ ] `MessageInput.tsx` - Text input with send button / 带发送按钮的文本输入
  - [ ] `ParticipantList.tsx` - Online users sidebar / 在线用户侧边栏
  - [ ] `TypingIndicator.tsx` - "User is typing..." / "用户正在输入..."

- [ ] Integrate with existing auth flow / 与现有认证流程集成
  - [ ] Get JWT token after login / 登录后获取JWT令牌
  - [ ] Pass token to WebSocket connection / 将令牌传递给WebSocket连接
  - [ ] Handle token refresh / 处理令牌刷新

- [ ] Style chat interface / 样式化聊天界面
  - [ ] Responsive design (mobile/desktop) / 响应式设计(移动/桌面)
  - [ ] Dark/light theme support / 深色/浅色主题支持
  - [ ] Message bubbles styled / 消息气泡样式
  - [ ] Smooth animations / 流畅动画

### CLI Client (Go + Bubble Tea) / CLI客户端(Go + Bubble Tea)

- [ ] Create `websocket_client.go` / 创建websocket_client.go
  - [ ] WebSocket dial and connect / WebSocket拨号和连接
  - [ ] Read loop for incoming messages / 传入消息的读取循环
  - [ ] Send message function / 发送消息函数
  - [ ] Reconnection logic / 重新连接逻辑

- [ ] Create `chat_handler.go` / 创建chat_handler.go
  - [ ] Bubble Tea model for chat / 聊天的Bubble Tea模型
  - [ ] Message list view / 消息列表视图
  - [ ] Input field for typing / 输入字段
  - [ ] Real-time message updates / 实时消息更新

- [ ] Integrate with existing CLI menu / 与现有CLI菜单集成
  - [ ] Add "Open Chat" option / 添加"打开聊天"选项
  - [ ] Navigate between chat sessions / 在聊天会话间导航
  - [ ] Return to main menu / 返回主菜单

### Cross-Platform Testing / 跨平台测试

- [ ] Test Web ↔ Web communication / 测试Web↔Web通信
  - [ ] Two browser tabs chatting / 两个浏览器标签页聊天
  - [ ] Messages delivered instantly / 消息即时传递
  - [ ] Both see read receipts / 双方看到已读回执

- [ ] Test CLI ↔ CLI communication / 测试CLI↔CLI通信
  - [ ] Two terminal instances chatting / 两个终端实例聊天
  - [ ] Real-time message sync / 实时消息同步
  - [ ] No message loss / 无消息丢失

- [ ] Test Web ↔ CLI communication / 测试Web↔CLI通信
  - [ ] Web user sends to CLI user / Web用户发送给CLI用户
  - [ ] CLI user responds to Web user / CLI用户回复Web用户
  - [ ] Both platforms work seamlessly / 两个平台无缝工作

- [ ] Test mixed scenarios / 测试混合场景
  - [ ] Group chat with Web + CLI users / Web+CLI用户的群聊
  - [ ] Online status visible across platforms / 跨平台可见在线状态
  - [ ] Read receipts work cross-platform / 跨平台已读回执正常工作

**Phase 4 Completion Criteria / 第四阶段完成标准:**
✅ Web chat interface fully functional / Web聊天界面功能完全正常  
✅ CLI chat interface fully functional / CLI聊天界面功能完全正常  
✅ Cross-platform compatibility verified / 跨平台兼容性已验证  
✅ UI/UX polished and responsive / UI/UX精美且响应迅速  

---

## Phase 5: Optimization & Deployment (Week 10) / 第五阶段：优化与部署(第10周)

### Performance Optimization / 性能优化

- [ ] Database optimization / 数据库优化
  - [ ] Analyze query performance with EXPLAIN / 用EXPLAIN分析查询性能
  - [ ] Add missing indexes if needed / 必要时添加缺失索引
  - [ ] Implement message archival for old data / 为旧数据实现消息归档
  - [ ] Configure connection pool settings / 配置连接池设置

- [ ] Redis caching / Redis缓存
  - [ ] Cache recent messages per session / 缓存每会话最近消息
  - [ ] Cache online status / 缓存在线状态
  - [ ] Implement rate limiting with token bucket / 用令牌桶实现速率限制

- [ ] WebSocket optimization / WebSocket优化
  - [ ] Enable Per-message Deflate compression / 启用逐消息压缩
  - [ ] Tune buffer sizes (SendChan, ReadBufferSize) / 调整缓冲区大小
  - [ ] Implement batch message sending / 实现批量消息发送

- [ ] Memory profiling / 内存分析
  - [ ] Use pprof to identify leaks / 使用pprof识别泄漏
  - [ ] Optimize connection cleanup / 优化连接清理
  - [ ] Monitor goroutine count / 监控goroutine数量

### Load Testing / 负载测试

- [ ] Set up load testing environment / 设置负载测试环境
  - [ ] Separate test database / 独立测试数据库
  - [ ] Monitoring tools (Prometheus + Grafana) / 监控工具

- [ ] Concurrent connection test / 并发连接测试
  - [ ] Simulate 10,000 WebSocket connections / 模拟10,000个WebSocket连接
  - [ ] Measure memory and CPU usage / 测量内存和CPU使用
  - [ ] Identify connection limits / 识别连接限制

- [ ] Message throughput test / 消息吞吐量测试
  - [ ] Send 1,000 messages/second / 每秒发送1,000条消息
  - [ ] Measure P95 and P99 latency / 测量P95和P99延迟
  - [ ] Check for message loss / 检查消息丢失

- [ ] Database stress test / 数据库压力测试
  - [ ] Insert 100,000 messages concurrently / 并发插入100,000条消息
  - [ ] Query message history under load / 负载下查询消息历史
  - [ ] Monitor connection pool usage / 监控连接池使用

### Security Audit / 安全审计

- [ ] Code review / 代码审查
  - [ ] Review encryption implementation / 审查加密实现
  - [ ] Check for hardcoded secrets / 检查硬编码的秘密
  - [ ] Verify input validation / 验证输入验证
  - [ ] Review error handling / 审查错误处理

- [ ] Penetration testing / 渗透测试
  - [ ] Test WebSocket injection attacks / 测试WebSocket注入攻击
  - [ ] Attempt unauthorized message access / 尝试未经授权的消息访问
  - [ ] Test rate limiting effectiveness / 测试速率限制有效性
  - [ ] Check for XSS vulnerabilities / 检查XSS漏洞

- [ ] Compliance check / 合规性检查
  - [ ] GDPR compliance (data retention) / GDPR合规性(数据保留)
  - [ ] Encryption standards met / 符合加密标准
  - [ ] Audit logging complete / 审计日志完整

### Production Deployment / 生产部署

- [ ] Configure production environment / 配置生产环境
  - [ ] Set up TLS certificates for WSS / 为WSS设置TLS证书
  - [ ] Configure firewall rules / 配置防火墙规则
  - [ ] Set up reverse proxy (nginx) / 设置反向代理(nginx)
  - [ ] Configure environment variables / 配置环境变量

- [ ] Set up monitoring and alerting / 设置监控和告警
  - [ ] Prometheus metrics collection / Prometheus指标收集
  - [ ] Grafana dashboards / Grafana仪表板
  - [ ] Alert rules configured / 配置告警规则
  - [ ] Log aggregation (ELK or similar) / 日志聚合(ELK或类似)

- [ ] Deployment automation / 部署自动化
  - [ ] Create deployment scripts / 创建部署脚本
  - [ ] Configure CI/CD pipeline / 配置CI/CD流水线
  - [ ] Set up rollback procedure / 设置回滚程序
  - [ ] Document operational procedures / 文档化操作流程

- [ ] Final verification / 最终验证
  - [ ] Smoke tests pass / 冒烟测试通过
  - [ ] All features working in production / 所有功能在生产中正常工作
  - [ ] Performance meets targets / 性能达到目标
  - [ ] Security audit passed / 通过安全审计

**Phase 5 Completion Criteria / 第五阶段完成标准:**
✅ Performance benchmarks met (10k connections, <100ms P95) / 达到性能基准  
✅ Security audit passed with no critical issues / 通过安全审计，无严重问题  
✅ Production deployment successful / 生产部署成功  
✅ Monitoring and alerting operational / 监控和告警正常运行  

---

## Post-Launch Checklist / 上线后检查清单

### Week 1 After Launch / 上线后第1周

- [ ] Monitor error rates / 监控错误率
  - [ ] WebSocket connection failures < 1% / WebSocket连接失败<1%
  - [ ] Message delivery failures < 0.1% / 消息交付失败<0.1%
  - [ ] Decryption failures = 0% / 解密失败=0%

- [ ] Monitor performance metrics / 监控性能指标
  - [ ] Average latency within target / 平均延迟在目标内
  - [ ] Memory usage stable / 内存使用稳定
  - [ ] Database query times acceptable / 数据库查询时间可接受

- [ ] Gather user feedback / 收集用户反馈
  - [ ] Survey users on chat experience / 调查用户聊天体验
  - [ ] Identify pain points / 识别痛点
  - [ ] Prioritize bug fixes / 优先修复错误

- [ ] Address critical bugs / 解决严重错误
  - [ ] Fix any message loss issues / 修复任何消息丢失问题
  - [ ] Resolve connectivity problems / 解决连接问题
  - [ ] Patch security vulnerabilities / 修补安全漏洞

### Month 1 After Launch / 上线后第1个月

- [ ] Analyze usage patterns / 分析使用模式
  - [ ] Peak usage times / 高峰使用时间
  - [ ] Most active users/sessions / 最活跃的用户/会话
  - [ ] Average messages per day / 每天平均消息数

- [ ] Optimize based on data / 基于数据优化
  - [ ] Adjust cache TTLs / 调整缓存TTL
  - [ ] Fine-tune database indexes / 微调数据库索引
  - [ ] Scale resources if needed / 必要时扩展资源

- [ ] Plan Phase 2 features / 规划第二阶段功能
  - [ ] End-to-End Encryption (E2EE) / 端到端加密(E2EE)
  - [ ] File/image sharing / 文件/图片分享
  - [ ] Voice/video calls / 语音/视频通话

- [ ] Document lessons learned / 记录经验教训
  - [ ] What worked well / 哪些工作良好
  - [ ] What could be improved / 哪些可以改进
  - [ ] Update implementation guide / 更新实现指南

---

## Troubleshooting Quick Reference / 故障排除快速参考

### Common Issues / 常见问题

**Issue: WebSocket connection fails / 问题：WebSocket连接失败**
- [ ] Check JWT token validity / 检查JWT令牌有效性
- [ ] Verify token in URL parameter / 验证URL参数中的令牌
- [ ] Check server logs for auth errors / 检查服务器日志中的认证错误
- [ ] Ensure CORS allows origin / 确保CORS允许来源

**Issue: Messages not delivered / 问题：消息未交付**
- [ ] Verify recipient is connected / 验证接收者已连接
- [ ] Check session participation / 检查会话参与
- [ ] Review Hub broadcast logs / 审查Hub广播日志
- [ ] Test with simple echo first / 首先用简单回显测试

**Issue: Decryption fails / 问题：解密失败**
- [ ] Verify MASTER_ENCRYPTION_KEY matches / 验证MASTER_ENCRYPTION_KEY匹配
- [ ] Check IV size (must be 12 bytes) / 检查IV大小(必须12字节)
- [ ] Look for data corruption in database / 查找数据库中的数据损坏
- [ ] Review encryption service logs / 审查加密服务日志

**Issue: High memory usage / 问题：高内存使用**
- [ ] Check for connection leaks / 检查连接泄漏
- [ ] Verify SendChan closed on disconnect / 验证断开连接时SendChan关闭
- [ ] Profile with pprof / 用pprof分析
- [ ] Monitor goroutine count / 监控goroutine数量

**Issue: Slow message delivery / 问题：消息传递缓慢**
- [ ] Check database query performance / 检查数据库查询性能
- [ ] Review Redis response times / 审查Redis响应时间
- [ ] Monitor network latency / 监控网络延迟
- [ ] Profile encryption/decryption time / 分析加密/解密时间

---

## Success Metrics / 成功指标

### Technical Metrics / 技术指标

| Metric / 指标 | Target / 目标 | Current / 当前 | Status / 状态 |
|---------------|---------------|----------------|---------------|
| Concurrent Connections / 并发连接 | 10,000+ | TBD / 待定 | ⬜ |
| Message Latency P95 / 消息延迟P95 | < 100ms | TBD / 待定 | ⬜ |
| Message Latency P99 / 消息延迟P99 | < 500ms | TBD / 待定 | ⬜ |
| Message Throughput / 消息吞吐量 | > 1,000 msg/s | TBD / 待定 | ⬜ |
| Error Rate / 错误率 | < 1% | TBD / 待定 | ⬜ |
| Uptime / 正常运行时间 | > 99.9% | TBD / 待定 | ⬜ |

### User Experience Metrics / 用户体验指标

| Metric / 指标 | Target / 目标 | Current / 当前 | Status / 状态 |
|---------------|---------------|----------------|---------------|
| User Satisfaction / 用户满意度 | > 4.5/5 | TBD / 待定 | ⬜ |
| Message Delivery Success / 消息交付成功 | > 99.9% | TBD / 待定 | ⬜ |
| Average Session Duration / 平均会话时长 | > 15 min | TBD / 待定 | ⬜ |
| Daily Active Users / 日活跃用户 | Growing / 增长 | TBD / 待定 | ⬜ |

---

## Sign-off / 签署

**Phase 1 Approval / 第一阶段批准:**
- Developer / 开发人员: _________________ Date / 日期: _______
- Reviewer / 审查员: _________________ Date / 日期: _______

**Phase 2 Approval / 第二阶段批准:**
- Developer / 开发人员: _________________ Date / 日期: _______
- Reviewer / 审查员: _________________ Date / 日期: _______

**Phase 3 Approval / 第三阶段批准:**
- Developer / 开发人员: _________________ Date / 日期: _______
- Reviewer / 审查员: _________________ Date / 日期: _______

**Phase 4 Approval / 第四阶段批准:**
- Developer / 开发人员: _________________ Date / 日期: _______
- Reviewer / 审查员: _________________ Date / 日期: _______

**Phase 5 Approval / 第五阶段批准:**
- Developer / 开发人员: _________________ Date / 日期: _______
- Reviewer / 审查员: _________________ Date / 日期: _______

**Production Launch Approval / 生产上线批准:**
- Project Manager / 项目经理: _________________ Date / 日期: _______
- Security Officer / 安全官: _________________ Date / 日期: _______
- Operations Lead / 运维负责人: _________________ Date / 日期: _______

---

**Document Version / 文档版本:** 1.0  
**Last Updated / 最后更新:** 2026-06-02  
**Next Review / 下次审查:** After each phase completion / 每个阶段完成后

**Instructions / 说明:**
- Check off items as completed / 完成项目时勾选
- Add notes for blockers or issues / 为阻碍或问题添加注释
- Update status column with ✅ (done), 🔄 (in progress), ⬜ (not started) / 用✅(完成)、🔄(进行中)、⬜(未开始)更新状态列
- Review with team weekly / 每周与团队审查
