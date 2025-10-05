package utils

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[TAMIS] ", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string) {
	l.Printf("[INFO] %s", msg)
}

func (l *Logger) Error(msg string) {
	l.Printf("[ERROR] %s", msg)
}

func (l *Logger) Debug(msg string) {
	l.Printf("[DEBUG] %s", msg)
}

func (l *Logger) Fatal(msg string) {
	l.Printf("[FATAL] %s", msg)
	os.Exit(1)
}
func (l *Logger) Warn(msg string) {
	l.Printf("[WARN] %s", msg)
}
