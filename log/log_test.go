package log

import (
	"bytes"
	"testing"
	"unicode/utf8"
)

// TestLogWriterUTF8 校验统一 UTF-8 输出：
//   - 控制台保留 ANSI 颜色且为合法 UTF-8 字节；
//   - 文件去除 ANSI 颜色且为合法 UTF-8 字节；
//   - 汉字在两端都能正确还原。
func TestLogWriterUTF8(t *testing.T) {
	var console, file bytes.Buffer
	w := &logWriter{console: &console, file: &file}

	msg := "\033[32m[INFO]\033[0m 你好，世界 boardCode=88002552\n"
	if _, err := w.Write([]byte(msg)); err != nil {
		t.Fatalf("Write 失败: %v", err)
	}

	// 控制台应原样输出（UTF-8 + ANSI）
	if console.String() != msg {
		t.Fatalf("控制台输出应原样保留, 期望 %q 实际 %q", msg, console.String())
	}
	if !utf8.Valid(console.Bytes()) {
		t.Fatalf("控制台输出不是合法 UTF-8: %q", console.String())
	}

	// 文件应去除 ANSI 且为合法 UTF-8
	want := "[INFO] 你好，世界 boardCode=88002552\n"
	if file.String() != want {
		t.Fatalf("文件输出错误, 期望 %q 实际 %q", want, file.String())
	}
	if !utf8.Valid(file.Bytes()) {
		t.Fatalf("文件输出不是合法 UTF-8: %q", file.String())
	}
}

// TestChineseRoundTrip 校验含汉字、base64、URL 的混合内容在 UTF-8 下不被破坏。
func TestChineseRoundTrip(t *testing.T) {
	var console, file bytes.Buffer
	w := &logWriter{console: &console, file: &file}

	msg := "处理设备控制器/handleAct, url=http://iot.7iot.cn, token=AsIWf fEGIQ== 中文正常\n"
	if _, err := w.Write([]byte(msg)); err != nil {
		t.Fatalf("Write 失败: %v", err)
	}
	if console.String() != msg {
		t.Fatalf("控制台内容被破坏, 期望 %q 实际 %q", msg, console.String())
	}
	if file.String() != msg {
		t.Fatalf("文件内容被破坏, 期望 %q 实际 %q", msg, file.String())
	}
}

// TestSetupConsoleNoPanic 校验 setupConsole 在当前环境可安全调用（不 panic）。
func TestSetupConsoleNoPanic(t *testing.T) {
	setupConsole()
}
