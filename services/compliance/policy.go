package compliance

import (
	"fmt"

	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/tz"
)

// PolicyService handles compliance policy operations
type PolicyService struct {
	store db.Store
}

// NewPolicyService creates a new policy service
func NewPolicyService(store db.Store) *PolicyService {
	return &PolicyService{
		store: store,
	}
}

// CreatePolicy creates a new compliance policy
func (s *PolicyService) CreatePolicy(projectID, userID int, name, profileID string, contentID int, attrs map[string]interface{}) (*db.CompliancePolicy, error) {
	// Validate that the content exists and has the specified profile
	profiles, err := s.store.GetScapProfilesByContent(contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles: %v", err)
	}

	var profileExists bool
	for _, profile := range profiles {
		if profile.ProfileID == profileID {
			profileExists = true
			break
		}
	}

	if !profileExists {
		return nil, fmt.Errorf("profile %s not found in content", profileID)
	}

	// Create policy
	policy := &db.CompliancePolicy{
		ProjectID: projectID,
		Name:      name,
		ContentID: contentID,
		ProfileID: profileID,
		Created:   tz.Now(),
		CreatedBy: userID,
	}

	// Set attributes
	if err := policy.SetAttrs(attrs); err != nil {
		return nil, fmt.Errorf("failed to set attributes: %v", err)
	}

	// Save to database
	if err := s.store.CreateCompliancePolicy(policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %v", err)
	}

	return policy, nil
}

// GetPoliciesByProject retrieves all policies for a project
func (s *PolicyService) GetPoliciesByProject(projectID int) ([]*db.CompliancePolicy, error) {
	return s.store.GetCompliancePoliciesByProject(projectID)
}

// GetPolicy retrieves a specific policy
func (s *PolicyService) GetPolicy(id int) (*db.CompliancePolicy, error) {
	return s.store.GetCompliancePolicy(id)
}

// UpdatePolicy updates an existing policy
func (s *PolicyService) UpdatePolicy(policy *db.CompliancePolicy) error {
	return s.store.UpdateCompliancePolicy(policy)
}

// DeletePolicy deletes a policy and its assignments
func (s *PolicyService) DeletePolicy(id int) error {
	// Delete policy assignments first
	if err := s.store.DeletePolicyAssignments(id); err != nil {
		return fmt.Errorf("failed to delete policy assignments: %v", err)
	}

	// Delete the policy
	if err := s.store.DeleteCompliancePolicy(id); err != nil {
		return fmt.Errorf("failed to delete policy: %v", err)
	}

	return nil
}

// AssignPolicy assigns a policy to targets (inventories, hosts, or groups)
func (s *PolicyService) AssignPolicy(policyID int, assignments []PolicyAssignmentRequest) error {
	// Delete existing assignments
	if err := s.store.DeletePolicyAssignments(policyID); err != nil {
		return fmt.Errorf("failed to delete existing assignments: %v", err)
	}

	// Create new assignments
	for _, assignment := range assignments {
		dbAssignment := &db.PolicyAssignment{
			PolicyID:   policyID,
			TargetType: assignment.TargetType,
			TargetID:   assignment.TargetID,
			Created:    tz.Now(),
		}

		if err := s.store.CreatePolicyAssignment(dbAssignment); err != nil {
			return fmt.Errorf("failed to create assignment: %v", err)
		}
	}

	return nil
}

// GetPolicyAssignments retrieves assignments for a policy
func (s *PolicyService) GetPolicyAssignments(policyID int) ([]*db.PolicyAssignment, error) {
	return s.store.GetPolicyAssignments(policyID)
}

// PolicyAssignmentRequest represents a request to assign a policy
type PolicyAssignmentRequest struct {
	TargetType string `json:"target_type"` // 'inventory', 'host', 'group'
	TargetID   int    `json:"target_id"`
}

// ResolvePolicyTargets resolves the actual hosts for a policy based on its assignments
func (s *PolicyService) ResolvePolicyTargets(policyID int) ([]string, error) {
	assignments, err := s.store.GetPolicyAssignments(policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy assignments: %v", err)
	}

	var hosts []string
	for _, assignment := range assignments {
		switch assignment.TargetType {
		case "inventory":
			// Get all hosts from the inventory
			inventory, err := s.store.GetInventory(assignment.TargetID, 0) // Assuming 0 as projectID for now
			if err != nil {
				return nil, fmt.Errorf("failed to get inventory %d: %v", assignment.TargetID, err)
			}

			// Parse inventory hosts
			inventoryHosts, err := s.parseInventoryHosts(&inventory)
			if err != nil {
				return nil, fmt.Errorf("failed to parse inventory hosts: %v", err)
			}

			hosts = append(hosts, inventoryHosts...)

		case "host":
			// Single host assignment - would need to implement host storage
			// For now, we'll assume hosts are identified by their ID as string
			hosts = append(hosts, fmt.Sprintf("host_%d", assignment.TargetID))

		case "group":
			// Group assignment - would need to implement group storage
			// For now, we'll assume groups contain hosts identified by their IDs
			hosts = append(hosts, fmt.Sprintf("group_%d", assignment.TargetID))
		}
	}

	return hosts, nil
}

// parseInventoryHosts parses hosts from an inventory
func (s *PolicyService) parseInventoryHosts(inventory *db.Inventory) ([]string, error) {
	// This is a simplified implementation
	// In a real scenario, you'd parse the inventory JSON/YAML to extract host information
	// For now, we'll return a placeholder
	return []string{"host1", "host2"}, nil
}
