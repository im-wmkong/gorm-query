package genprops

import (
	"log"
	"os"
)

// Logger 定义了 genprops 生成器使用的日志接口。
type Logger interface {
	// Debug 输出调试级别的日志，用于过程性细节信息。
	Debug(format string, a ...any)
	// Info 输出信息级别的日志，用于关键操作的结果通知。
	Info(format string, a ...any)
	// Warn 输出警告级别的日志，用于非致命的异常情况。
	Warn(format string, a ...any)
}

// defaultLogger 基于标准库 log 的默认实现，输出到 stderr。
type defaultLogger struct {
	l *log.Logger
}

// NewDefaultLogger 创建一个输出到 stderr 的默认 Logger。
func NewDefaultLogger() Logger {
	return &defaultLogger{l: log.New(os.Stderr, "[genprops] ", log.LstdFlags)}
}

func (d *defaultLogger) Debug(format string, a ...any) { d.l.Printf("[DEBUG] "+format, a...) }
func (d *defaultLogger) Info(format string, a ...any)  { d.l.Printf("[INFO]  "+format, a...) }
func (d *defaultLogger) Warn(format string, a ...any)  { d.l.Printf("[WARN]  "+format, a...) }

// nopLogger 不输出任何日志。
type nopLogger struct{}

// NopLogger 返回一个静默的 Logger，不输出任何日志。
func NopLogger() Logger                         { return nopLogger{} }
func (nopLogger) Debug(format string, a ...any) { _ = format; _ = a }
func (nopLogger) Info(format string, a ...any)  { _ = format; _ = a }
func (nopLogger) Warn(format string, a ...any)  { _ = format; _ = a }
