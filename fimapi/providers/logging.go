package providers

import "github.com/FimGroup/logging"

type LoggerManager interface {
	logging.LoggerManager
}

type Logger interface {
	logging.Logger
}
