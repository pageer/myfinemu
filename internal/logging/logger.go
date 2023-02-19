package logging

import (
	"bytes"
	"fmt"
	"log"
)

var logger *log.Logger
var buffer bytes.Buffer

func init() {
	logger = log.New(&buffer, "", 0)
}

func Clear() {
	buffer.Truncate(0)
}

func Flush() {
	fmt.Print(&buffer)
}

func LogDebug(format string, v ...any) {
	logger.Printf(format, v...)
}
