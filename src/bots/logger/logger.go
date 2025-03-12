package logger

import (
	"log"
	"os"
)

type BotLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewBotLogger() *BotLogger {
	return &BotLogger{
		debug: log.New(os.Stdout, "DEBUG: ", log.LstdFlags),
		info:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		warn:  log.New(os.Stdout, "WARN: ", log.LstdFlags),
		error: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}
}

func (l *BotLogger) Debug(v ...interface{}) {
	l.debug.Println(v...)
}

func (l *BotLogger) Info(v ...interface{}) {
	l.info.Println(v...)
}

func (l *BotLogger) Warn(v ...interface{}) {
	l.warn.Println(v...)
}

func (l *BotLogger) Error(v ...interface{}) {
	l.error.Println(v...)
}
