package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestValidatePostgresReplicas(t *testing.T) {
	tests := []struct {
		name     string
		ha       types.Bool
		replicas types.Int64
		wantErr  bool
	}{
		{
			name:     "standalone default replicas",
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
			msg := validatePostgresReplicas(tt.ha, tt.replicas)
			gotErr := msg != ""
			if gotErr != tt.wantErr {
				t.Fatalf("validatePostgresReplicas() error = %v (%q), wantErr %v", gotErr, msg, tt.wantErr)
			}
		})
	}
}

func TestValidatePostgresPassword(t *testing.T) {
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
			msg := validatePostgresPassword(tt.password, tt.username, tt.clusterName)
			gotErr := msg != ""
			if gotErr != tt.wantErr {
				t.Fatalf("validatePostgresPassword() error = %v (%q), wantErr %v", gotErr, msg, tt.wantErr)
			}
		})
	}
}

func TestValidatePostgresNames(t *testing.T) {
	nameTests := []struct {
		name     string
		value    types.String
		validate func(types.String) string
		wantErr  bool
	}{
		{
			name:     "valid cluster name",
			value:    types.StringValue("prod-db-01"),
			validate: validatePostgresClusterName,
		},
		{
			name:     "invalid cluster uppercase",
			value:    types.StringValue("Prod-db"),
			validate: validatePostgresClusterName,
			wantErr:  true,
		},
		{
			name:     "invalid cluster underscore",
			value:    types.StringValue("prod_db"),
			validate: validatePostgresClusterName,
			wantErr:  true,
		},
		{
			name:     "valid database name",
			value:    types.StringValue("db_main1"),
			validate: func(v types.String) string { return validatePostgresIdentifier(v, "database_name") },
		},
		{
			name:     "invalid database starts with number",
			value:    types.StringValue("1db_main"),
			validate: func(v types.String) string { return validatePostgresIdentifier(v, "database_name") },
			wantErr:  true,
		},
		{
			name:     "valid username",
			value:    types.StringValue("admin_user1"),
			validate: func(v types.String) string { return validatePostgresIdentifier(v, "postgres_username") },
		},
		{
			name:     "invalid username hyphen",
			value:    types.StringValue("admin-user"),
			validate: func(v types.String) string { return validatePostgresIdentifier(v, "postgres_username") },
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

func TestValidatePostgresBackupScheduleRequired(t *testing.T) {
	tests := []struct {
		name     string
		backup   *PostgresBackupModel
		wantTime bool
		wantDay  bool
	}{
		{
			name: "backup disabled allows missing schedule",
			backup: &PostgresBackupModel{
				Enabled:      types.BoolValue(false),
				ScheduleTime: types.StringNull(),
				ScheduleDay:  types.StringNull(),
			},
		},
		{
			name: "backup enabled requires both schedule fields",
			backup: &PostgresBackupModel{
				Enabled:      types.BoolValue(true),
				ScheduleTime: types.StringNull(),
				ScheduleDay:  types.StringNull(),
			},
			wantTime: true,
			wantDay:  true,
		},
		{
			name: "backup enabled with schedule set",
			backup: &PostgresBackupModel{
				Enabled:      types.BoolValue(true),
				ScheduleTime: types.StringValue("12:30"),
				ScheduleDay:  types.StringValue("Monday"),
			},
		},
		{
			name: "backup enabled with blank schedule",
			backup: &PostgresBackupModel{
				Enabled:      types.BoolValue(true),
				ScheduleTime: types.StringValue("   "),
				ScheduleDay:  types.StringValue(""),
			},
			wantTime: true,
			wantDay:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeMsg := validatePostgresBackupScheduleTimeRequired(tt.backup)
			dayMsg := validatePostgresBackupScheduleDayRequired(tt.backup)

			gotTimeErr := timeMsg != ""
			gotDayErr := dayMsg != ""

			if gotTimeErr != tt.wantTime {
				t.Fatalf("schedule_time required error = %v (%q), want %v", gotTimeErr, timeMsg, tt.wantTime)
			}
			if gotDayErr != tt.wantDay {
				t.Fatalf("schedule_day required error = %v (%q), want %v", gotDayErr, dayMsg, tt.wantDay)
			}
		})
	}
}

func TestApplyPostgresClusterToStateKeepsConfigOwnedFields(t *testing.T) {
	conn := "postgresql://admin123:****@host:5451/terra"
	cluster := &models.PostgresCluster{
		UUID:             "cluster-uuid",
		Name:             "cluster-label-cluster-uuid",
		LevelName:        "cluster-label",
		Version:          "17",
		Topology:         models.PostgresTopologyPrimaryStandby,
		NumReplicas:      1,
		Status:           models.PostgresStatusActive,
		CreatedAt:        "2026-09-02T06:13:36.433000",
		ConnectionString: &conn,
		AzNames:          []string{"S1"},
		ComputeStorageConfig: models.PostgresComputeStorageConfig{
			Flavor:        models.PostgresFlavorRef{Name: "db.postgres.uhper.ccs.xlarge"},
			StorageName:   "High Performance",
			StorageSizeGB: 200,
		},
	}

	data := PostgresResourceModel{
		ClusterName:      types.StringValue("cluster-label"),
		Password:         types.StringValue("secret"),
		AvailabilityZone: types.StringValue("S1"),
		StorageType:      types.StringValue("High Performance"),
		ComputeSize:      types.StringValue("db.postgres.uhper.ccs.xlarge"),
	}

	diags := applyPostgresClusterToState(context.Background(), &data, cluster, false)
	if diags.HasError() {
		t.Fatalf("applyPostgresClusterToState() diagnostics = %v", diags)
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
	if data.Topology.ValueString() != models.PostgresTopologyPrimaryStandby {
		t.Fatalf("topology = %q", data.Topology.ValueString())
	}
	if !data.HighAvailability.ValueBool() {
		t.Fatal("expected high_availability true from topology")
	}
}

func TestApplyPostgresClusterToStateFillsImportGaps(t *testing.T) {
	cluster := &models.PostgresCluster{
		UUID:             "cluster-uuid",
		LevelName:        "cluster-label",
		Version:          "17",
		Topology:         models.PostgresTopologyStandalone,
		DatabaseName:     "terra",
		PostgresUsername: "admin123",
		AzNames:          []string{"S1"},
		ComputeStorageConfig: models.PostgresComputeStorageConfig{
			Flavor:        models.PostgresFlavorRef{Name: "db.postgres.uhper.ccs.xlarge"},
			StorageName:   "High Performance",
			StorageSizeGB: 200,
		},
	}

	data := PostgresResourceModel{
		ID:               types.StringValue("cluster-uuid"),
		Password:         types.StringNull(),
		ClusterName:      types.StringNull(),
		AvailabilityZone: types.StringNull(),
		StorageType:      types.StringNull(),
		ComputeSize:      types.StringNull(),
	}

	diags := applyPostgresClusterToState(context.Background(), &data, cluster, true)
	if diags.HasError() {
		t.Fatalf("applyPostgresClusterToState() diagnostics = %v", diags)
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
	if data.ComputeSize.ValueString() != "db.postgres.uhper.ccs.xlarge" {
		t.Fatalf("compute_size = %q", data.ComputeSize.ValueString())
	}
	if !data.Password.IsNull() {
		t.Fatal("password must remain unset after import")
	}
}
