//go:build integration

package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// TestSecurityGroupIntegration_CreateGetDelete tests the full lifecycle of a security group
func TestSecurityGroupIntegration_CreateGetDelete(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Generate unique security group name
	sgName := fmt.Sprintf("test-sg-%d", time.Now().Unix())

	// Create security group
	createReq := &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	}

	t.Logf("Creating security group: %s", sgName)

	sg, err := client.CreateSecurityGroup(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created with ID: %d, UUID: %s", sg.ID, sg.UUID)

	// Verify created security group fields
	if sg.SecurityGroupName != sgName {
		t.Errorf("Expected security group name %s, got %s", sgName, sg.SecurityGroupName)
	}
	if sg.ID == 0 {
		t.Error("Security group ID should not be zero")
	}
	if sg.UUID == "" {
		t.Error("Security group UUID should not be empty")
	}

	// Cleanup: Delete security group at the end
	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		} else {
			t.Log("Security group deleted successfully")
		}
	}()

	// Get security group
	t.Logf("Getting security group: %d", sg.ID)
	fetchedSG, err := client.GetSecurityGroup(ctx, sg.ID)
	if err != nil {
		t.Fatalf("GetSecurityGroup failed: %v", err)
	}

	// Verify fetched security group
	if fetchedSG.ID != sg.ID {
		t.Errorf("Expected security group ID %d, got %d", sg.ID, fetchedSG.ID)
	}
	if fetchedSG.UUID != sg.UUID {
		t.Errorf("Expected security group UUID %s, got %s", sg.UUID, fetchedSG.UUID)
	}
	if fetchedSG.SecurityGroupName != sgName {
		t.Errorf("Expected security group name %s, got %s", sgName, fetchedSG.SecurityGroupName)
	}

	t.Log("GetSecurityGroup returned correct data")
}

// TestSecurityGroupIntegration_List tests listing security groups
func TestSecurityGroupIntegration_List(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization: %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	// List security groups
	sgs, err := client.ListSecurityGroupsDetailed(ctx)
	if err != nil {
		t.Fatalf("ListSecurityGroupsDetailed failed: %v", err)
	}

	t.Logf("Found %d security groups", len(sgs))

	// Log security group details
	for i, sg := range sgs {
		t.Logf("SG %d: ID=%d, UUID=%s, Name=%s, Status=%s, Rules=%d",
			i+1, sg.ID, sg.UUID, sg.SecurityGroupName, sg.Status, len(sg.Rules))
	}
}

// TestSecurityGroupIntegration_GetNonExistent tests getting a non-existent security group
func TestSecurityGroupIntegration_GetNonExistent(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	nonExistentID := 99999999

	t.Logf("Attempting to get non-existent security group: %d", nonExistentID)

	_, err := client.GetSecurityGroup(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent security group, got nil")
		return
	}

	// Check if it's a 404 error
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent security group")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// TestSecurityGroupIntegration_RuleLifecycle tests creating, getting, listing, and deleting a security group rule
func TestSecurityGroupIntegration_RuleLifecycle(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group first
	sgName := fmt.Sprintf("test-sg-rule-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	// Cleanup: Delete security group at the end
	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		} else {
			t.Log("Security group deleted successfully")
		}
	}()

	// Create ingress rule for SSH (tcp/22)
	ruleReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      "ingress",
		Protocol:       "tcp",
		PortRangeMin:   "22",
		PortRangeMax:   "22",
		RemoteIPPrefix: "0.0.0.0/0",
		Ethertype:      "IPv4",
	}

	t.Log("Creating SSH ingress rule")
	rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, ruleReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroupRule failed: %v", err)
	}

	t.Logf("Rule created: ID=%d, Direction=%s, Protocol=%s, PortMin=%s, PortMax=%s",
		rule.ID, rule.Direction, rule.Protocol, rule.PortRangeMin, rule.PortRangeMax)

	// Verify rule fields
	if rule.Direction != "ingress" {
		t.Errorf("Expected direction ingress, got %s", rule.Direction)
	}
	if rule.Protocol != "tcp" {
		t.Errorf("Expected protocol tcp, got %s", rule.Protocol)
	}

	// Get rule
	t.Logf("Getting rule: %d", rule.ID)
	fetchedRule, err := client.GetSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Fatalf("GetSecurityGroupRule failed: %v", err)
	}

	if fetchedRule.ID != rule.ID {
		t.Errorf("Expected rule ID %d, got %d", rule.ID, fetchedRule.ID)
	}

	// List rules
	t.Log("Listing rules")
	rules, err := client.ListSecurityGroupRules(ctx, sg.ID)
	if err != nil {
		t.Fatalf("ListSecurityGroupRules failed: %v", err)
	}

	t.Logf("Found %d rules", len(rules))

	// Delete rule
	t.Logf("Deleting rule: %d", rule.ID)
	err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Errorf("DeleteSecurityGroupRule failed: %v", err)
	} else {
		t.Log("Rule deleted successfully")
	}
}

