package models

import "encoding/json"

const (
	ASGStatusCreateInProgress = "CREATE_IN_PROGRESS"
	ASGStatusCreateComplete   = "CREATE_COMPLETE"
	ASGStatusDeleteInProgress = "DELETE_IN_PROGRESS"
)

type ASGScalingRule struct {
	MetricType      string `json:"metric_type"`
	AggregationType string `json:"aggregation_type"`
	TargetValue     int64  `json:"target_value"`
	RuleType        string `json:"rule_type,omitempty"`
}

type ASGMetricConfig struct {
	MetricName      string `json:"metric_name"`
	IsActive        bool   `json:"is_active"`
	AggregationType string `json:"aggregation_type"`
}

type ASGVSConfig struct {
	LBServiceID        string `json:"lb_service_id"`
	Name               string `json:"name"`
	VIPPortID          int    `json:"vip_port_id"`
	Protocol           string `json:"protocol"`
	Port               int64  `json:"port"`
	RoutingAlgorithm   string `json:"routing_algorithm"`
	PoolName           string `json:"pool_name"`
	PoolPort           int64  `json:"pool_port"`
	Interval           int64  `json:"interval"`
	Timeout            int64  `json:"timeout"`
	XForwardedFor      bool   `json:"x_forwarded_for"`
	PersistenceEnabled bool   `json:"persistence_enabled"`
	MonitorProtocol    string `json:"monitor_protocol"`
	MaxConn            int64  `json:"max_conn"`
}

type CreateASGRequest struct {
	Name                    string
	ImageID                 int64
	SnapshotID              int64
	FlavorID                int64
	VPCID                   string
	NetworkID               string
	SecurityGroupID         int64
	AvailabilityZone        string
	KeypairID               string
	ScalingGroupDesiredSize int64
	ScalingGroupMaxSize     int64
	DesiredCount            int64
	ScaleUpStepSize         int64
	ScaleDownStepSize       int64
	ScalingInterval         int64
	CooloffPeriod           int64
	TerminationPolicy       string
	DrainPeriod             int64
	BootVolumeSize          int64
	ScaleUpRules            []ASGScalingRule
	ScaleDownRules          []ASGScalingRule
	Labels                  []string
	VSConfig                *ASGVSConfig
}

type ASGFlavor struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ASGImage struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
}

type ASGLabels struct {
	Labels []string `json:"labels"`
}

type ASGVirtualServer struct {
	Name      string `json:"name"`
	LBService struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"lb_service"`
	ASGPool struct {
		Members []struct {
			Port    int64 `json:"port"`
			MaxConn int64 `json:"max_conn"`
		} `json:"members"`
	} `json:"asg_pool"`
}

type AutoScalingGroup struct {
	ID                      string           `json:"id"`
	Name                    string           `json:"name"`
	Status                  string           `json:"status"`
	VPCID                   string           `json:"vpc_id"`
	NetworkID               string           `json:"network_id"`
	AZName                  string           `json:"az_name"`
	AvailabilityZone        string           `json:"availability_zone"`
	FlavorID                int64            `json:"flavor_id"`
	Flavor                  ASGFlavor        `json:"flavor"`
	ImageID                 int64            `json:"image_id"`
	Image                   ASGImage         `json:"image"`
	SecurityGroupID         int64            `json:"sec_group_id"`
	KeypairID               string           `json:"keypair_id"`
	ScalingGroupDesiredSize int64            `json:"scaling_group_desired_size"`
	ScalingGroupMaxSize     int64            `json:"scaling_group_max_size"`
	ScaleUpStepSize         int64            `json:"scale_up_step_size"`
	ScaleDownStepSize       int64            `json:"scale_down_step_size"`
	ScalingInterval         int64            `json:"scaling_interval"`
	CooloffPeriod           int64            `json:"cooloff_period"`
	TerminationPolicy       string           `json:"termination_policy"`
	DrainPeriod             int64            `json:"drain_period"`
	ScaleUpRulesJSON        string           `json:"scaleup_rules"`
	ScaleDownRulesJSON      string           `json:"scaledown_rules"`
	Labels                  ASGLabels        `json:"labels"`
	VirtualServer           ASGVirtualServer `json:"virtual_server"`
}

func (a *AutoScalingGroup) ScaleUpRules() ([]ASGScalingRule, error) {
	var rules []ASGScalingRule
	err := json.Unmarshal([]byte(a.ScaleUpRulesJSON), &rules)
	return rules, err
}

func (a *AutoScalingGroup) ScaleDownRules() ([]ASGScalingRule, error) {
	var rules []ASGScalingRule
	err := json.Unmarshal([]byte(a.ScaleDownRulesJSON), &rules)
	return rules, err
}
