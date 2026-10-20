//go:build windows

package log

import (
	"syscall"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// systemConsoleEncoding 探测当前 Windows 控制台的输出代码页，
// 返回对应的编码，使日志字节编码追随系统，避免中文乱码。
// 返回 nil 表示无需转码（UTF-8 直出）。
func systemConsoleEncoding() encoding.Encoding {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetConsoleOutputCP")
	codePage, _, _ := proc.Call()

	switch uint32(codePage) {
	case 65001: // UTF-8
		return nil
	case 936: // 简体中文 GBK
		return simplifiedchinese.GBK
	case 54936: // GB18030
		return simplifiedchinese.GB18030
	case 20936: // 简体中文 HZ-GB2312
		return simplifiedchinese.HZGB2312
	case 950, 951: // 繁体中文 Big5
		return traditionalchinese.Big5
	case 1252: // Windows 西欧
		return charmap.Windows1252
	case 1251: // Windows 西里尔
		return charmap.Windows1251
	case 437: // DOS 美国
		return charmap.CodePage437
	case 850: // DOS  Multilingual
		return charmap.CodePage850
	default:
		// 未知代码页时按 UTF-8 处理
		return nil
	}
}
