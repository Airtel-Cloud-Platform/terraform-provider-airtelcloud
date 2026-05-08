package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// MockServer represents a mock HTTP server for testing
type MockServer struct {
	*httptest.Server
	Handlers map[string]http.HandlerFunc
}

// NewMockServer creates a new mock server with predefined handlers
func NewMockServer() *MockServer {
	ms := &MockServer{
		Handlers: make(map[string]http.HandlerFunc),
	}

	// Set up default handlers
	ms.setupDefaultHandlers()

	// Create the actual HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", ms.routeHandler)
	ms.Server = httptest.NewServer(mux)

	return ms
}

// setupDefaultHandlers sets up default API response handlers
func (ms *MockServer) setupDefaultHandlers() {
	// Compute handlers (v2.1 API)
	ms.Handlers["POST /api/v2.1/computes/domain/test-org/project/test-project/computes/"] = ms.createComputeHandler
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/"] = ms.getComputeHandler
	ms.Handlers["PUT /api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/"] = ms.updateComputeHandler
	ms.Handlers["DELETE /api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/"] = ms.deleteComputeHandler

	// Compute resize handler
	ms.Handlers["POST /api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/resize/2"] = ms.resizeComputeHandler

	// Volume handlers (v2.1 API)
	ms.Handlers["POST /api/v2.1/volumes/domain/test-org/project/test-project/volumes/create-and-attach/"] = ms.createVolumeHandler
	ms.Handlers["GET /api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/"] = ms.getVolumeHandler
	ms.Handlers["PUT /api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/"] = ms.updateVolumeHandler
	ms.Handlers["DELETE /api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/"] = ms.deleteVolumeHandler
	ms.Handlers["POST /api/v2.1/volumes/domain/test-org/project/test-project/volumes/volume_attach/test-uuid-1/"] = ms.attachVolumeHandler
	ms.Handlers["POST /api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/volume_detach/"] = ms.detachVolumeHandler
	ms.Handlers["POST /api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/snapshots/"] = ms.createSnapshotHandler

	// Volume types handler
	ms.Handlers["GET /api/v2.1/volumes/domain/test-org/project/test-project/volumes/volume_types"] = ms.listVolumeTypesHandler

	// List handlers
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/"] = ms.listComputesHandler
	ms.Handlers["GET /api/v2.1/volumes/domain/test-org/project/test-project/volumes/"] = ms.listVolumesHandler
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/flavors/"] = ms.listFlavorsHandler
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/images/"] = ms.listImagesHandler
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/keypairs/"] = ms.listKeypairsHandler
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/security-groups/"] = ms.listSecurityGroupsHandler
	ms.Handlers["GET /api/network-manager/v1/domain/test-org/project/test-project/networks"] = ms.listVPCsHandler
	ms.Handlers["GET /api/network-manager/v1/domain/test-org/project/test-project/network/test-network-id/subnets"] = ms.listSubnetsHandler

	// VPC Peering handlers
	ms.Handlers["POST /api/network-manager/v1/domain/test-org/project/test-project/vpc-peering"] = ms.createVPCPeeringHandler
	ms.Handlers["GET /api/network-manager/v1/domain/test-org/project/test-project/vpc-peering/test-peering-id"] = ms.getVPCPeeringHandler
	ms.Handlers["GET /api/network-manager/v1/domain/test-org/project/test-project/vpc-peerings"] = ms.listVPCPeeringsHandler
	ms.Handlers["DELETE /api/network-manager/v1/domain/test-org/project/test-project/vpc-peering/test-peering-id"] = ms.deleteVPCPeeringHandler

	// Compute Snapshot handlers
	ms.Handlers["POST /api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/snapshot/"] = ms.createComputeSnapshotHandler
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/snapshot/snap-uuid-1234/"] = ms.getComputeSnapshotHandler
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/snapshot/"] = ms.listComputeSnapshotsHandler
	ms.Handlers["DELETE /api/v2.1/computes/domain/test-org/project/test-project/computes/snapshot/snap-uuid-1234/"] = ms.deleteComputeSnapshotHandler

	// Protection handlers
	ms.Handlers["POST /api/v2.1/backups/domain/test-org/project/test-project/backups/protections/"] = ms.createProtectionHandler
	ms.Handlers["GET /api/v2.1/backups/domain/test-org/project/test-project/backups/protections/1/"] = ms.getProtectionHandler
	ms.Handlers["GET /api/v2.1/backups/domain/test-org/project/test-project/backups/protections/"] = ms.listProtectionsHandler
	ms.Handlers["PUT /api/v2.1/backups/domain/test-org/project/test-project/backups/protections/1/"] = ms.updateProtectionHandler
	ms.Handlers["DELETE /api/v2.1/backups/domain/test-org/project/test-project/backups/protections/1/"] = ms.deleteProtectionHandler

	// Protection Plan handlers
	ms.Handlers["POST /api/v2.1/backups/domain/test-org/project/test-project/backups/protection_plans/"] = ms.createProtectionPlanHandler
	ms.Handlers["GET /api/v2.1/backups/domain/test-org/project/test-project/backups/protection_plan/1"] = ms.getProtectionPlanHandler
	ms.Handlers["GET /api/v2.1/backups/domain/test-org/project/test-project/backups/protection_plans/"] = ms.listProtectionPlansHandler

	// LB Service handlers
	ms.Handlers["POST /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb_service/"] = ms.createLBServiceHandler
	ms.Handlers["GET /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb_service/lb-svc-1"] = ms.getLBServiceHandler
	ms.Handlers["GET /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb_service/"] = ms.listLBServicesHandler
	ms.Handlers["DELETE /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb_service/lb-svc-1"] = ms.deleteLBServiceHandler

	// LB VIP handlers (v1 API - flat path, org/project in headers)
	ms.Handlers["POST /api/v1/load-balancers/lb_service/lb-svc-1/vip"] = ms.createLBVipHandler
	ms.Handlers["GET /api/v1/load-balancers/lb_service/lb-svc-1/vip"] = ms.listLBVipsHandler
	ms.Handlers["DELETE /api/v1/load-balancers/lb_service/lb-svc-1/vip/1/"] = ms.deleteLBVipHandler

	// LB Certificate handlers
	ms.Handlers["POST /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/certificates"] = ms.createLBCertificateHandler
	ms.Handlers["GET /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/certificates"] = ms.listLBCertificatesHandler
	ms.Handlers["DELETE /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/certificates/1/"] = ms.deleteLBCertificateHandler

	// Public IP (IPAM) handlers
	ms.Handlers["POST /api/v1/ipam/domain/test-org/project/test-project"] = ms.createPublicIPHandler
	ms.Handlers["GET /api/v1/ipam/domain/test-org/project/test-project/test-public-ip-uuid"] = ms.getPublicIPHandler
	ms.Handlers["GET /api/v1/ipam/domain/test-org/project/test-project"] = ms.listPublicIPsHandler
	ms.Handlers["DELETE /api/v1/ipam/domain/test-org/project/test-project/test-public-ip-uuid"] = ms.deletePublicIPHandler

	// Public IP Policy Rule handlers
	ms.Handlers["POST /api/v1/admin/ipam_vip/nat_rule"] = ms.createPublicIPPolicyRuleHandler
	ms.Handlers["POST /api/v1/admin/ipam_vip/vip_object"] = ms.mapPublicIPHandler
	ms.Handlers["GET /api/v1/admin/ipam_vip/test-public-ip-uuid/rules"] = ms.listPublicIPPolicyRulesHandler
	ms.Handlers["DELETE /api/v1/admin/ipam_vip/nat_rule/test-public-ip-uuid-1"] = ms.deletePublicIPPolicyRuleHandler
	ms.Handlers["GET /api/v1/admin/ipam_vip/ipam_port"] = ms.listIPAMServicesHandler

	// LB Virtual Server handlers
	ms.Handlers["POST /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers"] = ms.createVirtualServerHandler
	ms.Handlers["GET /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers/vs-1"] = ms.getVirtualServerHandler
	ms.Handlers["GET /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers"] = ms.listVirtualServersHandler
	ms.Handlers["PATCH /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers/vs-1"] = ms.updateVirtualServerHandler
	ms.Handlers["DELETE /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers/vs-1"] = ms.deleteVirtualServerHandler

	// LB Flavor handler (shares compute flavors path, dispatched by ?type=lb query param)

	// Security Group handlers (network API v1)
	ms.Handlers["POST /api/v1/networks/securitygroup/"] = ms.createSecurityGroupHandler
	ms.Handlers["GET /api/v1/networks/securitygroup/1/"] = ms.getSecurityGroupHandler
	ms.Handlers["GET /api/v1/networks/securitygroup/"] = ms.listSecurityGroupsDetailedHandler
	ms.Handlers["DELETE /api/v1/networks/securitygroup/1/"] = ms.deleteSecurityGroupHandler

	// Security Group Rule handlers (integer ID based paths matching Swagger API)
	ms.Handlers["POST /api/v1/networks/securitygroup/1/bulksecuritygrouprule/"] = ms.createSecurityGroupRuleHandler
	ms.Handlers["GET /api/v1/networks/securitygroup/1/securitygrouprule/10/"] = ms.getSecurityGroupRuleHandler
	ms.Handlers["GET /api/v1/networks/securitygroup/1/securitygrouprule/"] = ms.listSecurityGroupRulesHandler
	ms.Handlers["DELETE /api/v1/networks/securitygroup/1/securitygrouprule/10/"] = ms.deleteSecurityGroupRuleHandler
}

