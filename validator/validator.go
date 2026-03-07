package validator

import (
	"fmt"
	"smartCmd/llm"
	"smartCmd/shell"
	"smartCmd/ui"
)

const maxRetries = 5

// ValidateAndRetry 验证命令并在失败时重试
func ValidateAndRetry(client *llm.Client, userInput, systemInfo, shellType string) (*llm.CommandResult, error) {
	ui.PrintSpinning("理解用户意图...")

	// 第一次解析
	result, err := client.ParseCommand(userInput, systemInfo, shellType)
	if err != nil {
		ui.ClearLine()
		ui.PrintError(fmt.Sprintf("解析失败: %v", err))
		return nil, err
	}

	ui.ClearLine()
	ui.PrintSuccess(fmt.Sprintf("意图：%s", result.Intent))

	if result.Note != "" {
		ui.PrintStep("ℹ", ui.ColorCyan, result.Note)
	}

	// 验证命令是否存在
	triedCommands := []string{}
	currentResult := result

	for attempt := 1; attempt <= maxRetries; attempt++ {
		cmdName := shell.ExtractCommandName(currentResult.Command)
		if cmdName == "" {
			ui.PrintError("无法提取命令名称")
			return nil, fmt.Errorf("无法提取命令名称")
		}

		triedCommands = append(triedCommands, currentResult.Command)

		// 检查命令是否存在
		if shell.CommandExists(cmdName) {
			ui.PrintSuccess(fmt.Sprintf("命令：%s", currentResult.Command))
			return currentResult, nil
		}

		// 命令不存在，尝试获取替代命令
		if attempt >= maxRetries {
			ui.PrintError(fmt.Sprintf("找不到可用命令，已尝试 %d 次", maxRetries))
			ui.PrintError(fmt.Sprintf("尝试过的命令：%v", triedCommands))
			return nil, fmt.Errorf("找不到可用命令")
		}

		ui.PrintWarning(fmt.Sprintf("命令 '%s' 不存在，尝试第 %d 次替代...", cmdName, attempt))

		// 请求替代命令
		ui.PrintSpinning("查找替代命令...")
		alternativeResult, err := client.RequestAlternative(
			currentResult.Command,
			fmt.Sprintf("命令 '%s' 不存在", cmdName),
			systemInfo,
			shellType,
			attempt+1,
		)

		if err != nil {
			ui.ClearLine()
			ui.PrintError(fmt.Sprintf("获取替代命令失败: %v", err))
			return nil, err
		}

		ui.ClearLine()
		currentResult = alternativeResult
	}

	return nil, fmt.Errorf("超过最大重试次数")
}
