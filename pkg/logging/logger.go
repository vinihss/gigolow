
// pkg/logging/logger.go

package logging

import (
	"os"
	"sync"
)

type Logger struct {
	file *os.File
	mu   sync.Mutex
}

func NewLogger(filePath string) *Logger {
	file, _ := os.Create(filePath)
	return &Logger{file: file}
}

func (l *Logger) Log(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.file.WriteString(message + "\n")
}

func (l *Logger) Close() {
	l.file.Close()
}