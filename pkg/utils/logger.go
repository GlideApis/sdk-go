package utils

import (
	"log"
)

type LogLevel int

const (
	UNSET LogLevel = iota // 0 - represents unset value
	DEBUG LogLevel = iota //1 - Most verbose (shows all logs)
	INFO                  //2 - Shows INFO, WARN, ERROR
	WARN                  //3 - Shows WARN, ERROR
	ERROR                 //4 - Shows only ERROR
)

var (
	currentLogLevel = ERROR // default log level
)

var logger = struct {
	Debug func(format string, v ...interface{})
	Info  func(format string, v ...interface{})
	Warn  func(format string, v ...interface{})
	Error func(format string, v ...interface{})
}{
	Debug: func(format string, v ...interface{}) {
		if currentLogLevel == DEBUG { // Show debug only at DEBUG level (1)
			log.Printf("DEBUG: "+format, v...)
		}
	},
	Info: func(format string, v ...interface{}) {
		if currentLogLevel <= INFO { // Show info at DEBUG or INFO level (1,2)
			log.Printf("INFO: "+format, v...)
		}
	},
	Warn: func(format string, v ...interface{}) {
		if currentLogLevel <= WARN { // Show warn if we're at DEBUG, INFO, or WARN level
			log.Printf("WARN: "+format, v...)
		}
	},
	Error: func(format string, v ...interface{}) {
		if currentLogLevel <= ERROR { // Show error for all levels
			log.Printf("ERROR: "+format, v...)
		}
	},
}

// Export the logger
var Logger = logger

func SetLogLevel(level LogLevel) {
	currentLogLevel = level
}
