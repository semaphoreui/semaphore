package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Digital-Data-Co/forge/api/helpers"
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/services/scc"
	"github.com/gorilla/mux"
)

// SCCController handles SCC-related API endpoints
type SCCController struct {
	store      db.Store
	sccService *scc.SCCService
}

// NewSCCController creates a new SCC controller
func NewSCCController(store db.Store, sccService *scc.SCCService) *SCCController {
	return &SCCController{
		store:      store,
		sccService: sccService,
	}
}

// GetAvailableBenchmarks returns the list of available SCAP benchmarks
func (c *SCCController) GetAvailableBenchmarks(w http.ResponseWriter, r *http.Request) {
	benchmarks, err := c.sccService.GetAvailableBenchmarks()
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, benchmarks)
}

// GetSupportedOS returns the list of supported operating systems
func (c *SCCController) GetSupportedOS(w http.ResponseWriter, r *http.Request) {
	osList := c.sccService.GetSupportedOS()
	helpers.WriteJSON(w, http.StatusOK, osList)
}

// GetBenchmarksByOS returns available benchmarks for a specific OS
func (c *SCCController) GetBenchmarksByOS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	os := vars["os"]

	benchmarks, err := c.sccService.GetBenchmarksByOS(os)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, benchmarks)
}

// DownloadBenchmark downloads a SCAP benchmark
func (c *SCCController) DownloadBenchmark(w http.ResponseWriter, r *http.Request) {
	var benchmark scc.SCCBenchmark
	if !helpers.Bind(w, r, &benchmark) {
		return
	}

	err := c.sccService.DownloadBenchmark(benchmark)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]string{"message": "Benchmark downloaded successfully"})
}

// GetBenchmarkProfiles returns available profiles for a benchmark
func (c *SCCController) GetBenchmarkProfiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	benchmarkPath := vars["benchmarkPath"]

	profiles, err := c.sccService.GetBenchmarkProfiles(benchmarkPath)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, profiles)
}

// RunSCCScan runs an SCC compliance scan
func (c *SCCController) RunSCCScan(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	user := helpers.GetFromContext(r, "user").(db.User)

	var req scc.SCCScanRequest
	if !helpers.Bind(w, r, &req) {
		return
	}

	req.ProjectID = project.ID
	req.UserID = user.ID

	if err := scc.ValidateSCCScanRequest(&req); err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Create and run the SCC scan task
	task, err := scc.CreateSCCTask(c.store, c.sccService, req.ProjectID, req.PolicyID, req.Host, req.Benchmark, req.Profile, req.UserID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, task)
}

// GetSCCScanHistory returns the scan history for a project
func (c *SCCController) GetSCCScanHistory(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	scans, err := scc.GetSCCScanHistory(c.store, project.ID, limit, offset)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, scans)
}

// GetSCCScanResults returns the results for a specific scan
func (c *SCCController) GetSCCScanResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanIDStr := vars["scanId"]

	scanID, err := strconv.Atoi(scanIDStr)
	if err != nil {
		helpers.WriteError(w, fmt.Errorf("invalid scan ID: %w", err))
		return
	}

	results, err := scc.GetSCCScanResults(c.store, scanID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, results)
}

// GetSCCRuleResults returns the rule results for a specific report
func (c *SCCController) GetSCCRuleResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportIDStr := vars["reportId"]

	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil {
		helpers.WriteError(w, fmt.Errorf("invalid report ID: %w", err))
		return
	}

	results, err := scc.GetSCCRuleResults(c.store, reportID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, results)
}

// CheckSCCAvailability checks if SCC is available on the system
func (c *SCCController) CheckSCCAvailability(w http.ResponseWriter, r *http.Request) {
	available, version, err := c.sccService.CheckSCCAvailability()
	if err != nil {
		helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"available": false,
			"version":   "",
			"error":     err.Error(),
		})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"available": available,
		"version":   version,
		"error":     nil,
	})
}

// GetSCCStatus returns the current status of SCC integration
func (c *SCCController) GetSCCStatus(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	// Get recent scan statistics
	recentScans, err := scc.GetSCCScanHistory(c.store, project.ID, 10, 0)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Calculate statistics
	var totalScans, completedScans, failedScans int
	for _, scan := range recentScans {
		totalScans++
		if scan.Status == "completed" {
			completedScans++
		} else if scan.Status == "failed" {
			failedScans++
		}
	}

	// Check SCC availability
	available, version, _ := c.sccService.CheckSCCAvailability()

	status := map[string]interface{}{
		"scc_available":   available,
		"scc_version":     version,
		"total_scans":     totalScans,
		"completed_scans": completedScans,
		"failed_scans":    failedScans,
		"supported_os":    c.sccService.GetSupportedOS(),
	}

	helpers.WriteJSON(w, http.StatusOK, status)
}