// TestSecurityGroupIntegration_MultipleRules tests creating and managing multiple rules
func TestSecurityGroupIntegration_MultipleRules(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-multi-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	// Cleanup: Delete security group at the end
	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		} else {
			t.Log("Security group deleted successfully")
		}
	}()

	// Define rules: SSH, HTTP, HTTPS
	ruleConfigs := []struct {
		name string
		req  *models.CreateSecurityGroupRuleRequest
	}{
		{
			name: "SSH",
			req: &models.CreateSecurityGroupRuleRequest{
				Direction:      "ingress",
				Protocol:       "tcp",
				PortRangeMin:   "22",
				PortRangeMax:   "22",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
			},
		},
		{
			name: "HTTP",
			req: &models.CreateSecurityGroupRuleRequest{
				Direction:      "ingress",
				Protocol:       "tcp",
				PortRangeMin:   "80",
				PortRangeMax:   "80",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
			},
		},
		{
			name: "HTTPS",
			req: &models.CreateSecurityGroupRuleRequest{
				Direction:      "ingress",
				Protocol:       "tcp",
				PortRangeMin:   "443",
				PortRangeMax:   "443",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
			},
		},
	}

	// Create all rules
	var createdRules []*models.SecurityGroupRuleDetail
	for _, rc := range ruleConfigs {
		t.Logf("Creating %s rule", rc.name)
		rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, rc.req)
		if err != nil {
			t.Fatalf("CreateSecurityGroupRule (%s) failed: %v", rc.name, err)
		}
		t.Logf("%s rule created: ID=%d", rc.name, rule.ID)
		createdRules = append(createdRules, rule)
	}

	// List rules and verify count
	rules, err := client.ListSecurityGroupRules(ctx, sg.ID)
	if err != nil {
		t.Fatalf("ListSecurityGroupRules failed: %v", err)
	}

	t.Logf("Found %d rules (expected at least %d)", len(rules), len(ruleConfigs))

	if len(rules) < len(ruleConfigs) {
		t.Errorf("Expected at least %d rules, got %d", len(ruleConfigs), len(rules))
	}

	// Delete all created rules
	for i, rule := range createdRules {
		t.Logf("Deleting rule %d: ID=%d", i+1, rule.ID)
		err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroupRule (ID=%d) failed: %v", rule.ID, err)
		}
	}

	t.Log("All rules deleted successfully")
}

// TestSecurityGroupIntegration_EgressRule tests creating an egress TCP rule
func TestSecurityGroupIntegration_EgressRule(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-egress-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		}
	}()

	// Create egress TCP rule for port 443
	ruleReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      "egress",
		Protocol:       "tcp",
		PortRangeMin:   "443",
		PortRangeMax:   "443",
		RemoteIPPrefix: "0.0.0.0/0",
		Ethertype:      "IPv4",
	}

	t.Log("Creating egress HTTPS rule")
	rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, ruleReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroupRule failed: %v", err)
	}

	t.Logf("Rule created: ID=%d, Direction=%s, Protocol=%s, PortMin=%s, PortMax=%s",
		rule.ID, rule.Direction, rule.Protocol, rule.PortRangeMin, rule.PortRangeMax)

	if rule.Direction != "egress" {
		t.Errorf("Expected direction egress, got %s", rule.Direction)
	}
	if rule.Protocol != "tcp" {
		t.Errorf("Expected protocol tcp, got %s", rule.Protocol)
	}
	if rule.PortRangeMin != "443" {
		t.Errorf("Expected port_range_min 443, got %s", rule.PortRangeMin)
	}

	// Re-fetch and verify direction persists
	fetched, err := client.GetSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Fatalf("GetSecurityGroupRule failed: %v", err)
	}

	if fetched.Direction != "egress" {
		t.Errorf("Expected direction egress after re-fetch, got %s", fetched.Direction)
	}

	// Cleanup rule
	err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Errorf("DeleteSecurityGroupRule failed: %v", err)
	}

	t.Log("Egress rule test passed")
}

