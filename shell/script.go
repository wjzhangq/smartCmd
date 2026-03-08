package shell

import (
	"fmt"
	"os"
	"runtime"
	"smartCmd/llm"
	"smartCmd/ui"
	"strings"
)

// ParseScript 解析脚本文件为命令行数组
func ParseScript(filepath string) ([]string, error) {
	// 读取文件内容
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("读取脚本文件失败: %w", err)
	}

	// 按行分割
	lines := strings.Split(string(content), "\n")
	var commands []string

	for _, line := range lines {
		// 去除首尾空白
		trimmed := strings.TrimSpace(line)

		// 跳过空行
		if trimmed == "" {
			continue
		}

		// 跳过注释行（以 # 开头）
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 保留有效命令
		commands = append(commands, trimmed)
	}

	return commands, nil
}

// ValidateScriptFile 验证脚本文件
func ValidateScriptFile(filepath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("脚本文件不存在: %s", filepath)
	}

	// 检查文件扩展名
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(strings.ToLower(filepath), ".ps1") {
			return fmt.Errorf("Windows 系统请使用 .ps1 脚本文件")
		}
	} else {
		if !strings.HasSuffix(filepath, ".sh") {
			return fmt.Errorf("Linux/macOS 系统请使用 .sh 脚本文件")
		}
	}

	return nil
}

// CheckDangerousCommands 检查脚本中的危险命令
func CheckDangerousCommands(commands []string) ([]string, bool) {
	var dangerous []string

	for _, cmd := range commands {
		if IsDangerousCommand(cmd) {
			dangerous = append(dangerous, cmd)
		}
	}

	return dangerous, len(dangerous) > 0
}

// ExecuteScriptWithRetry 执行脚本并自动修复错误
func ExecuteScriptWithRetry(commands []string, shellType string, llmClient *llm.Client, sysInfo string) error {
	totalSteps := len(commands)

	for i, originalCmd := range commands {
		stepNum := i + 1
		currentCmd := originalCmd
		retryCount := 0
		maxRetries := 5

		for {
			// 显示执行进度
			ui.PrintScriptStep(stepNum, totalSteps, currentCmd)

			// 执行命令
			stdout, stderr, err := ExecuteCommand(currentCmd, shellType)

			// 获取退出码
			exitCode := 0
			if err != nil {
				exitCode = 1 // 简化处理，实际可以从 err 中提取具体退出码
			}

			// 打印输出
			if stdout != "" {
				fmt.Print(stdout)
			}
			if stderr != "" {
				fmt.Fprint(os.Stderr, stderr)
			}

			// 检查是否成功
			if err == nil {
				ui.PrintSuccess(fmt.Sprintf("步骤 %d/%d 执行成功", stepNum, totalSteps))
				break // 成功，继续下一条命令
			}

			// 执行失败
			ui.PrintError(fmt.Sprintf("步骤 %d/%d 执行失败 (退出码: %d)", stepNum, totalSteps, exitCode))

			// 检查重试次数
			retryCount++
			if retryCount > maxRetries {
				return fmt.Errorf("步骤 %d/%d 超过最大重试次数 (%d 次)，终止执行\n失败命令: %s",
					stepNum, totalSteps, maxRetries, currentCmd)
			}

			// 调用 LLM 修复命令
			ui.PrintSpinning(fmt.Sprintf("AI 分析错误并修复 (第 %d 次重试)...", retryCount))
			fixedCmd, fixErr := llmClient.FixCommand(currentCmd, stdout, stderr, sysInfo, shellType, exitCode)
			ui.ClearLine()

			if fixErr != nil {
				ui.PrintWarning(fmt.Sprintf("AI 修复失败: %v", fixErr))
				// LLM 调用失败也计入重试次数，继续用原命令重试
				continue
			}

			// 显示修复建议
			if fixedCmd != currentCmd {
				ui.PrintRetry(fmt.Sprintf("AI 修复建议: %s", fixedCmd))
				currentCmd = fixedCmd
			} else {
				ui.PrintWarning("AI 无法提供有效修复，将重试原命令")
			}

			// 继续循环，重试修复后的命令
		}
	}

	return nil
}
