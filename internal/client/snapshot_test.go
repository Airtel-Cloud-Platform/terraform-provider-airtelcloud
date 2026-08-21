package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func newTestClient(t *testing.T, ms *testutil.MockServer) *Client {
	t.Helper()
	baseURL := strings.TrimSuffix(ms.URL, "/")
	c, err := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return c
}

// fastSnapshotPolling shrinks the poll interval for the duration of a test.
func fastSnapshotPolling(t *testing.T) {
	t.Helper()
	original := snapshotPollInterval
	snapshotPollInterval = time.Millisecond
	t.Cleanup(func() { snapshotPollInterval = original })
}

// TestComputeSnapshotUnmarshalRealShape is the regression test for the model
// type mismatches (labels was string vs object, image_id was string vs number):
// decoding the real wire shape must not error.
func TestComputeSnapshotUnmarshalRealShape(t *testing.T) {
	var snapshot models.ComputeSnapshot
	if err := json.Unmarshal([]byte(testutil.SnapshotJSON), &snapshot); err != nil {
		t.Fatalf("unmarshalling real snapshot JSON: %v", err)
	}

	if snapshot.UUID != "snap-uuid-1234" {
		t.Errorf("UUID = %q, want snap-uuid-1234", snapshot.UUID)
	}
	if snapshot.Status != "Active" {
		t.Errorf("Status = %q, want Active", snapshot.Status)
	}
	if got := snapshot.ImageIDString(); got != "1407" {
		t.Errorf("ImageIDString() = %q, want 1407", got)
	}
	if snapshot.Created != "1784023036" {
		t.Errorf("Created = %q, want 1784023036", snapshot.Created)
	}
	if len(snapshot.Labels) == 0 {
		t.Error("Labels is empty, want the raw object")
	}
	if got := snapshot.ComputeIDString(); got != "test-id" {
		t.Errorf("ComputeIDString() = %q, want test-id", got)
	}
}

func TestSnapshotClientForCompute(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	scoped, err := newTestClient(t, ms).SnapshotClientForCompute(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("SnapshotClientForCompute() error = %v", err)
	}
	if scoped.SubnetID != "network-1" {
		t.Errorf("scoped SubnetID = %q, want network-1 (the compute's network_id)", scoped.SubnetID)
	}
}

func TestCreateComputeSnapshot(t *testing.T) {
	t.Run("object response", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		scoped, _ := newTestClient(t, ms).SnapshotClientForCompute(context.Background(), "test-id")

		snapshot, err := scoped.CreateComputeSnapshot(context.Background(), "test-id", &models.CreateComputeSnapshotRequest{SnapshotName: "test-snapshot"})
		if err != nil {
			t.Fatalf("CreateComputeSnapshot() error = %v", err)
		}
		if snapshot == nil || snapshot.UUID != "snap-uuid-1234" {
			t.Fatalf("CreateComputeSnapshot() = %+v, want UUID snap-uuid-1234", snapshot)
		}
	})

	t.Run("array response is unwrapped", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		ms.AddHandler("POST", "/api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/snapshot/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[" + testutil.SnapshotJSON + "]"))
		})
		scoped, _ := newTestClient(t, ms).SnapshotClientForCompute(context.Background(), "test-id")

		snapshot, err := scoped.CreateComputeSnapshot(context.Background(), "test-id", &models.CreateComputeSnapshotRequest{SnapshotName: "test-snapshot"})
		if err != nil {
			t.Fatalf("CreateComputeSnapshot() error = %v", err)
		}
		if snapshot == nil || snapshot.UUID != "snap-uuid-1234" {
			t.Fatalf("CreateComputeSnapshot() = %+v, want UUID snap-uuid-1234", snapshot)
		}
	})

	// Regression for the empty-body bug: a create without snapshot_name must 422.
	t.Run("missing snapshot_name is a 422", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		scoped, _ := newTestClient(t, ms).SnapshotClientForCompute(context.Background(), "test-id")

		_, err := scoped.CreateComputeSnapshot(context.Background(), "test-id", &models.CreateComputeSnapshotRequest{})
		if err == nil {
			t.Fatal("CreateComputeSnapshot() with no name = nil error, want 422")
		}
	})
}

func TestResolveNewSnapshotUUID(t *testing.T) {
	fastSnapshotPolling(t)
	ms := testutil.NewMockServer()
	defer ms.Close()
	c := newTestClient(t, ms)

	// snap-uuid-1234 pre-exists; the caller wants the other one with the same name.
	exclude := map[string]struct{}{"snap-uuid-1234": {}}
	uuid, err := c.ResolveNewSnapshotUUID(context.Background(), "test-id", "test-snapshot-2", exclude, time.Second)
	if err != nil {
		t.Fatalf("ResolveNewSnapshotUUID() error = %v", err)
	}
	if uuid != "snap-uuid-5678" {
		t.Errorf("ResolveNewSnapshotUUID() = %q, want snap-uuid-5678", uuid)
	}
}

