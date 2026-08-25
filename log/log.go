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
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"gopkg.in/natefinch/lumberjack.v2"
)

type sLogger struct {
	logPath     string
	sysLog      *log.Logger
	filePath    string
	currentDate string // 当前日志文件对应的日期
	fileLogger  *lumberjack.Logger
}

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
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	// 控制台输出保留颜色
	_, err = w.console.Write(p)
	if err != nil {
		return 0, err
	}
	// 文件输出去除颜色码
	colorless := stripAnsiColors(string(p))
	_, err = w.file.Write([]byte(colorless))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

var (
	logs   *sLogger
	initMu sync.Mutex // 用于保护初始化过程的互斥锁
)

// cleanOldLogs 删除指定天数之前的日志文件（包括主文件和备份文件）
func cleanOldLogs(days int) {
	if !gfile.Exists(logs.logPath) {
		return
	}

	// 获取所有日志文件
	files, err := gfile.DirNames(logs.logPath)
	if err != nil {
		return
	}

	now := time.Now()
	for _, file := range files {
		path := filepath.Join(logs.logPath, file)
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
	if logs.currentDate != date {
		// 日期变化，需要重新初始化
		logs.currentDate = date
		logs.filePath = gfile.Join(logs.logPath, fmt.Sprintf("log-%s.log", logs.currentDate))
		logs.fileLogger = &lumberjack.Logger{
			Filename:   logs.filePath,
			MaxSize:    2,     // 单个文件最大 10MB
			MaxBackups: 5,     // 最多保留 5 个备份
			MaxAge:     30,    // 保留 30 天
			Compress:   false, // 启用压缩
		}
		// 创建新的 writer
		multiWriter := &logWriter{
			console: os.Stdout,
			file:    logs.fileLogger,
		}
		logs.sysLog = log.New(multiWriter, "", 0)

		// 清理 30 天前的旧日志
		cleanOldLogs(30)
	}
}

func Init() {
	// 如果已经初始化，检查是否需要切换文件
	if logs != nil && logs.sysLog != nil {
		checkAndRotateLogFile()
		return
	}

	// 加锁确保线程安全
	initMu.Lock()
	defer initMu.Unlock()

	// 双重检查，避免重复初始化
	if logs != nil && logs.sysLog != nil {
		return
	}

	// 初始化日志器
	currentDate := gtime.Date()
	logPath := gfile.Join(gfile.Pwd(), "logs")
	filePath := gfile.Join(logPath, fmt.Sprintf("log-%s.log", currentDate))

	// 创建 lumberjack 日志文件处理器
	fileLogger := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    2,     // 单个文件最大 2MB
		MaxBackups: 5,     // 最多保留 5 个备份
		MaxAge:     30,    // 保留 30 天
		Compress:   false, // 不启用压缩
	}

	// 使用自定义 writer 实现控制台带颜色、文件无颜色的输出
	multiWriter := &logWriter{
		console: os.Stdout,
		file:    fileLogger,
	}

	logs = &sLogger{
		logPath:     logPath,
		sysLog:      log.New(multiWriter, "", 0),
		fileLogger:  fileLogger,
		filePath:    filePath,
		currentDate: currentDate,
	}

	// 启动时清理 30 天前的旧日志
	cleanOldLogs(30)
}

func (s *sLogger) Info(v ...any) {
	Init()
	s.sysLog.SetPrefix(fmt.Sprintf("[%s] %s[INFO]%s ", time.Now().Format("2006-01-02 15:04:05"), Green, Reset))
	s.sysLog.Println(fmt.Sprint(v...))
}
func (s *sLogger) Error(v ...any) {
	Init()
	s.sysLog.SetPrefix(fmt.Sprintf("[%s] %s[ERROR]%s ", time.Now().Format("2006-01-02 15:04:05"), Red, Reset))
	msg := fmt.Sprint(v...)
	s.sysLog.Println(msg, strings.TrimSpace(string(debug.Stack())))
}
func (s *sLogger) Warn(v ...any) {
	Init()
	s.sysLog.SetPrefix(fmt.Sprintf("[%s] %s[WARN]%s ", time.Now().Format("2006-01-02 15:04:05"), Yellow, Reset))
	s.sysLog.Println(fmt.Sprint(v...))
}
func (s *sLogger) Debug(v ...any) {
	Init()
	s.sysLog.SetPrefix(fmt.Sprintf("[%s] %s[DEBUG]%s ", time.Now().Format("2006-01-02 15:04:05"), Blue, Reset))
	s.sysLog.Println(fmt.Sprint(v...))
}

func Info(v ...any) {
	if logs == nil {
		Init()
	}
	logs.Info(v...)
}
func Error(v ...any) {
	if logs == nil {
		Init()
	}
	logs.Error(v...)
}
func Warn(v ...any) {
	if logs == nil {
		Init()
	}
	logs.Warn(v...)
}
func Debug(v ...any) {
	if logs == nil {
		Init()
	}
	logs.Debug(v...)
}

// GetLogger 返回自定义日志器实例，实现 ILogger 接口
func GetLogger() ILogger {
	Init()
	return logs
}

type ILogger interface {
	Info(v ...any)
	Error(v ...any)
	Warn(v ...any)
	Debug(v ...any)
}
