package models

// BaremetalListResponse wraps the baremetal server list API response.
type BaremetalListResponse struct {
	Count int         `json:"count"`
	Items []Baremetal `json:"items"`
}

// NetworkInfo is the nested network block of the /server/{name} detail response.
type NetworkInfo struct {
	// PortID -> virtual server node backend_port_id. The API sends it as either a
	// JSON string ("17963") or a number, so use FlexInt64 to accept both.
	PortID FlexInt64 `json:"portId"`
}

// Baremetal represents a baremetal server (from the baremetal-manager API).
// It is used as a backend pool member for load balancer virtual servers.
type Baremetal struct {
	Name             string      `json:"name"`     // friendly name, e.g. "amd-test2"
	Hostname         string      `json:"hostname"` // e.g. "n1-copper-bm09"
	State            string      `json:"state"`    // "Ready" | "Failed" | ...
	IPAddr           []string    `json:"ipAddr"`
	UUID             string      `json:"uuid"` // -> resource_id
	Flavor           string      `json:"flavor,omitempty"`
	PowerState       string      `json:"powerState,omitempty"`
	AvailabilityZone string      `json:"availabilityZone,omitempty"`
	NetworkInfo      NetworkInfo `json:"networkInfo"` // carries the backend port id
	// PortID is derived from NetworkInfo.PortID (networkInfo.portId in the
	// /server/{name} detail response) and maps to a virtual server node's
	// backend_port_id. It is populated by ResolveBaremetalNode, not by the API.
	PortID int `json:"-"`
}

// PrimaryIP returns the first fixed IP of the baremetal server, or "".
func (b *Baremetal) PrimaryIP() string {
	if len(b.IPAddr) > 0 {
		return b.IPAddr[0]
	}
	return ""
}
