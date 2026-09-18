package client

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const kubernetesPollInterval = 15 * time.Second

func (c *Client) kubernetesBasePath() string {
	return fmt.Sprintf("/api/airtel/v1/domain/%s/project/%s/cluster", c.Organization, c.ProjectName)
}

func (c *Client) kubernetesClusterPath(name string) string {
	return fmt.Sprintf("%s/%s", c.kubernetesBasePath(), url.PathEscape(name))
}

func (c *Client) kubernetesHostGroupClustersPath(hostGroup string) string {
	return fmt.Sprintf("/api/airtel/v1/domain/%s/project/%s/hostgroup/%s/clusters",
		c.Organization, c.ProjectName, url.PathEscape(hostGroup))
}

func kubernetesNotFound(name string) error {
	return &APIError{StatusCode: 404, Message: fmt.Sprintf("kubernetes cluster %q not found", name), Code: 5}
}

// CreateKubernetesCluster creates a CKP cluster. The API returns an empty body.
func (c *Client) CreateKubernetesCluster(ctx context.Context, req *models.CreateKubernetesClusterRequest) error {
	return c.Post(ctx, c.kubernetesBasePath(), req, nil)
}

// ListKubernetesClustersByHostGroup lists clusters associated with a BYOH host group.
func (c *Client) ListKubernetesClustersByHostGroup(ctx context.Context, hostGroup string) (*models.KubernetesClusterList, error) {
	var list models.KubernetesClusterList
	if err := c.Get(ctx, c.kubernetesHostGroupClustersPath(hostGroup), &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetKubernetesCluster finds a cluster by name via host-group cluster lists.
// GET /cluster/{name} is not implemented by the API (HTTP 501).
func (c *Client) GetKubernetesCluster(ctx context.Context, name string, hostGroups []string) (*models.KubernetesCluster, error) {
	if len(hostGroups) == 0 {
		return nil, fmt.Errorf("at least one host group is required to look up kubernetes cluster %q", name)
	}

	var lastErr error
	for _, hostGroup := range hostGroups {
		if hostGroup == "" {
			continue
		}
		list, err := c.ListKubernetesClustersByHostGroup(ctx, hostGroup)
		if err != nil {
			lastErr = err
			continue
		}
		for i := range list.Items {
			item := list.Items[i]
			if item.ClusterName() == name {
				return &item, nil
			}
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, kubernetesNotFound(name)
}

// DeleteKubernetesCluster deletes a cluster by name. A 404 is treated as success.
func (c *Client) DeleteKubernetesCluster(ctx context.Context, name string) error {
	err := c.Delete(ctx, c.kubernetesClusterPath(name))
	if err != nil && !IsNotFoundError(err) {
		return err
	}
	return nil
}

// WaitForKubernetesReady polls host-group cluster lists until state is Ready.
func (c *Client) WaitForKubernetesReady(ctx context.Context, name string, hostGroups []string, timeout time.Duration) (*models.KubernetesCluster, error) {
	deadline := time.Now().Add(timeout)
	var cluster *models.KubernetesCluster

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		got, err := c.GetKubernetesCluster(ctx, name, hostGroups)
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
			switch cluster.State {
			case models.KubernetesStateReady:
				return cluster, nil
			case models.KubernetesStateDeleting:
				return nil, fmt.Errorf("kubernetes cluster %q entered Deleting state while waiting to become ready", name)
			}
		}

		if !time.Now().Add(kubernetesPollInterval).Before(deadline) {
			state := ""
			if cluster != nil {
				state = cluster.State
			}
			return nil, fmt.Errorf("kubernetes cluster %q did not become ready within %v (state %q)", name, timeout, state)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(kubernetesPollInterval):
		}
	}
}

// WaitForKubernetesDeleted polls until the cluster is gone from host-group lists or isDeleted is true.
func (c *Client) WaitForKubernetesDeleted(ctx context.Context, name string, hostGroups []string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		cluster, err := c.GetKubernetesCluster(ctx, name, hostGroups)
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