// routeHandler routes requests to the appropriate handler
func (ms *MockServer) routeHandler(w http.ResponseWriter, r *http.Request) {
	key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

	if handler, exists := ms.Handlers[key]; exists {
		handler(w, r)
		return
	}

	// Default 404 response
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
}

// Compute handlers
func (ms *MockServer) createComputeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := []models.Compute{{
		ID:                 "test-id",
		ProviderInstanceID: "provider-test-id",
		InstanceName:       "test-instance",
		InstanceType:       "t2.micro",
		ImageID:            "ubuntu-20.04",
		NetworkID:          "network-1",
		Status:             "ACTIVE",
		PublicIPs:          "192.168.1.100",
		FloatingIP:         "10.0.0.100",
		AvailabilityZone:   "south-1a",
		VolumeSize:         20,
		Ports:              []models.Port{{FixedIPs: []string{"10.0.0.100"}}},
	}}

	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) getComputeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	compute := models.Compute{
		ID:                 "test-id",
		ProviderInstanceID: "provider-test-id",
		InstanceName:       "test-instance",
		InstanceType:       "t2.micro",
		ImageID:            "ubuntu-20.04",
		NetworkID:          "network-1",
		Status:             "ACTIVE",
		PublicIPs:          "192.168.1.100",
		FloatingIP:         "10.0.0.100",
		AvailabilityZone:   "south-1a",
		VolumeSize:         20,
		Ports:              []models.Port{{FixedIPs: []string{"10.0.0.100"}}},
	}

	json.NewEncoder(w).Encode(compute)
}

