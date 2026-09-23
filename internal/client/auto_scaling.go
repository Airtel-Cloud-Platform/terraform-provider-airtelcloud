package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func (c *Client) autoScalingBasePath() string {
	return fmt.Sprintf("/api/v2.1/autoscaling-group/domain/%s/project/%s/autoscaling-group/auto_scaling_group_with_lb",
		c.Organization, c.ProjectName)
}

func (c *Client) ListASGMetricConfigs(ctx context.Context) ([]models.ASGMetricConfig, error) {
	var configs []models.ASGMetricConfig
	path := fmt.Sprintf("/api/v2.1/autoscaling-group/domain/%s/project/%s/autoscaling-group/metric-configs",
		c.Organization, c.ProjectName)
	if err := c.Get(ctx, path, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

func (c *Client) ValidateASGMetricRules(ctx context.Context, rules []models.ASGScalingRule) error {
	configs, err := c.ListASGMetricConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list autoscaling metric configs: %w", err)
	}
	for _, rule := range rules {
		found := false
		for _, config := range configs {
			if config.MetricName == rule.MetricType && config.AggregationType == rule.AggregationType {
				if !config.IsActive {
					return fmt.Errorf("metric %q with aggregation %q is inactive", rule.MetricType, rule.AggregationType)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("metric %q with aggregation %q is not supported", rule.MetricType, rule.AggregationType)
		}
	}
	return nil
}

func (c *Client) CreateAutoScalingGroup(ctx context.Context, req *models.CreateASGRequest) (*models.AutoScalingGroup, error) {
	form, err := autoScalingForm(req)
	if err != nil {
		return nil, err
	}
	var group models.AutoScalingGroup
	if err := c.PostURLEncodedForm(ctx, c.autoScalingBasePath(), form, &group); err != nil {
		return nil, err
	}
	if group.ID == "" {
		return nil, fmt.Errorf("autoscaling create response contained no id")
	}
	return &group, nil
}

func autoScalingForm(req *models.CreateASGRequest) (map[string]interface{}, error) {
	form := map[string]interface{}{
		"policy":                     "metrics",
		"name":                       req.Name,
		"flavor_id":                  req.FlavorID,
		"vpc_id":                     req.VPCID,
		"network_id":                 req.NetworkID,
		"sec_group_id":               req.SecurityGroupID,
		"availability_zone":          req.AvailabilityZone,
		"keypair_id":                 req.KeypairID,
		"scaling_group_desired_size": req.ScalingGroupDesiredSize,
		"scaling_group_max_size":     req.ScalingGroupMaxSize,
		"desired_count":              req.DesiredCount,
		"scale_up_step_size":         req.ScaleUpStepSize,
		"scale_down_step_size":       req.ScaleDownStepSize,
		"scaling_interval":           req.ScalingInterval,
		"cooloff_period":             req.CooloffPeriod,
		"termination_policy":         req.TerminationPolicy,
		"drain_period":               req.DrainPeriod,
		"boot_vol_size":              req.BootVolumeSize,
	}
	if req.ImageID != 0 {
		form["image_id"] = req.ImageID
	}
	if req.SnapshotID != 0 {
		form["snapshot_id"] = req.SnapshotID
	}
	if len(req.Labels) > 0 {
		form["labels"] = strings.Join(req.Labels, ",")
	}
	up, err := marshalASGRules(req.ScaleUpRules)
	if err != nil {
		return nil, fmt.Errorf("encoding scaleup_rules: %w", err)
	}
	down, err := marshalASGRules(req.ScaleDownRules)
	if err != nil {
		return nil, fmt.Errorf("encoding scaledown_rules: %w", err)
	}
	form["scaleup_rules"] = up
	form["scaledown_rules"] = down
	if req.VSConfig != nil {
		raw, err := json.Marshal(req.VSConfig)
		if err != nil {
			return nil, fmt.Errorf("encoding vs_config: %w", err)
		}
		form["vs_config"] = string(raw)
	}
	return form, nil
}

func marshalASGRules(rules []models.ASGScalingRule) ([]string, error) {
	out := make([]string, 0, len(rules))
	for _, rule := range rules {
		raw, err := json.Marshal(rule)
		if err != nil {
			return nil, err
		}
		out = append(out, string(raw))
	}
	return out, nil
}

func (c *Client) GetAutoScalingGroup(ctx context.Context, id string) (*models.AutoScalingGroup, error) {
	var group models.AutoScalingGroup
	if err := c.Get(ctx, fmt.Sprintf("%s/%s", c.autoScalingBasePath(), id), &group); err != nil {
		return nil, err
	}
	return &group, nil
}

func (c *Client) DeleteAutoScalingGroup(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("%s/%s", c.autoScalingBasePath(), id))
}

func (c *Client) WaitForAutoScalingGroupReady(ctx context.Context, id string, timeout time.Duration) (*models.AutoScalingGroup, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		group, err := c.GetAutoScalingGroup(ctx, id)
		if err != nil {
			return nil, err
		}
		switch group.Status {
		case models.ASGStatusCreateComplete:
			return group, nil
		case models.ASGStatusCreateInProgress:
		default:
			return nil, fmt.Errorf("autoscaling group %s entered unexpected status %q", id, group.Status)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	return nil, fmt.Errorf("autoscaling group %s did not become ready within %v", id, timeout)
}

func (c *Client) WaitForAutoScalingGroupDeleted(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		group, err := c.GetAutoScalingGroup(ctx, id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		if group.Status != models.ASGStatusDeleteInProgress {
			return fmt.Errorf("autoscaling group %s entered unexpected delete status %q", id, group.Status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	return fmt.Errorf("autoscaling group %s was not deleted within %v", id, timeout)
}

func (c *Client) ResolveSnapshotID(ctx context.Context, name string) (int64, error) {
	snapshots, err := c.ListComputeSnapshots(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list snapshots: %w", err)
	}
	var id int64
	for _, snapshot := range snapshots {
		if snapshot.SnapshotName != name {
			continue
		}
		if id != 0 {
			return 0, fmt.Errorf("multiple snapshots found with name %q", name)
		}
		id = int64(snapshot.ID)
	}
	if id == 0 {
		return 0, fmt.Errorf("snapshot with name %q not found", name)
	}
	return id, nil
}
