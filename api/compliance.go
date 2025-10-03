package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Digital-Data-Co/forge/api/helpers"
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/services/compliance"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// ComplianceController handles compliance-related API endpoints
type ComplianceController struct {
	store      db.Store
	contentSvc *compliance.ContentService
	policySvc  *compliance.PolicyService
	scannerSvc *compliance.ScannerService
}

// NewComplianceController creates a new compliance controller
func NewComplianceController(store db.Store, contentSvc *compliance.ContentService, policySvc *compliance.PolicyService, scannerSvc *compliance.ScannerService) *ComplianceController {
	return &ComplianceController{
		store:      store,
		contentSvc: contentSvc,
		policySvc:  policySvc,
		scannerSvc: scannerSvc,
	}
}

// PreflightCheck validates OpenSCAP installation
func (c *ComplianceController) PreflightCheck(w http.ResponseWriter, r *http.Request) {
	result := c.scannerSvc.PreflightCheck()
	helpers.WriteJSON(w, http.StatusOK, result)
}

// Content endpoints

// UploadContent uploads a SCAP DataStream file
func (c *ComplianceController) UploadContent(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	user := helpers.GetFromContext(r, "user").(*db.User)
	projectID := project.ID
	userID := user.ID

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to parse form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Read file content
	contentData, err := io.ReadAll(file)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to read file"})
		return
	}

	// Perform additional file content validation
	if err := ScanFileForMalware(contentData); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Get name from form or use filename
	name := r.FormValue("name")
	if name == "" {
		name = header.Filename
	}

	source := r.FormValue("source")

	// Upload and process content
	content, profiles, err := c.contentSvc.UploadContent(projectID, userID, name, source, contentData)
	if err != nil {
		log.WithError(err).Error("Failed to upload content")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"content":  content,
		"profiles": profiles,
	}

	helpers.WriteJSON(w, http.StatusCreated, response)
}

// GetContents retrieves all SCAP contents for a project
func (c *ComplianceController) GetContents(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	projectID := project.ID

	contents, err := c.contentSvc.GetContentsByProject(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get contents")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, contents)
}

// GetContent retrieves a specific SCAP content
func (c *ComplianceController) GetContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid content ID"})
		return
	}

	content, err := c.contentSvc.GetContent(contentID)
	if err != nil {
		if err == db.ErrNotFound {
			helpers.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Content not found"})
			return
		}
		log.WithError(err).Error("Failed to get content")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, content)
}

// GetContentProfiles retrieves profiles for a content
func (c *ComplianceController) GetContentProfiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid content ID"})
		return
	}

	profiles, err := c.contentSvc.GetProfilesByContent(contentID)
	if err != nil {
		log.WithError(err).Error("Failed to get profiles")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, profiles)
}

// DeleteContent deletes a SCAP content
func (c *ComplianceController) DeleteContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid content ID"})
		return
	}

	err = c.contentSvc.DeleteContent(contentID)
	if err != nil {
		log.WithError(err).Error("Failed to delete content")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusNoContent, nil)
}

// Policy endpoints

// CreatePolicy creates a new compliance policy
func (c *ComplianceController) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	user := helpers.GetFromContext(r, "user").(*db.User)
	projectID := project.ID
	userID := user.ID

	var request struct {
		Name      string                 `json:"name"`
		ContentID int                    `json:"content_id"`
		ProfileID string                 `json:"profile_id"`
		Attrs     map[string]interface{} `json:"attrs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	policy, err := c.policySvc.CreatePolicy(projectID, userID, request.Name, request.ProfileID, request.ContentID, request.Attrs)
	if err != nil {
		log.WithError(err).Error("Failed to create policy")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, policy)
}

// GetPolicies retrieves all policies for a project
func (c *ComplianceController) GetPolicies(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	projectID := project.ID

	policies, err := c.policySvc.GetPoliciesByProject(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get policies")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, policies)
}

// GetPolicy retrieves a specific policy
func (c *ComplianceController) GetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid policy ID"})
		return
	}

	policy, err := c.policySvc.GetPolicy(policyID)
	if err != nil {
		if err == db.ErrNotFound {
			helpers.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Policy not found"})
			return
		}
		log.WithError(err).Error("Failed to get policy")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, policy)
}

// UpdatePolicy updates an existing policy
func (c *ComplianceController) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid policy ID"})
		return
	}

	var policy db.CompliancePolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	policy.ID = policyID

	err = c.policySvc.UpdatePolicy(&policy)
	if err != nil {
		log.WithError(err).Error("Failed to update policy")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, policy)
}

// DeletePolicy deletes a policy
func (c *ComplianceController) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid policy ID"})
		return
	}

	err = c.policySvc.DeletePolicy(policyID)
	if err != nil {
		log.WithError(err).Error("Failed to delete policy")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusNoContent, nil)
}

// AssignPolicy assigns a policy to targets
func (c *ComplianceController) AssignPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid policy ID"})
		return
	}

	var request struct {
		Assignments []compliance.PolicyAssignmentRequest `json:"assignments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	err = c.policySvc.AssignPolicy(policyID, request.Assignments)
	if err != nil {
		log.WithError(err).Error("Failed to assign policy")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]string{"message": "Policy assigned successfully"})
}

