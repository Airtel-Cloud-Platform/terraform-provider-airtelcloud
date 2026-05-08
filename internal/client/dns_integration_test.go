//go:build integration

package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// helper to create a string pointer
func strPtr(s string) *string { return &s }

// helper to create an int pointer
func intPtr(i int) *int { return &i }

// TestDNSZoneIntegration_CreateGetDelete tests the full lifecycle of a DNS zone
func TestDNSZoneIntegration_CreateGetDelete(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Generate unique zone name
	zoneName := fmt.Sprintf("test-%d.example.com.", time.Now().Unix())

	// Create DNS zone
	createReq := &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	}

	t.Logf("Creating DNS zone: %s", zoneName)

	zone, err := client.CreateDNSZone(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created with UUID: %s", zone.UUID)

	// Verify created zone fields
	if zone.ZoneName != zoneName {
		t.Errorf("Expected zone name %s, got %s", zoneName, zone.ZoneName)
	}
	if zone.UUID == "" {
		t.Error("Zone UUID should not be empty")
	}
	if zone.ZoneType != "forward" {
		t.Errorf("Expected zone type forward, got %s", zone.ZoneType)
	}

	// Cleanup: Delete zone at the end
	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		} else {
			t.Log("DNS zone deleted successfully")
		}
	}()

	// Get zone
	t.Logf("Getting DNS zone: %s", zone.UUID)
	fetchedZone, err := client.GetDNSZone(ctx, zone.UUID)
	if err != nil {
		t.Fatalf("GetDNSZone failed: %v", err)
	}

	// Verify fetched zone
	if fetchedZone.UUID != zone.UUID {
		t.Errorf("Expected zone UUID %s, got %s", zone.UUID, fetchedZone.UUID)
	}
	if fetchedZone.ZoneName != zoneName {
		t.Errorf("Expected zone name %s, got %s", zoneName, fetchedZone.ZoneName)
	}

	t.Log("GetDNSZone returned correct data")
}

// TestDNSZoneIntegration_List tests listing DNS zones
func TestDNSZoneIntegration_List(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization: %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	// List DNS zones
	response, err := client.ListDNSZones(ctx)
	if err != nil {
		t.Fatalf("ListDNSZones failed: %v", err)
	}

	t.Logf("Found %d DNS zones", response.Count)

	// Log zone details
	for i, zone := range response.Items {
		desc := ""
		if zone.Description != nil {
			desc = *zone.Description
		}
		t.Logf("Zone %d: UUID=%s, Name=%s, Type=%s, Description=%s",
			i+1, zone.UUID, zone.ZoneName, zone.ZoneType, desc)
	}
}

// TestDNSZoneIntegration_Update tests updating a DNS zone
func TestDNSZoneIntegration_Update(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create a zone to update
	zoneName := fmt.Sprintf("test-update-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	// Cleanup: Delete zone at the end
	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		} else {
			t.Log("DNS zone deleted successfully")
		}
	}()

	// Update zone description
	updatedDesc := "Updated integration test zone"
	updateReq := &models.UpdateDNSZoneRequest{
		Description: strPtr(updatedDesc),
	}

	t.Logf("Updating DNS zone: %s", zone.UUID)
	updatedZone, err := client.UpdateDNSZone(ctx, zone.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateDNSZone failed: %v", err)
	}

	// Verify update
	if updatedZone.Description == nil || *updatedZone.Description != updatedDesc {
		got := ""
		if updatedZone.Description != nil {
			got = *updatedZone.Description
		}
		t.Errorf("Expected description %q, got %q", updatedDesc, got)
	}

	// Get and verify
	fetchedZone, err := client.GetDNSZone(ctx, zone.UUID)
	if err != nil {
		t.Fatalf("GetDNSZone failed: %v", err)
	}

	if fetchedZone.Description == nil || *fetchedZone.Description != updatedDesc {
		got := ""
		if fetchedZone.Description != nil {
			got = *fetchedZone.Description
		}
		t.Errorf("Expected description %q after re-fetch, got %q", updatedDesc, got)
	}

	t.Log("DNS zone updated successfully")
}

