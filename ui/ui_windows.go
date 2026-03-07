// +build windows

package ui

import (
	"syscall"
)

func init() {
	enableWindowsANSI()
}

// enableWindowsANSI 启用 Windows 控制台 ANSI 支持
func enableWindowsANSI() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	handle := syscall.Handle(syscall.Stdout)
	var mode uint32
	syscall.GetConsoleMode(handle, &mode)
	mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
	setConsoleMode.Call(uintptr(handle), uintptr(mode))
}
