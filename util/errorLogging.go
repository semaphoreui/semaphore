package util

import (
	log "github.com/sirupsen/logrus"
)

// LogWarning logs a warning with arbitrary field if error
func LogWarning(err error) {
	LogWarningF(err, log.Fields{"level": "Warn"})
}

// LogWarningF logs a warning with added field context if error
func LogWarningF(err error, fields log.Fields) {
	if err != nil {
		log.WithFields(fields).Warn(err.Error())
	}
}

// LogError logs an error with arbitrary field if error
func LogError(err error) {
	LogErrorF(err, log.Fields{"level": "Error"})
}

// LogErrorF logs a error with added field context if error
func LogErrorF(err error, fields log.Fields) {
	if err != nil {
		log.WithFields(fields).Error(err.Error())
	}
}

// LogPanic logs and panics with arbitrary field if error
func LogPanic(err error) {
	LogPanicF(err, log.Fields{"level": "Panic"})
}

// LogPanicF logs and panics with added field context if error
func LogPanicF(err error, fields log.Fields) {
	if err != nil {
		log.WithFields(fields).Panic(err.Error())
	}
}
