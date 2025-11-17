# DeepLX 翻译器 Go 包

> 🐱 浮浮酱用心制作的专业翻译包喵～

一个与 DeepLX API 兼容的 Go 语言翻译客户端，由大型语言模型（LLM）驱动。

## ✨ 特性

- ✅ **完全兼容** DeepLX API 规范
- ✅ **类型安全** 的 Go 接口设计
- ✅ **HTTP 客户端复用** 提高性能
- ✅ **超时保护** 避免长时间等待
- ✅ **灵活扩展** 支持自定义配置
- ✅ **完善测试** 单元测试 + 集成测试 + 基准测试

## 🎯 编程原则

本项目严格遵循以下编程原则：

- **SOLID 原则**：单一职责、开放封闭、接口隔离等
- **KISS (简单至上)**：代码简洁直观
- **DRY (杜绝重复)**：避免重复代码
- **YAGNI (精益求精)**：只实现必要功能

## 📦 安装

```bash
go get untitled/deeplx
```

## 🚀 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "untitled/deeplx"
)

func main() {
    // 创建翻译器实例
    translator, err := deeplx.NewTranslator("sk-your-api-key")
    if err != nil {
        panic(err)
    }

    // 执行翻译
    result := translator.Translate("Hello, world!", "ZH", "EN")

    if result.Success {
        fmt.Println(result.TranslatedText) // 输出: 你好，世界！
    } else {
        fmt.Println("翻译失败:", result.ErrorMessage)
    }
}
```

### 自动检测源语言

```go
// 不指定源语言，自动检测
result := translator.Translate("你好，世界！", "EN")
```

### 使用指定模型

```go
result := translator.TranslateWithModel(
    "Hello, world!",
    "ZH",
    "gpt-4",  // 模型名称
    "EN",
)
```

### 自定义配置

```go
import (
    "net/http"
    "time"
)

// 使用自定义 HTTP 客户端
client := &http.Client{
    Timeout: 60 * time.Second,
}
translator, _ := deeplx.NewTranslatorWithClient("sk-your-key", client)

// 设置自定义基础 URL（私有部署）
translator.SetBaseURL("https://your-custom-domain.com/api")
```

## 📖 API 文档

### 核心类型

#### `DeepLXTranslator`

主翻译器结构体。

**方法：**

- `NewTranslator(apiKey string) (*DeepLXTranslator, error)`
  - 创建新的翻译器实例

- `NewTranslatorWithClient(apiKey string, client *http.Client) (*DeepLXTranslator, error)`
  - 使用自定义 HTTP 客户端创建翻译器

- `Translate(text, targetLang string, sourceLang ...string) *TranslationResult`
  - 执行翻译，sourceLang 可选（留空自动检测）

- `TranslateWithModel(text, targetLang, model string, sourceLang ...string) *TranslationResult`
  - 使用指定模型执行翻译

- `SetBaseURL(baseURL string)`
  - 设置自定义基础 URL

#### `TranslationResult`

翻译结果结构体。

```go
type TranslationResult struct {
    Success        bool                  // 是否成功
    TranslatedText string                // 翻译后的文本
    SourceLang     string                // 检测到的源语言
    TargetLang     string                // 目标语言
    ErrorMessage   string                // 错误信息（失败时）
    RawResponse    *TranslationResponse  // 原始响应
}
```

### 支持的语言代码

常用语言代码示例：

| 代码 | 语言 | 代码 | 语言 |
|------|------|------|------|
| ZH / ZH-HANS | 中文（简体） | EN / EN-US | 英语（美式） |
| JA | 日语 | KO | 韩语 |
| FR | 法语 | DE | 德语 |
| ES | 西班牙语 | RU | 俄语 |

更多语言请参考 [DeepL 文档](https://translate.ai.jayogo.com/deeplx.html)。

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./deeplx

# 运行测试并显示详细信息
go test -v ./deeplx

# 运行测试并显示覆盖率
go test -cover ./deeplx

# 生成覆盖率报告
go test -coverprofile=coverage.out ./deeplx
go tool cover -html=coverage.out
```

### 运行基准测试

```bash
go test -bench=. ./deeplx
```

### 运行示例

```bash
# 设置 API 密钥环境变量
export DEEPLX_API_KEY="sk-your-api-key"

# 运行示例程序
go run examples/basic_usage.go
```

## 📁 项目结构

```
untitled/
├── deeplx/                    # 翻译器包
│   ├── translator.go          # 核心实现
│   ├── translator_test.go     # 测试文件
│   └── README.md              # 包文档
├── examples/                  # 使用示例
│   └── basic_usage.go         # 基本用法示例
├── go.mod                     # Go 模块定义
└── go.sum                     # 依赖校验和
```

## 🔒 最佳实践

### 1. API 密钥管理

**推荐方式：使用环境变量**

```go
import "os"

apiKey := os.Getenv("DEEPLX_API_KEY")
translator, _ := deeplx.NewTranslator(apiKey)
```

**不推荐：硬编码在代码中**

```go
// ❌ 不要这样做
translator, _ := deeplx.NewTranslator("sk-123456789")
```

### 2. 错误处理

始终检查翻译结果：

```go
result := translator.Translate("text", "ZH")
if !result.Success {
    log.Printf("翻译失败: %s", result.ErrorMessage)
    return
}
```

### 3. 复用翻译器实例

```go
// ✅ 推荐：创建一次，多次使用
translator, _ := deeplx.NewTranslator(apiKey)
for _, text := range texts {
    result := translator.Translate(text, "ZH")
    // 处理结果...
}

// ❌ 不推荐：每次都创建新实例
for _, text := range texts {
    translator, _ := deeplx.NewTranslator(apiKey)  // 浪费资源
    result := translator.Translate(text, "ZH")
}
```

## 🤝 贡献

欢迎贡献代码！请确保：

1. 遵循项目的编程原则（SOLID、KISS、DRY、YAGNI）
2. 添加必要的测试
3. 更新相关文档

## 📄 许可证

MIT License

---

*🐱 浮浮酱用心制作，祝您使用愉快喵～ o(*￣︶￣*)o*
