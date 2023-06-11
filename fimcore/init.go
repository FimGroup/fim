package fimcore

import (
	"github.com/FimGroup/fim/fimapi/providers"
	"github.com/FimGroup/fim/fimsupport/logging"
)

var loggerManager providers.LoggerManager

func Init() error {
	loggerManager = logging.GetLoggerManager()
	return nil
}
