package utils

import (
	"fmt"
	"log"
	"strings"
)

const (
	loggerPrefixDelimiter = " "
)

func LoggerExtendWithPrefix(l *log.Logger, prefix string) *log.Logger {
	currentPrefix := l.Prefix()
	currentPrefix = strings.TrimSpace(currentPrefix)

	newPrefix := fmt.Sprintf("%s%s%s%s",
		currentPrefix, loggerPrefixDelimiter, prefix, loggerPrefixDelimiter)

	return log.New(l.Writer(), newPrefix, l.Flags())
}
