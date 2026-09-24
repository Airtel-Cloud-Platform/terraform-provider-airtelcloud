package models

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalKubernetesCreateRequest(t *testing.T) {
	raw := []byte(`{
		"cluster":"test-kms",
		"desc":"test",
		"k8sInfo":{"k8sName":"CKP","k8sVersion":"v1.33.7","cniName":"calico","cniVersion":"v3.30.6","masterNodes":3,"workerNodes":1},
		"ckpProvider":{"provider":"BringYourOwnHost","byoh":{"masterHostGroup":"","nodePools":{"md0":{"hostGroup":"ccd.xLarge","groupName":"Compute dense","subnet":"c27614a2-0c1b-4eda-9377-dd9f8e24f3e3","osDistribution":"Ubuntu","az":"S1","count":0,"autoscaling":{"enabled":true,"maxNodes":2},"labels":[{"key":"test","value":"test","op":"MetaOpSet"}],"annotations":[{"key":"new","value":"new","op":"MetaOpSet"}],"taints":[{"key":"test","value":"test","effect":"TaintEffectPreferNoSchedule","op":"MetaOpSet"}]}},"controlPlaneProvider":"Kamaji"}},
		"requiredSchedulingTags":{"availabilityZone":"S1"}
	}`)

	var req CreateKubernetesClusterRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Cluster != "test-kms" || req.K8sInfo.K8sVersion != "v1.33.7" {
		t.Fatalf("unexpected identity: %+v", req)
	}
	pool, ok := req.CKPProvider.BYOH.NodePools["md0"]
	if !ok || pool.HostGroup != "ccd.xLarge" || pool.GroupName != "Compute dense" || pool.Count != 0 {
		t.Fatalf("unexpected node pool: %#v", req.CKPProvider.BYOH.NodePools)
	}
	if pool.Autoscaling == nil || !pool.Autoscaling.Enabled || pool.Autoscaling.MaxNodes != 2 {
		t.Fatalf("unexpected autoscaling: %#v", pool.Autoscaling)
	}
	if len(pool.Labels) != 1 || pool.Labels[0].Key != "test" || len(pool.Taints) != 1 {
		t.Fatalf("unexpected metadata: labels=%#v taints=%#v", pool.Labels, pool.Taints)
	}
	if req.RequiredSchedulingTags["availabilityZone"] != "S1" {
		t.Fatalf("unexpected tags: %v", req.RequiredSchedulingTags)
	}
}

func TestKubernetesClusterIsReady(t *testing.T) {
	for _, state := range []string{"Connected", "connected", "Ready", "Created"} {
		if !KubernetesClusterIsReady(state) {
			t.Fatalf("KubernetesClusterIsReady(%q) = false", state)
		}
	}
	for _, state := range []string{"Creating", "Unknown", "Deleting", ""} {
		if KubernetesClusterIsReady(state) {
			t.Fatalf("KubernetesClusterIsReady(%q) = true", state)
		}
	}
}

func TestKubernetesClusterNameFallback(t *testing.T) {
	c := KubernetesCluster{Cluster: "from-cluster"}
	if c.ClusterName() != "from-cluster" {
		t.Fatalf("ClusterName = %q", c.ClusterName())
	}
	c.Name = "from-name"
	if c.ClusterName() != "from-name" {
		t.Fatalf("ClusterName = %q", c.ClusterName())
	}
	c = KubernetesCluster{Key: KubernetesClusterKey{Cluster: "from-key"}}
	if c.ClusterName() != "from-key" {
		t.Fatalf("ClusterName = %q", c.ClusterName())
	}
}

func TestUnmarshalKubernetesClusterStatus(t *testing.T) {
	raw := []byte(`{
		"cluster": {
			"key": {"domain":"elements","project":"copper","cluster":"test-kms"},
			"desc":"terraform kubernetes test",
			"k8sInfo":{"k8sName":"CKP","k8sVersion":"v1.33.7","cniName":"calico","cniVersion":"v3.30.6","masterNodes":3,"workerNodes":1},
			"ckpProvider":{"provider":"BringYourOwnHost","byoh":{"nodePools":{"md0":{"hostGroup":"ccd.xLarge","count":0,"labels":{},"annotations":{},"taints":[]}}}},
			"isDeleted":false,
			"lifeCycleState":"Creating",
			"failedStateError":"",
			"createdBy":"vinay.patel@airtel.com"
		},
		"state":"Not Registered"
	}`)

	var status KubernetesClusterStatus
	if err := json.Unmarshal(raw, &status); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if status.State != KubernetesStateNotRegistered {
		t.Fatalf("state = %q", status.State)
	}
	if status.Cluster.ClusterName() != "test-kms" {
		t.Fatalf("cluster name = %q", status.Cluster.ClusterName())
	}
	if status.Cluster.LifeCycleState != KubernetesStateCreating {
		t.Fatalf("lifecycle state = %q", status.Cluster.LifeCycleState)
	}
}
