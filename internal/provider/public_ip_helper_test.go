package provider

import (
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestGetPublicIPAddr(t *testing.T) {
	cases := []struct {
		name string
		in   *models.PublicIP
		want string
	}{
		{"nil_input", nil, ""},
		{"public_ip_set", &models.PublicIP{PublicIP: "203.0.113.5", IP: ""}, "203.0.113.5"},
		{"ip_set_only", &models.PublicIP{PublicIP: "", IP: "103.239.168.100"}, "103.239.168.100"},
		{"both_set_prefers_public_ip", &models.PublicIP{PublicIP: "203.0.113.5", IP: "103.239.168.100"}, "203.0.113.5"},
		{"none_set", &models.PublicIP{PublicIP: "", IP: ""}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := getPublicIPAddr(tc.in)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
