package ui

import (
	"fmt"
	"time"
)

// 颜色代码
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
)

// 符号
const (
	SymbolSpinner  = "⟳"
	SymbolSuccess  = "✓"
	SymbolWarning  = "⚠"
	SymbolError    = "✗"
	SymbolExecute  = "▶"
)

// PrintStep 打印步骤信息
func PrintStep(symbol, color, message string) {
	fmt.Printf("%s%s%s %s\n", color, symbol, ColorReset, message)
}

// PrintSpinning 打印进行中的步骤
func PrintSpinning(message string) {
	fmt.Printf("%s%s%s %s", ColorYellow, SymbolSpinner, ColorReset, message)
}

// PrintSuccess 打印成功信息
func PrintSuccess(message string) {
	PrintStep(SymbolSuccess, ColorGreen, message)
}

// PrintWarning 打印警告信息
func PrintWarning(message string) {
	PrintStep(SymbolWarning, ColorYellow, message)
}

// PrintError 打印错误信息
func PrintError(message string) {
	PrintStep(SymbolError, ColorRed, message)
}

// PrintExecute 打印执行信息
func PrintExecute(message string) {
	PrintStep(SymbolExecute, ColorCyan, message)
}

// Spinner 等待动画
type Spinner struct {
	stop chan struct{}
}

// NewSpinner 创建新的等待动画
func NewSpinner() *Spinner {
	return &Spinner{
		stop: make(chan struct{}),
	}
}

// Start 启动等待动画
func (s *Spinner) Start() {
	go func() {
		frames := []string{".", "..", "...", "   "}
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K") // 清除当前行
				return
			case <-time.After(400 * time.Millisecond):
				fmt.Printf("\r%s", frames[i%len(frames)])
				i++
			}
		}
	}()
}

// Stop 停止等待动画
func (s *Spinner) Stop() {
	close(s.stop)
	time.Sleep(50 * time.Millisecond) // 等待清除完成
}

// ClearLine 清除当前行
func ClearLine() {
	fmt.Print("\r\033[K")
}

// PrintSeparator 打印分隔线
func PrintSeparator(title string) {
	if title == "" {
		fmt.Println("────────────────────────────────────────")
	} else {
		fmt.Printf("─────────────── %s ───────────────\n", title)
	}
}

// PrintRetry 打印重试信息
func PrintRetry(message string) {
	PrintStep("↻", ColorYellow, message)
}

// PrintScriptStep 打印脚本执行进度
func PrintScriptStep(stepNum, totalSteps int, command string) {
	fmt.Printf("%s[%d/%d]%s %s\n", ColorCyan, stepNum, totalSteps, ColorReset, command)
}
