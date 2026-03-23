package api

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

func accessLogFormatter(_ io.Writer, params handlers.LogFormatterParams) {
	if strings.HasSuffix(params.URL.Path, "/api/ping") {
		return
	}
	log.WithFields(log.Fields{
		"method":   params.Request.Method,
		"path":     params.URL.Path,
		"status":   params.StatusCode,
		"size":     params.Size,
		"duration": time.Since(params.TimeStamp).String(),
		"remote":   params.Request.RemoteAddr,
	}).Info("http request")
}

func AccessLogMiddleware(next http.Handler) http.Handler {
	return handlers.CustomLoggingHandler(nil, next, accessLogFormatter)
}