// TestSecurityGroupIntegration_UDPRule tests creating an ingress UDP rule
func TestSecurityGroupIntegration_UDPRule(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-udp-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		}
	}()

	// Create ingress UDP rule for DNS (port 53)
	ruleReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      "ingress",
		Protocol:       "udp",
		PortRangeMin:   "53",
		PortRangeMax:   "53",
		RemoteIPPrefix: "0.0.0.0/0",
		Ethertype:      "IPv4",
	}

	t.Log("Creating UDP DNS ingress rule")
	rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, ruleReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroupRule failed: %v", err)
	}

	t.Logf("Rule created: ID=%d, Protocol=%s, PortMin=%s, PortMax=%s",
		rule.ID, rule.Protocol, rule.PortRangeMin, rule.PortRangeMax)

	if rule.Protocol != "udp" {
		t.Errorf("Expected protocol udp, got %s", rule.Protocol)
	}
	if rule.PortRangeMin != "53" {
		t.Errorf("Expected port_range_min 53, got %s", rule.PortRangeMin)
	}
	if rule.PortRangeMax != "53" {
		t.Errorf("Expected port_range_max 53, got %s", rule.PortRangeMax)
	}

	// Cleanup rule
	err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Errorf("DeleteSecurityGroupRule failed: %v", err)
	}

	t.Log("UDP rule test passed")
}

// TestSecurityGroupIntegration_ICMPRule tests creating an ingress ICMP rule
func TestSecurityGroupIntegration_ICMPRule(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-icmp-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		}
	}()

	// Create ingress ICMP rule (no port ranges)
	ruleReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      "ingress",
		Protocol:       "icmp",
		RemoteIPPrefix: "0.0.0.0/0",
		Ethertype:      "IPv4",
	}

	t.Log("Creating ICMP ingress rule")
	rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, ruleReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroupRule failed: %v", err)
	}

	t.Logf("Rule created: ID=%d, Protocol=%s", rule.ID, rule.Protocol)

	if rule.Protocol != "icmp" {
		t.Errorf("Expected protocol icmp, got %s", rule.Protocol)
	}

	// Cleanup rule
	err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Errorf("DeleteSecurityGroupRule failed: %v", err)
	}

	t.Log("ICMP rule test passed")
}

// TestSecurityGroupIntegration_PortRange tests creating a rule with a port range
func TestSecurityGroupIntegration_PortRange(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-portrange-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		}
	}()

	// Create ingress TCP rule with port range 8000-9000
	ruleReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      "ingress",
		Protocol:       "tcp",
		PortRangeMin:   "8000",
		PortRangeMax:   "9000",
		RemoteIPPrefix: "0.0.0.0/0",
		Ethertype:      "IPv4",
	}

	t.Log("Creating TCP port range rule (8000-9000)")
	rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, ruleReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroupRule failed: %v", err)
	}

	t.Logf("Rule created: ID=%d, PortMin=%s, PortMax=%s", rule.ID, rule.PortRangeMin, rule.PortRangeMax)

	if rule.PortRangeMin != "8000" {
		t.Errorf("Expected port_range_min 8000, got %s", rule.PortRangeMin)
	}
	if rule.PortRangeMax != "9000" {
		t.Errorf("Expected port_range_max 9000, got %s", rule.PortRangeMax)
	}

	// Cleanup rule
	err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Errorf("DeleteSecurityGroupRule failed: %v", err)
	}

	t.Log("Port range rule test passed")
}

// TestSecurityGroupIntegration_RuleWithDescription tests creating a rule with a description
func TestSecurityGroupIntegration_RuleWithDescription(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-desc-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		}
	}()

	// Create ingress TCP rule for MySQL with description
	desc := "MySQL access from private network"
	ruleReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      "ingress",
		Protocol:       "tcp",
		PortRangeMin:   "3306",
		PortRangeMax:   "3306",
		RemoteIPPrefix: "10.0.0.0/8",
		Ethertype:      "IPv4",
		Description:    desc,
	}

	t.Log("Creating MySQL rule with description")
	rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, ruleReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroupRule failed: %v", err)
	}

	t.Logf("Rule created: ID=%d, Description=%s", rule.ID, rule.Description)

	if rule.Description != desc {
		t.Errorf("Expected description %q, got %q", desc, rule.Description)
	}

	// Re-fetch and verify description persists
	fetched, err := client.GetSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Fatalf("GetSecurityGroupRule failed: %v", err)
	}

	if fetched.Description != desc {
		t.Errorf("Expected description %q after re-fetch, got %q", desc, fetched.Description)
	}

	// Cleanup rule
	err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Errorf("DeleteSecurityGroupRule failed: %v", err)
	}

	t.Log("Rule with description test passed")
}

