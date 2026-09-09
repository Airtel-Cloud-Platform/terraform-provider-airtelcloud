package models

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalPostgresCreateRequest(t *testing.T) {
	raw := []byte(`{
		"name": "cluster-label",
		"description": "Hi I ma description",
		"version": "17",
		"topology": "primary-standby",
		"num_replicas": 1,
		"database_name": "terra",
		"postgres_username": "admin123",
		"password": "secret",
		"is_superuser": true,
		"network_type": "private",
		"compute_storage_config": {
			"flavor": {"name": "db.postgres.uhper.ccs.xlarge", "ram": 16},
			"flavor_ids": [{"az_id": "S1", "id": 249}],
			"storage_size_gb": 200,
			"storage_type": "s1_wkld_ntp02_5iops_backend",
			"storage_name": "High Performance",
			"data_volume_type_id": 10
		},
		"az_ids": ["S1"],
		"pg_extensions": ["pgvector"],
		"labels": ["Label"],
		"backup": {
			"enabled": true,
			"protection_plan": "weekly-full-daily-incr",
			"compression_level": 6,
			"retention": 15,
			"schedule_time": "12:00",
			"schedule_day": "tuesday"
		},
		"security_group": {"allowed_ips": ["192.168.1.0/24"]}
	}`)

	var req CreatePostgresClusterRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal create request: %v", err)
	}

	if req.Name != "cluster-label" || req.Topology != PostgresTopologyPrimaryStandby {
		t.Fatalf("unexpected identity: name=%q topology=%q", req.Name, req.Topology)
	}
	if req.ComputeStorageConfig.Flavor.RAM != 16 || req.ComputeStorageConfig.FlavorIDs[0].ID != 249 {
		t.Fatalf("unexpected flavor mapping: %+v", req.ComputeStorageConfig)
	}
	if req.AZIDs[0] != "S1" || req.Backup == nil || req.Backup.ProtectionPlan != "weekly-full-daily-incr" {
		t.Fatalf("unexpected az/backup mapping: az=%v backup=%+v", req.AZIDs, req.Backup)
	}
}

func TestUnmarshalPostgresCatalogs(t *testing.T) {
	var zones PostgresZoneListResponse
	if err := json.Unmarshal([]byte(`{"count":1,"items":[{"name":"South-AZ1","azCode":"S1","regionCode":"south","isActive":true}]}`), &zones); err != nil {
		t.Fatalf("zones: %v", err)
	}
	if zones.Items[0].AZCode != "S1" {
		t.Fatalf("azCode = %q", zones.Items[0].AZCode)
	}

	var plans PostgresProtectionPlanListResponse
	if err := json.Unmarshal([]byte(`{"protection_plans":[{"value":"weekly-full-daily-incr","label":"Weekly"}]}`), &plans); err != nil {
		t.Fatalf("protection plans: %v", err)
	}
	if plans.ProtectionPlans[0].Value != "weekly-full-daily-incr" {
		t.Fatalf("plan value = %q", plans.ProtectionPlans[0].Value)
	}
}
