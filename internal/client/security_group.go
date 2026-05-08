package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// securityGroupBasePath returns the base path for security group endpoints
func (c *Client) securityGroupBasePath() string {
	return "/api/v1/networks/securitygroup"
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

// GetSecurityGroup retrieves a security group by ID
func (c *Client) GetSecurityGroup(ctx context.Context, id int) (*models.SecurityGroupDetail, error) {
	var sg models.SecurityGroupDetail
	err := c.Get(ctx, fmt.Sprintf("%s/%d/", c.securityGroupBasePath(), id), &sg)
	if err != nil {
		return nil, err
	}
	return &sg, nil
}

// ListSecurityGroupsDetailed retrieves all security groups with full details
// Named to avoid collision with ListSecurityGroups in compute.go
func (c *Client) ListSecurityGroupsDetailed(ctx context.Context) ([]models.SecurityGroupDetail, error) {
	var sgs []models.SecurityGroupDetail
	err := c.Get(ctx, fmt.Sprintf("%s/", c.securityGroupBasePath()), &sgs)
	if err != nil {
		return nil, err
	}
	return sgs, nil
}

// DeleteSecurityGroup deletes a security group
func (c *Client) DeleteSecurityGroup(ctx context.Context, id int) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%d/", c.securityGroupBasePath(), id))
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
func (c *Client) CreateSecurityGroupRule(ctx context.Context, sgID int, req *models.CreateSecurityGroupRuleRequest) (*models.SecurityGroupRuleDetail, error) {
	// JSON-marshal the request as a single-element array
	rulesJSON, err := json.Marshal([]models.CreateSecurityGroupRuleRequest{*req})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rule request: %w", err)
	}

	formData := map[string]interface{}{
		"security_group_data": string(rulesJSON),
	}

	var rules []models.SecurityGroupRuleDetail
	err = c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%d/bulksecuritygrouprule/", c.securityGroupBasePath(), sgID), formData, &rules)
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

// GetSecurityGroupRule retrieves a specific rule from a security group
func (c *Client) GetSecurityGroupRule(ctx context.Context, sgID int, ruleID int) (*models.SecurityGroupRuleDetail, error) {
	var rule models.SecurityGroupRuleDetail
	err := c.Get(ctx, fmt.Sprintf("%s/%d/securitygrouprule/%d/", c.securityGroupBasePath(), sgID, ruleID), &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// ListSecurityGroupRules retrieves all rules for a security group
func (c *Client) ListSecurityGroupRules(ctx context.Context, sgID int) ([]models.SecurityGroupRuleDetail, error) {
	var rules []models.SecurityGroupRuleDetail
	err := c.Get(ctx, fmt.Sprintf("%s/%d/securitygrouprule/", c.securityGroupBasePath(), sgID), &rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// DeleteSecurityGroupRule deletes a rule from a security group
func (c *Client) DeleteSecurityGroupRule(ctx context.Context, sgID int, ruleID int) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%d/securitygrouprule/%d/", c.securityGroupBasePath(), sgID, ruleID))
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}
	return nil
}
