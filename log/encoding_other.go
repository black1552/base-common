//go:build !windows

package log

// setupConsole 在非 Windows 平台（Linux/macOS 等）无需任何处理，
// 终端默认使用 UTF-8，日志字节直接以 UTF-8 输出即可正常显示汉字。
func setupConsole() {}
