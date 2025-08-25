package pro_interfaces

import "net/http"

type ProjectRunnerController interface {
	GetRunners(w http.ResponseWriter, r *http.Request)
	AddRunner(w http.ResponseWriter, r *http.Request)
	RunnerMiddleware(next http.Handler) http.Handler
	GetRunner(w http.ResponseWriter, r *http.Request)
	UpdateRunner(w http.ResponseWriter, r *http.Request)
	DeleteRunner(w http.ResponseWriter, r *http.Request)
	SetRunnerActive(w http.ResponseWriter, r *http.Request)
	ClearRunnerCache(w http.ResponseWriter, r *http.Request)
	GetRunnerTags(w http.ResponseWriter, r *http.Request)
}
