// Package logger 提供统一的日志记录功能
// 支持同时输出到控制台和文件，文件按日期自动分类
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Level 日志级别
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志记录器
type Logger struct {
	level      Level
	consoleLog *log.Logger
	fileLog    *log.Logger
	file       *os.File
	logDir     string
	currentDate string
}

// Config 日志配置
type Config struct {
	Level      Level  // 日志级别
	LogDir     string // 日志目录
	EnableFile bool   // 是否启用文件日志
}

var (
	// DefaultLogger 默认日志记录器
	DefaultLogger *Logger
)

// Init 初始化默认日志记录器
func Init(config Config) error {
	logger, err := New(config)
	if err != nil {
		return err
	}
	DefaultLogger = logger
	return nil
}

// New 创建一个新的日志记录器
func New(config Config) (*Logger, error) {
	logger := &Logger{
		level:  config.Level,
		logDir: config.LogDir,
	}

	// 控制台日志（输出到stdout）
	logger.consoleLog = log.New(os.Stdout, "", 0)

	// 文件日志
	if config.EnableFile {
		if err := logger.rotateLogFile(); err != nil {
			return nil, fmt.Errorf("failed to create log file: %w", err)
		}
	}

	return logger, nil
}

// rotateLogFile 按日期轮转日志文件
func (l *Logger) rotateLogFile() error {
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// 如果是同一天，不需要轮转
	if l.currentDate == dateStr && l.file != nil {
		return nil
	}

	// 关闭旧文件
	if l.file != nil {
		l.file.Close()
	}

	// 确保日志目录存在
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 创建新的日志文件
	filename := filepath.Join(l.logDir, fmt.Sprintf("canpulse-%s.log", dateStr))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	l.currentDate = dateStr

	// 创建文件日志记录器（同时写入文件，不写入控制台）
	l.fileLog = log.New(file, "", 0)

	return nil
}

// log 写入日志（内部方法）
func (l *Logger) log(level Level, format string, args ...interface{}) {
	// 检查日志级别
	if level < l.level {
		return
	}

	// 轮转日志文件（如果需要）
	if l.file != nil {
		l.rotateLogFile()
	}

	// 格式化消息
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := level.String()
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, levelStr, message)

	// 输出到控制台
	if l.consoleLog != nil {
		l.consoleLog.Println(logLine)
	}

	// 输出到文件
	if l.fileLog != nil {
		l.fileLog.Println(logLine)
	}
}

// Debug 输出Debug级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info 输出Info级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn 输出Warn级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error 输出Error级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Printf 输出不带级别的日志（兼容旧代码）
func (l *Logger) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	// 输出到控制台（不带时间戳和级别）
	if l.consoleLog != nil {
		fmt.Print(message)
	}

	// 输出到文件（带时间戳）
	if l.fileLog != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		l.fileLog.Printf("[%s] %s", timestamp, message)
	}
}

// Println 输出不带级别的日志行（兼容旧代码）
func (l *Logger) Println(args ...interface{}) {
	message := fmt.Sprintln(args...)

	// 输出到控制台（不带时间戳和级别）
	if l.consoleLog != nil {
		fmt.Print(message)
	}

	// 输出到文件（带时间戳）
	if l.fileLog != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		l.fileLog.Printf("[%s] %s", timestamp, message)
	}
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Writer 返回一个io.Writer，用于其他库的日志输出
func (l *Logger) Writer() io.Writer {
	if l.file != nil {
		return io.MultiWriter(os.Stdout, l.file)
	}
	return os.Stdout
}

// 全局日志函数（使用DefaultLogger）

// Debug 输出Debug级别日志
func Debug(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Debug(format, args...)
	}
}

// Info 输出Info级别日志
func Info(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Info(format, args...)
	}
}

// Warn 输出Warn级别日志
func Warn(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Warn(format, args...)
	}
}

// Error 输出Error级别日志
func Error(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Error(format, args...)
	}
}

// Printf 输出不带级别的日志
func Printf(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Printf(format, args...)
	} else {
		fmt.Printf(format, args...)
	}
}

// Println 输出不带级别的日志行
func Println(args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Println(args...)
	} else {
		fmt.Println(args...)
	}
}
