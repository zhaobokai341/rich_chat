Go Gin 的“语法”本质上就是 Go 语言结合 Gin 框架提供的 API 调用方式。掌握 Gin 的基础语法，主要就是掌握如何初始化引擎、定义路由、获取参数、返回响应以及使用中间件。

以下是 Go Gin 最核心的 5 个基础语法板块，我为你整理成了代码示例和说明：

### 1. 初始化与启动

这是所有 Gin 程序的起点。

```go
package main

import "github.com/gin-gonic/gin"

func main() {
	// 1. 创建引擎
	// gin.Default() 包含 Logger（日志）和 Recovery（ panic 捕获）中间件
	// gin.New() 是一个没有任何中间件的空引擎
	r := gin.Default()

	// ... 定义路由 ...

	// 2. 启动 HTTP 服务
	// 默认监听本机 0.0.0.0:8080
	r.Run() 
}
```

### 2. 路由定义

Gin 支持所有的 HTTP 方法，并且支持路径参数。

```go
// 基础 RESTful 风格路由
r.GET("/get", func(c *gin.Context) { ... })
r.POST("/post", func(c *gin.Context) { ... })
r.PUT("/put", func(c *gin.Context) { ... })
r.DELETE("/delete", func(c *gin.Context) { ... })

// 路径参数 (例如: /user/zhangsan)
// 注意：冒号 : 是关键语法
r.GET("/user/:name", func(c *gin.Context) {
    name := c.Param("name") // 获取路径中的 name
    c.String(200, "Hello %s", name)
})

// 通配符参数 (例如: /files/logo.jpg, /files/static/css/style.css)
// *action 会匹配包括 / 的路径
r.GET("/files/*filepath", func(c *gin.Context) {
    filepath := c.Param("filepath")
    c.String(200, "%s", filepath)
})
```

### 3. 获取请求参数

这是处理业务逻辑最常用的部分，分为三种主要来源。

```go
r.POST("/post", func(c *gin.Context) {
    
    // 1. 获取 Query 参数 (URL 中 ? 后面的参数)
    // 例如: /post?name=tom&age=20
    name := c.Query("name")
    age := c.DefaultQuery("age", "18") // 带默认值

    // 2. 获取 Post Form 参数 (表单提交)
    // 例如: body 中 message=hello
    message := c.PostForm("message")

    // 3. 获取 JSON/XML 参数并绑定到结构体 (最常用)
    // 定义结构体，使用 binding 标签进行验证
    type Person struct {
        Name    string `json:"name" binding:"required"` // 必填
        Address string `json:"address"`
    }

    var person Person
    // ShouldBindJSON 会自动解析 Body 中的 JSON 并验证
    if err := c.ShouldBindJSON(&person); err == nil {
        c.JSON(200, gin.H{"data": person})
    } else {
        c.JSON(400, gin.H{"error": err.Error()})
    }
})
```

### 4. 返回响应

Gin 提供了多种格式的响应方法。

```go
r.GET("/response", func(c *gin.Context) {
    
    // 1. 返回字符串
    c.String(200, "纯文本响应")

    // 2. 返回 JSON (最常用)
    c.JSON(200, gin.H{
        "status": "success",
        "code":   200,
    })
    // 或者直接传结构体
    // c.JSON(200, userObj)

    // 3. 返回 HTML
    // 需要先使用 r.LoadHTMLGlob("templates/*") 加载模板
    // c.HTML(200, "index.html", gin.H{"title": "首页"})

    // 4. 返回文件
    // c.File("./uploads/photo.jpg")

    // 5. 重定向
    // c.Redirect(301, "/login")
})
```

### 5. 路由分组

当项目变大时，使用路由分组可以更好地管理 API（例如按版本或模块划分）。

```go
// v1 组
v1 := r.Group("/v1")
{
    // 实际访问路径是 /v1/login
    v1.POST("/login", loginEndpoint) 
    v1.POST("/submit", submitEndpoint)
}

// v2 组，并且给这个组单独使用中间件
v2 := r.Group("/v2")
v2.Use(AuthMiddleware()) // 只有 v2 下的路由会经过此中间件
{
    v2.POST("/login", loginEndpoint)
    v2.POST("/submit", submitEndpoint)
}
```

### 6. 中间件

中间件语法本质上是一个函数，它接收 `gin.Context`，并调用 `c.Next()`。

```go
// 定义中间件
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        fmt.Println("请求前...")
        
        c.Next() // 处理请求
        
        fmt.Println("请求后...")
    }
}

// 全局使用
r.Use(Logger())

// 单个路由使用
r.GET("/test", Logger(), func(c *gin.Context) {
    // ...
})
```

### 总结

学习 Go Gin 的基础语法，建议按以下顺序练习：
1.  先会用 `r.GET/POST` 和 `c.JSON` 跑通一个接口。
2.  学会 `c.ShouldBindJSON` 处理前端传来的复杂数据。
3.  学会 `r.Group` 整理代码结构。
4.  最后学习中间件处理登录鉴权等通用逻辑。