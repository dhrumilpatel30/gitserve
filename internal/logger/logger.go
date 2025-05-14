package logger

import (
	"fmt"
	"gitserve/internal/termui"
	"io"
	"os"
	"strings"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

var logLevelStrings = map[LogLevel]string{
	LogLevelDebug:   "DEBUG",
	LogLevelInfo:    "INFO",
	LogLevelWarning: "WARNING",
	LogLevelError:   "ERROR",
}

func (l LogLevel) String() string {
	return logLevelStrings[l]
}

type Service interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
	SetLevel(level LogLevel)
	SetOutput(writer io.Writer)
}

type loggerService struct {
	level  LogLevel
	output io.Writer
}

func NewService(defaultLevel LogLevel) Service {
	return &loggerService{
		level:  defaultLevel,
		output: os.Stdout, // Default to Stdout
	}
}

func (s *loggerService) SetLevel(level LogLevel) {
	s.level = level
}

func (s *loggerService) SetOutput(writer io.Writer) {
	s.output = writer
}

func (s *loggerService) log(level LogLevel, color string, format string, args ...interface{}) {
	if level < s.level {
		return
	}

	var builder strings.Builder
	builder.WriteString(color)
	builder.WriteString(termui.ColorBold)
	builder.WriteString(fmt.Sprintf("[%s] ", level.String()))
	builder.WriteString(termui.ColorReset)
	builder.WriteString(color)
	builder.WriteString(fmt.Sprintf(format, args...))
	builder.WriteString(termui.ColorReset)
	builder.WriteString("\n")

	fmt.Fprint(s.output, builder.String())
}

func (s *loggerService) Debug(format string, args ...interface{}) {
	s.log(LogLevelDebug, termui.ColorCyan, format, args...)
}

func (s *loggerService) Info(format string, args ...interface{}) {
	s.log(LogLevelInfo, termui.ColorGreen, format, args...)
}

func (s *loggerService) Warning(format string, args ...interface{}) {
	s.log(LogLevelWarning, termui.ColorYellow, format, args...)
}

func (s *loggerService) Error(format string, args ...interface{}) {
	// Errors should go to Stderr
	originalOutput := s.output
	s.output = os.Stderr
	s.log(LogLevelError, termui.ColorRed, format, args...)
	s.output = originalOutput
}

// Helper function to create a new logger for testing or specific configurations
func NewTestLogger(level LogLevel, writer io.Writer) Service {
	return &loggerService{
		level:  level,
		output: writer,
	}
}
