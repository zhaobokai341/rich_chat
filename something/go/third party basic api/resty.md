Go Resty 是一个 Go 语言中非常流行且简单易用的 HTTP 和 REST 客户端库。它旨在让发送 HTTP 请求变得像喝水一样简单，同时提供了丰富的功能，如自动 JSON/XML 序列化、重试机制、中间件支持等。

下面我将为你介绍 Go Resty 的核心概念、安装方法以及几个常用的代码示例。

### 1. 安装

在你的 Go 项目目录下，使用 `go get` 命令安装：

```bash
go get github.com/go-resty/resty/v2
```

### 2. 基础用法

最简单的 GET 请求：

```go
package main

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
)

func main() {
	// 创建一个 Resty 客户端
	client := resty.New()

	// 发送 GET 请求
	resp, err := client.R().
		EnableTrace().
		Get("https://httpbin.org/get")

	// 处理错误
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	// 输出结果
	fmt.Println("Response Status:", resp.Status())
	fmt.Println("Response Body:", resp)
}
```

### 3. POST 请求 (发送 JSON)

Resty 可以自动将结构体或 Map 序列化为 JSON，并自动设置 `Content-Type: application/json` 头部。

```go
package main

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
)

// 定义一个用于发送的结构体
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	client := resty.New()

	user := User{
		Username: "codegeeX",
		Email:    "codegeeX@example.com",
	}

	// 发送 POST 请求
	resp, err := client.R().
		SetBody(user) // 自动序列化为 JSON
		Post("https://httpbin.org/post")

	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	fmt.Println("Response Status:", resp.Status())
	fmt.Println("Response Body:", resp)
}
```

### 4. 设置查询参数和路径参数

```go
client := resty.New()

resp, err := client.R().
		SetQueryParams(map[string]string{
			"page_no": "1",
			"limit":   "20",
		}).
		SetHeader("Accept", "application/json").
		SetPathParam("userId", "1001").
		Get("https://httpbin.org/users/{userId}")

// ... 处理响应
```

### 5. 高级特性：重试机制

Resty 内置了强大的重试机制，可以轻松配置在特定错误下自动重试。

```go
client := resty.New()

// 设置重试：最多重试 3 次，每次等待 1 秒
client.
    SetRetryCount(3).
    SetRetryWaitTime(1 * time.Second).
    SetRetryMaxWaitTime(5 * time.Second)

// 你可以添加条件来决定何时重试
client.AddRetryCondition(
    func(r *resty.Response, err error) bool {
        // 仅当服务器返回 500 状态码或发生网络错误时重试
        return r.StatusCode() >= 500 || err != nil
    },
)
```

### 6. 自动解析响应

Resty 可以自动将 JSON 响应体解析到你的结构体中，省去了 `json.Unmarshal` 的步骤。

```go
type Result struct {
	Args struct {
		Foo string `json:"foo"`
		Bar string `json:"bar"`
	} `json:"args"`
	Headers struct {
		Accept string `json:"Accept"`
	} `json:"headers"`
}

func main() {
	client := resty.New()

	var result Result

	// 发送请求并自动将结果解析到 result 对象中
	resp, err := client.R().
		SetResult(&result).
		SetQueryParam("foo", "bar").
		Get("https://httpbin.org/get")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Result Args:", result.Args.Foo)
	fmt.Println("Result Headers:", result.Headers.Accept)
}
```

### 7. 中间件 (拦截器)

你可以在请求发送前或响应接收后执行自定义逻辑，例如统一添加 Token 或记录日志。

```go
client := resty.New()

// 请求拦截器
client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
	fmt.Printf("正在请求: %s %s\n", req.Method, req.URL)
	return nil // 如果返回 error，请求将中止
})

// 响应拦截器
client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
	fmt.Printf("收到响应: %s\n", resp.Status())
	return nil
})

client.R().Get("https://httpbin.org/get")
```

### 总结

Go Resty 相比于 Go 标准库 `net/http` 的优势在于：
1.  **链式调用**：代码可读性极高。
2.  **自动化处理**：自动处理 JSON 编解码。
3.  **调试方便**：内置 Trace 功能。
4.  **企业级功能**：如重试、代理、Auth 等配置简单。

如果你需要开发复杂的 HTTP 客户端应用，Go Resty 是一个非常棒的选择。