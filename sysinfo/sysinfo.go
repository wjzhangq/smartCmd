package sysinfo

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SystemInfo 存储系统信息
type SystemInfo struct {
	OS         string
	OSVersion  string
	Shell      string
	ShellVer   string
	BinPath    string
	Username   string
}

// Collect 采集系统信息
func Collect() (*SystemInfo, error) {
	info := &SystemInfo{
		OS: runtime.GOOS,
	}

	// 获取 OS 版本
	info.OSVersion = getOSVersion()

	// 获取 Shell 类型和版本
	info.Shell, info.ShellVer = getShellInfo()

	// 获取可执行文件路径
	binPath, err := os.Executable()
	if err != nil {
		binPath = "smartCmd"
	}
	info.BinPath = binPath

	// 获取用户名
	if runtime.GOOS == "windows" {
		info.Username = os.Getenv("USERNAME")
	} else {
		info.Username = os.Getenv("USER")
	}

	return info, nil
}

// getOSVersion 获取操作系统版本
func getOSVersion() string {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("cmd", "/c", "ver").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
		return "Windows"
	case "darwin":
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err == nil {
			return "macOS " + strings.TrimSpace(string(out))
		}
		return "macOS"
	default: // linux
		// 尝试读取 /etc/os-release
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					ver := strings.TrimPrefix(line, "PRETTY_NAME=")
					ver = strings.Trim(ver, "\"")
					return ver
				}
			}
		}
		// 回退到 uname
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			return "Linux " + strings.TrimSpace(string(out))
		}
		return "Linux"
	}
}

// getShellInfo 获取 Shell 类型和版本
func getShellInfo() (string, string) {
	// Windows 默认 PowerShell
	if runtime.GOOS == "windows" {
		out, err := exec.Command("powershell", "-Command", "$PSVersionTable.PSVersion.ToString()").Output()
		if err == nil {
			return "powershell", strings.TrimSpace(string(out))
		}
		return "powershell", ""
	}

	// 检测 ZSH
	if zshVer := os.Getenv("ZSH_VERSION"); zshVer != "" {
		return "zsh", zshVer
	}

	// 检测 Bash
	if bashVer := os.Getenv("BASH_VERSION"); bashVer != "" {
		return "bash", bashVer
	}

	// 检测 Fish
	if shell := os.Getenv("SHELL"); strings.Contains(shell, "fish") {
		out, err := exec.Command("fish", "--version").Output()
		if err == nil {
			parts := strings.Fields(string(out))
			if len(parts) >= 3 {
				return "fish", parts[2]
			}
		}
		return "fish", ""
	}

	// 默认 sh
	return "sh", ""
}

// ToEnvString 转换为环境变量格式字符串
func (s *SystemInfo) ToEnvString() string {
	return fmt.Sprintf("OS=%s;VER=%s;SHELL=%s;SHELL_VER=%s;BIN=%s",
		s.OS, s.OSVersion, s.Shell, s.ShellVer, s.BinPath)
}

// ToReadableString 转换为可读字符串
func (s *SystemInfo) ToReadableString() string {
	return fmt.Sprintf("当前系统：%s (%s)，Shell: %s %s，smartCmd 路径: %s",
		s.OS, s.OSVersion, s.Shell, s.ShellVer, s.BinPath)
}
