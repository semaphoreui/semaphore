package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

func NewHostController(hostRepo db.HostRepository, subscription pro_interfaces.SubscriptionService) pro_interfaces.HostController {
	return &HostControllerImpl{}
}

type HostControllerImpl struct {
}

func (c *HostControllerImpl) GetProjectHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Total-Count", "0")
	helpers.WriteJSON(w, http.StatusOK, []any{})
}

func (c *HostControllerImpl) GetProjectHost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *HostControllerImpl) GetProjectHostTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Has-Next", "false")
	helpers.WriteJSON(w, http.StatusOK, []any{})
}

func (c *HostControllerImpl) GetHosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Total-Count", "0")
	helpers.WriteJSON(w, http.StatusOK, []any{})
}

func (c *HostControllerImpl) GetHost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *HostControllerImpl) GetHostTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Has-Next", "false")
	helpers.WriteJSON(w, http.StatusOK, []any{})
}

func (c *HostControllerImpl) DecommissionHost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
