package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// CommandExists 检查命令是否存在
func CommandExists(name string) bool {
	// 处理内置命令
	if isBuiltinCommand(name) {
		return true
	}

	_, err := exec.LookPath(name)
	return err == nil
}

// isBuiltinCommand 检查是否为 Shell 内置命令
func isBuiltinCommand(name string) bool {
	builtins := []string{
		"cd", "echo", "pwd", "exit", "export", "set", "unset",
		"alias", "source", ".", "eval", "exec", "read", "test",
		"[", "[[", "type", "command", "builtin", "hash",
	}

	for _, builtin := range builtins {
		if name == builtin {
			return true
		}
	}

	// Windows PowerShell 内置命令
	if runtime.GOOS == "windows" {
		psBuiltins := []string{
			"Get-Command", "Set-Location", "Get-ChildItem", "Write-Host",
			"Get-Content", "Set-Content", "Copy-Item", "Remove-Item",
		}
		for _, builtin := range psBuiltins {
			if strings.EqualFold(name, builtin) {
				return true
			}
		}
	}

	return false
}

// ExtractCommandName 从命令字符串中提取主命令名
func ExtractCommandName(cmdStr string) string {
	// 移除前导空格
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return ""
	}

	// 处理管道、重定向等
	cmdStr = strings.Split(cmdStr, "|")[0]
	cmdStr = strings.Split(cmdStr, ">")[0]
	cmdStr = strings.Split(cmdStr, "<")[0]
	cmdStr = strings.TrimSpace(cmdStr)

	// 提取第一个单词
	fields := strings.Fields(cmdStr)
	if len(fields) == 0 {
		return ""
	}

	return fields[0]
}

// ExecuteCommand 执行命令并返回输出
func ExecuteCommand(cmdStr string, shellType string) (stdout, stderr string, err error) {
	var cmd *exec.Cmd

	switch shellType {
	case "powershell":
		cmd = exec.Command("powershell", "-Command", cmdStr)
	case "bash":
		cmd = exec.Command("bash", "-c", cmdStr)
	case "zsh":
		cmd = exec.Command("zsh", "-c", cmdStr)
	case "fish":
		cmd = exec.Command("fish", "-c", cmdStr)
	default:
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	return
}

// ExecuteCommandStreaming 执行命令并实时输出
func ExecuteCommandStreaming(cmdStr string, shellType string) error {
	var cmd *exec.Cmd

	switch shellType {
	case "powershell":
		cmd = exec.Command("powershell", "-Command", cmdStr)
	case "bash":
		cmd = exec.Command("bash", "-c", cmdStr)
	case "zsh":
		cmd = exec.Command("zsh", "-c", cmdStr)
	case "fish":
		cmd = exec.Command("fish", "-c", cmdStr)
	default:
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// IsDangerousCommand 检查是否为危险命令
func IsDangerousCommand(cmdStr string) bool {
	dangerous := []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=",
		"mkfs",
		"format",
		":(){:|:&};:",  // fork bomb
		"> /dev/sda",
		"mv / ",
	}

	cmdLower := strings.ToLower(cmdStr)
	for _, pattern := range dangerous {
		if strings.Contains(cmdLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// ConfirmDangerousCommand 确认危险命令执行
func ConfirmDangerousCommand(cmdStr string) bool {
	fmt.Printf("\n\033[31m⚠ 警告：检测到潜在危险命令\033[0m\n")
	fmt.Printf("命令：%s\n", cmdStr)
	fmt.Print("确认执行？(yes/no): ")

	var response string
	fmt.Scanln(&response)

	return strings.ToLower(response) == "yes"
}
