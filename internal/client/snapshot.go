package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// BillingUnitHourly is the billing unit the API expects for compute snapshots.
const BillingUnitHourly = "HRC"

// snapshotPollInterval is a var rather than a const so tests can shrink it.
var snapshotPollInterval = 10 * time.Second

// SnapshotClientForCompute returns a copy of the client carrying the subnet-id
// header that the snapshot detail endpoints require to resolve the backing
// provider; without it they answer 404 "Failed to get provider". The value is
// the compute's network_id, read from the compute detail endpoint, which needs
// no special headers itself.
//
// Callers should reuse the returned client for the whole operation so that
// polling loops do not re-fetch the compute on every iteration.
func (c *Client) SnapshotClientForCompute(ctx context.Context, computeID string) (*Client, error) {
	compute, err := c.GetCompute(ctx, computeID)
	if err != nil {
		return nil, fmt.Errorf("resolving subnet-id for compute %s: %w", computeID, err)
	}

	subnetID := compute.NetworkID
	if subnetID == "" {
		subnetID = compute.SubnetID
	}
	if subnetID == "" {
		return nil, fmt.Errorf("compute %s returned no network_id, cannot derive the subnet-id header required by the snapshot API", computeID)
	}

	return c.WithSubnetID(subnetID), nil
}

// snapshotDetailPath is the per-snapshot route. It is not nested under the
// compute: {base}/{compute_id}/snapshot/{uuid}/ does not exist.
func (c *Client) snapshotDetailPath(snapshotUUID string) string {
	return fmt.Sprintf("%s/snapshot/%s/", c.computeBasePath(), snapshotUUID)
}

// snapshotCollectionPath lists every snapshot in the project.
func (c *Client) snapshotCollectionPath() string {
	return fmt.Sprintf("%s/snapshot/", c.computeBasePath())
}

// computeSnapshotCollectionPath lists and creates snapshots of one compute.
func (c *Client) computeSnapshotCollectionPath(computeID string) string {
	return fmt.Sprintf("%s/%s/snapshot/", c.computeBasePath(), computeID)
}

// snapshotStatusReady reports whether status is terminal-good. The API
// capitalises its statuses ("Active"), so compare case-insensitively.
func snapshotStatusReady(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "available", "created", "completed":
		return true
	}
	return false
}

// snapshotStatusFailed reports whether status is terminal-bad. Any status that
// is neither ready nor failed is treated as transient and polled again.
func snapshotStatusFailed(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "error", "error_deleting", "failed":
		return true
	}
	return false
}

// IsProviderLookupError reports whether err is the API's provider-resolution
// failure: a 404 whose message is "Failed to get provider". It means the
// subnet-id header was missing or wrong, NOT that the snapshot is gone, so it
// must never be mistaken for a successful deletion.
func IsProviderLookupError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "failed to get provider")
}

