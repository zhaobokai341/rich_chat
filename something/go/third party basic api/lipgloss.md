`lipgloss` 是一个非常流行的 Go 语言库，用于在终端中构建美观、现代化的用户界面（TUI）。它类似于 CSS，允许你设置颜色、样式、边距、边框等，但作用对象是终端文本。

下面我将为你详细介绍 `lipgloss` 的基础语法和常用功能。

### 1. 安装

首先，你需要安装这个库：

```bash
go get github.com/charmbracelet/lipgloss
```

### 2. 基础概念：Style

`lipgloss` 的核心是 `Style` 结构体。你可以定义一个样式，然后将其应用到字符串上。

```go
package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// 定义一个样式
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")). // 前景色（文字颜色）
		Background(lipgloss.Color("#7D56F4")). // 背景色
		Padding(1, 2)                          // 内边距 (上下, 左右)

	// 应用样式
	str := style.Render("Hello, Lipgloss!")
	fmt.Println(str)
}
```

### 3. 颜色

`lipgloss` 支持多种颜色设置方式：

*   **ANSI 16色**: 使用标准颜色名称，如 `red`, `blue`, `white` 等。
*   **ANSI 256色**: 使用数字 `0`-`255`。
*   **真彩色 (Hex)**: 使用十六进制字符串，如 `#FF0000`。

```go
// 1. 使用颜色名称
s1 := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))

// 2. 使用 ANSI 256 色代码
s2 := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))

// 3. 使用 Hex 颜色代码
s3 := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

// 4. 使用自适应颜色 (AdaptiveColor)
// 根据终端背景是深色还是浅色自动切换颜色
lightColor := "#FFFFFF"
darkColor := "#000000"
adaptive := lipgloss.AdaptiveColor{Light: lightColor, Dark: darkColor}
s4 := lipgloss.NewStyle().Foreground(adaptive)
```

### 4. 布局与间距

你可以像 CSS 一样控制边距、内边距和对齐方式。

```go
style := lipgloss.NewStyle().
	Width(20).              // 设置宽度
	Height(5).              // 设置高度
	Align(lipgloss.Center). // 水平对齐
	Padding(1, 2).          // 内边距：上下1，左右2
	Margin(1, 0).           // 外边距：上下1，左右0
	MarginTop(2).           // 单独设置上外边距
	MarginLeft(4)           // 单独设置左外边距
```

### 5. 边框

`lipgloss` 提供了非常强大的边框功能。

```go
style := lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()). // 设置边框样式：圆角
	BorderForeground(lipgloss.Color("63")). // 边框颜色
	BorderBackground(lipgloss.Color("235")). // 边框背景色（仅影响边框线条）
```

**内置边框样式：**
*   `NormalBorder()`: 普通边框
*   `RoundedBorder()`: 圆角边框
*   `ThickBorder()`: 粗边框
*   `DoubleBorder()`: 双线边框
*   `HiddenBorder()`: 隐藏边框

你也可以自定义边框字符：

```go
customBorder := lipgloss.Border{
	Top:    "_",
	Bottom: "-",
	Left:   "|",
	Right:  "|",
}
style := lipgloss.NewStyle().Border(customBorder, true) // true 表示显示所有边
```

### 6. 文本样式

控制粗体、斜体、下划线、删除线等。

```go
style := lipgloss.NewStyle().
	Bold(true).      // 粗体
	Italic(true).    // 斜体
	Underline(true). // 下划线
	Strikethrough(true). // 删除线
	Blink(true).     // 闪烁 (部分终端支持)
	Faint(true)      // 暗淡 (部分终端支持)
```

### 7. 继承与组合

你可以基于已有的样式创建新样式，或者将两个样式合并。

```go
// 基础样式
baseStyle := lipgloss.NewStyle().
	Foreground(lipgloss.Color("white")).
	Background(lipgloss.Color("blue"))

// 继承并扩展
derivedStyle := baseStyle.Copy().
	Bold(true).
	Padding(1)

// 合并样式
styleA := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
styleB := lipgloss.NewStyle().Background(lipgloss.Color("black"))
mergedStyle := styleA.Inherit(styleB) // styleB 的属性会覆盖 styleA 的同名属性
```

### 8. 渲染与拼接

通常你会将多个组件拼接在一起。`lipgloss` 提供了垂直和水平拼接的辅助函数。

```go
var (
	headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("blue")).Bold(true)
	contentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("white"))
	footerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))
)

header  := headerStyle.Render("Header")
	content := contentStyle.Render("This is the main content.")
	footer  := footerStyle.Render("Footer")

// 垂直拼接 (换行)
vertical := lipgloss.JoinVertical(lipgloss.Left, header, content, footer)

// 水平拼接 (空格分隔)
col1 := lipgloss.NewStyle().Width(10).Render("Col 1")
col2 := lipgloss.NewStyle().Width(20).Render("Col 2")
horizontal := lipgloss.JoinHorizontal(lipgloss.Top, col1, col2)

fmt.Println(vertical)
fmt.Println(horizontal)
```

### 9. 完整示例

下面是一个结合了上述功能的完整示例，模拟一个简单的卡片组件：

```go
package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// 定义通用变量
	var (
		// 边框样式
		borderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(1, 2)

		// 标题样式
		titleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("228")).
				Background(lipgloss.Color("62")).
				Padding(0, 1).
				Bold(true)

		// 内容样式
		bodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Width(30) // 限制内容宽度，强制换行
	)

	// 渲染各个部分
	title := titleStyle.Render("Welcome")
	body := bodyStyle.Render("Lipgloss makes styling terminal output easy and fun. It's like CSS for the console!")

	// 拼接内容：标题在上，正文在下
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body)

	// 应用边框
	card := borderStyle.Render(content)

	// 输出
	fmt.Println(card)
}
```

### 10. 提示

*   **性能**: `Style.Render` 是相对廉价的操作，但在循环中大量调用时，建议尽可能复用 `Style` 对象，而不是每次都创建新的。
*   **宽度计算**: 在处理多行文本时，`lipgloss` 会自动处理 ANSI 转义序列的宽度，确保对齐正确，这是它比手动拼接字符串强大的地方。
*   **环境变量**: `NO_COLOR` 环境变量会强制禁用颜色。`TERM` 环境变量会影响颜色支持检测。

通过这些基础语法，你就可以开始构建复杂的终端 UI 了。