func (ms *MockServer) updateComputeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	compute := models.Compute{
		ID:                 "test-id",
		ProviderInstanceID: "provider-test-id",
		InstanceName:       "updated-instance",
		InstanceType:       "t2.small",
		ImageID:            "ubuntu-20.04",
		NetworkID:          "network-1",
		Status:             "ACTIVE",
		PublicIPs:          "192.168.1.100",
		FloatingIP:         "10.0.0.100",
		AvailabilityZone:   "south-1a",
		VolumeSize:         20,
		Ports:              []models.Port{{FixedIPs: []string{"10.0.0.100"}}},
	}

	json.NewEncoder(w).Encode(compute)
}

func (ms *MockServer) deleteComputeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	// After deletion, subsequent GET requests should return 404
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (ms *MockServer) resizeComputeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Volume handlers
func (ms *MockServer) createVolumeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	volume := models.Volume{
		ID:               1,
		UUID:             "test-uuid-1",
		ProviderVolumeID: "provider-vol-id",
		VolumeName:       "test-volume",
		VolumeSize:       10,
		Status:           "available",
		AvailabilityZone: "south-1a",
		VolumeTypeID:     json.RawMessage(`"gp2"`),
		VolumeType:       models.VolumeType{ID: 1, Name: "gp2"},
	}

	json.NewEncoder(w).Encode(volume)
}

func (ms *MockServer) getVolumeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	volume := models.Volume{
		ID:               1,
		UUID:             "test-uuid-1",
		ProviderVolumeID: "provider-vol-id",
		VolumeName:       "test-volume",
		VolumeSize:       10,
		Status:           "available",
		AvailabilityZone: "south-1a",
		VolumeTypeID:     json.RawMessage(`"gp2"`),
		VolumeAttachments: []models.VolumeAttachment{{
			ID:                         1,
			VolumeID:                   json.RawMessage(`"1"`),
			ComputeID:                  "test-id",
			Compute:                    json.RawMessage(`"test-instance"`),
			ProviderVolumeAttachmentID: "attach-id",
			VolumeAttachmentDeviceName: "/dev/sdb",
		}},
		VolumeType: models.VolumeType{
			ID:   1,
			Name: "gp2",
		},
	}

	json.NewEncoder(w).Encode(volume)
}

