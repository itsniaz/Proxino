package logger

import (
	"log"
	"os"
	"strings"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var (
	currentLevel Level
	debugLogger  = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger   = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger   = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger  = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func Init(level string) {
	switch strings.ToLower(level) {
	case "debug":
		currentLevel = DEBUG
	case "info":
		currentLevel = INFO
	case "warn":
		currentLevel = WARN
	case "error":
		currentLevel = ERROR
	default:
		currentLevel = INFO
	}
}

func Debug(v ...interface{}) {
	if currentLevel <= DEBUG {
		debugLogger.Println(v...)
	}
}

func Info(v ...interface{}) {
	if currentLevel <= INFO {
		infoLogger.Println(v...)
	}
}

func Warn(v ...interface{}) {
	if currentLevel <= WARN {
		warnLogger.Println(v...)
	}
}

func Error(v ...interface{}) {
	if currentLevel <= ERROR {
		errorLogger.Println(v...)
	}
} 