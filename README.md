# smartCmd - 智能命令行工具

smartCmd 是一个使用 Go 语言开发的跨平台智能命令行工具，通过 AI 将自然语言转换为系统命令。

## 功能特性

- **自然语言转命令**: 使用自然语言描述需求，AI 自动生成对应的系统命令
- **跨平台支持**: 支持 Windows (PowerShell)、macOS (zsh/bash) 和 Linux (bash/sh)
- **智能验证**: 自动检查命令是否存在，最多 5 轮替代命令查找
- **命令执行与分析**: 可选择自动执行命令并 AI 分析输出结果
- **零配置感知**: 自动识别系统环境和 Shell 类型
- **安全保护**: 危险命令二次确认机制

## 快速开始

### 1. 构建

```bash
# 设置 LLM 配置并构建
make build BASE_URL=https://your-api-url API_KEY=your-api-key MODEL=gpt-4o

# 或使用环境变量
export LLM_BASE_URL=https://your-api-url
export LLM_API_KEY=your-api-key
export LLM_MODEL=gpt-4o
make build
```

### 2. 初始化

```bash
# 运行初始化（会输出 shell 脚本）
./dist/smartCmd

# 将初始化脚本加入 shell 配置
# bash/zsh
echo 'eval "$(./dist/smartCmd)"' >> ~/.bashrc  # 或 ~/.zshrc

# 重新加载配置
source ~/.bashrc
```

### 3. 使用

初始化后，可以使用 `?` 和 `??` 别名：

```bash
# 解析命令（不执行）
? 查看本机所有开放的端口

# 解析并执行命令
?? 我的 IP 地址是什么
```

## 运行模式

### 初始化模式

```bash
smartCmd
```

采集系统信息，验证 LLM 配置，输出 shell 初始化脚本。

### 命令解析模式

```bash
smartCmd parse <自然语言命令>
```

分析用户输入，输出转换后的命令（不执行）。

### 解析并执行模式

```bash
smartCmd parseAndRun <自然语言命令>
```

分析、执行命令，并 AI 解析执行结果。

## 配置

### 编译时配置

通过 `-ldflags` 在编译时注入默认配置：

```bash
go build -ldflags \
  "-X smartCmd/config.DefaultBaseURL=https://api.openai.com/v1 \
   -X smartCmd/config.DefaultAPIKey=sk-xxxx \
   -X smartCmd/config.DefaultModel=gpt-4o" \
  -o smartCmd .
```

### 运行时配置

环境变量会覆盖编译时配置：

- `LLM_BASE_URL`: LLM API 基础 URL（兼容 OpenAI 协议）
- `LLM_API_KEY`: API 密钥
- `LLM_MODEL`: 模型名称（如 gpt-4o, claude-3-5-sonnet）
- `LLM_PROMPT`: 附加系统提示词

## 构建命令

```bash
# 本地构建
make build

# 构建所有平台
make build-all

# 单独构建
make build-linux   # Linux
make build-mac     # macOS
make build-win     # Windows

# 清理
make clean
```

## 使用示例

### 示例 1: 查看端口

```bash
$ ? 查看本机所有开放的端口
⟳ 理解用户意图...
✓ 意图：查询本机监听的网络端口
✓ 命令：netstat -tuln
```

### 示例 2: 获取 IP 地址

```bash
$ ?? 我的 IP 地址是什么
⟳ 理解用户意图...
✓ 意图：获取本机公网 IP
✓ 命令：curl ifconfig.me
▶ 正在执行：curl ifconfig.me
─────────────── 原始输出 ───────────────
203.0.113.42
────────────────────────────────────────
✓ 分析：命令成功执行，输出显示了你的公网 IP 地址。
```

### 示例 3: 命令不存在时自动查找替代

```bash
$ ? 安装 openclaw
⟳ 理解用户意图...
✓ 意图：安装 OpenClaw 游戏
⚠ 命令 'openclaw' 不存在，尝试第 1 次替代...
✓ 命令：brew install openclaw
```

## 安全注意事项

- 危险命令（如 `rm -rf /`）会触发二次确认
- 命令输出超过 4000 字符会自动截断
- 所有命令以当前用户权限执行

## 项目结构

```
smartCmd/
├── config/         # 配置管理
├── sysinfo/        # 系统信息采集
├── shell/          # Shell 类型检测和命令执行
├── ui/             # 终端交互界面
├── llm/            # LLM API 调用
├── validator/      # 命令验证和重试
├── init/           # 初始化脚本生成
├── main.go         # 主入口
├── Makefile        # 构建脚本
└── README.md       # 项目文档
```

## 退出码

- `0`: 成功
- `1`: 参数错误
- `2`: LLM 请求失败
- `3`: 5 轮后仍未找到可执行命令
- `4`: 命令执行失败

## 许可证

MIT License
