# UI Handler 重构文档

## 📁 目录结构

```
client/
├── ui_handler/                    # UI 处理模块
│   ├── interfaces.go              # 接口定义 + 数据结构
│   ├── factory.go                 # 工厂函数（预留扩展）
│   ├── mocks.go                   # Mock 实现（测试用）
│   ├── ui_handler.go              # 主 UIHandler + 核心流程 + 菜单渲染器
│   ├── message_printer.go         # 消息打印（lipgloss 美化）
│   ├── message_printer_test.go    # 消息打印测试
│   ├── validators.go              # 输入验证逻辑
│   ├── validators_test.go         # 验证逻辑测试
│   ├── auth_menu.go               # 认证菜单（登录/注册/登出/删除账户）
│   ├── auth_menu_test.go          # 认证菜单测试
│   ├── user_menu.go               # 用户信息管理菜单
│   ├── user_menu_test.go          # 用户菜单测试
│   ├── chat_menu.go               # 聊天菜单
│   └── chat_menu_test.go          # 聊天菜单测试
├── adapters.go                    # 类型适配器（桥接 main 包和 ui_handler 包）
├── main.go                        # 主入口
└── ui_handler.go.bak              # 旧文件备份（供参考）
```

## 🎯 设计原则

### 1. 单一职责原则
每个文件只负责一件事：
- `interfaces.go` - 定义接口和数据结构
- `message_printer.go` - 只负责打印消息
- `validators.go` - 只负责验证输入
- `auth_menu.go` - 只负责认证相关菜单
- `user_menu.go` - 只负责用户信息菜单
- `chat_menu.go` - 只负责聊天菜单

### 2. 依赖注入
所有依赖通过构造函数注入：
```go
func NewUIHandler(
    authService AuthService,
    userService UserService,
    apiClient APIClient,
    chatAPIClient ChatAPIClient,
    chatService ChatService,
    configMgr ConfigManager,
    languagePack LanguagePack,
    configDir string,
) *UIHandler
```

### 3. 接口隔离
每个接口只定义必要的方法：
- `AuthService` - 认证相关
- `UserService` - 用户操作相关
- `APIClient` - REST API 相关
- `ChatAPIClient` - 聊天 API 相关
- `ChatService` - 聊天服务相关
- `ConfigManager` - 配置管理相关

## 📝 代码规范

### 命名规范
- **接口**: 使用 `er` 后缀，如 `AuthService`, `UserService`
- **结构体**: 使用名词，如 `UIHandler`, `MessagePrinter`, `MenuRenderer`
- **函数**: 使用动词开头，如 `handleLogin`, `printInfo`, `validatePassword`
- **测试函数**: `Test + 被测试对象 + 场景`，如 `TestUIHandler_HandleLogin_Success`

### 文件组织
1. **接口定义** 在 `interfaces.go`
2. **实现** 按功能拆分到不同文件
3. **测试** 每个实现文件对应一个 `_test.go`
4. **Mock** 统一放在 `mocks.go`

### 错误处理
- 使用 `printError()` 统一处理错误输出
- 错误消息通过 `languagePack.Get()` 国际化
- 使用 `fmt.Errorf("%s: %w", msg, err)` 包装错误

### 消息打印
```go
// 正确用法
h.printInfo("key")           // 从语言包获取消息
h.printError(err)            // 直接打印 error
h.printError("key")          // 从语言包获取消息
h.printSuccess("key")        // 从语言包获取消息

// 错误用法（已修复）
h.printError(fmt.Sprintf("key: %v", err))  // ❌ 会尝试查找 "key: xxx" 这个 key
```

## 🔧 如何添加新功能

### 1. 添加新的菜单功能
**步骤：**
1. 在对应的菜单文件中添加函数（如 `auth_menu.go`、`user_menu.go`、`chat_menu.go`）
2. 在 `interfaces.go` 中添加需要的接口方法（如果需要）
3. 在 `mocks.go` 中添加 Mock 实现
4. 编写测试文件

**示例：添加新的用户菜单功能**
```go
// user_menu.go
func (h *UIHandler) handleNewFeature() {
    h.printInfo("new_feature_start")
    // 实现逻辑
    h.printSuccess("new_feature_success")
}
```

### 2. 添加新的验证规则
**步骤：**
1. 在 `validators.go` 中添加验证函数
2. 在 `validators_test.go` 中添加测试

**示例：**
```go
// validators.go
func validatePhoneNumber(phone string) error {
    if phone == "" {
        return errors.New("phone cannot be empty")
    }
    if len(phone) < 11 {
        return errors.New("phone number too short")
    }
    return nil
}
```

