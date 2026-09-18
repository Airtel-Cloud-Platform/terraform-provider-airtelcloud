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
		"ckpProvider":{"provider":"BringYourOwnHost","byoh":{"masterHostGroup":"","nodePools":{"md0":{"hostGroup":"ccd.xLarge","groupName":"Compute dense","subnet":"c27614a2-0c1b-4eda-9377-dd9f8e24f3e3","osDistribution":"Ubuntu","az":"S1","count":1}},"controlPlaneProvider":"Kamaji"}},
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
	if !ok || pool.HostGroup != "ccd.xLarge" || pool.GroupName != "Compute dense" || pool.Count != 1 {
		t.Fatalf("unexpected node pool: %#v", req.CKPProvider.BYOH.NodePools)
	}
	if req.RequiredSchedulingTags["availabilityZone"] != "S1" {
		t.Fatalf("unexpected tags: %v", req.RequiredSchedulingTags)
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
}
