package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestKubernetesMetadataTypeName(t *testing.T) {
	t.Parallel()

	var resp resource.MetadataResponse
	(&KubernetesResource{}).Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "airtelcloud"}, &resp)
	if resp.TypeName != "airtelcloud_kubernetes" {
		t.Fatalf("TypeName = %q", resp.TypeName)
	}
}

func TestKubernetesSchemaRequiredFields(t *testing.T) {
	t.Parallel()

	var resp resource.SchemaResponse
	(&KubernetesResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	for _, name := range []string{"name", "kubernetes_version", "networking_name", "networking_version", "availability_zone", "os_distribution", "node_pools"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("missing attribute %s", name)
		}
		if !attr.IsRequired() {
			t.Fatalf("%s must be required", name)
		}
	}

	pools, ok := resp.Schema.Attributes["node_pools"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("node_pools is not a list nested attribute")
	}
	if !pools.NestedObject.Attributes["flavor"].IsRequired() {
		t.Fatal("node_pools.flavor must be required")
	}
	if !pools.NestedObject.Attributes["subnet_name"].IsRequired() {
		t.Fatal("node_pools.subnet_name must be required")
	}
	if _, ok := pools.NestedObject.Attributes["name"]; ok {
		t.Fatal("node_pools.name must not be a Terraform attribute")
	}
	if _, ok := resp.Schema.Attributes["master_nodes"]; ok {
		t.Fatal("master_nodes must not be a Terraform attribute")
	}
	if _, ok := resp.Schema.Attributes["master_host_group"]; ok {
		t.Fatal("master_host_group must not be a Terraform attribute")
	}
}

func TestBuildKubernetesCreateRequest(t *testing.T) {
	data := KubernetesResourceModel{
		Name:                 types.StringValue("test-kms"),
		Description:          types.StringValue("test"),
		K8sName:              types.StringValue("CKP"),
		KubernetesVersion:    types.StringValue("v1.33.7"),
		NetworkingName:       types.StringValue("calico"),
		NetworkingVersion:    types.StringValue("v3.30.6"),
		AvailabilityZone:     types.StringValue("S1"),
		OSDistribution:       types.StringValue("Ubuntu"),
		ProviderType:         types.StringValue(models.KubernetesProviderBYOH),
		ControlPlaneProvider: types.StringValue(models.KubernetesControlPlaneKamaji),
		NodePools: []KubernetesNodePoolModel{{
			Flavor:     types.StringValue("ccd.xLarge"),
			FlavorType: types.StringValue("Compute dense"),
			SubnetName: types.StringValue("app-subnet"),
			Count:      types.Int64Value(0),
			Autoscaling: &KubernetesAutoscalingModel{
				Enabled:  types.BoolValue(true),
				MaxNodes: types.Int64Value(2),
			},
			Labels: []KubernetesMetaOpModel{{
				Key:   types.StringValue("test"),
				Value: types.StringValue("test"),
				Op:    types.StringValue("MetaOpSet"),
			}},
			Annotations: []KubernetesMetaOpModel{{
				Key:   types.StringValue("new"),
				Value: types.StringValue("new"),
				Op:    types.StringValue("MetaOpSet"),
			}},
			Taints: []KubernetesTaintModel{{
				Key:    types.StringValue("test"),
				Value:  types.StringValue("test"),
				Effect: types.StringValue("PreferNoSchedule"),
				Op:     types.StringValue("MetaOpSet"),
			}},
		}},
	}

	req := buildKubernetesCreateRequest(&data, map[string]string{"md0": "c27614a2-0c1b-4eda-9377-dd9f8e24f3e3"})
	if req.Cluster != "test-kms" || req.K8sInfo.WorkerNodes != 1 {
		t.Fatalf("unexpected request: %+v", req)
	}
	if req.K8sInfo.K8sVersion != "v1.33.7" || req.K8sInfo.CNIName != "calico" || req.K8sInfo.CNIVersion != "v3.30.6" {
		t.Fatalf("unexpected k8sInfo: %+v", req.K8sInfo)
	}
	pool := req.CKPProvider.BYOH.NodePools["md0"]
	if pool.AZ != "S1" || pool.HostGroup != "ccd.xLarge" || pool.GroupName != "Compute dense" || pool.Subnet != "c27614a2-0c1b-4eda-9377-dd9f8e24f3e3" || pool.Count != 0 {
		t.Fatalf("unexpected pool: %+v", pool)
	}
	if pool.Autoscaling == nil || !pool.Autoscaling.Enabled || pool.Autoscaling.MaxNodes != 2 {
		t.Fatalf("unexpected autoscaling: %+v", pool.Autoscaling)
	}
	if len(pool.Labels) != 1 || pool.Labels[0].Key != "test" || pool.Labels[0].Op != "MetaOpSet" {
		t.Fatalf("unexpected labels: %+v", pool.Labels)
	}
	if len(pool.Annotations) != 1 || pool.Annotations[0].Key != "new" {
		t.Fatalf("unexpected annotations: %+v", pool.Annotations)
	}
	if len(pool.Taints) != 1 || pool.Taints[0].Effect != "TaintEffectPreferNoSchedule" {
		t.Fatalf("unexpected taints: %+v", pool.Taints)
	}
	if req.RequiredSchedulingTags["availabilityZone"] != "S1" {
		t.Fatalf("tags = %v", req.RequiredSchedulingTags)
	}
	if req.K8sInfo.MasterNodes != 3 || req.CKPProvider.BYOH.MasterHostGroup != "" {
		t.Fatalf("unexpected control-plane defaults: %+v", req)
	}
	if data.WorkerNodes.ValueInt64() != 1 {
		t.Fatalf("worker_nodes not filled: %v", data.WorkerNodes)
	}
}

func TestKubernetesTaintEffectAPIValue(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"NoSchedule":       "TaintEffectNoSchedule",
		"PreferNoSchedule": "TaintEffectPreferNoSchedule",
		"NoExecute":        "TaintEffectNoExecute",
		"":                 "TaintEffectPreferNoSchedule",
	}
	for input, want := range tests {
		if got := kubernetesTaintEffectAPIValue(input); got != want {
			t.Errorf("kubernetesTaintEffectAPIValue(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestKubernetesNodePoolAPIName(t *testing.T) {
	t.Parallel()
	if got := kubernetesNodePoolAPIName(0); got != "md0" {
		t.Fatalf("kubernetesNodePoolAPIName(0) = %q", got)
	}
	if got := kubernetesNodePoolAPIName(1); got != "md1" {
		t.Fatalf("kubernetesNodePoolAPIName(1) = %q", got)
	}
}