func (ms *MockServer) updateVolumeHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid form data"})
		return
	}

	// Validate required fields are present
	requiredFields := []string{"volume_name", "volume_size", "volume_type", "billing_unit"}
	for _, field := range requiredFields {
		if r.FormValue(field) == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("missing required field: %s", field)})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (ms *MockServer) deleteVolumeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	// After deletion, subsequent GET requests should return 404
	ms.Handlers["GET /api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (ms *MockServer) attachVolumeHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid form data"})
		return
	}

	if r.FormValue("compute_id") == "" || r.FormValue("volume_id") == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "compute_id and volume_id are required"})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ms *MockServer) detachVolumeHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid form data"})
		return
	}

	if r.FormValue("compute_id") == "" || r.FormValue("volume_id") == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "compute_id and volume_id are required"})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ms *MockServer) createSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid form data"})
		return
	}

	if r.FormValue("snapshot_name") == "" || r.FormValue("billing_unit") == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "snapshot_name and billing_unit are required"})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// List handlers

func (ms *MockServer) listComputesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := []models.Compute{
		{
			ID:           "compute-1",
			InstanceName: "instance-1",
			InstanceType: "t2.micro",
			Status:       "ACTIVE",
			Ports:        []models.Port{{ID: 101, FixedIPs: []string{"10.1.99.172"}}},
		},
		{
			ID:           "compute-2",
			InstanceName: "instance-2",
			InstanceType: "t2.small",
			Status:       "ACTIVE",
			Ports:        []models.Port{{ID: 202, FixedIPs: []string{"10.1.99.200"}}},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) listVolumesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	computeID := r.URL.Query().Get("compute_id")
	volumes := []models.Volume{
		{
			ID:         1,
			UUID:       "test-uuid-1",
			VolumeName: "vol-1",
			VolumeSize: 10,
			Status:     "available",
		},
		{
			ID:         2,
			UUID:       "test-uuid-2",
			VolumeName: "vol-2",
			VolumeSize: 20,
			Status:     "in-use",
			VolumeAttachments: []models.VolumeAttachment{{
				ComputeID: "test-compute-id",
			}},
		},
	}
	if computeID != "" {
		var filtered []models.Volume
		for _, v := range volumes {
			for _, a := range v.VolumeAttachments {
				if a.ComputeID == computeID {
					filtered = append(filtered, v)
					break
				}
			}
		}
		json.NewEncoder(w).Encode(filtered)
		return
	}
	json.NewEncoder(w).Encode(volumes)
}

func (ms *MockServer) listFlavorsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Query().Get("type") == "lb" {
		flavors := []models.LBFlavor{
			{ID: 1, Name: "lb-small"},
			{ID: 2, Name: "lb-medium"},
		}
		json.NewEncoder(w).Encode(flavors)
		return
	}
	response := []models.Flavor{
		{ID: 1, Name: "t2.micro", VCPU: 1, RAM: 1024, Disk: 20},
		{ID: 2, Name: "t2.small", VCPU: 2, RAM: 2048, Disk: 40},
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) listImagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := []models.Image{
		{ID: 1, Name: "ubuntu-20.04", OSType: "linux"},
		{ID: 2, Name: "centos-8", OSType: "linux"},
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) listKeypairsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := []models.Keypair{
		{ID: 1, Name: "keypair-1", Fingerprint: "aa:bb:cc:dd"},
		{ID: 2, Name: "keypair-2", Fingerprint: "ee:ff:00:11"},
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) listSecurityGroupsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := []models.SecurityGroup{
		{ID: 1, Name: "default"},
		{ID: 2, Name: "web-sg"},
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) listVPCsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Return raw JSON matching real API format (networkId field) instead of
	// relying on Go struct serialization, so tests catch JSON tag mismatches.
	response := `{
		"count": 2,
		"items": [
			{"networkId": "vpc-1", "name": "vpc-default", "cidr_block": "10.0.0.0/16"},
			{"networkId": "vpc-2", "name": "vpc-custom", "cidr_block": "172.16.0.0/16"}
		]
	}`
	w.Write([]byte(response))
}

