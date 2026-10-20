package log

import (
	"bytes"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// decodeWith 用给定编码把字节解码回 UTF-8 字符串；enc 为 nil 时视为 UTF-8 原样返回。
func decodeWith(t *testing.T, enc encoding.Encoding, b []byte) string {
	t.Helper()
	if enc == nil {
		return string(b)
	}
	out, err := enc.NewDecoder().Bytes(b)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	return string(out)
}

// TestEncodingForCodePage 校验代码页映射：常见编码可识别，UTF-8 及未知代码页直出。
func TestEncodingForCodePage(t *testing.T) {
	if got := encodingForCodePage(65001); got != nil {
		t.Fatalf("UTF-8(65001) 应返回 nil，实际非 nil")
	}
	if got := encodingForCodePage(12345); got != nil {
		t.Fatalf("未知代码页应返回 nil，实际非 nil")
	}
	if got := encodingForCodePage(936); got != simplifiedchinese.GBK {
		t.Fatalf("CP936 应映射为 GBK")
	}
}

// TestLogWriterUTF8 校验 enc 为 nil 时，控制台与文件均以 UTF-8 输出，且文件去除颜色码。
func TestLogWriterUTF8(t *testing.T) {
	var console, file bytes.Buffer
	w := &logWriter{console: &console, file: &file, enc: nil}

	msg := "\033[31m[ERROR]\033[0m 中文错误信息\n"
	if _, err := w.Write([]byte(msg)); err != nil {
		t.Fatalf("Write 失败: %v", err)
	}

	// 控制台保留 ANSI 且为 UTF-8 原样字节
	if console.String() != msg {
		t.Fatalf("控制台输出应原样保留, 期望 %q 实际 %q", msg, console.String())
	}
	// 文件应为去除 ANSI 的 UTF-8 文本
	want := "[ERROR] 中文错误信息\n"
	if file.String() != want {
		t.Fatalf("文件输出编码错误, 期望 %q 实际 %q", want, file.String())
	}
}

// TestLogWriterConsoleGBK 校验 Windows 控制台场景：控制台字节编码为系统 GBK，
// 用 GBK 解码可还原出正确文本；文件仍保存 UTF-8。
func TestLogWriterConsoleGBK(t *testing.T) {
	enc := encodingForCodePage(936) // GBK
	if enc == nil {
		t.Skip("当前构建未包含 GBK 编码")
	}

	var console, file bytes.Buffer
	w := &logWriter{console: &console, file: &file, enc: enc}

	msg := "\033[32m[INFO]\033[0m 你好，世界\n"
	if _, err := w.Write([]byte(msg)); err != nil {
		t.Fatalf("Write 失败: %v", err)
	}

	// 控制台字节必须可用 GBK 正确解码回原始文本（关键：追随系统编码）
	if got := decodeWith(t, enc, console.Bytes()); got != msg {
		t.Fatalf("GBK 控制台往返解码失败, 期望 %q 实际 %q", msg, got)
	}
	// 控制台不应是 UTF-8 字节（GBK 编码与 UTF-8 不同）
	if bytes.Equal(console.Bytes(), []byte(msg)) {
		t.Fatalf("CP936 场景控制台应转码为 GBK, 不应是 UTF-8 原字节")
	}

	// 文件必须是 UTF-8 且去除 ANSI
	want := "[INFO] 你好，世界\n"
	if file.String() != want {
		t.Fatalf("文件应为 UTF-8 去色文本, 期望 %q 实际 %q", want, file.String())
	}
}

// TestSystemConsoleEncodingRedirected 在 go test 环境下 stdout 被重定向（非真实控制台），
// 应返回 nil 即以 UTF-8 输出，保证 IDE / 管道 / 文件场景正常显示。
func TestSystemConsoleEncodingRedirected(t *testing.T) {
	if got := systemConsoleEncoding(); got != nil {
		t.Fatalf("重定向环境下应使用 UTF-8(nil), 实际返回了编码: %T", got)
	}
}
