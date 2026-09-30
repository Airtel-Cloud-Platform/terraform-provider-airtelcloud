package models

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

const (
	KubernetesProviderBYOH        = "BringYourOwnHost"
	KubernetesControlPlaneKamaji  = "Kamaji"
	KubernetesControlPlaneKubeadm = "Kubeadm"
	KubernetesDistributionCKP     = "CKP"

	KubernetesStateUnknown       = "Unknown"
	KubernetesStateCreating      = "Creating"
	KubernetesStateDeleting      = "Deleting"
	KubernetesStateReady         = "Ready"
	KubernetesStateConnected     = "Connected"
	KubernetesStateCreated       = "Created"
	KubernetesStateNotRegistered = "Not Registered"
)

// CreateKubernetesClusterRequest is the POST body for creating a CKP cluster.
// Field names match the console and Cloud Compass Cluster Manager API.
type CreateKubernetesClusterRequest struct {
	Cluster                string                `json:"cluster"`
	Desc                   string                `json:"desc,omitempty"`
	ClusterType            string                `json:"clusterType,omitempty"`
	Registry               string                `json:"registry,omitempty"`
	K8sInfo                KubernetesKubeInfo    `json:"k8sInfo"`
	CKPProvider            KubernetesCKPProvider `json:"ckpProvider"`
	RequiredSchedulingTags map[string]string     `json:"requiredSchedulingTags,omitempty"`
}

// KubernetesKubeInfo is k8sInfo on create/get.
type KubernetesKubeInfo struct {
	K8sName        string `json:"k8sName,omitempty"`
	K8sVersion     string `json:"k8sVersion"`
	CNIName        string `json:"cniName,omitempty"`
	CNIVersion     string `json:"cniVersion,omitempty"`
	VirtualIP      string `json:"virtualIP,omitempty"`
	MasterNodes    int    `json:"masterNodes,omitempty"`
	WorkerNodes    int    `json:"workerNodes,omitempty"`
	HAEnabled      bool   `json:"haEnabled,omitempty"`
	PodNetworkCIDR string `json:"podNetworkCidr,omitempty"`
	ServiceCIDR    string `json:"serviceCidr,omitempty"`
}

// KubernetesCKPProvider is ckpProvider on create/get.
type KubernetesCKPProvider struct {
	Provider string                  `json:"provider"`
	BYOH     *KubernetesBYOHProvider `json:"byoh,omitempty"`
}

// KubernetesBYOHProvider is ckpProvider.byoh.
type KubernetesBYOHProvider struct {
	MasterHostGroup      string                        `json:"masterHostGroup"`
	VirtualIP            string                        `json:"virtualIP,omitempty"`
	WorkerHostGroup      string                        `json:"workerHostGroup,omitempty"`
	ControlPlaneProvider string                        `json:"controlPlaneProvider,omitempty"`
	NodePools            map[string]KubernetesNodePool `json:"nodePools,omitempty"`
}

// KubernetesNodePool is one BYOH node pool. groupName is sent by the console
// even though it is not in the swagger definition.
type KubernetesNodePool struct {
	HostGroup      string                 `json:"hostGroup"`
	GroupName      string                 `json:"groupName,omitempty"`
	Count          int                    `json:"count"`
	Subnet         string                 `json:"subnet,omitempty"`
	AZ             string                 `json:"az,omitempty"`
	OSDistribution string                 `json:"osDistribution,omitempty"`
	Autoscaling    *KubernetesAutoscaling `json:"autoscaling,omitempty"`
	Labels         KubernetesMetaOps      `json:"labels,omitempty"`
	Annotations    KubernetesMetaOps      `json:"annotations,omitempty"`
	Taints         []KubernetesTaintOp    `json:"taints,omitempty"`
}

// KubernetesAutoscaling is node pool autoscaling in the create payload.
type KubernetesAutoscaling struct {
	Enabled  bool `json:"enabled"`
	MaxNodes int  `json:"maxNodes"`
}

