package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// securityGroupBasePath returns the base path for security group endpoints
func (c *Client) securityGroupBasePath() string {
	return fmt.Sprintf("/api/v2.1/networks/domain/%s/project/%s/networks/security-groups",
		c.Organization, c.ProjectName)
}

// CreateSecurityGroup creates a new security group
func (c *Client) CreateSecurityGroup(ctx context.Context, req *models.CreateSecurityGroupRequest) (*models.SecurityGroupDetail, error) {
	formData := structToFormData(req)

	var sg models.SecurityGroupDetail
	err := c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/", c.securityGroupBasePath()), formData, &sg)
	if err != nil {
		return nil, err
	}

	return &sg, nil
}

// GetSecurityGroup retrieves a security group by UUID.
// The v2.1 API addresses security groups by UUID in the path; numeric IDs return 404.
func (c *Client) GetSecurityGroup(ctx context.Context, uuid string) (*models.SecurityGroupDetail, error) {
	var sg models.SecurityGroupDetail
	err := c.Get(ctx, fmt.Sprintf("%s/%s/", c.securityGroupBasePath(), uuid), &sg)
	if err != nil {
		return nil, err
	}
	return &sg, nil
}

// ListSecurityGroupsDetailed retrieves all security groups with full details
func (c *Client) ListSecurityGroupsDetailed(ctx context.Context) ([]models.SecurityGroupDetail, error) {
	var sgs []models.SecurityGroupDetail
	err := c.Get(ctx, fmt.Sprintf("%s/", c.securityGroupBasePath()), &sgs)
	if err != nil {
		return nil, err
	}
	return sgs, nil
}

// ResolveSecurityGroupID resolves a security group name to its ID
func (c *Client) ResolveSecurityGroupID(ctx context.Context, name string) (int, error) {
	groups, err := c.ListSecurityGroupsDetailed(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list security groups: %w", err)
	}
	for _, sg := range groups {
		if sg.SecurityGroupName == name {
			return sg.ID, nil
		}
	}
	return 0, fmt.Errorf("security group with name %q not found", name)
}

// ResolveSecurityGroupUUID resolves a security group's numeric ID to its UUID.
// The v2.1 API addresses security groups by UUID, but Terraform tracks the numeric ID,
// so we look it up via the list endpoint (which takes no per-group identifier). Returns a
// 404 APIError when no group matches, so callers can treat it as a missing resource.
func (c *Client) ResolveSecurityGroupUUID(ctx context.Context, id int) (string, error) {
	groups, err := c.ListSecurityGroupsDetailed(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list security groups: %w", err)
	}
	for _, sg := range groups {
		if sg.ID == id {
			return sg.UUID, nil
		}
	}
	return "", &APIError{StatusCode: 404, Message: fmt.Sprintf("security group with id %d not found", id)}
}

// DeleteSecurityGroup deletes a security group by UUID.
func (c *Client) DeleteSecurityGroup(ctx context.Context, uuid string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s/", c.securityGroupBasePath(), uuid))
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}
	return nil
}

// CreateSecurityGroupRule creates a new rule in a security group using the bulk endpoint.
// The bulk endpoint returns ALL rules for the security group (including system defaults),
// so we find the newly created rule by selecting the one with the highest ID.
func (c *Client) CreateSecurityGroupRule(ctx context.Context, sgUUID string, req *models.CreateSecurityGroupRuleRequest) (*models.SecurityGroupRuleDetail, error) {
	// The v2.1 bulk endpoint addresses the security group by UUID in the path, not by
	// numeric ID. Passing the numeric ID makes the backend return 500 "Failed to create rules".

	// JSON-marshal the request as a single-element array
	rulesJSON, err := json.Marshal([]models.CreateSecurityGroupRuleRequest{*req})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rule request: %w", err)
	}

	formData := map[string]interface{}{
		"security_group_data": string(rulesJSON),
	}

	var rules []models.SecurityGroupRuleDetail
	err = c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/bulk-security-group-rules/", c.securityGroupBasePath(), sgUUID), formData, &rules)
	if err != nil {
		return nil, err
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("bulk create returned empty response")
	}

	// Find the rule with the highest ID — that's the newly created one
	best := &rules[0]
	for i := range rules {
		if rules[i].ID > best.ID {
			best = &rules[i]
		}
	}

	return best, nil
}

// GetSecurityGroupRule retrieves a specific rule from a security group by UUID.
// The v2.1 API returns rules embedded in the security group detail (addressed by UUID),
// so we fetch the group and locate the rule by its UUID. Returns a 404 APIError when the
// rule is absent, so callers can treat it as a missing resource.
func (c *Client) GetSecurityGroupRule(ctx context.Context, sgUUID, ruleUUID string) (*models.SecurityGroupRuleDetail, error) {
	sg, err := c.GetSecurityGroup(ctx, sgUUID)
	if err != nil {
		return nil, err
	}
	for i := range sg.Rules {
		if sg.Rules[i].UUID == ruleUUID {
			return &sg.Rules[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("security group rule %s not found", ruleUUID)}
}

// ListSecurityGroupRules retrieves all rules for a security group by UUID.
// The v2.1 API returns rules embedded in the security group detail.
func (c *Client) ListSecurityGroupRules(ctx context.Context, sgUUID string) ([]models.SecurityGroupRuleDetail, error) {
	sg, err := c.GetSecurityGroup(ctx, sgUUID)
	if err != nil {
		return nil, err
	}
	return sg.Rules, nil
}

// DeleteSecurityGroupRule deletes a rule from a security group by UUID.
func (c *Client) DeleteSecurityGroupRule(ctx context.Context, sgUUID, ruleUUID string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s/security-group-rules/%s/", c.securityGroupBasePath(), sgUUID, ruleUUID))
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}
	return nil
}