// snapshotSleep waits one poll interval, honouring ctx and the deadline.
func snapshotSleep(ctx context.Context, deadline time.Time) error {
	if !time.Now().Add(snapshotPollInterval).Before(deadline) {
		return fmt.Errorf("timed out")
	}

	timer := time.NewTimer(snapshotPollInterval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// decodeSnapshotResponse tolerates the create endpoint answering with a bare
// object, an array (as the sibling compute create does), or something else
// entirely. It returns (nil, nil) when the body carries no snapshot object, in
// which case the caller falls back to locating the snapshot by name.
func decodeSnapshotResponse(raw json.RawMessage) (*models.ComputeSnapshot, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, nil
	}

	switch trimmed[0] {
	case '[':
		var snapshots []models.ComputeSnapshot
		if err := json.Unmarshal(trimmed, &snapshots); err != nil {
			return nil, fmt.Errorf("decoding snapshot array response: %w", err)
		}
		if len(snapshots) == 0 {
			return nil, nil
		}
		return &snapshots[0], nil
	case '{':
		var snapshot models.ComputeSnapshot
		if err := json.Unmarshal(trimmed, &snapshot); err != nil {
			return nil, fmt.Errorf("decoding snapshot object response: %w", err)
		}
		return &snapshot, nil
	default:
		return nil, nil
	}
}

// CreateComputeSnapshot creates a snapshot of a compute instance. The receiver
// should be compute-scoped; see SnapshotClientForCompute.
func (c *Client) CreateComputeSnapshot(ctx context.Context, computeID string, req *models.CreateComputeSnapshotRequest) (*models.ComputeSnapshot, error) {
	formData := structToFormData(req)

	var raw json.RawMessage
	if err := c.PostURLEncodedForm(ctx, c.computeSnapshotCollectionPath(computeID), formData, &raw); err != nil {
		return nil, err
	}

	return decodeSnapshotResponse(raw)
}

// GetComputeSnapshot retrieves a snapshot by UUID. The receiver MUST be
// compute-scoped: without the subnet-id header this endpoint answers
// 404 "Failed to get provider".
func (c *Client) GetComputeSnapshot(ctx context.Context, snapshotUUID string) (*models.ComputeSnapshot, error) {
	var snapshot models.ComputeSnapshot
	if err := c.Get(ctx, c.snapshotDetailPath(snapshotUUID), &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// ListComputeSnapshots retrieves every snapshot in the project. It needs no
// subnet-id header, which makes it the fallback whenever the owning compute
// (and therefore the header) is unknown or gone.
func (c *Client) ListComputeSnapshots(ctx context.Context) ([]models.ComputeSnapshot, error) {
	var snapshots []models.ComputeSnapshot
	if err := c.Get(ctx, c.snapshotCollectionPath(), &snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}

// ResolveSnapshotImageID resolves a snapshot name to its backing image_id using
// the project-wide snapshot list endpoint.
func (c *Client) ResolveSnapshotImageID(ctx context.Context, snapshotName string) (string, error) {
	name := strings.TrimSpace(snapshotName)
	if name == "" {
		return "", fmt.Errorf("snapshot name is empty")
	}

	snapshots, err := c.ListComputeSnapshots(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list snapshots: %w", err)
	}

	bestID := -1
	bestImageID := ""
	matches := 0
	for i := range snapshots {
		snapshot := &snapshots[i]
		if snapshot.SnapshotName != name {
			continue
		}

		imageID := strings.TrimSpace(snapshot.ImageIDString())
		if imageID == "" {
			continue
		}

		matches++
		if snapshot.ID > bestID {
			bestID = snapshot.ID
			bestImageID = imageID
		}
	}

	if bestImageID == "" {
		return "", fmt.Errorf("snapshot with name %q not found or has no image_id", snapshotName)
	}

	if matches > 1 {
		return "", fmt.Errorf("multiple snapshots found with name %q; use image_id/image_name for deterministic selection", snapshotName)
	}

	return bestImageID, nil
}

// ListComputeSnapshotsForCompute retrieves the snapshots of one compute. It
// needs no subnet-id header.
func (c *Client) ListComputeSnapshotsForCompute(ctx context.Context, computeID string) ([]models.ComputeSnapshot, error) {
	var snapshots []models.ComputeSnapshot
	if err := c.Get(ctx, c.computeSnapshotCollectionPath(computeID), &snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}

// FindComputeSnapshotByUUID scans the project-wide list. It returns (nil, nil)
// when the snapshot is absent.
func (c *Client) FindComputeSnapshotByUUID(ctx context.Context, snapshotUUID string) (*models.ComputeSnapshot, error) {
	snapshots, err := c.ListComputeSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	for i := range snapshots {
		if snapshots[i].UUID == snapshotUUID {
			return &snapshots[i], nil
		}
	}
	return nil, nil
}

// ResolveNewSnapshotUUID locates a just-created snapshot when the create
// response carried no UUID. exclude holds the UUIDs that already existed on the
// compute before the create, so a name collision with an older snapshot cannot
// produce a false match. Creation is asynchronous, so this polls.
func (c *Client) ResolveNewSnapshotUUID(ctx context.Context, computeID, snapshotName string, exclude map[string]struct{}, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)

	for {
		snapshots, err := c.ListComputeSnapshotsForCompute(ctx, computeID)
		if err != nil {
			return "", err
		}

		newest := ""
		newestID := -1
		for i := range snapshots {
			snapshot := &snapshots[i]
			if snapshot.UUID == "" || snapshot.SnapshotName != snapshotName {
				continue
			}
			if _, existed := exclude[snapshot.UUID]; existed {
				continue
			}
			if snapshot.ID > newestID {
				newestID, newest = snapshot.ID, snapshot.UUID
			}
		}
		if newest != "" {
			return newest, nil
		}

		if err := snapshotSleep(ctx, deadline); err != nil {
			return "", fmt.Errorf("snapshot %q did not appear on compute %s within %v: %w", snapshotName, computeID, timeout, err)
		}
	}
}

// WaitForSnapshotReady polls until the snapshot reaches a terminal status. The
// receiver MUST be compute-scoped.
func (c *Client) WaitForSnapshotReady(ctx context.Context, snapshotUUID string, timeout time.Duration) (*models.ComputeSnapshot, error) {
	deadline := time.Now().Add(timeout)

	for {
		snapshot, err := c.GetComputeSnapshot(ctx, snapshotUUID)
		if err != nil {
			return nil, err
		}

		if snapshotStatusReady(snapshot.Status) {
			return snapshot, nil
		}
		if snapshotStatusFailed(snapshot.Status) {
			return nil, fmt.Errorf("snapshot %s entered error state: %s", snapshotUUID, snapshot.Status)
		}

		if err := snapshotSleep(ctx, deadline); err != nil {
			return nil, fmt.Errorf("snapshot %s did not become ready within %v (last status: %q): %w", snapshotUUID, timeout, snapshot.Status, err)
		}
	}
}

// DeleteComputeSnapshot deletes a snapshot and waits for it to disappear. The
// receiver MUST be compute-scoped for the DELETE itself; completion is then
// confirmed against the project-wide list, which needs no header, so a
// "Failed to get provider" 404 can never be read as a successful deletion.
func (c *Client) DeleteComputeSnapshot(ctx context.Context, snapshotUUID string, timeout time.Duration) error {
	if err := c.Delete(ctx, c.snapshotDetailPath(snapshotUUID)); err != nil {
		if IsNotFoundError(err) && !IsProviderLookupError(err) {
			return nil
		}
		return err
	}

	deadline := time.Now().Add(timeout)
	for {
		snapshot, err := c.FindComputeSnapshotByUUID(ctx, snapshotUUID)
		if err != nil {
			return err
		}
		if snapshot == nil {
			return nil
		}

		if err := snapshotSleep(ctx, deadline); err != nil {
			return fmt.Errorf("snapshot %s was not deleted within %v: %w", snapshotUUID, timeout, err)
		}
	}
}