// TestDNSZoneIntegration_GetNonExistent tests getting a non-existent DNS zone
func TestDNSZoneIntegration_GetNonExistent(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	nonExistentID := "00000000-0000-0000-0000-000000000000"

	t.Logf("Attempting to get non-existent DNS zone: %s", nonExistentID)

	_, err := client.GetDNSZone(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent DNS zone, got nil")
		return
	}

	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent DNS zone")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// TestDNSRecordIntegration_CreateGetDelete tests the full lifecycle of a DNS record
func TestDNSRecordIntegration_CreateGetDelete(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone first
	zoneName := fmt.Sprintf("test-rec-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	// Cleanup: Delete zone at the end
	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		} else {
			t.Log("DNS zone deleted successfully")
		}
	}()

	// Create A record
	recordReq := &models.CreateDNSRecordRequest{
		Owner:      strPtr("www"),
		Data:       strPtr("192.168.1.1"),
		RecordType: "A",
		TTL:        intPtr(300),
	}

	t.Log("Creating A record")
	record, err := client.CreateDNSRecord(ctx, zone.UUID, recordReq)
	if err != nil {
		t.Fatalf("CreateDNSRecord failed: %v", err)
	}

	t.Logf("Record created: UUID=%s, Owner=%s, Data=%s, Type=%s, TTL=%d",
		record.UUID, record.Owner, record.Data, record.RecordType, record.TTL)

	// Verify record fields
	if record.RecordType != "A" {
		t.Errorf("Expected record type A, got %s", record.RecordType)
	}
	if record.UUID == "" {
		t.Error("Record UUID should not be empty")
	}

	// Cleanup: Delete record before zone
	defer func() {
		t.Logf("Deleting DNS record: %s", record.UUID)
		err := client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
		if err != nil {
			t.Errorf("DeleteDNSRecord failed: %v", err)
		} else {
			t.Log("DNS record deleted successfully")
		}
	}()

	// Get record
	t.Logf("Getting DNS record: %s", record.UUID)
	fetchedRecord, err := client.GetDNSRecord(ctx, zone.UUID, record.UUID)
	if err != nil {
		t.Fatalf("GetDNSRecord failed: %v", err)
	}

	// Verify fetched record
	if fetchedRecord.UUID != record.UUID {
		t.Errorf("Expected record UUID %s, got %s", record.UUID, fetchedRecord.UUID)
	}
	if fetchedRecord.RecordType != "A" {
		t.Errorf("Expected record type A, got %s", fetchedRecord.RecordType)
	}

	t.Log("GetDNSRecord returned correct data")
}