### 3. 修改消息样式
**步骤：**
1. 在 `message_printer.go` 中修改样式
2. 运行测试确保样式正确

**示例：**
```go
// 修改 Info 样式
infoStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#00FF00")).  // 改为绿色
    Bold(true).
    Border(lipgloss.NormalBorder(), true).
    BorderForeground(lipgloss.Color("#00FF00"))
```

## 🧪 测试规范

### 测试文件命名
- 每个实现文件对应一个 `_test.go`
- 如：`auth_menu.go` → `auth_menu_test.go`

### 测试函数命名
- `Test + 被测试对象 + 方法 + 场景`
- 如：`TestUIHandler_HandleLogin_Success`
- 如：`TestUIHandler_HandleLogin_Error`

### Mock 使用
```go
func TestUIHandler_HandleLogin_Success(t *testing.T) {
    // 1. 创建 Mock
    mockAuth := new(MockAuthService)
    mockAuth.On("Login", "user", "pass").Return(nil)

    // 2. 创建 Handler
    handler := &UIHandler{
        authService: mockAuth,
        languagePack: new(MockLanguagePack),
        printer: NewMessagePrinter(),
    }

    // 3. 调用方法
    handler.handleLogin()

    // 4. 验证 Mock 调用
    mockAuth.AssertExpectations(t)
}
```

### 测试注意事项
1. **需要 stdin 的测试**：如 `input()`、`readPassword()`，只验证函数存在，不实际调用
2. **Mock 期望**：确保所有 `languagePack.Get()` 调用都有对应的 Mock
3. **错误场景**：测试成功和失败两种情况

## 🐛 已知修复的 Bug

1. **类型断言 panic**：移除了 `h.apiClient.(*RestAPIClient)` 的危险类型断言
2. **错误消息格式化**：修复了 `printError(fmt.Sprintf(...))` 导致的语言包查找失败
3. **密码验证重复**：提取到 `validators.go` 统一处理
4. **parseUserID 错误检查**：完善了负数和零的检查

## 🎨 lipgloss 样式规范

### 消息样式
- **Info**: 蓝色 + `[i]` + 左边框
- **Warning**: 黄色 + `[!]` + 左边框
- **Error**: 红色 + `[x]` + 左边框
- **Success**: 绿色 + `[ok]` + 左边框

### 菜单样式
- **标题**: 粗体 + 下划线 + 黄色
- **菜单框**: 圆角边框 + 蓝色边框 + 内边距
- **菜单项**: 左缩进 2 空格 + 编号

### 卡片样式
- **用户信息**: 圆角边框 + 绿色边框
- **会话列表**: 圆角边框 + 蓝色边框

## 📦 依赖关系

```
main.go
  └── adapters.go (类型适配)
       └── ui_handler/
            ├── ui_handler.go (主流程)
            │    ├── auth_menu.go
            │    ├── user_menu.go
            │    └── chat_menu.go
            ├── message_printer.go
            ├── validators.go
            ├── menu_renderer (内嵌在 ui_handler.go)
            └── interfaces.go (接口定义)
```

## 🚀 快速上手

### 1. 理解主流程
阅读 `ui_handler.go` 的 `Start()` 方法：
```
Start() → checkServer() → handleAuthenticated() → showMainMenu()
```

### 2. 理解菜单流程
- **认证菜单**: `auth_menu.go` - 登录/注册/登出/删除账户
- **用户菜单**: `user_menu.go` - 查看/修改用户信息/修改密码
- **聊天菜单**: `chat_menu.go` - 聊天会话/新建聊天/生成密钥

### 3. 理解依赖注入
所有依赖通过 `NewUIHandler()` 注入，便于测试和替换实现。

### 4. 理解测试
每个功能文件都有对应的测试文件，使用 Mock 隔离依赖。

## 💡 最佳实践

1. **添加新功能时**：先确定属于哪个模块（auth/user/chat），然后在对应文件中添加
2. **修改样式时**：只修改 `message_printer.go` 或 `MenuRenderer`
3. **添加验证时**：只修改 `validators.go`
4. **测试时**：使用 Mock 隔离外部依赖，只测试当前单元的逻辑
5. **错误处理时**：使用 `printError()` 统一处理，不要直接 `fmt.Println()`

## 📊 重构成果

| 指标 | 重构前 | 重构后 |
|------|--------|--------|
| 文件数 | 1 | 15 |
| 总行数 | 974 | 2519 (含测试) |
| 测试覆盖 | 0 | 40+ 测试用例 |
| 代码重复 | 高 | 低 |
| 可测试性 | 差 | 优秀 |
| 可维护性 | 差 | 优秀 |