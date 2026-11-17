package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestReplacesUnlessAdopting(t *testing.T) {
	tests := []struct {
		name          string
		state, config attr.Value
		requireConfig bool
		want          bool
	}{
		{"null state adopts configuration", types.StringNull(), types.StringValue("subnet1213"), false, false},
		{"changed value replaces", types.StringValue("a"), types.StringValue("b"), false, true},
		{"removed argument replaces", types.StringValue("a"), types.StringNull(), false, true},
		{"computed value not configured is left alone", types.StringValue("a"), types.StringNull(), true, false},
		{"computed value configured and changed replaces", types.StringValue("a"), types.StringValue("b"), true, true},
		{"null state with computed attribute adopts", types.StringNull(), types.StringValue("b"), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replacesUnlessAdopting(tt.state, tt.config, tt.requireConfig); got != tt.want {
				t.Fatalf("replacesUnlessAdopting() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlannedASGOSType(t *testing.T) {
	r := &ASGResource{}
	ctx := context.Background()

	t.Run("snapshot has no image OS", func(t *testing.T) {
		os, ok := r.plannedASGOSType(ctx, &ASGResourceModel{SnapshotName: types.StringValue("snap"), OSType: types.StringUnknown()}, nil)
		if !ok || os != "" {
			t.Fatalf("got (%q, %v), want (\"\", true)", os, ok)
		}
	})
	t.Run("OS carried from state needs no lookup", func(t *testing.T) {
		os, ok := r.plannedASGOSType(ctx, &ASGResourceModel{OSType: types.StringValue("windows")}, nil)
		if !ok || os != "windows" {
			t.Fatalf("got (%q, %v), want (windows, true)", os, ok)
		}
	})
	t.Run("configured image name resolves while image_id is unknown", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			if strings.HasSuffix(req.URL.Path, "/images/") {
				_, _ = fmt.Fprint(w, `[{"id":1,"name":"Other","os":"ubuntu"},{"id":2,"name":"Win2022","os":"windows"}]`)
				return
			}
			http.NotFound(w, req)
		})
		ts := httptest.NewServer(mux)
		defer ts.Close()
		c, err := client.NewClient(ts.URL, "test-api-key", "test-api-secret", "north", "test-org", "test-project", "")
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}

		os, ok := r.plannedASGOSType(ctx, &ASGResourceModel{
			OSType: types.StringUnknown(), ImageID: types.Int64Unknown(), ImageName: types.StringValue("Win2022"),
		}, c)
		if !ok || os != "windows" {
			t.Fatalf("got (%q, %v), want (windows, true)", os, ok)
		}
		if got := defaultASGDiskSize(os); got != 200 {
			t.Fatalf("default disk = %d, want 200 for windows", got)
		}
	})
	t.Run("unresolvable without a scoped client", func(t *testing.T) {
		_, ok := r.plannedASGOSType(ctx, &ASGResourceModel{OSType: types.StringUnknown(), ImageName: types.StringValue("Ubuntu"), ImageID: types.Int64Unknown()}, nil)
		if ok {
			t.Fatal("expected ok=false so disk_size stays unknown")
		}
	})
}