// GetPolicyAssignments retrieves assignments for a policy
func (c *ComplianceController) GetPolicyAssignments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid policy ID"})
		return
	}

	assignments, err := c.policySvc.GetPolicyAssignments(policyID)
	if err != nil {
		log.WithError(err).Error("Failed to get policy assignments")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, assignments)
}

// Scan endpoints

// ScanPolicy initiates a compliance scan
func (c *ComplianceController) ScanPolicy(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	user := helpers.GetFromContext(r, "user").(*db.User)
	projectID := project.ID
	userID := user.ID

	vars := mux.Vars(r)
	policyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid policy ID"})
		return
	}

	scan, err := c.scannerSvc.ScanPolicy(projectID, userID, policyID)
	if err != nil {
		log.WithError(err).Error("Failed to scan policy")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, scan)
}

// GetScans retrieves scans for a project
func (c *ComplianceController) GetScans(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	projectID := project.ID

	// Parse query parameters for pagination
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	scans, err := c.scannerSvc.GetScansByProject(projectID, limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to get scans")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, scans)
}

// GetScan retrieves a specific scan
func (c *ComplianceController) GetScan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid scan ID"})
		return
	}

	scan, err := c.scannerSvc.GetScan(scanID)
	if err != nil {
		if err == db.ErrNotFound {
			helpers.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Scan not found"})
			return
		}
		log.WithError(err).Error("Failed to get scan")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, scan)
}

// CancelScan cancels a running scan
func (c *ComplianceController) CancelScan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid scan ID"})
		return
	}

	err = c.scannerSvc.CancelScan(scanID)
	if err != nil {
		log.WithError(err).Error("Failed to cancel scan")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]string{"message": "Scan cancelled successfully"})
}

// GetScanReports retrieves reports for a scan
func (c *ComplianceController) GetScanReports(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid scan ID"})
		return
	}

	reports, err := c.scannerSvc.GetReportsByScan(scanID)
	if err != nil {
		log.WithError(err).Error("Failed to get scan reports")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, reports)
}

// Report endpoints

// GetReports retrieves reports for a project with filters
func (c *ComplianceController) GetReports(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	projectID := project.ID

	// Parse query parameters for filters and pagination
	filters := make(map[string]interface{})
	if policyID := r.URL.Query().Get("policy_id"); policyID != "" {
		if id, err := strconv.Atoi(policyID); err == nil {
			filters["policy_id"] = id
		}
	}
	if host := r.URL.Query().Get("host"); host != "" {
		filters["host"] = host
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}

	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	reports, err := c.scannerSvc.GetReportsByProject(projectID, filters, limit, offset)
	if err != nil {
		log.WithError(err).Error("Failed to get reports")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, reports)
}

// GetReport retrieves a specific report
func (c *ComplianceController) GetReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid report ID"})
		return
	}

	report, err := c.scannerSvc.GetReport(reportID)
	if err != nil {
		if err == db.ErrNotFound {
			helpers.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Report not found"})
			return
		}
		log.WithError(err).Error("Failed to get report")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, report)
}

// DownloadArf downloads the ARF file for a report
func (c *ComplianceController) DownloadArf(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid report ID"})
		return
	}

	report, err := c.scannerSvc.GetReport(reportID)
	if err != nil {
		if err == db.ErrNotFound {
			helpers.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Report not found"})
			return
		}
		log.WithError(err).Error("Failed to get report")
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if report.ArfPath == nil {
		helpers.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "ARF file not available"})
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.arf.xml\"", report.Host))

	// Serve the file
	http.ServeFile(w, r, *report.ArfPath)
}
