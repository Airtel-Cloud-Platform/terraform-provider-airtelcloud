package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestMariaDBSchemaNetworkAccess(t *testing.T) {
	t.Parallel()

	var resp resource.SchemaResponse
	(&MariaDBResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if _, ok := resp.Schema.Attributes["security_group"]; ok {
		t.Fatal("security_group must not be exposed")
	}
	for _, name := range []string{"vpc", "subnet", "custom_cidr"} {
		attribute, ok := resp.Schema.Attributes[name]
		if !ok || !attribute.IsOptional() {
			t.Fatalf("%s must be optional", name)
		}
	}
}

func TestBuildMariaDBCreateRequestNetworkAccess(t *testing.T) {
	for _, test := range []struct {
		name       string
		vpc        string
		subnet     string
		customCIDR []string
		subnetCIDR string
		wantIPs    []string
		wantError  string
	}{
		{name: "custom only", customCIDR: []string{"192.168.1.0/24"}, wantIPs: []string{"192.168.1.0/24"}},
		{name: "subnet only", vpc: "app-vpc", subnet: "app-subnet", subnetCIDR: "10.0.0.0/24", wantIPs: []string{"10.0.0.0/24"}},
		{name: "both", vpc: "app-vpc", subnet: "app-subnet", subnetCIDR: "10.0.0.0/24", customCIDR: []string{"192.168.1.0/24"}, wantIPs: []string{"192.168.1.0/24", "10.0.0.0/24"}},
		{name: "duplicate subnet CIDR", vpc: "app-vpc", subnet: "app-subnet", subnetCIDR: "10.0.0.0/24", customCIDR: []string{"10.0.0.0/24"}, wantIPs: []string{"10.0.0.0/24"}},
		{name: "missing access", wantError: "Invalid Configuration"},
		{name: "missing VPC", vpc: "missing", subnet: "app-subnet", wantError: "VPC Resolution Error"},
		{name: "missing subnet", vpc: "app-vpc", subnet: "missing", wantError: "Subnet Resolution Error"},
		{name: "empty subnet CIDR", vpc: "app-vpc", subnet: "app-subnet", subnetCIDR: " ", wantError: "Subnet Resolution Error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := testutil.NewMockServer()
			defer server.Close()
			const databaseBase = "/api/v1/dbaas/domain/test-org/project/test-project"
			const networkBase = "/api/network-manager/v1/domain/test-org/project/test-project"
			responses := map[string]interface{}{
				"/api/auth-mgmt/v1/zones":               models.MariaDBZoneListResponse{Items: []models.MariaDBZone{{AZCode: "S1", IsActive: true}}},
				databaseBase + "/mariadb/versions":      []models.MariaDBVersion{{Version: "10.6", Supported: true}},
				databaseBase + "/mariadb/flavors":       []models.MariaDBFlavor{{ID: 290, Name: "db.maria.uhper.ccs.xlarge", RAM: 16, IsActive: true}},
				databaseBase + "/mariadb/topologies":    []models.MariaDBTopology{{Topology: models.MariaDBTopologyMaster}},
				databaseBase + "/mssql/volumetypes":     []models.MariaDBVolumeType{{ID: 10, Name: "backend", Label: "High Performance", IsActive: true}},
				networkBase + "/networks":               models.VPCListResponse{Items: []models.VPC{{ID: "vpc-id", Name: "app-vpc"}}},
				networkBase + "/network/vpc-id/subnets": models.SubnetListResponse{Items: []models.Subnet{{Name: "app-subnet", IPv4AddressSpace: test.subnetCIDR}}},
			}
			for endpoint, response := range responses {
				server.AddHandler("GET", endpoint, func(writer http.ResponseWriter, request *http.Request) {
					if test.vpc == "" && request.URL.Path == networkBase+"/networks" {
						t.Error("custom CIDRs must not trigger VPC lookup")
					}
					writer.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(writer).Encode(response); err != nil {
						t.Errorf("encode response: %v", err)
					}
				})
			}
			apiClient, err := client.NewClient(server.URL, "test-api-key", "test-api-secret", "south", "test-org", "test-project", "")
			if err != nil {
				t.Fatal(err)
			}
			cidrs, diags := types.ListValueFrom(context.Background(), types.StringType, test.customCIDR)
			if diags.HasError() {
				t.Fatal(diags)
			}
			data := MariaDBResourceModel{
				ClusterName: types.StringValue("app-db"), Version: types.StringValue("10.6"),
				AvailabilityZone: types.StringValue("S1"), ComputeSize: types.StringValue("db.maria.uhper.ccs.xlarge"),
				StorageType: types.StringValue("High Performance"), StorageSize: types.Int64Value(200),
				VPC: types.StringValue(test.vpc), Subnet: types.StringValue(test.subnet), CustomCIDR: cidrs,
			}
			createRequest, diags := (&MariaDBResource{client: apiClient}).buildMariaDBCreateRequest(context.Background(), &data)
			if test.wantError != "" {
				if !diags.HasError() || diags[0].Summary() != test.wantError {
					t.Fatalf("diagnostics = %v, want %s", diags, test.wantError)
				}
				return
			}
			if diags.HasError() {
				t.Fatal(diags)
			}
			if createRequest.SecurityGroup == nil || !reflect.DeepEqual(createRequest.SecurityGroup.AllowedIPs, test.wantIPs) {
				t.Fatalf("security group = %+v, want allowed IPs %v", createRequest.SecurityGroup, test.wantIPs)
			}
			if !data.CustomCIDR.Equal(cidrs) {
				t.Fatal("resolved subnet must not modify configured custom CIDRs")
			}
		})
	}
}

