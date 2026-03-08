# Prompt 模板说明

本目录包含 smartCmd 使用的所有 AI 提示词模板。这些模板在编译时通过 Go embed 嵌入到二进制文件中。

## 模板文件

### system.txt
命令解析的系统提示词，用于将用户的自然语言转换为系统命令。

**可用变量:**
- `{{.Shell}}`: Shell 类型 (bash/zsh/powershell/fish/sh)
- `{{.SystemInfo}}`: 系统信息描述
- `{{.ExtraPrompt}}`: 用户通过 LLM_PROMPT 环境变量提供的额外提示

**输出格式:** JSON
```json
{
  "intent": "用户意图描述",
  "command": "可执行命令字符串",
  "exists": true,
  "note": "补充说明（可选）"
}
```

### alternative.txt
替代命令查找的提示词，当原命令不存在时使用。

**可用变量:**
- `{{.Shell}}`: Shell 类型
- `{{.SystemInfo}}`: 系统信息描述
- `{{.Command}}`: 原始命令
- `{{.Error}}`: 错误信息
- `{{.Attempt}}`: 当前尝试次数

**输出格式:** 同 system.txt

### analyze.txt
命令输出分析的提示词，用于分析命令执行结果。

**可用变量:**
- `{{.Command}}`: 执行的命令
- `{{.Stdout}}`: 标准输出（自动截断至 4000 字符）
- `{{.Stderr}}`: 错误输出（自动截断至 1000 字符）

**输出格式:** 自然语言文本（3-5句）

## 修改提示词

1. 直接编辑对应的 `.txt` 文件
2. 使用 Go template 语法引用变量：`{{.VariableName}}`
3. 重新构建项目：`make build`
4. 新的提示词会在编译时嵌入到二进制文件中

## 注意事项

- 保持 JSON 输出格式的一致性（system.txt 和 alternative.txt）
- 避免在提示词中包含代码块标记（如 \`\`\`json），程序会自动清理
- 提示词应该清晰、简洁，避免歧义
- 测试修改后的提示词以确保 AI 能正确理解和响应
