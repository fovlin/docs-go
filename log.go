package main

import (
	"fmt"

	"acovia.net/docs-go/internal/record"
)

var logQuiet bool

func logInfo(format string, value ...any) {
	if logQuiet {
		return
	}

	record.Info(fmt.Sprintf(format, value...))
}

func logWarn(format string, value ...any) {
	if logQuiet {
		return
	}

	record.Warn(fmt.Sprintf(format, value...))
}

func logError(format string, value ...any) {
	record.ErrorNoExit(fmt.Sprintf(format, value...))
}

func logPlain(format string, value ...any) {
	record.Printf(format, value...)
}