func (ms *MockServer) listSubnetsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := models.SubnetListResponse{
		Count: 2,
		Items: []models.Subnet{
			{SubnetID: "subnet-1", Name: "subnet-a", IPv4AddressSpace: "10.0.1.0/24"},
			{SubnetID: "subnet-2", Name: "subnet-b", IPv4AddressSpace: "10.0.2.0/24"},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) listVolumeTypesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	allTypes := []models.VolumeType{
		{ID: 1, Name: "SSD", Group: "BLOCK_STORAGE", IsActive: true},
		{ID: 2, Name: "HDD", Group: "BLOCK_STORAGE", IsActive: true},
		{ID: 3, Name: "NFS-SSD", Group: "FILE_STORAGE", IsActive: true},
		{ID: 4, Name: "Deprecated", Group: "BLOCK_STORAGE", IsActive: false},
	}

	group := r.URL.Query().Get("group")
	if group != "" {
		var filtered []models.VolumeType
		for _, vt := range allTypes {
			if vt.Group == group {
				filtered = append(filtered, vt)
			}
		}
		json.NewEncoder(w).Encode(filtered)
		return
	}
	json.NewEncoder(w).Encode(allTypes)
}

// AddHandler adds a custom handler for a specific method and path
func (ms *MockServer) AddHandler(method, path string, handler http.HandlerFunc) {
	key := fmt.Sprintf("%s %s", method, path)
	ms.Handlers[key] = handler
}

// Security Group handlers

func (ms *MockServer) createSecurityGroupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	sg := models.SecurityGroupDetail{
		ID:                1,
		UUID:              "sg-uuid-1234",
		SecurityGroupName: "test-sg",
		Status:            "ACTIVE",
		ProjectID:         1,
		AZName:            "south-1a",
		AZRegion:          "south-1",
	}
	json.NewEncoder(w).Encode(sg)
}

func (ms *MockServer) getSecurityGroupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sg := models.SecurityGroupDetail{
		ID:                1,
		UUID:              "sg-uuid-1234",
		SecurityGroupName: "test-sg",
		Status:            "ACTIVE",
		ProjectID:         1,
		AZName:            "south-1a",
		AZRegion:          "south-1",
		Rules: []models.SecurityGroupRuleDetail{
			{
				ID:             10,
				UUID:           "rule-uuid-1",
				Direction:      "ingress",
				Protocol:       "tcp",
				PortRangeMin:   "22",
				PortRangeMax:   "22",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
				Status:         "ACTIVE",
				Description:    "Allow SSH",
			},
		},
	}
	json.NewEncoder(w).Encode(sg)
}

func (ms *MockServer) listSecurityGroupsDetailedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sgs := []models.SecurityGroupDetail{
		{
			ID:                1,
			UUID:              "sg-uuid-1234",
			SecurityGroupName: "test-sg",
			Status:            "ACTIVE",
		},
		{
			ID:                2,
			UUID:              "sg-uuid-5678",
			SecurityGroupName: "default-sg",
			Status:            "ACTIVE",
		},
	}
	json.NewEncoder(w).Encode(sgs)
}

func (ms *MockServer) deleteSecurityGroupHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	// After deletion, subsequent GET requests should return 404
	ms.Handlers["GET /api/v1/networks/securitygroup/1/"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

// Security Group Rule handlers

func (ms *MockServer) createSecurityGroupRuleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	rules := []models.SecurityGroupRuleDetail{{
		ID:                          10,
		UUID:                        "rule-uuid-1",
		Direction:                   "ingress",
		Protocol:                    "tcp",
		PortRangeMin:                "22",
		PortRangeMax:                "22",
		RemoteIPPrefix:              "0.0.0.0/0",
		Ethertype:                   "IPv4",
		Status:                      "ACTIVE",
		Description:                 "Allow SSH",
		ProviderSecurityGroupRuleID: "provider-rule-1",
	}}
	json.NewEncoder(w).Encode(rules)
}

func (ms *MockServer) getSecurityGroupRuleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rule := models.SecurityGroupRuleDetail{
		ID:                          10,
		UUID:                        "rule-uuid-1",
		Direction:                   "ingress",
		Protocol:                    "tcp",
		PortRangeMin:                "22",
		PortRangeMax:                "22",
		RemoteIPPrefix:              "0.0.0.0/0",
		Ethertype:                   "IPv4",
		Status:                      "ACTIVE",
		Description:                 "Allow SSH",
		ProviderSecurityGroupRuleID: "provider-rule-1",
	}
	json.NewEncoder(w).Encode(rule)
}

func (ms *MockServer) listSecurityGroupRulesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rules := []models.SecurityGroupRuleDetail{
		{
			ID:             10,
			UUID:           "rule-uuid-1",
			Direction:      "ingress",
			Protocol:       "tcp",
			PortRangeMin:   "22",
			PortRangeMax:   "22",
			RemoteIPPrefix: "0.0.0.0/0",
			Ethertype:      "IPv4",
			Status:         "ACTIVE",
			Description:    "Allow SSH",
		},
		{
			ID:             11,
			UUID:           "rule-uuid-2",
			Direction:      "egress",
			Protocol:       "tcp",
			PortRangeMin:   "443",
			PortRangeMax:   "443",
			RemoteIPPrefix: "0.0.0.0/0",
			Ethertype:      "IPv4",
			Status:         "ACTIVE",
			Description:    "Allow HTTPS outbound",
		},
	}
	json.NewEncoder(w).Encode(rules)
}

