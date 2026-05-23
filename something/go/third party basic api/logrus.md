`logrus` 是 Go 语言中目前最流行、使用最广泛的结构化日志库之一。它完全兼容 Go 标准库 `log` 的 API，并提供了可插拔的 Hook（钩子）架构、自定义格式化器等功能。

下面我将详细介绍 `logrus` 的安装、基本用法、进阶配置以及如何与第三方库集成。

### 1. 安装

在使用之前，你需要先安装它：

```bash
go get github.com/sirupsen/logrus
```

> **注意**：导入时路径是 `github.com/sirupsen/logrus` (不是 `Sirupsen`)，这是为了修正 GitHub 上的重命名问题。

### 2. 基础用法

`logrus` 支持日志级别：`Trace`, `Debug`, `Info`, `Warning`, `Error`, `Fatal` 和 `Panic`。

```go
package main

import (
    log "github.com/sirupsen/logrus"
)

func main() {
    log.Info("这是一条普通信息")
    log.Warn("这是一条警告信息")
    log.Error("这是一条错误信息")

    // Fatal 会打印日志并退出程序 (os.Exit(1))
    // log.Fatal("严重错误，程序终止")

    // Panic 会打印日志并抛出 panic
    // log.Panic("发生异常")
}
```

### 3. 设置日志级别

你可以设置日志级别，低于该级别的日志将不会输出。这对于生产环境非常有用（例如关闭 Debug 日志）。

```go
package main

import (
    log "github.com/sirupsen/logrus"
)

func main() {
    // 设置日志级别为 Warn
    log.SetLevel(log.WarnLevel)

    log.Debug("这条不会显示，因为级别低于 Warn")
    log.Info("这条也不会显示")
    log.Warn("这条会显示")
    log.Error("这条也会显示")
}
```

### 4. 结构化日志

这是 `logrus` 最强大的功能。推荐使用 `WithFields` 来记录结构化数据，而不是拼接字符串。

```go
package main

import (
    "os"
    log "github.com/sirupsen/logrus"
)

func init() {
    // 设置为 JSON 格式输出
    log.SetFormatter(&log.JSONFormatter{})
    // 设置输出到标准输出 (默认也是 os.Stderr)
    log.SetOutput(os.Stdout)
}

func main() {
    // 使用 WithFields 记录上下文信息
    log.WithFields(log.Fields{
        "event":     "user_login",
        "user_id":   1001,
        "ip":        "192.168.1.1",
        "status":    "success",
    }).Info("用户登录成功")
}
```

**输出结果示例：**
```json
{"event":"user_login","ip":"192.168.1.1","level":"info","msg":"用户登录成功","status":"success","time":"2023-10-27T10:00:00+08:00","user_id":1001}
```

### 5. 自定义 Logger 实例

在大型项目中，我们可能需要不同的 Logger 实例（例如一个用于 HTTP 请求日志，一个用于数据库日志），而不是使用全局的 `log`。

```go
package main

import (
    "os"
    "github.com/sirupsen/logrus"
)

func main() {
    // 创建一个新的 Logger 实例
    var httpLogger = logrus.New()
    
    // 可以独立配置这个实例
    httpLogger.SetFormatter(&logrus.TextFormatter{
        ForceColors: true,
        FullTimestamp: true,
    })
    
    httpLogger.Out = os.Stdout

    // 使用自定义实例记录日志
    httpLogger.WithFields(logrus.Fields{
        "method": "GET",
        "path":   "/api/users",
    }).Info("收到请求")
}
```

### 6. 常用 Formatter (格式化器)

除了默认的文本格式，`logrus` 内置了 `JSONFormatter`，你也可以自定义格式。

#### 文本模式 (TextFormatter) 带颜色和时间

```go
log.SetFormatter(&logrus.TextFormatter{
    FullTimestamp: true,
    TimestampFormat: "2006-01-02 15:04:05", // Go 的特定时间格式
    ForceColors: true,                     // 强制显示颜色
})
```

### 7. 日志分割与归档

`logrus` 本身不负责日志文件的分割（如按天切割、按大小切割）。通常建议配合第三方库使用，最常用的是 **lumberjack**。

**安装：**
```bash
go get gopkg.in/natefinch/lumberjack.v2
```

**使用示例：**

```go
package main

import (
    "gopkg.in/natefinch/lumberjack.v2"
    log "github.com/sirupsen/logrus"
)

func main() {
    // 设置 lumberjack
    log.SetOutput(&lumberjack.Logger{
        Filename:   "./app.log", // 日志文件名
        MaxSize:    10,          // 单个文件最大尺寸 (MB)
        MaxBackups: 3,           // 保留的旧文件最大数量
        MaxAge:     28,          // 保留旧文件的最大天数
        Compress:   true,        // 是否压缩旧文件
    })

    log.Info("这条日志会被写入文件，并自动进行切割管理")
}
```

### 8. Hooks (钩子) - 进阶功能

Hooks 允许你在日志记录的特定阶段（如记录完毕后）执行一些逻辑。常见的用法包括：
*   将错误日志发送到 Slack/钉钉群。
*   将日志发送到 Sentry (错误追踪)。
*   将日志写入数据库。

这里展示一个简单的自定义 Hook 示例（例如只记录 Error 级别以上的日志到特定文件）：

```go
package main

import (
    "os"
    "github.com/sirupsen/logrus"
)

// MyHook 实现 logrus.Hook 接口
type MyHook struct{}

func (hook *MyHook) Fire(entry *logrus.Entry) error {
    // 只有 Error 级别以上才处理
    if entry.Level <= logrus.ErrorLevel {
        line, err := entry.String()
        if err != nil {
            return err
        }
        // 这里可以写入文件或发送网络请求
        // 简单演示：写入另一个文件
        f, _ := os.OpenFile("errors_only.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        defer f.Close()
        f.WriteString(line)
    }
    return nil
}

func (hook *MyHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

func main() {
    log.AddHook(&MyHook{})
    log.Info("普通信息")
    log.Error("这条错误信息会被 MyHook 捕获并写入 errors_only.log")
}
```

### 总结

1.  **开始**：直接使用 `log.Info`, `log.Error`。
2.  **结构化**：使用 `WithFields` 代替字符串拼接。
3.  **生产环境**：设置 `log.SetLevel`，使用 `JSONFormatter`，配合 `lumberjack` 进行日志文件管理。
4.  **扩展**：使用 Hooks 实现日志告警或第三方平台集成。

`logrus` 是一个非常成熟的库，足以应对绝大多数 Go 项目的日志需求。