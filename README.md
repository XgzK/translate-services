# Translate Services

一个基于 Echo 与自定义 DeepLX 适配层的轻量化翻译服务，提供与 Google Translate 接口兼容的 `/translate_a/single` 以及 `/translate_a/t` 端点，内建健康检查与 Prometheus 指标暴露能力，方便快速集成到现有工具链。

## 特性

- **协议兼容**：复刻 Google Translate 请求/响应格式，可被常见浏览器插件或脚本直接调用。
- **多提供商抽象**：通过 `internal/translator` 提供可插拔的翻译后端，目前内置 DeepLX。
- **稳健服务**：支持请求日志、超时、Body 限流、优雅停机与健康检查。
- **监控可观测**：内建 `/metrics`，以 Prometheus 形式导出关键指标。

## 环境要求

- Go `1.25` 或以上。
- 可用的 DeepLX/第三方翻译服务（需有效 API Key）。

## 快速开始

```bash
# 1. 安装依赖
go mod download

# 2. 准备配置（可直接复制模板后编辑）
cp config.example.yaml config.yaml

# 3. 运行服务
go run ./cmd/...   # 当前仓库直接执行: go run .
```

> 默认监听 `:8080`。生产环境中建议使用 `CONFIG_FILE` 或环境变量管理密钥，避免将 Key 写死在仓库中。

## 配置说明

配置文件默认读取根目录 `config.yaml`，可通过 `CONFIG_FILE` 指定其他路径。核心字段如下：

```yaml
port: "8080"            # 服务监听端口，亦可用环境变量 PORT 覆盖
debug: false            # 控制日志级别
translation:
  service_type: deeplx  # 当前支持 deeplx
  api_key: "xxx"        # 必填，DeepLX 访问密钥
  base_url: ""          # 可选，自定义 DeepLX/代理地址
```

环境变量覆盖优先于文件，支持：

| 变量 | 作用 |
| ---- | ---- |
| `PORT` / `DEBUG` | 覆盖监听端口与调试开关 |
| `TRANSLATION_SERVICE` / `DEEPLX_SERVICE` | 指定翻译后端类型 |
| `TRANSLATION_API_KEY` / `DEEPLX_API_KEY` | 配置 API Key |
| `TRANSLATION_BASE_URL` / `DEEPLX_BASE_URL` | 覆盖翻译后端地址 |

## API 参考

### `POST /translate_a/single`

- **请求体**：`application/json` 或 `application/x-www-form-urlencoded`
- **字段**：
  - `q`：待翻译文本（必填）
  - `sl`：源语言代码，留空自动检测
  - `tl`：目标语言代码
  - `dt`：数组，可重复，控制返回块（默认 `["t"]`）
- **示例**：

```bash
curl -X POST http://localhost:8080/translate_a/single \
  -H "Content-Type: application/json" \
  -d '{"q":"Hello","sl":"auto","tl":"zh-CN"}'
```

### `POST /translate_a/t`

- 用于 HTML 文档翻译，需保证 `format=html`。
- Query 参数：`client, sl, tl, format, tk`。
- Body：`form-data` 中包含 `q`（原文 HTML）。
- 若缺失任何必填字段将返回 `400`。

### 其他端点

| 方法 | 路径 | 描述 |
| ---- | ---- | ---- |
| `GET` | `/healthz` | 返回 `status` 与 `uptime`，供探活使用 |
| `GET` | `/metrics` | 暴露 Prometheus 指标（需配合 `echoprometheus` 中间件） |

## IntelliJ TranslationPlugin（谷歌自定义服务器）接入指南

> 适用于 IntelliJ 平台的 TranslationPlugin（https://github.com/YiiGuxing/TranslationPlugin），以自建谷歌翻译兼容接口方式使用。

1. **启动本服务**  
   - 确保 `translation.api_key` 配置有效（DeepLX 侧要求）。  
   - 执行 `go run .`（或部署后的可执行/容器）并确认监听端口，例如 `http://localhost:8080`。
2. **在 IDE 中设置**  
   - 打开 *Settings/Preferences → Tools → Translation*。  
   - 提供商选择 **Google**，勾选或切换为 **自定义服务器/Custom API**。  
   - 将自定义地址填为 `http://<你的主机>:<端口>`（路径无需填写，插件会按谷歌协议调用 `/translate_a/single` 等路径）。  
   - 若使用 HTTP（非 HTTPS），请确认 IDE 允许该地址；如走反向代理，请在代理上配置 TLS。
3. **验证**  
   - 在翻译窗口输入任意句子，来源/目标语言保持默认（自动检测）。  
   - 访问 `http://<主机>:<端口>/healthz` 应返回 `{"status":"ok"}`，`/metrics` 可用于监控。

常见问题：
- 出现 401/403：检查 `translation.api_key` 是否正确、DeepLX 侧是否可用。  
- 网络超时：确认 IDE 能访问到自建服务的主机与端口（必要时在 IDE 代理例外中加入该地址）。  
- 仍访问官方谷歌域：确认已选择“自定义服务器”而非默认 Google 域名。

## 日志与监控

- 使用 Zerolog 记录结构化请求日志，自动附带 `request_id`。
- Echo 中间件提供 `2MB` Body 限制、`12s` 超时与 panic 恢复。
- Prometheus 中间件自动统计 HTTP 指标，可直接 scrape `/metrics`。

## 项目结构速览

```
.
├── main.go                # 服务入口，加载配置并启动 Echo
├── internal/config        # 配置解析与校验
├── internal/server        # Echo 服务、路由、中间件与 Handler
├── internal/translation   # Google Translate 兼容结构、构造器
└── internal/translator    # DeepLX 实现与接口定义
```

## 开发与测试

- 运行单元测试：`go test ./...`
- 推荐为自定义翻译提供商实现 `internal/translator` 下的接口，并通过 `NewFactory` 注册。
- 提交前请确保 `go fmt ./...`、`go vet ./...` 能顺利通过，以维持代码质量。

## 部署建议

- 生产环境需通过环境变量或密钥管理服务注入 `TRANSLATION_API_KEY`。
- 搭配反向代理（Nginx、Caddy）处理 TLS，或直接将服务纳入容器编排（Docker/K8s）。
- 若放置在公网，建议额外接入认证/速率限制组件，防止滥用。

> 欢迎基于该服务扩展更多翻译后端，只需实现 `TranslationService` 接口并注册即可。
