package models

const (
	KubernetesProviderBYOH        = "BringYourOwnHost"
	KubernetesControlPlaneKamaji  = "Kamaji"
	KubernetesControlPlaneKubeadm = "Kubeadm"
	KubernetesDistributionCKP     = "CKP"

	KubernetesStateUnknown  = "Unknown"
	KubernetesStateCreating = "Creating"
	KubernetesStateDeleting = "Deleting"
	KubernetesStateReady    = "Ready"
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
	HostGroup      string `json:"hostGroup"`
	GroupName      string `json:"groupName,omitempty"`
	Count          int    `json:"count"`
	Subnet         string `json:"subnet,omitempty"`
	AZ             string `json:"az,omitempty"`
	OSDistribution string `json:"osDistribution,omitempty"`
}

// KubernetesCluster is the GET/list representation of a cluster.
type KubernetesCluster struct {
	Name                   string                 `json:"name,omitempty"`
	Cluster                string                 `json:"cluster,omitempty"`
	Project                string                 `json:"project,omitempty"`
	Desc                   string                 `json:"desc,omitempty"`
	K8sVersion             string                 `json:"k8sVersion,omitempty"`
	K8sInfo                *KubernetesKubeInfo    `json:"k8sInfo,omitempty"`
	CKPProvider            *KubernetesCKPProvider `json:"ckpProvider,omitempty"`
	State                  string                 `json:"state,omitempty"`
	IsDeleted              bool                   `json:"isDeleted,omitempty"`
	CreateTime             string                 `json:"createTime,omitempty"`
	CreatedBy              string                 `json:"createdBy,omitempty"`
	RequiredSchedulingTags map[string]string      `json:"requiredSchedulingTags,omitempty"`
}

// ClusterName returns the Terraform ID (cluster name).
func (c *KubernetesCluster) ClusterName() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Cluster
}

// KubernetesClusterList is GET collection / host-group cluster list.
type KubernetesClusterList struct {
	Count int                 `json:"count"`
	Items []KubernetesCluster `json:"items"`
}