func (ms *MockServer) deleteSecurityGroupRuleHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	// After deletion, subsequent GET requests should return 404
	ms.Handlers["GET /api/v1/networks/securitygroup/1/securitygrouprule/10/"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

// VPC Peering handlers

func (ms *MockServer) createVPCPeeringHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (ms *MockServer) getVPCPeeringHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	peering := models.VPCPeering{
		ID:          "test-peering-id",
		Name:        "test-peering",
		Description: "Test peering connection",
		VPCSourceID: "vpc-source-1",
		VPCTargetID: "vpc-target-1",
		AZ:          "south-1a",
		Region:      "south-1",
		State:       "ACTIVE",
	}
	json.NewEncoder(w).Encode(peering)
}

func (ms *MockServer) listVPCPeeringsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := models.VPCPeeringListResponse{
		Count: 1,
		Items: []models.VPCPeering{
			{
				ID:          "test-peering-id",
				Name:        "test-peering",
				Description: "Test peering connection",
				VPCSourceID: "vpc-source-1",
				VPCTargetID: "vpc-target-1",
				AZ:          "south-1a",
				Region:      "south-1",
				State:       "ACTIVE",
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) deleteVPCPeeringHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	// After deletion, subsequent GET requests should return 404
	ms.Handlers["GET /api/network-manager/v1/domain/test-org/project/test-project/vpc-peering/test-peering-id"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

// Compute Snapshot handlers

func (ms *MockServer) createComputeSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	snapshot := models.ComputeSnapshot{
		ID:           1,
		UUID:         "snap-uuid-1234",
		SnapshotName: "test-snapshot",
		Status:       "active",
		IsActive:     true,
		IsImage:      false,
		Created:      "2026-03-26T10:00:00Z",
	}
	json.NewEncoder(w).Encode(snapshot)
}

func (ms *MockServer) getComputeSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	snapshot := models.ComputeSnapshot{
		ID:           1,
		UUID:         "snap-uuid-1234",
		SnapshotName: "test-snapshot",
		Status:       "active",
		IsActive:     true,
		IsImage:      false,
		Created:      "2026-03-26T10:00:00Z",
	}
	json.NewEncoder(w).Encode(snapshot)
}

func (ms *MockServer) listComputeSnapshotsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	snapshots := []models.ComputeSnapshot{
		{
			ID:           1,
			UUID:         "snap-uuid-1234",
			SnapshotName: "test-snapshot",
			Status:       "active",
			IsActive:     true,
		},
		{
			ID:           2,
			UUID:         "snap-uuid-5678",
			SnapshotName: "test-snapshot-2",
			Status:       "active",
			IsActive:     true,
		},
	}
	json.NewEncoder(w).Encode(snapshots)
}

func (ms *MockServer) deleteComputeSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	ms.Handlers["GET /api/v2.1/computes/domain/test-org/project/test-project/computes/snapshot/snap-uuid-1234/"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

// Protection handlers

func (ms *MockServer) createProtectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	protection := models.VeritasProtection{
		ID:             1,
		Name:           "test-protection",
		Description:    "Test protection policy",
		Status:         "ACTIVE",
		ComputeID:      "compute-1",
		ProtectionPlan: "daily-plan",
		Region:         "south-1",
		AZName:         "south-1a",
		Created:        "2026-03-26T10:00:00Z",
	}
	json.NewEncoder(w).Encode(protection)
}

func (ms *MockServer) getProtectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	protection := models.VeritasProtection{
		ID:             1,
		Name:           "test-protection",
		Description:    "Test protection policy",
		Status:         "ACTIVE",
		ComputeID:      "compute-1",
		ProtectionPlan: "daily-plan",
		Region:         "south-1",
		AZName:         "south-1a",
		Created:        "2026-03-26T10:00:00Z",
	}
	json.NewEncoder(w).Encode(protection)
}

func (ms *MockServer) listProtectionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	protections := []models.VeritasProtection{
		{
			ID:             1,
			Name:           "test-protection",
			Status:         "ACTIVE",
			ComputeID:      "compute-1",
			ProtectionPlan: "daily-plan",
		},
	}
	json.NewEncoder(w).Encode(protections)
}

