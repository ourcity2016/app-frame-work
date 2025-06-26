package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type Logger interface {
	Info(format string, msg ...any)
	Error(format string, msg ...any)
	Warn(format string, msg ...any)
	Debug(format string, msg ...any)
	SetOutput(w io.Writer)
}

type myLogger struct {
	logger *log.Logger
	skip   int
	debug  bool
}

var (
	instance *myLogger
	once     sync.Once
)

func BuildMyLogger() Logger {
	once.Do(func() {
		instance = &myLogger{
			logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds),
			skip:   3, // 根据实际调用栈调整
			debug:  true,
		}
	})
	return instance
}

func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "???:0"
	}
	// 简化文件路径
	dir, filename := filepath.Split(file)
	dirs := strings.Split(strings.Trim(dir, string(filepath.Separator)), string(filepath.Separator))
	if len(dirs) > 2 {
		dir = filepath.Join(dirs[len(dirs)-2:]...)
	}
	return filepath.Join(dir, filename) + ":" + strconv.Itoa(line)
}

func (l *myLogger) output(level, format string, msg ...any) {
	caller := getCallerInfo(l.skip)
	l.logger.SetPrefix(level + " " + caller + " ")

	if len(msg) > 0 {
		l.logger.Printf(format, msg...)
	} else {
		l.logger.Println(format)
	}

	l.logger.SetPrefix("")
}

func (l *myLogger) Info(format string, msg ...any) {
	l.output("INFO", format, msg...)
}

func (l *myLogger) Error(format string, msg ...any) {
	l.output("ERROR", format, msg...)
}

func (l *myLogger) Warn(format string, msg ...any) {
	l.output("WARN", format, msg...)
}

func (l *myLogger) Debug(format string, msg ...any) {
	if l.debug {
		l.output("DEBUG", format, msg...)
	}
}

func (l *myLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}
