//go:build windows

package log

import (
	"io"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// setupConsole 保留为空实现：
// 不再强制把控制台代码页切到 UTF-8（SetConsoleOutputCP(65001)），
// 因为在中文 Windows 的 cmd/conhost 上强切 UTF-8 常因字体/回退导致汉字乱码。
// 改为由 newConsoleWriter 按控制台“当前真实代码页”对输出做转码，更可靠。
func setupConsole() {}

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle       = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode     = kernel32.NewProc("GetConsoleMode")
	procGetConsoleOutputCP = kernel32.NewProc("GetConsoleOutputCP")
)

// stdOutputHandle 为 (DWORD)STD_OUTPUT_HANDLE，即 -11 的无符号表示 0xFFFFFFF5。
const stdOutputHandle = uintptr(0xFFFFFFF5)

// consoleOutputCP 返回标准输出所依附控制台的代码页。
// 若 stdout 被重定向到管道/文件（如 GoLand 运行窗口、`> file`），
// GetConsoleMode 会失败，此时返回 (0, false)，表示不应做转码。
func consoleOutputCP() (uint32, bool) {
	handle, _, _ := procGetStdHandle.Call(stdOutputHandle)
	if handle == 0 || handle == ^uintptr(0) { // 无效句柄
		return 0, false
	}
	var mode uint32
	if ret, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode))); ret == 0 {
		return 0, false // 不是真正的控制台
	}
	cp, _, _ := procGetConsoleOutputCP.Call()
	return uint32(cp), true
}

// encodingForCP 将常见中文 Windows 控制台代码页映射为对应编码。
// UTF-8(65001)、纯 ASCII 或其它未知代码页返回 nil，表示直接输出 UTF-8 无需转码。
func encodingForCP(cp uint32) encoding.Encoding {
	switch cp {
	case 936: // GBK / CP936，简体中文 Windows 默认控制台代码页
		return simplifiedchinese.GBK
	case 54936: // GB18030
		return simplifiedchinese.GB18030
	default:
		return nil
	}
}

// consoleEncoder 把 UTF-8 字节转码为目标代码页后再写给真正的 Windows 控制台。
type consoleEncoder struct {
	inner io.Writer
	enc   encoding.Encoding
}

func (w *consoleEncoder) Write(p []byte) (int, error) {
	dst, _, err := transform.Bytes(w.enc.NewEncoder(), p)
	if err != nil {
		dst = p // 转码失败则退回原始 UTF-8 字节，绝不丢日志
	}
	if _, err = w.inner.Write(dst); err != nil {
		return 0, err
	}
	return len(p), nil
}

// newConsoleWriter 依据真实控制台代码页决定是否包装 stdout 做转码：
//   - 真正的控制台且代码页为 GBK/GB18030 -> 转码输出，修复 cmd/PowerShell 汉字乱码；
//   - 控制台已是 UTF-8(65001) 或其它 -> 原样输出 UTF-8；
//   - stdout 被重定向（IDE 运行窗口 / 管道 / 文件）-> 原样输出 UTF-8。
func newConsoleWriter(stdout *os.File) io.Writer {
	cp, ok := consoleOutputCP()
	if !ok {
		return stdout
	}
	if enc := encodingForCP(cp); enc != nil {
		return &consoleEncoder{inner: stdout, enc: enc}
	}
	return stdout
}
