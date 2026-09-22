package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const kubernetesPollInterval = 15 * time.Second

func (c *Client) kubernetesBasePath() string {
	return fmt.Sprintf("/api/airtel/v1/domain/%s/project/%s/cluster", c.Organization, c.ProjectName)
}

func (c *Client) kubernetesCompassBasePath() string {
	return fmt.Sprintf("/api/compass/v1/domain/%s/project/%s", c.Organization, c.ProjectName)
}

func (c *Client) kubernetesClusterStatusPath(name string) string {
	return fmt.Sprintf("%s/cluster/%s/status", c.kubernetesCompassBasePath(), url.PathEscape(name))
}

func (c *Client) kubernetesClusterDeletePath(name string) string {
	return fmt.Sprintf("%s/cluster/%s?forceDelete=false", c.kubernetesCompassBasePath(), url.PathEscape(name))
}

func kubernetesNotFound(name string) error {
	return &APIError{StatusCode: 404, Message: fmt.Sprintf("kubernetes cluster %q not found", name), Code: 5}
}

// CreateKubernetesCluster creates a CKP cluster. The API returns an empty body.
func (c *Client) CreateKubernetesCluster(ctx context.Context, req *models.CreateKubernetesClusterRequest) error {
	return c.Post(ctx, c.kubernetesBasePath(), req, nil)
}

// GetKubernetesCluster reads the cluster and registration state shown by the
// Compass UI.
func (c *Client) GetKubernetesCluster(ctx context.Context, name string) (*models.KubernetesCluster, error) {
	var status models.KubernetesClusterStatus
	if err := c.Get(ctx, c.kubernetesClusterStatusPath(name), &status); err != nil {
		return nil, err
	}
	status.Cluster.State = strings.TrimSpace(status.State)
	if status.Cluster.ClusterName() == "" {
		status.Cluster.Name = name
	}
	return &status.Cluster, nil
}

// DeleteKubernetesCluster deletes a cluster by name via Compass.
// A 404 is treated as success.
func (c *Client) DeleteKubernetesCluster(ctx context.Context, name string) error {
	err := c.Delete(ctx, c.kubernetesClusterDeletePath(name))
	if err != nil && !IsNotFoundError(err) {
		return err
	}
	return nil
}

// WaitForKubernetesReady polls Compass until the cluster registration state is
// Connected, matching the status shown by the UI.
func (c *Client) WaitForKubernetesReady(ctx context.Context, name string, timeout time.Duration) (*models.KubernetesCluster, error) {
	deadline := time.Now().Add(timeout)
	var cluster *models.KubernetesCluster

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		got, err := c.GetKubernetesCluster(ctx, name)
		if err != nil {
			if IsNotFoundError(err) {
				if !time.Now().Add(kubernetesPollInterval).Before(deadline) {
					return nil, fmt.Errorf("kubernetes cluster %q was not found while waiting to become ready", name)
				}
			} else {
				return nil, err
			}
		} else {
			cluster = got
			if cluster.IsDeleted {
				return nil, fmt.Errorf("kubernetes cluster %q was deleted while waiting to become ready", name)
			}
			state := cluster.LifecycleState()
			if models.KubernetesClusterIsReady(state) {
				return cluster, nil
			}
			if failedStateError := strings.TrimSpace(cluster.FailedStateError); failedStateError != "" {
				return nil, fmt.Errorf("kubernetes cluster %q failed while waiting to become ready: %s", name, failedStateError)
			}
			provisioningState := strings.TrimSpace(cluster.LifeCycleState)
			if strings.EqualFold(state, models.KubernetesStateDeleting) ||
				strings.EqualFold(provisioningState, models.KubernetesStateDeleting) {
				return nil, fmt.Errorf("kubernetes cluster %q entered Deleting state while waiting to become ready", name)
			}
			if models.KubernetesClusterIsFailed(state) || models.KubernetesClusterIsFailed(provisioningState) {
				return nil, fmt.Errorf(
					"kubernetes cluster %q entered error state %q while waiting to become ready",
					name,
					provisioningState,
				)
			}
		}

		if !time.Now().Add(kubernetesPollInterval).Before(deadline) {
			state := ""
			if cluster != nil {
				state = cluster.LifecycleState()
			}
			provisioningState := ""
			if cluster != nil {
				provisioningState = cluster.LifeCycleState
			}
			return nil, fmt.Errorf(
				"kubernetes cluster %q did not become ready within %v (state %q, lifecycle state %q)",
				name,
				timeout,
				state,
				provisioningState,
			)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(kubernetesPollInterval):
		}
	}
}

// WaitForKubernetesDeleted polls Compass until the cluster is gone or marked deleted.
func (c *Client) WaitForKubernetesDeleted(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		cluster, err := c.GetKubernetesCluster(ctx, name)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		if cluster.IsDeleted {
			return nil
		}

		if !time.Now().Add(kubernetesPollInterval).Before(deadline) {
			return fmt.Errorf("kubernetes cluster %q deletion timed out after %v (state %q)", name, timeout, cluster.State)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(kubernetesPollInterval):
		}
	}
}