// TestSecurityGroupIntegration_RestrictedCIDR tests creating a rule with a restricted CIDR
func TestSecurityGroupIntegration_RestrictedCIDR(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-cidr-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		}
	}()

	// Create ingress TCP SSH rule with restricted CIDR
	ruleReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      "ingress",
		Protocol:       "tcp",
		PortRangeMin:   "22",
		PortRangeMax:   "22",
		RemoteIPPrefix: "10.0.0.0/8",
		Ethertype:      "IPv4",
	}

	t.Log("Creating SSH rule with restricted CIDR 10.0.0.0/8")
	rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, ruleReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroupRule failed: %v", err)
	}

	t.Logf("Rule created: ID=%d, RemoteIPPrefix=%s", rule.ID, rule.RemoteIPPrefix)

	if rule.RemoteIPPrefix != "10.0.0.0/8" {
		t.Errorf("Expected remote_ip_prefix 10.0.0.0/8, got %s", rule.RemoteIPPrefix)
	}

	// Cleanup rule
	err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
	if err != nil {
		t.Errorf("DeleteSecurityGroupRule failed: %v", err)
	}

	t.Log("Restricted CIDR rule test passed")
}

// TestSecurityGroupIntegration_GetNonExistentRule tests getting a non-existent security group rule
func TestSecurityGroupIntegration_GetNonExistentRule(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-noexist-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		}
	}()

	nonExistentRuleID := 99999999

	t.Logf("Attempting to get non-existent rule: %d", nonExistentRuleID)

	_, err = client.GetSecurityGroupRule(ctx, sg.ID, nonExistentRuleID)
	if err == nil {
		t.Error("Expected error for non-existent security group rule, got nil")
		return
	}

	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent security group rule")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// TestSecurityGroupIntegration_MultipleProtocolRules tests creating rules with different protocols
func TestSecurityGroupIntegration_MultipleProtocolRules(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create security group
	sgName := fmt.Sprintf("test-sg-multiproto-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}

	t.Logf("Security group created: ID=%d, UUID=%s", sg.ID, sg.UUID)

	defer func() {
		t.Logf("Deleting security group: %d", sg.ID)
		err := client.DeleteSecurityGroup(ctx, sg.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroup failed: %v", err)
		}
	}()

	// Define 4 rules: TCP/SSH ingress, UDP/DNS ingress, ICMP ingress, TCP/HTTPS egress
	ruleConfigs := []struct {
		name     string
		protocol string
		req      *models.CreateSecurityGroupRuleRequest
	}{
		{
			name:     "TCP-SSH-ingress",
			protocol: "tcp",
			req: &models.CreateSecurityGroupRuleRequest{
				Direction:      "ingress",
				Protocol:       "tcp",
				PortRangeMin:   "22",
				PortRangeMax:   "22",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
			},
		},
		{
			name:     "UDP-DNS-ingress",
			protocol: "udp",
			req: &models.CreateSecurityGroupRuleRequest{
				Direction:      "ingress",
				Protocol:       "udp",
				PortRangeMin:   "53",
				PortRangeMax:   "53",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
			},
		},
		{
			name:     "ICMP-ingress",
			protocol: "icmp",
			req: &models.CreateSecurityGroupRuleRequest{
				Direction:      "ingress",
				Protocol:       "icmp",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
			},
		},
		{
			name:     "TCP-HTTPS-egress",
			protocol: "tcp",
			req: &models.CreateSecurityGroupRuleRequest{
				Direction:      "egress",
				Protocol:       "tcp",
				PortRangeMin:   "443",
				PortRangeMax:   "443",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
			},
		},
	}

	// Create all rules
	var createdRules []*models.SecurityGroupRuleDetail
	for _, rc := range ruleConfigs {
		t.Logf("Creating %s rule", rc.name)
		rule, err := client.CreateSecurityGroupRule(ctx, sg.ID, rc.req)
		if err != nil {
			t.Fatalf("CreateSecurityGroupRule (%s) failed: %v", rc.name, err)
		}
		t.Logf("%s rule created: ID=%d, Protocol=%s", rc.name, rule.ID, rule.Protocol)
		createdRules = append(createdRules, rule)
	}

	// List rules and verify count
	rules, err := client.ListSecurityGroupRules(ctx, sg.ID)
	if err != nil {
		t.Fatalf("ListSecurityGroupRules failed: %v", err)
	}

	t.Logf("Found %d rules (expected at least 4)", len(rules))

	if len(rules) < 4 {
		t.Errorf("Expected at least 4 rules, got %d", len(rules))
	}

	// Verify each created rule's protocol matches expectations
	for i, rc := range ruleConfigs {
		if createdRules[i].Protocol != rc.protocol {
			t.Errorf("Rule %s: expected protocol %s, got %s", rc.name, rc.protocol, createdRules[i].Protocol)
		}
	}

	// Delete all created rules
	for i, rule := range createdRules {
		t.Logf("Deleting rule %d: ID=%d", i+1, rule.ID)
		err = client.DeleteSecurityGroupRule(ctx, sg.ID, rule.ID)
		if err != nil {
			t.Errorf("DeleteSecurityGroupRule (ID=%d) failed: %v", rule.ID, err)
		}
	}

	t.Log("Multiple protocol rules test passed")
}