func TestValidateMariaDBReplicas(t *testing.T) {
	tests := []struct {
		name     string
		ha       types.Bool
		replicas types.Int64
		wantErr  bool
	}{
		{
			name:     "master default replicas",
			ha:       types.BoolValue(false),
			replicas: types.Int64Value(0),
		},
		{
			name:     "ha with one replica",
			ha:       types.BoolValue(true),
			replicas: types.Int64Value(1),
		},
		{
			name:     "ha with two replicas",
			ha:       types.BoolValue(true),
			replicas: types.Int64Value(2),
		},
		{
			name:     "ha with ten replicas",
			ha:       types.BoolValue(true),
			replicas: types.Int64Value(10),
		},
		{
			name:     "ha with zero replicas",
			ha:       types.BoolValue(true),
			replicas: types.Int64Value(0),
			wantErr:  true,
		},
		{
			name:     "ha with eleven replicas",
			ha:       types.BoolValue(true),
			replicas: types.Int64Value(11),
			wantErr:  true,
		},
		{
			name:     "ha with replicas unset",
			ha:       types.BoolValue(true),
			replicas: types.Int64Null(),
			wantErr:  true,
		},
		{
			name:     "ha unknown skips validation",
			ha:       types.BoolUnknown(),
			replicas: types.Int64Value(0),
		},
		{
			name:     "ha true replicas unknown skips validation",
			ha:       types.BoolValue(true),
			replicas: types.Int64Unknown(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := validateMariaDBReplicas(tt.ha, tt.replicas)
			gotErr := msg != ""
			if gotErr != tt.wantErr {
				t.Fatalf("validateMariaDBReplicas() error = %v (%q), wantErr %v", gotErr, msg, tt.wantErr)
			}
		})
	}
}

func TestValidateMariaDBPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    types.String
		username    types.String
		clusterName types.String
		wantErr     bool
	}{
		{
			name:        "valid password",
			password:    types.StringValue("Strong#Pass1"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
		},
		{
			name:        "too short",
			password:    types.StringValue("Aa1#abc"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
		{
			name:        "missing uppercase",
			password:    types.StringValue("strong#pass1"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
		{
			name:        "missing lowercase",
			password:    types.StringValue("STRONG#PASS1"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
		{
			name:        "missing number",
			password:    types.StringValue("Strong#Pass"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
		{
			name:        "missing special character",
			password:    types.StringValue("StrongPass1"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
		{
			name:        "contains at",
			password:    types.StringValue("Strong@Pass1"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
		{
			name:        "contains underscore",
			password:    types.StringValue("Strong_Pass1!"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
		{
			name:        "contains username word",
			password:    types.StringValue("Strong#admin1"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
		{
			name:        "contains cluster word",
			password:    types.StringValue("Strong#cluster1"),
			username:    types.StringValue("admin_user"),
			clusterName: types.StringValue("cluster-name"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := validateMariaDBPassword(tt.password, tt.username, tt.clusterName)
			gotErr := msg != ""
			if gotErr != tt.wantErr {
				t.Fatalf("validateMariaDBPassword() error = %v (%q), wantErr %v", gotErr, msg, tt.wantErr)
			}
		})
	}
}

func TestValidateMariaDBNames(t *testing.T) {
	nameTests := []struct {
		name     string
		value    types.String
		validate func(types.String) string
		wantErr  bool
	}{
		{
			name:     "valid cluster name",
			value:    types.StringValue("prod-db-01"),
			validate: validateMariaDBClusterName,
		},
		{
			name:     "invalid cluster uppercase",
			value:    types.StringValue("Prod-db"),
			validate: validateMariaDBClusterName,
			wantErr:  true,
		},
		{
			name:     "invalid cluster underscore",
			value:    types.StringValue("prod_db"),
			validate: validateMariaDBClusterName,
			wantErr:  true,
		},
		{
			name:     "valid database name",
			value:    types.StringValue("db_main1"),
			validate: func(v types.String) string { return validateMariaDBIdentifier(v, "database_name") },
		},
		{
			name:     "invalid database starts with number",
			value:    types.StringValue("1db_main"),
			validate: func(v types.String) string { return validateMariaDBIdentifier(v, "database_name") },
			wantErr:  true,
		},
		{
			name:     "valid username",
			value:    types.StringValue("admin_user1"),
			validate: func(v types.String) string { return validateMariaDBIdentifier(v, "mariadb_username") },
		},
		{
			name:     "invalid username hyphen",
			value:    types.StringValue("admin-user"),
			validate: func(v types.String) string { return validateMariaDBIdentifier(v, "mariadb_username") },
			wantErr:  true,
		},
	}

	for _, tt := range nameTests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.validate(tt.value)
			gotErr := msg != ""
			if gotErr != tt.wantErr {
				t.Fatalf("name validation error = %v (%q), wantErr %v", gotErr, msg, tt.wantErr)
			}
		})
	}
}

func TestValidateMariaDBBackupScheduleRequired(t *testing.T) {
	tests := []struct {
		name          string
		backup        *MariaDBBackupModel
		wantTime      bool
		wantDay       bool
		wantRetention bool
	}{
		{
			name: "backup disabled allows missing schedule and retention",
			backup: &MariaDBBackupModel{
				Enabled:      types.BoolValue(false),
				ScheduleTime: types.StringNull(),
				ScheduleDay:  types.StringNull(),
				Retention:    types.Int64Null(),
			},
		},
		{
			name: "backup enabled requires schedule and retention",
			backup: &MariaDBBackupModel{
				Enabled:      types.BoolValue(true),
				ScheduleTime: types.StringNull(),
				ScheduleDay:  types.StringNull(),
				Retention:    types.Int64Null(),
			},
			wantTime:      true,
			wantDay:       true,
			wantRetention: true,
		},
		{
			name: "backup enabled with schedule and retention set",
			backup: &MariaDBBackupModel{
				Enabled:      types.BoolValue(true),
				ScheduleTime: types.StringValue("12:30"),
				ScheduleDay:  types.StringValue("Monday"),
				Retention:    types.Int64Value(15),
			},
		},
		{
			name: "backup enabled with blank schedule and zero retention",
			backup: &MariaDBBackupModel{
				Enabled:      types.BoolValue(true),
				ScheduleTime: types.StringValue("   "),
				ScheduleDay:  types.StringValue(""),
				Retention:    types.Int64Value(0),
			},
			wantTime:      true,
			wantDay:       true,
			wantRetention: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeMsg := validateMariaDBBackupScheduleTimeRequired(tt.backup)
			dayMsg := validateMariaDBBackupScheduleDayRequired(tt.backup)
			retentionMsg := validateMariaDBBackupRetentionRequired(tt.backup)

			gotTimeErr := timeMsg != ""
			gotDayErr := dayMsg != ""
			gotRetentionErr := retentionMsg != ""

			if gotTimeErr != tt.wantTime {
				t.Fatalf("schedule_time required error = %v (%q), want %v", gotTimeErr, timeMsg, tt.wantTime)
			}
			if gotDayErr != tt.wantDay {
				t.Fatalf("schedule_day required error = %v (%q), want %v", gotDayErr, dayMsg, tt.wantDay)
			}
			if gotRetentionErr != tt.wantRetention {
				t.Fatalf("retention required error = %v (%q), want %v", gotRetentionErr, retentionMsg, tt.wantRetention)
			}
		})
	}
}

func TestApplyMariaDBClusterToStateKeepsConfigOwnedFields(t *testing.T) {
	conn := "mysql -u admin123 -p'****' -h host -P 3350 -D terra"
	cluster := &models.MariaDBCluster{
		UUID:             "cluster-uuid",
		Name:             "cluster-label-cluster-uuid",
		LevelName:        "cluster-label",
		Version:          "10.6",
		Topology:         models.MariaDBTopologyMasterSlave,
		NumReplicas:      1,
		MariaDBUsername:  "admin123",
		Status:           models.MariaDBStatusActive,
		CreatedAt:        "2026-09-30T18:06:32.159043",
		ConnectionString: &conn,
		Port:             3350,
		AzNames:          []string{"S1"},
		ComputeStorageConfig: models.MariaDBComputeStorageConfig{
			Flavor:        models.MariaDBFlavorRef{Name: "db.maria.uhper.ccs.xlarge", VCPUs: "4", RAM: 16},
			StorageName:   "High Performance",
			StorageSizeGB: 200,
		},
	}

	data := MariaDBResourceModel{
		ClusterName:      types.StringValue("cluster-label"),
		Password:         types.StringValue("secret"),
		AvailabilityZone: types.StringValue("S1"),
		StorageType:      types.StringValue("High Performance"),
		ComputeSize:      types.StringValue("db.maria.uhper.ccs.xlarge"),
	}

	diags := applyMariaDBClusterToState(context.Background(), &data, cluster, false)
	if diags.HasError() {
		t.Fatalf("applyMariaDBClusterToState() diagnostics = %v", diags)
	}
	if data.ID.ValueString() != "cluster-uuid" {
		t.Fatalf("id = %q", data.ID.ValueString())
	}
	if data.ClusterName.ValueString() != "cluster-label" {
		t.Fatalf("cluster_name overwritten: %q", data.ClusterName.ValueString())
	}
	if data.Password.ValueString() != "secret" {
		t.Fatal("password must stay from plan")
	}
	if data.Topology.ValueString() != models.MariaDBTopologyMasterSlave {
		t.Fatalf("topology = %q", data.Topology.ValueString())
	}
	if !data.HighAvailability.ValueBool() {
		t.Fatal("expected high_availability true from topology")
	}
	if data.Port.ValueInt64() != 3350 {
		t.Fatalf("port = %d", data.Port.ValueInt64())
	}
	if data.Users.ValueString() != "admin123" {
		t.Fatalf("users = %q", data.Users.ValueString())
	}
}

func TestApplyMariaDBClusterToStateFillsImportGaps(t *testing.T) {
	cluster := &models.MariaDBCluster{
		UUID:            "cluster-uuid",
		LevelName:       "cluster-label",
		Version:         "10.6",
		Topology:        models.MariaDBTopologyMaster,
		DatabaseName:    "terra",
		MariaDBUsername: "admin123",
		Port:            3350,
		AzNames:         []string{"S1"},
		ComputeStorageConfig: models.MariaDBComputeStorageConfig{
			Flavor:        models.MariaDBFlavorRef{Name: "db.maria.uhper.ccs.xlarge", VCPUs: "4", RAM: 16},
			StorageName:   "High Performance",
			StorageSizeGB: 200,
		},
	}

	data := MariaDBResourceModel{
		ID:               types.StringValue("cluster-uuid"),
		Password:         types.StringNull(),
		ClusterName:      types.StringNull(),
		AvailabilityZone: types.StringNull(),
		StorageType:      types.StringNull(),
		ComputeSize:      types.StringNull(),
	}

	diags := applyMariaDBClusterToState(context.Background(), &data, cluster, true)
	if diags.HasError() {
		t.Fatalf("applyMariaDBClusterToState() diagnostics = %v", diags)
	}
	if data.ClusterName.ValueString() != "cluster-label" {
		t.Fatalf("cluster_name = %q", data.ClusterName.ValueString())
	}
	if data.AvailabilityZone.ValueString() != "S1" {
		t.Fatalf("availability_zone = %q", data.AvailabilityZone.ValueString())
	}
	if data.StorageType.ValueString() != "High Performance" {
		t.Fatalf("storage_type = %q", data.StorageType.ValueString())
	}
	if data.ComputeSize.ValueString() != "db.maria.uhper.ccs.xlarge" {
		t.Fatalf("compute_size = %q", data.ComputeSize.ValueString())
	}
	if data.MariaDBUsername.ValueString() != "admin123" {
		t.Fatalf("mariadb_username = %q", data.MariaDBUsername.ValueString())
	}
	if !data.Password.IsNull() {
		t.Fatal("password must remain unset after import")
	}
}
