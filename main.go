package main

import (
	"fmt"
	"os"
	"smartCmd/config"
	initpkg "smartCmd/init"
	"smartCmd/llm"
	"smartCmd/shell"
	"smartCmd/sysinfo"
	"smartCmd/ui"
	"smartCmd/validator"
	"strings"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 解析命令行参数
	if len(os.Args) == 1 {
		// 无参数：初始化模式
		runInitMode(cfg)
		return
	}

	command := os.Args[1]
	switch command {
	case "parse":
		if len(os.Args) < 3 {
			ui.PrintError("用法: smartCmd parse <命令>")
			os.Exit(1)
		}
		userInput := strings.Join(os.Args[2:], " ")
		runParseMode(cfg, userInput)

	case "parseAndRun":
		if len(os.Args) < 3 {
			ui.PrintError("用法: smartCmd parseAndRun <命令>")
			os.Exit(1)
		}
		userInput := strings.Join(os.Args[2:], " ")
		runParseAndRunMode(cfg, userInput)

	default:
		ui.PrintError(fmt.Sprintf("未知命令: %s", command))
		ui.PrintError("可用命令: parse, parseAndRun")
		os.Exit(1)
	}
}

// runInitMode 初始化模式
func runInitMode(cfg *config.Config) {
	// 采集系统信息
	info, err := sysinfo.Collect()
	if err != nil {
		ui.PrintError(fmt.Sprintf("采集系统信息失败: %v", err))
		os.Exit(1)
	}

	// 验证 LLM 配置
	if cfg.IsValid() {
		ui.PrintSpinning("验证 LLM 配置...")
		client := llm.NewClient(cfg)
		response, err := client.TestConnection()
		ui.ClearLine()

		if err != nil {
			ui.PrintWarning(fmt.Sprintf("LLM 验证失败: %v", err))
			ui.PrintWarning("请检查环境变量 LLM_BASE_URL, LLM_API_KEY, LLM_MODEL")
		} else {
			ui.PrintSuccess(fmt.Sprintf("LLM 验证成功: %s", response))
		}
	} else {
		ui.PrintWarning("LLM 配置不完整，请设置环境变量：")
		ui.PrintWarning("  LLM_BASE_URL - API 基础 URL")
		ui.PrintWarning("  LLM_API_KEY  - API 密钥")
		ui.PrintWarning("  LLM_MODEL    - 模型名称")
	}

	// 生成并输出初始化脚本
	fmt.Println()
	fmt.Println("要在当前 shell 中启用 smartCmd，请执行：")
	fmt.Printf("  %s\n", initpkg.GetInitCommand(info))
	fmt.Println()
	script := initpkg.GenerateInitScript(info)
	fmt.Print(script)
}

// runParseMode 命令解析模式
func runParseMode(cfg *config.Config, userInput string) {
	if !cfg.IsValid() {
		ui.PrintError("LLM 配置不完整，请先运行 smartCmd 进行初始化")
		os.Exit(2)
	}

	// 获取系统信息
	info, err := sysinfo.Collect()
	if err != nil {
		ui.PrintError(fmt.Sprintf("采集系统信息失败: %v", err))
		os.Exit(1)
	}

	// 创建 LLM 客户端
	client := llm.NewClient(cfg)

	// 验证并获取命令
	result, err := validator.ValidateAndRetry(client, userInput, info.ToReadableString(), info.Shell)
	if err != nil {
		os.Exit(3)
	}

	// 输出最终命令
	fmt.Println()
	fmt.Println(result.Command)
}

// runParseAndRunMode 解析并执行模式
func runParseAndRunMode(cfg *config.Config, userInput string) {
	if !cfg.IsValid() {
		ui.PrintError("LLM 配置不完整，请先运行 smartCmd 进行初始化")
		os.Exit(2)
	}

	// 获取系统信息
	info, err := sysinfo.Collect()
	if err != nil {
		ui.PrintError(fmt.Sprintf("采集系统信息失败: %v", err))
		os.Exit(1)
	}

	// 创建 LLM 客户端
	client := llm.NewClient(cfg)

	// 验证并获取命令
	result, err := validator.ValidateAndRetry(client, userInput, info.ToReadableString(), info.Shell)
	if err != nil {
		os.Exit(3)
	}

	// 检查危险命令
	if shell.IsDangerousCommand(result.Command) {
		if !shell.ConfirmDangerousCommand(result.Command) {
			ui.PrintWarning("用户取消执行")
			os.Exit(0)
		}
	}

	// 执行命令
	fmt.Println()
	ui.PrintExecute(fmt.Sprintf("正在执行：%s", result.Command))
	ui.PrintSeparator("原始输出")

	stdout, stderr, execErr := shell.ExecuteCommand(result.Command, info.Shell)

	// 打印原始输出
	if stdout != "" {
		fmt.Print(stdout)
	}
	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}

	ui.PrintSeparator("")

	// 分析执行结果
	ui.PrintSpinning("AI 分析结果...")
	analysis, err := client.AnalyzeOutput(result.Command, stdout, stderr)
	ui.ClearLine()

	if err != nil {
		ui.PrintWarning(fmt.Sprintf("分析失败: %v", err))
	} else {
		ui.PrintSuccess("分析：")
		fmt.Println(analysis)
	}

	// 如果命令执行失败，返回非零退出码
	if execErr != nil {
		os.Exit(4)
	}
}