// TestDNSRecordIntegration_Update tests updating a DNS record
func TestDNSRecordIntegration_Update(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-recup-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	// Create A record
	record, err := client.CreateDNSRecord(ctx, zone.UUID, &models.CreateDNSRecordRequest{
		Owner:      strPtr("api"),
		Data:       strPtr("10.0.0.1"),
		RecordType: "A",
		TTL:        intPtr(300),
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord failed: %v", err)
	}

	t.Logf("Record created: UUID=%s, TTL=%d, Data=%s", record.UUID, record.TTL, record.Data)

	defer func() {
		t.Logf("Deleting DNS record: %s", record.UUID)
		err := client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
		if err != nil {
			t.Errorf("DeleteDNSRecord failed: %v", err)
		}
	}()

	// Update TTL and data
	updateReq := &models.UpdateDNSRecordRequest{
		Owner:      strPtr("api"),
		Data:       strPtr("10.0.0.2"),
		RecordType: "A",
		TTL:        intPtr(600),
	}

	t.Log("Updating DNS record")
	updatedRecord, err := client.UpdateDNSRecord(ctx, zone.UUID, record.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateDNSRecord failed: %v", err)
	}

	// Verify update
	if updatedRecord.TTL != 600 {
		t.Errorf("Expected TTL 600, got %d", updatedRecord.TTL)
	}
	if updatedRecord.Data != "10.0.0.2" {
		t.Errorf("Expected data 10.0.0.2, got %s", updatedRecord.Data)
	}

	// Re-fetch and verify
	fetchedRecord, err := client.GetDNSRecord(ctx, zone.UUID, record.UUID)
	if err != nil {
		t.Fatalf("GetDNSRecord failed: %v", err)
	}

	if fetchedRecord.TTL != 600 {
		t.Errorf("Expected TTL 600 after re-fetch, got %d", fetchedRecord.TTL)
	}
	if fetchedRecord.Data != "10.0.0.2" {
		t.Errorf("Expected data 10.0.0.2 after re-fetch, got %s", fetchedRecord.Data)
	}

	t.Log("DNS record updated successfully")
}

// TestDNSRecordIntegration_MultipleTypes tests creating records of multiple types
func TestDNSRecordIntegration_MultipleTypes(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-multi-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	// Define records: A, CNAME, MX
	recordConfigs := []struct {
		name string
		req  *models.CreateDNSRecordRequest
	}{
		{
			name: "A",
			req: &models.CreateDNSRecordRequest{
				Owner:      strPtr("www"),
				Data:       strPtr("192.168.1.1"),
				RecordType: "A",
				TTL:        intPtr(300),
			},
		},
		{
			name: "CNAME",
			req: &models.CreateDNSRecordRequest{
				Owner:      strPtr("blog"),
				Data:       strPtr("www." + zoneName),
				RecordType: "CNAME",
				TTL:        intPtr(300),
			},
		},
		{
			name: "MX",
			req: &models.CreateDNSRecordRequest{
				Owner:      strPtr("@"),
				Data:       strPtr("mail." + zoneName),
				RecordType: "MX",
				TTL:        intPtr(300),
				Preference: intPtr(10),
			},
		},
	}

	// Create all records
	var createdRecords []*models.DNSRecord
	for _, rc := range recordConfigs {
		t.Logf("Creating %s record", rc.name)
		record, err := client.CreateDNSRecord(ctx, zone.UUID, rc.req)
		if err != nil {
			t.Fatalf("CreateDNSRecord (%s) failed: %v", rc.name, err)
		}
		t.Logf("%s record created: UUID=%s", rc.name, record.UUID)
		createdRecords = append(createdRecords, record)
	}

	// List records and verify count
	response, err := client.ListDNSRecords(ctx, zone.UUID)
	if err != nil {
		t.Fatalf("ListDNSRecords failed: %v", err)
	}

	t.Logf("Found %d records (expected at least %d)", response.Count, len(recordConfigs))

	if response.Count < len(recordConfigs) {
		t.Errorf("Expected at least %d records, got %d", len(recordConfigs), response.Count)
	}

	// Log all records
	for i, record := range response.Items {
		t.Logf("Record %d: UUID=%s, Type=%s, Owner=%s, Data=%s, TTL=%d",
			i+1, record.UUID, record.RecordType, record.Owner, record.Data, record.TTL)
	}

	// Cleanup: Delete all created records
	for _, record := range createdRecords {
		t.Logf("Deleting record: %s", record.UUID)
		err = client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
		if err != nil {
			t.Errorf("DeleteDNSRecord (UUID=%s) failed: %v", record.UUID, err)
		}
	}

	t.Log("All records deleted successfully")
}

// TestDNSRecordIntegration_AAAARecord tests creating an AAAA record with an IPv6 address
func TestDNSRecordIntegration_AAAARecord(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-aaaa-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	// Create AAAA record
	recordReq := &models.CreateDNSRecordRequest{
		Owner:      strPtr("ipv6host"),
		Data:       strPtr("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
		RecordType: "AAAA",
		TTL:        intPtr(300),
	}

	t.Log("Creating AAAA record")
	record, err := client.CreateDNSRecord(ctx, zone.UUID, recordReq)
	if err != nil {
		t.Fatalf("CreateDNSRecord failed: %v", err)
	}

	t.Logf("Record created: UUID=%s, Type=%s, Data=%s", record.UUID, record.RecordType, record.Data)

	if record.RecordType != "AAAA" {
		t.Errorf("Expected record type AAAA, got %s", record.RecordType)
	}
	if record.UUID == "" {
		t.Error("Record UUID should not be empty")
	}

	defer func() {
		t.Logf("Deleting DNS record: %s", record.UUID)
		err := client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
		if err != nil {
			t.Errorf("DeleteDNSRecord failed: %v", err)
		}
	}()

	// Re-fetch and verify
	fetched, err := client.GetDNSRecord(ctx, zone.UUID, record.UUID)
	if err != nil {
		t.Fatalf("GetDNSRecord failed: %v", err)
	}

	if fetched.RecordType != "AAAA" {
		t.Errorf("Expected record type AAAA after re-fetch, got %s", fetched.RecordType)
	}
	if fetched.UUID != record.UUID {
		t.Errorf("Expected UUID %s, got %s", record.UUID, fetched.UUID)
	}

	t.Log("AAAA record test passed")
}

// TestDNSRecordIntegration_TXTRecord tests creating a TXT record with SPF data
func TestDNSRecordIntegration_TXTRecord(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-txt-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	// Create TXT record with SPF value
	txtData := "v=spf1 include:_spf.example.com ~all"
	recordReq := &models.CreateDNSRecordRequest{
		Owner:      strPtr("@"),
		Data:       strPtr(txtData),
		RecordType: "TXT",
		TTL:        intPtr(3600),
	}

	t.Log("Creating TXT record")
	record, err := client.CreateDNSRecord(ctx, zone.UUID, recordReq)
	if err != nil {
		t.Fatalf("CreateDNSRecord failed: %v", err)
	}

	t.Logf("Record created: UUID=%s, Type=%s, Data=%s", record.UUID, record.RecordType, record.Data)

	if record.RecordType != "TXT" {
		t.Errorf("Expected record type TXT, got %s", record.RecordType)
	}

	defer func() {
		t.Logf("Deleting DNS record: %s", record.UUID)
		err := client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
		if err != nil {
			t.Errorf("DeleteDNSRecord failed: %v", err)
		}
	}()

	// Re-fetch and verify data round-trips correctly
	fetched, err := client.GetDNSRecord(ctx, zone.UUID, record.UUID)
	if err != nil {
		t.Fatalf("GetDNSRecord failed: %v", err)
	}

	if fetched.RecordType != "TXT" {
		t.Errorf("Expected record type TXT after re-fetch, got %s", fetched.RecordType)
	}
	if fetched.Data != txtData {
		t.Errorf("Expected TXT data %q, got %q", txtData, fetched.Data)
	}

	t.Log("TXT record test passed")
}

// TestDNSRecordIntegration_NSRecord tests creating an NS record for subdomain delegation
func TestDNSRecordIntegration_NSRecord(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-ns-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	// Create NS record for subdomain delegation
	recordReq := &models.CreateDNSRecordRequest{
		Owner:      strPtr("sub"),
		Data:       strPtr("ns1.example.com."),
		RecordType: "NS",
		TTL:        intPtr(3600),
	}

	t.Log("Creating NS record")
	record, err := client.CreateDNSRecord(ctx, zone.UUID, recordReq)
	if err != nil {
		t.Fatalf("CreateDNSRecord failed: %v", err)
	}

	t.Logf("Record created: UUID=%s, Type=%s, Data=%s", record.UUID, record.RecordType, record.Data)

	if record.RecordType != "NS" {
		t.Errorf("Expected record type NS, got %s", record.RecordType)
	}

	defer func() {
		t.Logf("Deleting DNS record: %s", record.UUID)
		err := client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
		if err != nil {
			t.Errorf("DeleteDNSRecord failed: %v", err)
		}
	}()

	// Re-fetch and verify
	fetched, err := client.GetDNSRecord(ctx, zone.UUID, record.UUID)
	if err != nil {
		t.Fatalf("GetDNSRecord failed: %v", err)
	}

	if fetched.RecordType != "NS" {
		t.Errorf("Expected record type NS after re-fetch, got %s", fetched.RecordType)
	}
	if fetched.Data != "ns1.example.com." {
		t.Errorf("Expected NS data %q, got %q", "ns1.example.com.", fetched.Data)
	}

	t.Log("NS record test passed")
}

// TestDNSRecordIntegration_WithDescription tests creating a record with a description
func TestDNSRecordIntegration_WithDescription(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-desc-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	// Create A record with description
	desc := "Integration test record with description"
	recordReq := &models.CreateDNSRecordRequest{
		Owner:       strPtr("described"),
		Data:        strPtr("10.0.0.1"),
		RecordType:  "A",
		TTL:         intPtr(300),
		Description: strPtr(desc),
	}

	t.Log("Creating A record with description")
	record, err := client.CreateDNSRecord(ctx, zone.UUID, recordReq)
	if err != nil {
		t.Fatalf("CreateDNSRecord failed: %v", err)
	}

	t.Logf("Record created: UUID=%s, Description=%v", record.UUID, record.Description)

	if record.Description == nil || *record.Description != desc {
		got := ""
		if record.Description != nil {
			got = *record.Description
		}
		t.Errorf("Expected description %q on creation, got %q", desc, got)
	}

	defer func() {
		t.Logf("Deleting DNS record: %s", record.UUID)
		err := client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
		if err != nil {
			t.Errorf("DeleteDNSRecord failed: %v", err)
		}
	}()

	// Re-fetch and verify description persists
	fetched, err := client.GetDNSRecord(ctx, zone.UUID, record.UUID)
	if err != nil {
		t.Fatalf("GetDNSRecord failed: %v", err)
	}

	if fetched.Description == nil || *fetched.Description != desc {
		got := ""
		if fetched.Description != nil {
			got = *fetched.Description
		}
		t.Errorf("Expected description %q after re-fetch, got %q", desc, got)
	}

	t.Log("Description test passed")
}

// TestDNSRecordIntegration_ListRecords tests listing multiple records in a zone
func TestDNSRecordIntegration_ListRecords(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-list-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	// Create 3 records: 2 A + 1 CNAME
	recordConfigs := []struct {
		name string
		req  *models.CreateDNSRecordRequest
	}{
		{
			name: "A-www",
			req: &models.CreateDNSRecordRequest{
				Owner:      strPtr("www"),
				Data:       strPtr("192.168.1.1"),
				RecordType: "A",
				TTL:        intPtr(300),
			},
		},
		{
			name: "A-api",
			req: &models.CreateDNSRecordRequest{
				Owner:      strPtr("api"),
				Data:       strPtr("192.168.1.2"),
				RecordType: "A",
				TTL:        intPtr(300),
			},
		},
		{
			name: "CNAME-blog",
			req: &models.CreateDNSRecordRequest{
				Owner:      strPtr("blog"),
				Data:       strPtr("www." + zoneName),
				RecordType: "CNAME",
				TTL:        intPtr(300),
			},
		},
	}

	var createdRecords []*models.DNSRecord
	for _, rc := range recordConfigs {
		t.Logf("Creating %s record", rc.name)
		record, err := client.CreateDNSRecord(ctx, zone.UUID, rc.req)
		if err != nil {
			t.Fatalf("CreateDNSRecord (%s) failed: %v", rc.name, err)
		}
		t.Logf("%s record created: UUID=%s", rc.name, record.UUID)
		createdRecords = append(createdRecords, record)
	}

	// Cleanup records
	defer func() {
		for _, record := range createdRecords {
			t.Logf("Deleting record: %s", record.UUID)
			err := client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
			if err != nil {
				t.Errorf("DeleteDNSRecord (UUID=%s) failed: %v", record.UUID, err)
			}
		}
	}()

	// List records and verify count
	response, err := client.ListDNSRecords(ctx, zone.UUID)
	if err != nil {
		t.Fatalf("ListDNSRecords failed: %v", err)
	}

	t.Logf("Found %d records (expected at least 3)", response.Count)

	if response.Count < 3 {
		t.Errorf("Expected at least 3 records, got %d", response.Count)
	}

	// Verify each created record UUID appears in the list
	for _, created := range createdRecords {
		found := false
		for _, listed := range response.Items {
			if listed.UUID == created.UUID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Created record UUID %s not found in list", created.UUID)
		}
	}

	// Log all records for visibility
	for i, record := range response.Items {
		t.Logf("Record %d: UUID=%s, Type=%s, Owner=%s, Data=%s, TTL=%d",
			i+1, record.UUID, record.RecordType, record.Owner, record.Data, record.TTL)
	}

	t.Log("ListRecords test passed")
}

// TestDNSRecordIntegration_GetNonExistent tests getting a non-existent DNS record
func TestDNSRecordIntegration_GetNonExistent(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-noexist-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	nonExistentID := "00000000-0000-0000-0000-000000000000"

	t.Logf("Attempting to get non-existent DNS record: %s", nonExistentID)

	_, err = client.GetDNSRecord(ctx, zone.UUID, nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent DNS record, got nil")
		return
	}

	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent DNS record")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// TestDNSRecordIntegration_DifferentTTLValues tests creating records with different TTL values
func TestDNSRecordIntegration_DifferentTTLValues(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create zone
	zoneName := fmt.Sprintf("test-ttl-%d.example.com.", time.Now().Unix())
	zone, err := client.CreateDNSZone(ctx, &models.CreateDNSZoneRequest{
		ZoneName: zoneName,
		ZoneType: "forward",
	})
	if err != nil {
		t.Fatalf("CreateDNSZone failed: %v", err)
	}

	t.Logf("DNS zone created: UUID=%s", zone.UUID)

	defer func() {
		t.Logf("Deleting DNS zone: %s", zone.UUID)
		err := client.DeleteDNSZone(ctx, zone.UUID)
		if err != nil {
			t.Errorf("DeleteDNSZone failed: %v", err)
		}
	}()

	// Create 3 A records with different TTLs
	ttlConfigs := []struct {
		owner string
		ip    string
		ttl   int
	}{
		{"short-ttl", "10.0.0.1", 60},
		{"medium-ttl", "10.0.0.2", 300},
		{"long-ttl", "10.0.0.3", 86400},
	}

	var createdRecords []*models.DNSRecord
	for _, tc := range ttlConfigs {
		t.Logf("Creating A record: owner=%s, TTL=%d", tc.owner, tc.ttl)
		record, err := client.CreateDNSRecord(ctx, zone.UUID, &models.CreateDNSRecordRequest{
			Owner:      strPtr(tc.owner),
			Data:       strPtr(tc.ip),
			RecordType: "A",
			TTL:        intPtr(tc.ttl),
		})
		if err != nil {
			t.Fatalf("CreateDNSRecord (owner=%s) failed: %v", tc.owner, err)
		}
		t.Logf("Record created: UUID=%s, TTL=%d", record.UUID, record.TTL)

		if record.TTL != tc.ttl {
			t.Errorf("Expected TTL %d on creation, got %d (owner=%s)", tc.ttl, record.TTL, tc.owner)
		}

		createdRecords = append(createdRecords, record)
	}

	// Cleanup records
	defer func() {
		for _, record := range createdRecords {
			t.Logf("Deleting record: %s", record.UUID)
			err := client.DeleteDNSRecord(ctx, zone.UUID, record.UUID)
			if err != nil {
				t.Errorf("DeleteDNSRecord (UUID=%s) failed: %v", record.UUID, err)
			}
		}
	}()

	// Re-fetch each record and verify TTL persists
	for i, tc := range ttlConfigs {
		fetched, err := client.GetDNSRecord(ctx, zone.UUID, createdRecords[i].UUID)
		if err != nil {
			t.Fatalf("GetDNSRecord (owner=%s) failed: %v", tc.owner, err)
		}

		if fetched.TTL != tc.ttl {
			t.Errorf("Expected TTL %d after re-fetch, got %d (owner=%s)", tc.ttl, fetched.TTL, tc.owner)
		}
	}

	t.Log("Different TTL values test passed")
}
