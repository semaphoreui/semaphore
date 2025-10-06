package projects

import (
	"net/http"
	"sort"
	"strings"

	"github.com/Digital-Data-Co/forge/api/helpers"
	"github.com/Digital-Data-Co/forge/db"
)

// GetTemplateTags returns all unique tags used in templates for a project
func GetTemplateTags(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	// Get all templates for the project
	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{}, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Extract unique tags
	tagMap := make(map[string]bool)
	for _, template := range templates {
		for _, tag := range template.Tags {
			if tag != "" {
				tagMap[tag] = true
			}
		}
	}

	// Convert to slice and sort
	var tags []string
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"tags": tags,
	})
}

// GetTemplateLabels returns all unique labels used in templates for a project
func GetTemplateLabels(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	// Get all templates for the project
	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{}, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Extract unique labels
	labelMap := make(map[string]bool)
	for _, template := range templates {
		for _, label := range template.Labels {
			if label != "" {
				labelMap[label] = true
			}
		}
	}

	// Convert to slice and sort
	var labels []string
	for label := range labelMap {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"labels": labels,
	})
}

// GetTemplateSearchSuggestions returns search suggestions based on templates
func GetTemplateSearchSuggestions(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	query := r.URL.Query().Get("q")
	
	if query == "" {
		helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"suggestions": []string{},
		})
		return
	}

	// Get all templates for the project
	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{}, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Build search suggestions
	suggestions := make(map[string]bool)
	queryLower := strings.ToLower(query)

	for _, template := range templates {
		// Search in name
		if strings.Contains(strings.ToLower(template.Name), queryLower) {
			suggestions[template.Name] = true
		}

		// Search in description
		if template.Description != nil && strings.Contains(strings.ToLower(*template.Description), queryLower) {
			suggestions[*template.Description] = true
		}

		// Search in tags
		for _, tag := range template.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				suggestions[tag] = true
			}
		}

		// Search in labels
		for _, label := range template.Labels {
			if strings.Contains(strings.ToLower(label), queryLower) {
				suggestions[label] = true
			}
		}
	}

	// Convert to slice and sort
	var suggestionList []string
	for suggestion := range suggestions {
		suggestionList = append(suggestionList, suggestion)
	}
	sort.Strings(suggestionList)

	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"suggestions": suggestionList,
	})
}