func (ms *MockServer) updateProtectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	protection := models.VeritasProtection{
		ID:             1,
		Name:           "updated-protection",
		Description:    "Updated protection policy",
		Status:         "ACTIVE",
		ComputeID:      "compute-1",
		ProtectionPlan: "daily-plan",
		Region:         "south-1",
		AZName:         "south-1a",
	}
	json.NewEncoder(w).Encode(protection)
}

func (ms *MockServer) deleteProtectionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Protection Plan handlers

func (ms *MockServer) createProtectionPlanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// The real API returns a success message string, not a plan object
	json.NewEncoder(w).Encode("Protection plan created and assigned successfully")
}

func (ms *MockServer) getProtectionPlanHandler(w http.ResponseWriter, r *http.Request) {
	// Single plan GET is not available in the real API (returns 404).
	// This handler is kept for mock testing but the real code uses list + filter.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "Not Found"})
}

func (ms *MockServer) listProtectionPlansHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := models.ProtectionPlanListResponse{
		PolicyAttributeList: []models.ProtectionPlan{
			{
				ID:          "plan-uuid-1234",
				Name:        "S1-TEST-ORG-TEST-PROJECT-TEST-PLAN-BKP-PP",
				ProjectID:   "test-project-id",
				ProjectName: "test-project",
				CreatedAt:   "2026-04-10T14:25:52.000000",
			},
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// LB Service handlers

func (ms *MockServer) createLBServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	svc := models.LBService{
		ID:              "lb-svc-1",
		Name:            "test-lb",
		Description:     "Test LB service",
		FlavorID:        1,
		NetworkID:       "net-1",
		VPCID:           "vpc-1",
		VPCName:         "test-vpc",
		Status:          "Active",
		OperatingStatus: "ONLINE",
		AZName:          "south-1a",
		Created:         "2026-03-26T10:00:00Z",
	}
	json.NewEncoder(w).Encode(svc)
}

func (ms *MockServer) getLBServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	svc := models.LBService{
		ID:              "lb-svc-1",
		Name:            "test-lb",
		Description:     "Test LB service",
		FlavorID:        1,
		NetworkID:       "net-1",
		VPCID:           "vpc-1",
		VPCName:         "test-vpc",
		Status:          "Active",
		OperatingStatus: "ONLINE",
		AZName:          "south-1a",
		Created:         "2026-03-26T10:00:00Z",
	}
	json.NewEncoder(w).Encode(svc)
}

func (ms *MockServer) listLBServicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	services := []models.LBService{
		{
			ID:       "lb-svc-1",
			Name:     "test-lb",
			FlavorID: 1,
			Status:   "Active",
		},
	}
	json.NewEncoder(w).Encode(services)
}

func (ms *MockServer) deleteLBServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	ms.Handlers["GET /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb_service/lb-svc-1"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

// LB VIP handlers

func (ms *MockServer) createLBVipHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	vip := models.LBVip{
		ID:             1,
		Name:           "vip-1",
		Status:         "Active",
		FixedIPs:       []string{"10.0.0.100"},
		PublicIP:       "203.0.113.10",
		ProviderPortID: "port-abc123",
	}
	json.NewEncoder(w).Encode(vip)
}

func (ms *MockServer) listLBVipsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vips := []models.LBVip{
		{
			ID:             1,
			Name:           "vip-1",
			Status:         "Active",
			FixedIPs:       []string{"10.0.0.100"},
			PublicIP:       "203.0.113.10",
			ProviderPortID: "port-abc123",
		},
	}
	json.NewEncoder(w).Encode(vips)
}

func (ms *MockServer) deleteLBVipHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// LB Certificate handlers

func (ms *MockServer) createLBCertificateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	cert := models.LBCertificate{
		ID:      1,
		Name:    "test-cert",
		Status:  "Active",
		Created: "2026-03-26T10:00:00Z",
	}
	json.NewEncoder(w).Encode(cert)
}

func (ms *MockServer) listLBCertificatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	certs := []models.LBCertificate{
		{
			ID:     1,
			Name:   "test-cert",
			Status: "Active",
		},
	}
	json.NewEncoder(w).Encode(certs)
}

func (ms *MockServer) deleteLBCertificateHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// LB Virtual Server handlers

func (ms *MockServer) createVirtualServerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	vs := models.LBVirtualServer{
		ID:               "vs-1",
		Name:             "test-vs",
		Protocol:         "HTTP",
		Port:             80,
		RoutingAlgorithm: "ROUND_ROBIN",
		Status:           "Active",
		VIP:              "10.0.0.100",
		Created:          "2026-03-26T10:00:00Z",
	}
	json.NewEncoder(w).Encode(vs)
}

