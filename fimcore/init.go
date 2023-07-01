package fimcore

import (
	"github.com/FimGroup/logging"
)

var loggerManager logging.LoggerManager

func Init() error {
	loggerManager = logging.GetLoggerManager()
	return nil
}
