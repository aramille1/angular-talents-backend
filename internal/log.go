package internal

import (
	log "github.com/sirupsen/logrus"
)

func LogError(receivedError *CustomError, contextData map[string]interface{}) {
	fields := map[string]interface{}{"error_data": receivedError.ErrorData(), "context_data": contextData}
	log.WithFields(log.Fields(fields)).Error(receivedError.message)
}

func LogInfo(message string, additionalContext map[string]interface{}) {
	log.WithFields(log.Fields(additionalContext)).Info(message)
}