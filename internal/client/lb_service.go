package client

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// lbBasePath returns the base path for load balancer endpoints
func (c *Client) lbBasePath() string {
	return fmt.Sprintf("/api/v2.1/load-balancers/domain/%s/project/%s/load-balancers",
		c.Organization, c.ProjectName)
}

// --- LB Flavors ---

// ListLBFlavors retrieves all available LB flavors
func (c *Client) ListLBFlavors(ctx context.Context) ([]models.LBFlavor, error) {
	var flavors []models.LBFlavor
	err := c.Get(ctx, fmt.Sprintf("%s/flavors/?type=lb", c.lbBasePath()), &flavors)
	if err != nil {
		return nil, err
	}
	return flavors, nil
}

// --- LB Service CRUD ---

// CreateLBService creates a new load balancer service
func (c *Client) CreateLBService(ctx context.Context, req *models.CreateLBServiceRequest) (*models.LBService, error) {
	formData := structToFormData(req)

	var lbService models.LBService
	err := c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/lb_service/", c.lbBasePath()), formData, &lbService)
	if err != nil {
		return nil, err
	}
	return &lbService, nil
}

// GetLBService retrieves an LB service by ID
func (c *Client) GetLBService(ctx context.Context, id string) (*models.LBService, error) {
	var lbService models.LBService
	err := c.Get(ctx, fmt.Sprintf("%s/lb_service/%s", c.lbBasePath(), id), &lbService)
	if err != nil {
		return nil, err
	}
	return &lbService, nil
}

// ListLBServices retrieves all LB services
func (c *Client) ListLBServices(ctx context.Context) ([]models.LBService, error) {
	var services []models.LBService
	err := c.Get(ctx, fmt.Sprintf("%s/lb_service/", c.lbBasePath()), &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// DeleteLBService deletes an LB service
func (c *Client) DeleteLBService(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("%s/lb_service/%s", c.lbBasePath(), id))
}

// ResolveLBServiceID resolves a load balancer service name to its ID
func (c *Client) ResolveLBServiceID(ctx context.Context, name string) (string, error) {
	services, err := c.ListLBServices(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list LB services: %w", err)
	}
	tflog.Debug(ctx, "ResolveLBServiceID: listing LB services", map[string]interface{}{
		"service_count": len(services),
		"searched_name": name,
	})
	for _, svc := range services {
		if svc.Name == name {
			if svc.ID == "" {
				return "", fmt.Errorf("LB service %q found but has empty ID (API response field mismatch)", name)
			}
			tflog.Debug(ctx, "ResolveLBServiceID: resolved LB service", map[string]interface{}{
				"name": name,
				"id":   svc.ID,
			})
			return svc.ID, nil
		}
	}
	return "", fmt.Errorf("LB service with name %q not found", name)
}

// WaitForLBServiceReady polls until the LB service reaches Active status
func (c *Client) WaitForLBServiceReady(ctx context.Context, id string, timeout time.Duration) (*models.LBService, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		svc, err := c.GetLBService(ctx, id)
		if err != nil {
			return nil, err
		}

		switch svc.Status {
		case "Active", "active", "ACTIVE", "Created", "created", "CREATED":
			return svc, nil
		case "Error", "error", "ERROR":
			return nil, fmt.Errorf("LB service entered error state")
		}

		time.Sleep(15 * time.Second)
	}

	return nil, fmt.Errorf("LB service did not become ready within %v", timeout)
}

// WaitForLBServiceDeleted polls until the LB service is deleted (404 or Deleted status)
func (c *Client) WaitForLBServiceDeleted(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		svc, err := c.GetLBService(ctx, id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}

		// Also treat "Deleted" status as successful deletion
		switch svc.Status {
		case "Deleted", "deleted", "DELETED":
			return nil
		}

		time.Sleep(15 * time.Second)
	}

	return fmt.Errorf("LB service deletion timed out after %v", timeout)
}

// --- LB VIP ---

// lbVipBasePath returns the v1 base path for LB VIP endpoints
func (c *Client) lbVipBasePath() string {
	return "/api/v1/load-balancers/lb_service"
}

// CreateLBVip creates a VIP port for an LB service
func (c *Client) CreateLBVip(ctx context.Context, lbServiceID string) (*models.LBVip, error) {
	var vip models.LBVip
	err := c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/vip", c.lbVipBasePath(), lbServiceID), nil, &vip)
	if err != nil {
		return nil, err
	}
	return &vip, nil
}

// ListLBVips lists all VIPs for an LB service
func (c *Client) ListLBVips(ctx context.Context, lbServiceID string) ([]models.LBVip, error) {
	var vips []models.LBVip
	err := c.Get(ctx, fmt.Sprintf("%s/%s/vip", c.lbVipBasePath(), lbServiceID), &vips)
	if err != nil {
		return nil, err
	}
	return vips, nil
}

// DeleteLBVip deletes a VIP port
func (c *Client) DeleteLBVip(ctx context.Context, lbServiceID string, vipID int) error {
	return c.Delete(ctx, fmt.Sprintf("%s/%s/vip/%s/", c.lbVipBasePath(), lbServiceID, strconv.Itoa(vipID)))
}

// ResolveVipPortID resolves an LB VIP's port id by matching its fixed_ips against
// the given fixed IP. The caller must scope the client with the LB service's
// network (WithSubnetID) so ListLBVips returns the VIPs.
func (c *Client) ResolveVipPortID(ctx context.Context, lbServiceID, fixedIP string) (int, error) {
	vips, err := c.ListLBVips(ctx, lbServiceID)
	if err != nil {
		return 0, fmt.Errorf("failed to list LB VIPs to resolve vip_port_id for fixed IP %s: %w", fixedIP, err)
	}
	tflog.Debug(ctx, "ResolveVipPortID: listing LB VIPs", map[string]interface{}{
		"vip_count":     len(vips),
		"searched_ip":   fixedIP,
		"lb_service_id": lbServiceID,
	})
	for _, vip := range vips {
		for _, ip := range vip.FixedIPs {
			if ip == fixedIP {
				return vip.ID, nil
			}
		}
	}
	return 0, fmt.Errorf("no LB VIP found with fixed IP %q on LB service %s", fixedIP, lbServiceID)
}

// --- LB Certificate ---

// CreateLBCertificate creates an SSL certificate for an LB service
func (c *Client) CreateLBCertificate(ctx context.Context, lbServiceID string, req *models.CreateLBCertificateRequest) (*models.LBCertificate, error) {
	formData := structToFormData(req)

	var cert models.LBCertificate
	err := c.PostForm(ctx, fmt.Sprintf("%s/%s/certificates", c.lbBasePath(), lbServiceID), formData, &cert)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// ListLBCertificates lists all certificates for an LB service
func (c *Client) ListLBCertificates(ctx context.Context, lbServiceID string) ([]models.LBCertificate, error) {
	var certs []models.LBCertificate
	err := c.Get(ctx, fmt.Sprintf("%s/%s/certificates", c.lbBasePath(), lbServiceID), &certs)
	if err != nil {
		return nil, err
	}
	return certs, nil
}

// DeleteLBCertificate deletes an SSL certificate
func (c *Client) DeleteLBCertificate(ctx context.Context, lbServiceID string, certID int) error {
	return c.Delete(ctx, fmt.Sprintf("%s/%s/certificates/%s/", c.lbBasePath(), lbServiceID, strconv.Itoa(certID)))
}
