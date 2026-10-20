//go:build !windows

package log

import "golang.org/x/text/encoding"

// systemConsoleEncoding 在非 Windows 平台（Linux/macOS 等）终端默认使用 UTF-8，
// 返回 nil 表示无需转码，日志字节直接以 UTF-8 输出。
func systemConsoleEncoding() encoding.Encoding {
	return nil
}
