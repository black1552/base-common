//go:build windows

package log

import (
	"syscall"
	"unsafe"

	"golang.org/x/text/encoding"
)

const (
	stdOutputHandle    = ^uintptr(0) - 10 // STD_OUTPUT_HANDLE = -11
	invalidHandleValue = ^uintptr(0)      // INVALID_HANDLE_VALUE
)

// isConsole 判断标准输出是否连接到真实的 Windows 控制台。
// 若被重定向（IDE、管道、文件等），GetConsoleMode 会失败并返回 false。
func isConsole() bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")

	handle, _, _ := getStdHandle.Call(stdOutputHandle)
	if handle == 0 || handle == invalidHandleValue {
		return false
	}

	var mode uint32
	ret, _, _ := getConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	return ret != 0
}

// getConsoleOutputCP 获取当前控制台输出代码页。
func getConsoleOutputCP() uint32 {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetConsoleOutputCP")
	ret, _, _ := proc.Call()
	return uint32(ret)
}

// systemConsoleEncoding 使日志的控制台字节编码追随系统：
//   - 真实控制台：按当前代码页转码（如中文环境的 GBK），避免乱码；
//   - 被重定向（IDE/管道/文件）：输出 UTF-8，因为接收方普遍按 UTF-8 解析。
func systemConsoleEncoding() encoding.Encoding {
	if !isConsole() {
		return nil
	}
	return encodingForCodePage(getConsoleOutputCP())
}
