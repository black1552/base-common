//go:build windows

package log

import "syscall"

// setupConsole 将 Windows 控制台的输入/输出代码页设置为 UTF-8(65001)。
// 日志统一以 UTF-8 字节输出，配合此设置可让真实控制台（cmd/PowerShell）正常显示汉字；
// 若 stdout 被重定向到管道/文件（如 GoLand 运行窗口），调用会静默失败，不影响 UTF-8 输出。
func setupConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleCP := kernel32.NewProc("SetConsoleCP")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	const cpUTF8 = 65001
	setConsoleCP.Call(cpUTF8)
	setConsoleOutputCP.Call(cpUTF8)
}
