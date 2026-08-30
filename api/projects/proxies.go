package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// ProxyMiddleware ensures a proxy exists and loads it to the context
func ProxyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		proxyID, err := helpers.GetIntParam("proxy_id", w, r)
		if err != nil {
			return
		}

		proxy, err := helpers.Store(r).GetProxy(project.ID, proxyID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "proxy", proxy)
		next.ServeHTTP(w, r)
	})
}

// GetProxy returns a single proxy, or all proxies of the project
func GetProxy(w http.ResponseWriter, r *http.Request) {
	if proxy := helpers.GetFromContext(r, "proxy"); proxy != nil {
		helpers.WriteJSON(w, http.StatusOK, proxy.(db.Proxy))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)

	proxies, err := helpers.Store(r).GetProxies(project.ID, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, proxies)
}

// GetProxyRefs returns entities which reference the proxy
func GetProxyRefs(w http.ResponseWriter, r *http.Request) {
	proxy := helpers.GetFromContext(r, "proxy").(db.Proxy)

	refs, err := helpers.Store(r).GetProxyRefs(proxy.ProjectID, proxy.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// AddProxy creates a proxy in the project
func AddProxy(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var proxy db.Proxy
	if !helpers.Bind(w, r, &proxy) {
		return
	}

	if proxy.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if err := proxy.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := db.ValidateProxy(helpers.Store(r), &proxy); err != nil {
		helpers.WriteError(w, err)
		return
	}

	newProxy, err := helpers.Store(r).CreateProxy(proxy)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   newProxy.ProjectID,
		ObjectType:  db.EventProxy,
		ObjectID:    newProxy.ID,
		Description: fmt.Sprintf("Proxy %s created", newProxy.Name),
	})

	helpers.WriteJSON(w, http.StatusCreated, newProxy)
}

// UpdateProxy writes a proxy to the database
func UpdateProxy(w http.ResponseWriter, r *http.Request) {
	oldProxy := helpers.GetFromContext(r, "proxy").(db.Proxy)

	var proxy db.Proxy
	if !helpers.Bind(w, r, &proxy) {
		return
	}

	if proxy.ID != oldProxy.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Proxy ID in URL and in body must be the same",
		})
		return
	}

	if proxy.ProjectID != oldProxy.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if err := proxy.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := db.ValidateProxy(helpers.Store(r), &proxy); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := helpers.Store(r).UpdateProxy(proxy); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   proxy.ProjectID,
		ObjectType:  db.EventProxy,
		ObjectID:    proxy.ID,
		Description: fmt.Sprintf("Proxy %s updated", proxy.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}

// RemoveProxy deletes a proxy from the database
func RemoveProxy(w http.ResponseWriter, r *http.Request) {
	proxy := helpers.GetFromContext(r, "proxy").(db.Proxy)

	if err := helpers.Store(r).DeleteProxy(proxy.ProjectID, proxy.ID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   proxy.ProjectID,
		ObjectType:  db.EventProxy,
		ObjectID:    proxy.ID,
		Description: fmt.Sprintf("Proxy %s deleted", proxy.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}
