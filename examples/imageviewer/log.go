package main

import (
	"log"
	"os"
)

var logDisabled = os.Getenv("LOG_DISABLED") == "1"

func logPrintln(v ...any) {
	if !logDisabled {
		log.Println(v...)
	}
}

func logPrintf(format string, v ...any) {
	if !logDisabled {
		log.Printf(format, v...)
	}
}