func (ms *MockServer) getVirtualServerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vs := models.LBVirtualServer{
		ID:                 "vs-1",
		Name:               "test-vs",
		Protocol:           "HTTP",
		Port:               80,
		RoutingAlgorithm:   "ROUND_ROBIN",
		Status:             "Active",
		VIP:                "10.0.0.100",
		PersistenceEnabled: false,
		XForwardedFor:      true,
	}
	json.NewEncoder(w).Encode(vs)
}

func (ms *MockServer) listVirtualServersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	servers := []models.LBVirtualServer{
		{
			ID:               "vs-1",
			Name:             "test-vs",
			Protocol:         "HTTP",
			Port:             80,
			RoutingAlgorithm: "ROUND_ROBIN",
			Status:           "Active",
		},
	}
	json.NewEncoder(w).Encode(servers)
}

func (ms *MockServer) updateVirtualServerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vs := models.LBVirtualServer{
		ID:               "vs-1",
		Name:             "test-vs",
		Protocol:         "HTTP",
		Port:             80,
		RoutingAlgorithm: "LEAST_CONNECTIONS",
		Status:           "Active",
		VIP:              "10.0.0.100",
		XForwardedFor:    true,
	}
	json.NewEncoder(w).Encode(vs)
}

func (ms *MockServer) deleteVirtualServerHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	ms.Handlers["GET /api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers/vs-1"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

// Public IP (IPAM) handlers
func (ms *MockServer) createPublicIPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := models.PublicIP{
		UUID:     "test-public-ip-uuid",
		PublicIP: "103.239.168.100",
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) getPublicIPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := models.PublicIP{
		UUID:          "test-public-ip-uuid",
		IP:            "103.239.168.100",
		Domain:        "airtelcloud.itm",
		ObjectName:    "test-public-ip",
		TargetVIP:     "10.1.99.172",
		AllocatedTime: "1775021475.878893",
		AZName:        "S1",
		ProjectName:   "test-project",
		Region:        "south",
		Status:        "Created",
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) listPublicIPsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := models.PublicIPListResponse{
		Items: []models.PublicIP{
			{
				UUID:          "test-public-ip-uuid",
				IP:            "103.239.168.100",
				Domain:        "airtelcloud.itm",
				ObjectName:    "test-public-ip",
				TargetVIP:     "10.1.99.172",
				AllocatedTime: "1775021475.878893",
				AZName:        "S1",
				ProjectName:   "test-project",
				Region:        "south",
				Status:        "Created",
			},
		},
		Count: 1,
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) deletePublicIPHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	ms.Handlers["GET /api/v1/ipam/domain/test-org/project/test-project/test-public-ip-uuid"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (ms *MockServer) mapPublicIPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": ""})
}

// Public IP Policy Rule handlers
func (ms *MockServer) createPublicIPPolicyRuleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": ""})
}

func (ms *MockServer) listPublicIPPolicyRulesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := models.PublicIPPolicyRuleListResponse{
		Items: []models.PublicIPPolicyRule{
			{
				DisplayName: "test-rule",
				UUID:        "test-public-ip-uuid-1",
				OrgID:       "test-org-id",
				OrgName:     "test-org",
				AZName:      "S1",
				SourceIP:    "any",
				TargetVIP:   "10.1.99.172",
				State:       "create_adom_policy",
				Services:    []string{"HTTP", "HTTPS"},
				Action:      "accept",
			},
		},
		Count: 1,
	}
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) deletePublicIPPolicyRuleHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	ms.Handlers["GET /api/v1/admin/ipam_vip/test-public-ip-uuid/rules"] = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.PublicIPPolicyRuleListResponse{Items: []models.PublicIPPolicyRule{}, Count: 0})
	}
}

func (ms *MockServer) listIPAMServicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	services := []models.IPAMService{
		{UUID: "uuid-http", Name: "HTTP", IsDefault: true},
		{UUID: "uuid-https", Name: "HTTPS", IsDefault: true},
		{UUID: "uuid-ssh", Name: "SSH", IsDefault: true},
		{UUID: "uuid-dns", Name: "DNS", IsDefault: true},
	}
	json.NewEncoder(w).Encode(services)
}

// SetErrorResponse sets up an error response for a specific endpoint
func (ms *MockServer) SetErrorResponse(method, path string, statusCode int, message string) {
	key := fmt.Sprintf("%s %s", method, path)
	ms.Handlers[key] = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": message})
	}
}
