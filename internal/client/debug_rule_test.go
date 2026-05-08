//go:build integration

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestSecurityGroupIntegration_DebugRuleCreate(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	sgName := fmt.Sprintf("test-debug-rule-%d", time.Now().Unix())
	sg, err := client.CreateSecurityGroup(ctx, &models.CreateSecurityGroupRequest{
		SecurityGroupName: sgName,
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}
	t.Logf("SG created: ID=%d", sg.ID)
	defer func() {
		client.DeleteSecurityGroup(ctx, sg.ID)
	}()

	// Create a rule via bulk endpoint and observe ALL returned rules
	ruleReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      "ingress",
		Protocol:       "tcp",
		PortRangeMin:   "22",
		PortRangeMax:   "22",
		RemoteIPPrefix: "10.0.0.0/8",
		Ethertype:      "IPv4",
		Description:    "SSH test rule",
	}

	rulesJSON, _ := json.Marshal([]models.CreateSecurityGroupRuleRequest{*ruleReq})
	formData := map[string]interface{}{
		"security_group_data": string(rulesJSON),
	}

	var rules []models.SecurityGroupRuleDetail
	err = client.PostURLEncodedForm(ctx, fmt.Sprintf("/api/v1/networks/securitygroup/%d/bulksecuritygrouprule/", sg.ID), formData, &rules)
	if err != nil {
		t.Fatalf("Bulk create failed: %v", err)
	}

	t.Logf("Bulk create returned %d rules:", len(rules))
	for i, r := range rules {
		t.Logf("  Rule[%d]: ID=%d, Direction=%s, Protocol=%s, PortMin=%s, PortMax=%s, RemoteIP=%s, Desc=%s",
			i, r.ID, r.Direction, r.Protocol, r.PortRangeMin, r.PortRangeMax, r.RemoteIPPrefix, r.Description)
	}
}
