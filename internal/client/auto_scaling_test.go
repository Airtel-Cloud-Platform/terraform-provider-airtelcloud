package client

import (
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestAutoScalingForm(t *testing.T) {
	req := &models.CreateASGRequest{
		Name: "app-asg", ImageID: 2435, FlavorID: 101, VPCID: "vpc", NetworkID: "network",
		SecurityGroupID: 3661, AvailabilityZone: "S1", ScalingGroupDesiredSize: 1,
		ScalingGroupMaxSize: 3, DesiredCount: 1, ScaleUpStepSize: 1, ScaleDownStepSize: 1,
		ScalingInterval: 2, CooloffPeriod: 10, TerminationPolicy: "oldest",
		BootVolumeSize: 100, Labels: []string{"app", "prod"},
		ScaleUpRules:   []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 60}},
		ScaleDownRules: []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
		VSConfig: &models.ASGVSConfig{
			LBServiceID: "lb-id", Name: "app-vs", VIPPortID: 75469, Protocol: "TCP",
			Port: 1234, RoutingAlgorithm: "ROUND_ROBIN", PoolName: "pool", PoolPort: 12,
			Interval: 5, Timeout: 16, XForwardedFor: true, MonitorProtocol: "TCP", MaxConn: 100,
		},
	}

	form, err := autoScalingForm(req)
	if err != nil {
		t.Fatalf("autoScalingForm() error = %v", err)
	}
	if form["policy"] != "metrics" {
		t.Fatalf("policy = %v, want metrics", form["policy"])
	}
	if form["labels"] != "app,prod" {
		t.Fatalf("labels = %v, want app,prod", form["labels"])
	}
	up, ok := form["scaleup_rules"].([]string)
	if !ok || len(up) != 1 || !strings.Contains(up[0], `"metric_type":"cpu"`) {
		t.Fatalf("scaleup_rules = %#v", form["scaleup_rules"])
	}
	vs, ok := form["vs_config"].(string)
	if !ok || !strings.Contains(vs, `"x_forwarded_for":true`) || !strings.Contains(vs, `"vip_port_id":75469`) {
		t.Fatalf("vs_config = %#v", form["vs_config"])
	}
}

func TestAutoScalingFormOmitsOptionalSources(t *testing.T) {
	form, err := autoScalingForm(&models.CreateASGRequest{})
	if err != nil {
		t.Fatalf("autoScalingForm() error = %v", err)
	}
	for _, field := range []string{"image_id", "snapshot_id", "vs_config", "labels"} {
		if _, ok := form[field]; ok {
			t.Fatalf("field %q unexpectedly present", field)
		}
	}
}
