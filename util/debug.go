package util

import (
	log "github.com/sirupsen/logrus"
	"runtime"
	"strconv"
	"strings"
)

func Goid() (int, error) {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		log.Debug("Cannot get goroutine id: ", err)
		return -1, err
	}
	return id, nil
}

func LogGoid(msg string) {
	id, err := Goid()
	if err == nil {
		log.Info(msg, ", goid=", id)
	}
}