func TestGetComputeSnapshot(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	c := newTestClient(t, ms)

	// Regression for the missing subnet-id header: unscoped GET must fail.
	t.Run("unscoped client is refused", func(t *testing.T) {
		_, err := c.GetComputeSnapshot(context.Background(), "snap-uuid-1234")
		if err == nil {
			t.Fatal("GetComputeSnapshot() without subnet-id = nil error, want provider-lookup failure")
		}
		if !IsProviderLookupError(err) {
			t.Errorf("expected a provider-lookup error, got %v", err)
		}
	})

	t.Run("scoped client succeeds", func(t *testing.T) {
		scoped, _ := c.SnapshotClientForCompute(context.Background(), "test-id")
		snapshot, err := scoped.GetComputeSnapshot(context.Background(), "snap-uuid-1234")
		if err != nil {
			t.Fatalf("GetComputeSnapshot() error = %v", err)
		}
		if snapshot.SnapshotName != "test-snapshot" {
			t.Errorf("SnapshotName = %q, want test-snapshot", snapshot.SnapshotName)
		}
	})
}

func TestListComputeSnapshots(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	snapshots, err := newTestClient(t, ms).ListComputeSnapshots(context.Background())
	if err != nil {
		t.Fatalf("ListComputeSnapshots() error = %v", err)
	}
	if len(snapshots) != 2 {
		t.Errorf("ListComputeSnapshots() count = %d, want 2", len(snapshots))
	}
}

func TestResolveSnapshotImageID(t *testing.T) {
	t.Run("resolves snapshot name", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		imageID, err := newTestClient(t, ms).ResolveSnapshotImageID(context.Background(), "test-snapshot")
		if err != nil {
			t.Fatalf("ResolveSnapshotImageID() error = %v", err)
		}
		if imageID != "1407" {
			t.Fatalf("ResolveSnapshotImageID() = %q, want 1407", imageID)
		}
	})

	t.Run("errors when duplicate snapshot names exist", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		ms.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/snapshot/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[
				{"id": 1, "uuid": "snap-1", "snapshot_name": "dup-snap", "image_id": 1001},
				{"id": 2, "uuid": "snap-2", "snapshot_name": "dup-snap", "image_id": 1002}
			]`))
		})

		_, err := newTestClient(t, ms).ResolveSnapshotImageID(context.Background(), "dup-snap")
		if err == nil {
			t.Fatal("ResolveSnapshotImageID() with duplicate snapshot names = nil error, want failure")
		}
	})
}

func TestFindComputeSnapshotByUUID(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	c := newTestClient(t, ms)

	found, err := c.FindComputeSnapshotByUUID(context.Background(), "snap-uuid-1234")
	if err != nil {
		t.Fatalf("FindComputeSnapshotByUUID() error = %v", err)
	}
	if found == nil || found.UUID != "snap-uuid-1234" {
		t.Fatalf("FindComputeSnapshotByUUID() = %+v, want snap-uuid-1234", found)
	}

	missing, err := c.FindComputeSnapshotByUUID(context.Background(), "does-not-exist")
	if err != nil {
		t.Fatalf("FindComputeSnapshotByUUID() error = %v", err)
	}
	if missing != nil {
		t.Errorf("FindComputeSnapshotByUUID() = %+v, want nil", missing)
	}
}

func TestWaitForSnapshotReady(t *testing.T) {
	fastSnapshotPolling(t)

	// Regression for the case-sensitive status bug: "Active" must be accepted.
	t.Run("capitalised Active is ready", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		scoped, _ := newTestClient(t, ms).SnapshotClientForCompute(context.Background(), "test-id")

		snapshot, err := scoped.WaitForSnapshotReady(context.Background(), "snap-uuid-1234", time.Second)
		if err != nil {
			t.Fatalf("WaitForSnapshotReady() error = %v", err)
		}
		if snapshot.UUID != "snap-uuid-1234" {
			t.Errorf("UUID = %q, want snap-uuid-1234", snapshot.UUID)
		}
	})

	t.Run("error status fails fast", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		ms.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/snapshot/snap-uuid-1234/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"uuid":"snap-uuid-1234","status":"Error"}`))
		})
		scoped, _ := newTestClient(t, ms).SnapshotClientForCompute(context.Background(), "test-id")

		if _, err := scoped.WaitForSnapshotReady(context.Background(), "snap-uuid-1234", time.Second); err == nil {
			t.Fatal("WaitForSnapshotReady() on Error status = nil error, want failure")
		}
	})
}

func TestDeleteComputeSnapshot(t *testing.T) {
	fastSnapshotPolling(t)
	ms := testutil.NewMockServer()
	defer ms.Close()
	scoped, _ := newTestClient(t, ms).SnapshotClientForCompute(context.Background(), "test-id")

	if err := scoped.DeleteComputeSnapshot(context.Background(), "snap-uuid-1234", time.Second); err != nil {
		t.Fatalf("DeleteComputeSnapshot() error = %v", err)
	}
	if !ms.DeletedSnapshots["snap-uuid-1234"] {
		t.Error("DeleteComputeSnapshot() did not delete the snapshot on the server")
	}
}
