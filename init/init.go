package init

import (
	"fmt"
	"smartCmd/sysinfo"
)

// GenerateInitScript 生成初始化脚本
func GenerateInitScript(info *sysinfo.SystemInfo) string {
	switch info.Shell {
	case "powershell":
		return generatePowerShellScript(info)
	case "fish":
		return generateFishScript(info)
	default:
		return generatePosixScript(info)
	}
}

// GetInitCommand 获取初始化命令提示
func GetInitCommand(info *sysinfo.SystemInfo) string {
	switch info.Shell {
	case "powershell":
		return fmt.Sprintf("iex (& '%s')", info.BinPath)
	case "fish":
		return fmt.Sprintf("eval (%s)", info.BinPath)
	default:
		return fmt.Sprintf("eval $(%s)", info.BinPath)
	}
}

// generatePowerShellScript 生成 PowerShell 初始化脚本
func generatePowerShellScript(info *sysinfo.SystemInfo) string {
	return fmt.Sprintf(`# smartCmd init output for PowerShell
function global:sc { & '%s' parse $args }
function global:scc { & '%s' parseAndRun $args }
$env:_smartCmd = '%s'
`, info.BinPath, info.BinPath, info.ToEnvString())
}

// generateFishScript 生成 Fish Shell 初始化脚本
func generateFishScript(info *sysinfo.SystemInfo) string {
	return fmt.Sprintf(`# smartCmd init output for fish
function sc; %s parse $argv; end
function scc; %s parseAndRun $argv; end
set -x _smartCmd '%s'
`, info.BinPath, info.BinPath, info.ToEnvString())
}

// generatePosixScript 生成 POSIX 兼容脚本 (bash/zsh/sh)
func generatePosixScript(info *sysinfo.SystemInfo) string {
	return fmt.Sprintf(`# smartCmd init output for %s
alias sc='%s parse'
alias scc='%s parseAndRun'
export _smartCmd='%s'
`, info.Shell, info.BinPath, info.BinPath, info.ToEnvString())
}