// KubernetesMetaOp is a label or annotation selector.
type KubernetesMetaOp struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	Op    string `json:"op,omitempty"`
}

// KubernetesMetaOps accepts both create payload arrays and the key/value object
// shape returned by the Compass status endpoint.
type KubernetesMetaOps []KubernetesMetaOp

func (m *KubernetesMetaOps) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) || bytes.Equal(data, []byte("{}")) {
		*m = nil
		return nil
	}
	if len(data) > 0 && data[0] == '[' {
		var items []KubernetesMetaOp
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		*m = items
		return nil
	}

	var values map[string]string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	items := make(KubernetesMetaOps, 0, len(keys))
	for _, key := range keys {
		items = append(items, KubernetesMetaOp{Key: key, Value: values[key]})
	}
	*m = items
	return nil
}

// KubernetesTaintOp is a node taint selector.
type KubernetesTaintOp struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect,omitempty"`
	Op     string `json:"op,omitempty"`
}

// KubernetesClusterKey identifies a cluster in Compass.
type KubernetesClusterKey struct {
	Domain  string `json:"domain,omitempty"`
	Project string `json:"project,omitempty"`
	Cluster string `json:"cluster,omitempty"`
}

// KubernetesCluster is the cluster object nested in the Compass status response.
type KubernetesCluster struct {
	Key                    KubernetesClusterKey   `json:"key"`
	Name                   string                 `json:"name,omitempty"`
	Cluster                string                 `json:"cluster,omitempty"`
	Project                string                 `json:"project,omitempty"`
	Desc                   string                 `json:"desc,omitempty"`
	K8sVersion             string                 `json:"k8sVersion,omitempty"`
	K8sInfo                *KubernetesKubeInfo    `json:"k8sInfo,omitempty"`
	CKPProvider            *KubernetesCKPProvider `json:"ckpProvider,omitempty"`
	State                  string                 `json:"state,omitempty"`
	Status                 string                 `json:"status,omitempty"`
	LifeCycleState         string                 `json:"lifeCycleState,omitempty"`
	FailedStateError       string                 `json:"failedStateError,omitempty"`
	IsDeleted              bool                   `json:"isDeleted,omitempty"`
	CreateTime             string                 `json:"createTime,omitempty"`
	CreatedBy              string                 `json:"createdBy,omitempty"`
	RequiredSchedulingTags map[string]string      `json:"requiredSchedulingTags,omitempty"`
}

// KubernetesClusterStatus is returned by the Compass cluster status endpoint.
// State is the registration state shown in the UI, while LifeCycleState tracks
// provisioning progress.
type KubernetesClusterStatus struct {
	Cluster KubernetesCluster `json:"cluster"`
	State   string            `json:"state"`
}

// ClusterName returns the Terraform ID (cluster name).
func (c *KubernetesCluster) ClusterName() string {
	if c.Name != "" {
		return c.Name
	}
	if c.Cluster != "" {
		return c.Cluster
	}
	return c.Key.Cluster
}

// LifecycleState is the cluster lifecycle value from state or status.
func (c *KubernetesCluster) LifecycleState() string {
	if c == nil {
		return ""
	}
	if s := strings.TrimSpace(c.State); s != "" {
		return s
	}
	return strings.TrimSpace(c.Status)
}

// KubernetesClusterIsReady reports whether the API lifecycle value means the
// cluster finished provisioning successfully. Compass reports Connected.
func KubernetesClusterIsReady(state string) bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "connected", "ready", "created", "active", "running":
		return true
	default:
		return false
	}
}

// KubernetesClusterIsFailed reports a terminal create failure.
func KubernetesClusterIsFailed(state string) bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "failed", "failure", "error", "timed_out", "timeout", "disconnected":
		return true
	default:
		return false
	}
}

// KubernetesClusterList is GET collection / host-group cluster list.
type KubernetesClusterList struct {
	Count int                 `json:"count"`
	Items []KubernetesCluster `json:"items"`
}
