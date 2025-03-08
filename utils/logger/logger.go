package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var levelFromString = map[string]LogLevel{
	"debug": DEBUG,
	"info":  INFO,
	"warn":  WARN,
	"error": ERROR,
	"fatal": FATAL,
}

// Logger 日志记录器
type Logger struct {
	level      LogLevel
	logger     *log.Logger
	fileHandle *os.File
	console    bool
}

var defaultLogger *Logger

// Config 日志配置
type Config struct {
	Level   string `json:"level"`
	File    string `json:"file"`
	Console bool   `json:"console"`
}

// Init 初始化日志系统
func Init(config Config) error {
	level, ok := levelFromString[strings.ToLower(config.Level)]
	if !ok {
		level = INFO
	}

	var writers []io.Writer

	// 如果指定了日志文件，则创建文件写入器
	var fileHandle *os.File
	if config.File != "" {
		// 确保日志目录存在
		dir := filepath.Dir(config.File)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %v", err)
		}

		var err error
		fileHandle, err = os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %v", err)
		}
		writers = append(writers, fileHandle)
	}

	// 如果启用控制台输出，则添加标准输出
	if config.Console {
		writers = append(writers, os.Stdout)
	}

	// 创建多输出写入器
	var writer io.Writer
	if len(writers) > 1 {
		writer = io.MultiWriter(writers...)
	} else if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = os.Stdout // 默认输出到标准输出
	}

	// 创建日志记录器
	logger := log.New(writer, "", 0)

	defaultLogger = &Logger{
		level:      level,
		logger:     logger,
		fileHandle: fileHandle,
		console:    config.Console,
	}

	return nil
}

// Close 关闭日志文件
func Close() {
	if defaultLogger != nil && defaultLogger.fileHandle != nil {
		defaultLogger.fileHandle.Close()
	}
}

// formatLog 格式化日志消息
func formatLog(level LogLevel, format string, args ...interface{}) string {
	now := time.Now().Format("2006-01-02 15:04:05.000")

	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2)
	fileInfo := "???"
	if ok {
		fileInfo = filepath.Base(file)
	}

	// 格式化消息
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	return fmt.Sprintf("[%s] [%s] [%s:%d] %s", now, levelNames[level], fileInfo, line, msg)
}

// Debug 输出调试级别日志
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= DEBUG {
		defaultLogger.logger.Println(formatLog(DEBUG, format, args...))
	}
}

// Info 输出信息级别日志
func Info(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= INFO {
		defaultLogger.logger.Println(formatLog(INFO, format, args...))
	}
}

// Warn 输出警告级别日志
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= WARN {
		defaultLogger.logger.Println(formatLog(WARN, format, args...))
	}
}

// Error 输出错误级别日志
func Error(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= ERROR {
		defaultLogger.logger.Println(formatLog(ERROR, format, args...))
	}
}

// Fatal 输出致命错误日志并退出程序
func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= FATAL {
		defaultLogger.logger.Println(formatLog(FATAL, format, args...))
	}
	os.Exit(1)
}
