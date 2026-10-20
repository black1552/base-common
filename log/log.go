package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"golang.org/x/text/encoding"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logPath     string
	sysLog      *log.Logger
	filePath    string
	currentDate string // 当前日志文件对应的日期
	fileLogger  *lumberjack.Logger
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
)

// 正则表达式匹配 ANSI 颜色码
var ansiColorRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripAnsiColors 去除字符串中的 ANSI 颜色码
func stripAnsiColors(s string) string {
	return ansiColorRegex.ReplaceAllString(s, "")
}

// logWriter 自定义 writer，用于分别处理控制台和文件输出
type logWriter struct {
	console io.Writer
	file    io.Writer
	enc     encoding.Encoding // 控制台输出编码，追随系统；nil 表示 UTF-8 直出
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	// 控制台输出保留颜色，并按系统编码转码（如 Windows 中文控制台的 GBK）
	consoleBytes := p
	if w.enc != nil {
		if dst, e := w.enc.NewEncoder().Bytes(p); e == nil {
			consoleBytes = dst
		}
	}
	_, err = w.console.Write(consoleBytes)
	if err != nil {
		return 0, err
	}
	// 文件输出保持 UTF-8 并去除颜色码
	colorless := stripAnsiColors(string(p))
	_, err = w.file.Write([]byte(colorless))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// cleanOldLogs 删除指定天数之前的日志文件（包括主文件和备份文件）
func cleanOldLogs(days int) {
	if !gfile.Exists(logPath) {
		return
	}

	// 获取所有日志文件
	files, err := gfile.DirNames(logPath)
	if err != nil {
		return
	}

	now := time.Now()
	for _, file := range files {
		path := filepath.Join(logPath, file)
		if gfile.IsDir(path) {
			continue
		}

		var dateStr string
		var matched bool

		// 匹配主日志文件格式：log-YYYY-MM-DD.log
		if strings.HasPrefix(file, "log-") && strings.HasSuffix(file, ".log") {
			// 检查是否是主文件（没有备份时间戳）
			// 主文件格式：log-2026-04-25.log
			// 备份文件格式：log-2026-04-25-2026-04-25T10-30-45.123.log
			parts := strings.Split(strings.TrimSuffix(file, ".log"), "-")
			if len(parts) == 4 {
				// 主文件：log-YYYY-MM-DD
				dateStr = parts[1] + "-" + parts[2] + "-" + parts[3]
				matched = true
			} else if len(parts) > 4 {
				// 备份文件：log-YYYY-MM-DD-YYYY-MM-DDTHH-MM-SS.mmm
				// 提取主日期部分（第一个日期）
				dateStr = parts[1] + "-" + parts[2] + "-" + parts[3]
				matched = true
			}
		}

		if !matched {
			continue
		}

		// 解析日期
		fileTime, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue // 日期格式不正确，跳过
		}

		// 计算文件年龄
		tage := now.Sub(fileTime)
		if tage.Hours() > float64(days*24) {
			// 超过指定天数，删除文件
			err = os.Remove(path)
			if err == nil {
				Info(fmt.Sprintf("已删除过期日志文件：%s", file))
			}
		}
	}
}

// checkAndRotateLogFile 检查是否需要切换日志文件（跨天时）
func checkAndRotateLogFile() {
	date := gtime.Date()
	if currentDate != date {
		// 日期变化，需要重新初始化
		currentDate = date
		filePath = gfile.Join(logPath, fmt.Sprintf("log-%s.log", currentDate))
		fileLogger = &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    2,     // 单个文件最大 10MB
			MaxBackups: 5,     // 最多保留 5 个备份
			MaxAge:     30,    // 保留 30 天
			Compress:   false, // 启用压缩
		}
		// 创建新的 writer
		multiWriter := &logWriter{
			console: os.Stdout,
			file:    fileLogger,
			enc:     systemConsoleEncoding(),
		}
		sysLog = log.New(multiWriter, "", 0)

		// 清理 30 天前的旧日志
		cleanOldLogs(30)
	}
}

func Init() {
	if sysLog != nil {
		checkAndRotateLogFile() // 检查是否需要切换文件
		return
	}
	logPath = gfile.Join(gfile.Pwd(), "logs")
	currentDate = gtime.Date()
	filePath = gfile.Join(logPath, fmt.Sprintf("log-%s.log", currentDate))
	fileLogger = &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    2,     // 单个文件最大 10MB
		MaxBackups: 5,     // 最多保留 5 个备份
		MaxAge:     30,    // 保留 30 天
		Compress:   false, // 启用压缩
	}
	// 使用自定义 writer 实现控制台带颜色、文件无颜色的输出
	multiWriter := &logWriter{
		console: os.Stdout,
		file:    fileLogger,
		enc:     systemConsoleEncoding(),
	}
	sysLog = log.New(multiWriter, "", 0)

	// 启动时清理 30 天前的旧日志
	cleanOldLogs(30)
}

func Info(v ...any) {
	Init()
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[INFO]%s ", time.Now().Format("2006-01-02 15:04:05"), Green, Reset))
	sysLog.Println(fmt.Sprint(v...))
}
func Error(v ...any) error {
	Init()
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[ERROR]%s ", time.Now().Format("2006-01-02 15:04:05"), Red, Reset))
	msg := fmt.Sprint(v...)
	sysLog.Println(msg, strings.TrimSpace(string(debug.Stack())))
	return fmt.Errorf(msg)
}
func Warn(v ...any) {
	Init()
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[WARN]%s ", time.Now().Format("2006-01-02 15:04:05"), Yellow, Reset))
	sysLog.Println(fmt.Sprint(v...))
}
func Debug(v ...any) {
	Init()
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[DEBUG]%s ", time.Now().Format("2006-01-02 15:04:05"), Blue, Reset))
	sysLog.Println(fmt.Sprint(v...))
}
func Infof(format string, v ...any) {
	Init()
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[INFO]%s ", time.Now().Format("2006-01-02 15:04:05"), Green, Reset))
	sysLog.Println(fmt.Sprintf(format, v...))
}
func Errorf(format string, v ...any) error {
	Init()
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[ERROR]%s ", time.Now().Format("2006-01-02 15:04:05"), Red, Reset))
	sysLog.Println(fmt.Sprintf(format, v...), strings.TrimSpace(string(debug.Stack())))
	return fmt.Errorf(format, v...)
}
func Warnf(format string, v ...any) {
	Init()
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[WARN]%s ", time.Now().Format("2006-01-02 15:04:05"), Yellow, Reset))
	sysLog.Println(fmt.Sprintf(format, v...))
}
func Debugf(format string, v ...any) {
	Init()
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[DEBUG]%s ", time.Now().Format("2006-01-02 15:04:05"), Blue, Reset))
	sysLog.Println(fmt.Sprintf(format, v...))
}
