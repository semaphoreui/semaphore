package pro_interfaces

import "net/http"

type HostController interface {
	GetProjectHosts(w http.ResponseWriter, r *http.Request)
	GetProjectHost(w http.ResponseWriter, r *http.Request)
	GetProjectHostTasks(w http.ResponseWriter, r *http.Request)
	GetHosts(w http.ResponseWriter, r *http.Request)
	GetHost(w http.ResponseWriter, r *http.Request)
	GetHostTasks(w http.ResponseWriter, r *http.Request)
}