// asgImportClient fakes the lookups used to recover an imported group. The
// counter shows when the API was (not) called.
func asgImportClient(t *testing.T) (*client.Client, *int32) {
	t.Helper()
	var calls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		p := r.URL.Path
		switch {
		case strings.HasSuffix(p, "/network/vpc-1"):
			_, _ = fmt.Fprint(w, `{"networkId":"vpc-1","name":"copper-vpc1"}`)
		case strings.HasSuffix(p, "/network/vpc-1/subnets"):
			_, _ = fmt.Fprint(w, `{"count":2,"items":[{"subnetId":"net-other","name":"other"},{"subnetId":"net-1","name":"subnet1213"}]}`)
		case strings.Contains(p, "/security-groups"):
			_, _ = fmt.Fprint(w, `[{"id":1,"security_group_name":"other-sg"},{"id":3663,"security_group_name":"networksg-2"}]`)
		case strings.Contains(p, "/keypairs"):
			_, _ = fmt.Fprint(w, `{"data":{"items":[{"uuid":"kp-0","name":"other"},{"uuid":"kp-1","name":"Vinay"}]}}`)
		case strings.HasSuffix(p, "/lb_service/lb-1"):
			_, _ = fmt.Fprint(w, `{"id":"lb-1","name":"lb","network_id":"lb-net"}`)
		case strings.HasSuffix(p, "/lb-1/virtual-servers/vs-1"):
			_, _ = fmt.Fprint(w, `{"id":"vs-1","vip":"10.29.2.221","protocol":"TCP","port":28080,"routing_algorithm":"ROUND_ROBIN","persistence_enabled":true,"pools":[{"id":"p","pool_name":"app-pool"}]}`)
		default:
			http.NotFound(w, r)
		}
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	c, err := client.NewClient(ts.URL, "test-api-key", "test-api-secret", "north", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return c, &calls
}

func importedASGModel() *ASGResourceModel {
	return &ASGResourceModel{
		VPCName:            types.StringNull(),
		Subnet:             types.StringNull(),
		SecurityGroupNames: types.ListNull(types.StringType),
		KeypairID:          types.StringValue("kp-1"),
		KeypairName:        types.StringNull(),
		EnablePersistence:  types.BoolNull(),
		VIP:                types.StringNull(),
		Protocol:           types.StringNull(),
		Port:               types.Int64Null(),
		RoutingAlgorithm:   types.StringNull(),
		PoolName:           types.StringNull(),
	}
}

func TestFillASGImportedFields(t *testing.T) {
	ctx := context.Background()
	group := &models.AutoScalingGroup{
		VPCID: "vpc-1", NetworkID: "net-1", AZName: "S2", SecurityGroupID: 3663,
	}

	t.Run("recovers names after import", func(t *testing.T) {
		c, _ := asgImportClient(t)
		r := &ASGResource{client: c}
		data := importedASGModel()

		diags := r.fillASGImportedFields(ctx, data, group)

		if diags.HasError() || diags.WarningsCount() != 0 {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if got := data.VPCName.ValueString(); got != "copper-vpc1" {
			t.Errorf("vpc_name = %q, want copper-vpc1", got)
		}
		if got := data.Subnet.ValueString(); got != "subnet1213" {
			t.Errorf("subnet = %q, want subnet1213", got)
		}
		if got := data.KeypairName.ValueString(); got != "Vinay" {
			t.Errorf("keypair_name = %q, want Vinay", got)
		}
		var names []string
		data.SecurityGroupNames.ElementsAs(ctx, &names, false)
		if len(names) != 1 || names[0] != "networksg-2" {
			t.Errorf("security_group_names = %v, want [networksg-2]", names)
		}
		if data.EnablePersistence.IsNull() || data.EnablePersistence.ValueBool() {
			t.Errorf("enable_persistance = %v, want false for a group with no load balancer", data.EnablePersistence)
		}
	})

	t.Run("recovers virtual server settings", func(t *testing.T) {
		c, _ := asgImportClient(t)
		r := &ASGResource{client: c}
		data := importedASGModel()
		withVS := *group
		withVS.VirtualServer.ID = "vs-1"
		withVS.VirtualServer.LBService.ID = "lb-1"

		diags := r.fillASGImportedFields(ctx, data, &withVS)

		if diags.HasError() || diags.WarningsCount() != 0 {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if data.VIP.ValueString() != "10.29.2.221" || data.Protocol.ValueString() != "TCP" ||
			data.Port.ValueInt64() != 28080 || data.RoutingAlgorithm.ValueString() != "ROUND_ROBIN" ||
			data.PoolName.ValueString() != "app-pool" {
			t.Errorf("virtual server settings not recovered: %+v", data)
		}
		if !data.EnablePersistence.ValueBool() {
			t.Errorf("enable_persistance = false, want true from the virtual server")
		}
	})

	t.Run("makes no calls when state is populated", func(t *testing.T) {
		c, calls := asgImportClient(t)
		r := &ASGResource{client: c}
		data := importedASGModel()
		data.VPCName = types.StringValue("copper-vpc1")
		data.Subnet = types.StringValue("subnet1213")
		data.SecurityGroupNames = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("networksg-2")})
		data.KeypairName = types.StringValue("Vinay")
		data.EnablePersistence = types.BoolValue(false)
		withVS := *group
		withVS.VirtualServer.ID = "vs-1"
		withVS.VirtualServer.LBService.ID = "lb-1"
		data.VIP = types.StringValue("10.29.2.221")

		if diags := r.fillASGImportedFields(ctx, data, &withVS); len(diags) != 0 {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if n := atomic.LoadInt32(calls); n != 0 {
			t.Fatalf("API called %d times, want 0", n)
		}
	})

	t.Run("lookup failure warns and does not error", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"detail":"boom"}`, http.StatusInternalServerError)
		})
		ts := httptest.NewServer(mux)
		defer ts.Close()
		c, err := client.NewClient(ts.URL, "test-api-key", "test-api-secret", "north", "test-org", "test-project", "")
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		r := &ASGResource{client: c}
		data := importedASGModel()

		diags := r.fillASGImportedFields(ctx, data, group)

		if diags.HasError() {
			t.Fatalf("lookup failures must not be errors: %v", diags)
		}
		if diags.WarningsCount() == 0 {
			t.Fatal("expected warnings for failed lookups")
		}
		if stringSet(data.VPCName) {
			t.Errorf("vpc_name = %q, want it left unset for adoption", data.VPCName.ValueString())
		}
	})
}

func TestValidateRulePairs(t *testing.T) {
	tests := []struct {
		name    string
		up      []models.ASGScalingRule
		down    []models.ASGScalingRule
		wantErr bool
	}{
		{
			name: "valid pair",
			up:   []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 60}},
			down: []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
		},
		{
			name:    "missing scale down",
			up:      []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 60}},
			wantErr: true,
		},
		{
			name:    "invalid target ordering",
			up:      []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
			down:    []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
			wantErr: true,
		},
		{
			name:    "aggregation mismatch",
			up:      []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "max", TargetValue: 60}},
			down:    []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			validateRulePairs(&diags, tt.up, tt.down)
			if diags.HasError() != tt.wantErr {
				t.Fatalf("HasError() = %v, want %v; diagnostics: %v", diags.HasError(), tt.wantErr, diags)
			}
		})
	}
}

func TestSubnetAvailabilityZoneMismatch(t *testing.T) {
	if msg := subnetAvailabilityZoneMismatch("subnet1213", "S2", "S1"); msg == "" {
		t.Fatal("expected mismatch error")
	}
	if msg := subnetAvailabilityZoneMismatch("subnet1213", "S1", "S1"); msg != "" {
		t.Fatalf("unexpected mismatch: %s", msg)
	}
	if msg := subnetAvailabilityZoneMismatch("subnet1213", "", "S1"); msg != "" {
		t.Fatalf("empty subnet zone should skip: %s", msg)
	}
}

func TestApplyASGAvailabilityZone(t *testing.T) {
	t.Run("keeps configured zone and warns on mismatch", func(t *testing.T) {
		data := &ASGResourceModel{AvailabilityZone: types.StringValue("S1")}

		diags := applyASGAvailabilityZone(data, "S2")

		if diags.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", diags)
		}
		if diags.WarningsCount() != 1 {
			t.Fatalf("warnings = %d, want 1", diags.WarningsCount())
		}
		if got := data.AvailabilityZone.ValueString(); got != "S1" {
			t.Fatalf("availability_zone = %q, want S1", got)
		}
	})

	t.Run("adopts API zone on import", func(t *testing.T) {
		data := &ASGResourceModel{AvailabilityZone: types.StringNull()}

		diags := applyASGAvailabilityZone(data, "S2")

		if diags.WarningsCount() != 0 {
			t.Fatalf("warnings = %d, want 0", diags.WarningsCount())
		}
		if got := data.AvailabilityZone.ValueString(); got != "S2" {
			t.Fatalf("availability_zone = %q, want S2", got)
		}
	})

	t.Run("no diagnostics when zones match", func(t *testing.T) {
		data := &ASGResourceModel{AvailabilityZone: types.StringValue("S1")}

		if diags := applyASGAvailabilityZone(data, "S1"); len(diags) != 0 {
			t.Fatalf("diagnostics = %v, want none", diags)
		}
	})
}

func TestSettleASGComputed(t *testing.T) {
	data := &ASGResourceModel{
		KeypairID:          types.StringUnknown(),
		KeypairName:        types.StringUnknown(),
		OSType:             types.StringUnknown(),
		ImageID:            types.Int64Unknown(),
		SecurityGroupNames: types.ListUnknown(types.StringType),
		FlavorName:         types.StringValue("ccd.Large"),
	}

	settleASGComputed(data)

	if data.KeypairID.IsUnknown() || data.KeypairName.IsUnknown() {
		t.Fatalf("keypair values still unknown: id=%v name=%v", data.KeypairID, data.KeypairName)
	}
	if data.OSType.IsUnknown() || data.ImageID.IsUnknown() {
		t.Fatalf("image values still unknown: os_type=%v image_id=%v", data.OSType, data.ImageID)
	}
	if data.SecurityGroupNames.IsUnknown() {
		t.Fatalf("security_group_names still unknown")
	}
	if data.FlavorName.ValueString() != "ccd.Large" {
		t.Fatalf("flavor_name = %v, want ccd.Large", data.FlavorName)
	}
}

func TestImageOSFamily(t *testing.T) {
	if got := (models.Image{OSType: "linux", OS: "windows"}).OSFamily(); got != "linux" {
		t.Fatalf("OSType should win, got %q", got)
	}
	if got := (models.Image{OS: "windows"}).OSFamily(); got != "windows" {
		t.Fatalf("OS fallback = %q, want windows", got)
	}
}

func TestValidateOneOfSkipsUnknown(t *testing.T) {
	var diags diag.Diagnostics
	validateOneOfIntString(&diags, "flavor_id", types.Int64Unknown(), "flavor_name", types.StringValue("ccd.Large"), true)
	if diags.HasError() {
		t.Fatalf("unknown flavor_id should skip XOR, got %v", diags)
	}
	validateListXOR(&diags, "security_group_ids", types.ListUnknown(types.Int64Type), "security_group_names", types.ListValueMust(types.StringType, []attr.Value{types.StringValue("networksg-1")}), true)
	if diags.HasError() {
		t.Fatalf("unknown security_group_ids should skip XOR, got %v", diags)
	}
}

func TestDefaultASGDiskSize(t *testing.T) {
	if got := defaultASGDiskSize("ubuntu"); got != 100 {
		t.Fatalf("linux default = %d, want 100", got)
	}
	if got := defaultASGDiskSize("windows"); got != 200 {
		t.Fatalf("windows default = %d, want 200", got)
	}
	if got := defaultASGDiskSize(""); got != 100 {
		t.Fatalf("empty OS default = %d, want 100", got)
	}
}
