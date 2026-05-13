package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client represents the Airtel Cloud API client
type Client struct {
	BaseURL          *url.URL
	APIKey           string
	APISecret        string
	Region           string
	Organization     string
	ProjectName      string
	SubnetID         string
	AvailabilityZone string
	HTTPClient       *http.Client
	UserAgent        string
}

// APIError represents an error response from the API
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Code       int    `json:"code"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s (code: %d)", e.StatusCode, e.Message, e.Code)
}

// NewClient creates a new Airtel Cloud API client
func NewClient(endpoint, apiKey, apiSecret, region, organization, projectName, subnetID string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if apiSecret == "" {
		return nil, fmt.Errorf("API secret is required")
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid API endpoint: %w", err)
	}

	return &Client{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		APISecret:    apiSecret,
		Region:       region,
		Organization: organization,
		ProjectName:  projectName,
		SubnetID:     subnetID,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		UserAgent: "terraform-provider-airtelcloud/0.2.0",
	}, nil
}

// WithSubnetID returns a shallow copy of the client with SubnetID set.
func (c *Client) WithSubnetID(subnetID string) *Client {
	copy := *c
	copy.SubnetID = subnetID
	return &copy
}

// WithAvailabilityZone returns a shallow copy of the client with AvailabilityZone set.
func (c *Client) WithAvailabilityZone(az string) *Client {
	copy := *c
	copy.AvailabilityZone = az
	return &copy
}

// generateHMACAuth generates the Ce-Auth header value using HMAC-SHA256
// Format: apiKey.expiry.signature
func (c *Client) generateHMACAuth() string {
	// Generate expiry timestamp (current time + 120 seconds)
	expiry := time.Now().Unix() + 120

	// Create the message: apiKey.expiry
	data := c.APIKey + "." + strconv.FormatInt(expiry, 10)

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(c.APISecret))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))

	// Combine data and signature: apiKey.expiry.signature
	return data + "." + signature
}

// doRequest performs an HTTP request with proper authentication and error handling
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Ensure path starts with /api/v1
	if !strings.HasPrefix(path, "/api") {
		path = "/api" + path
	}

	// Parse the path to properly handle query strings
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	var contentType string

	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
		contentType = "application/json"
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Ce-Auth", c.generateHMACAuth())
	req.Header.Set("ce-region", c.Region)

	// Add organization header if specified
	if c.Organization != "" {
		req.Header.Set("organisation-name", c.Organization)
	}

	// Add project name header if specified
	if c.ProjectName != "" {
		req.Header.Set("Project-Name", c.ProjectName)
	}

	// Add subnet-id header if specified (required by volume API for provider lookup)
	if c.SubnetID != "" {
		req.Header.Set("subnet-id", c.SubnetID)
	}

	// Add availability zone header if specified
	if c.AvailabilityZone != "" {
		req.Header.Set("ce-availability-zone", c.AvailabilityZone)
	}

	// Log request details in debug mode
	tflog.Debug(ctx, "Making API request", map[string]interface{}{
		"method":       method,
		"url":          u.String(),
		"headers":      req.Header,
		"content_type": contentType,
	})

	// Log request body in debug mode (excluding sensitive data)
	if body != nil && buf != nil {
		if bodyBytes, ok := buf.(*bytes.Buffer); ok {
			tflog.Debug(ctx, "Request body", map[string]interface{}{
				"body": bodyBytes.String(),
			})
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		tflog.Error(ctx, "HTTP request failed", map[string]interface{}{
			"error": err.Error(),
			"url":   u.String(),
		})
		return nil, err
	}

	// Log response details in debug mode
	tflog.Debug(ctx, "Received API response", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     resp.Header,
	})

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		// Read response body for logging and error parsing
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read error response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		// Log error response body in debug mode
		tflog.Debug(ctx, "Error response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		var apiErr APIError
		if err := json.Unmarshal(bodyBytes, &apiErr); err != nil {
			tflog.Error(ctx, "Failed to parse error response", map[string]interface{}{
				"error": err.Error(),
				"body":  string(bodyBytes),
			})
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}

	return resp, nil
}

// doFormRequest performs an HTTP request with form-data encoding
func (c *Client) doFormRequest(ctx context.Context, method, path string, formData map[string]interface{}) (*http.Response, error) {
	// Ensure path starts with /api/v1
	if !strings.HasPrefix(path, "/api") {
		path = "/api" + path
	}

	// Parse the path to properly handle query strings
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	u := c.BaseURL.ResolveReference(rel)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields
	for key, value := range formData {
		if value == nil {
			continue
		}

		switch v := value.(type) {
		case string:
			if v != "" {
				writer.WriteField(key, v)
			}
		case []string:
			for _, item := range v {
				if item != "" {
					writer.WriteField(key, item)
				}
			}
		case int:
			writer.WriteField(key, fmt.Sprintf("%d", v))
		case int64:
			writer.WriteField(key, fmt.Sprintf("%d", v))
		case bool:
			writer.WriteField(key, fmt.Sprintf("%t", v))
		case fmt.Stringer:
			writer.WriteField(key, v.String())
		default:
			writer.WriteField(key, fmt.Sprintf("%v", v))
		}
	}

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Ce-Auth", c.generateHMACAuth())
	req.Header.Set("ce-region", c.Region)

	// Add organization header if specified
	if c.Organization != "" {
		req.Header.Set("organisation-name", c.Organization)
	}

	// Add project name header if specified
	if c.ProjectName != "" {
		req.Header.Set("Project-Name", c.ProjectName)
	}

	// Add subnet-id header if specified (required by volume API for provider lookup)
	if c.SubnetID != "" {
		req.Header.Set("subnet-id", c.SubnetID)
	}

	// Add availability zone header if specified
	if c.AvailabilityZone != "" {
		req.Header.Set("ce-availability-zone", c.AvailabilityZone)
	}

	// Log form request details in debug mode
	tflog.Debug(ctx, "Making form API request", map[string]interface{}{
		"method":       method,
		"url":          u.String(),
		"headers":      req.Header,
		"form_fields":  formData,
		"content_type": writer.FormDataContentType(),
	})

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		tflog.Error(ctx, "Form HTTP request failed", map[string]interface{}{
			"error": err.Error(),
			"url":   u.String(),
		})
		return nil, err
	}

	// Log response details in debug mode
	tflog.Debug(ctx, "Received form API response", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     resp.Header,
	})

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		// Read response body for logging and error parsing
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read form error response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		// Log error response body in debug mode
		tflog.Debug(ctx, "Form error response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		var apiErr APIError
		if err := json.Unmarshal(bodyBytes, &apiErr); err != nil {
			tflog.Error(ctx, "Failed to parse form error response", map[string]interface{}{
				"error": err.Error(),
				"body":  string(bodyBytes),
			})
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}

	return resp, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, v interface{}) error {
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		// Read response body for logging
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return readErr
		}

		// Log successful response body in debug mode
		tflog.Debug(ctx, "GET response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		return json.Unmarshal(bodyBytes, v)
	}

	return nil
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}, v interface{}) error {
	resp, err := c.doRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		// Read response body for logging
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read POST response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return readErr
		}

		// Log successful response body in debug mode
		tflog.Debug(ctx, "POST response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		return json.Unmarshal(bodyBytes, v)
	}

	return nil
}

// PostForm performs a POST request with form-data encoding
func (c *Client) PostForm(ctx context.Context, path string, formData map[string]interface{}, v interface{}) error {
	resp, err := c.doFormRequest(ctx, "POST", path, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		// Read response body for logging
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read PostForm response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return readErr
		}

		// Log successful response body in debug mode
		tflog.Debug(ctx, "PostForm response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		return json.Unmarshal(bodyBytes, v)
	}

	return nil
}

// PutForm performs a PUT request with form-data encoding
func (c *Client) PutForm(ctx context.Context, path string, formData map[string]interface{}, v interface{}) error {
	resp, err := c.doFormRequest(ctx, "PUT", path, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		// Read response body for logging
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read PutForm response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return readErr
		}

		// Log successful response body in debug mode
		tflog.Debug(ctx, "PutForm response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		return json.Unmarshal(bodyBytes, v)
	}

	return nil
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}, v interface{}) error {
	resp, err := c.doRequest(ctx, "PUT", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		// Read response body for logging
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read PUT response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return readErr
		}

		// Log successful response body in debug mode
		tflog.Debug(ctx, "PUT response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		return json.Unmarshal(bodyBytes, v)
	}

	return nil
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body interface{}, v interface{}) error {
	resp, err := c.doRequest(ctx, "PATCH", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		// Read response body for logging
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read PATCH response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return readErr
		}

		// Log successful response body in debug mode
		tflog.Debug(ctx, "PATCH response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		return json.Unmarshal(bodyBytes, v)
	}

	return nil
}

// doURLEncodedFormRequest performs an HTTP request with application/x-www-form-urlencoded encoding
func (c *Client) doURLEncodedFormRequest(ctx context.Context, method, path string, formData map[string]interface{}) (*http.Response, error) {
	// Ensure path starts with /api
	if !strings.HasPrefix(path, "/api") {
		path = "/api" + path
	}

	// Parse the path to properly handle query strings
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	u := c.BaseURL.ResolveReference(rel)

	values := url.Values{}

	// Add form fields
	for key, value := range formData {
		if value == nil {
			continue
		}

		switch v := value.(type) {
		case string:
			if v != "" {
				values.Add(key, v)
			}
		case []string:
			for _, item := range v {
				if item != "" {
					values.Add(key, item)
				}
			}
		case int:
			values.Add(key, fmt.Sprintf("%d", v))
		case int64:
			values.Add(key, fmt.Sprintf("%d", v))
		case bool:
			values.Add(key, fmt.Sprintf("%t", v))
		case fmt.Stringer:
			values.Add(key, v.String())
		default:
			values.Add(key, fmt.Sprintf("%v", v))
		}
	}

	body := strings.NewReader(values.Encode())

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Ce-Auth", c.generateHMACAuth())
	req.Header.Set("ce-region", c.Region)

	// Add organization header if specified
	if c.Organization != "" {
		req.Header.Set("organisation-name", c.Organization)
	}

	// Add project name header if specified
	if c.ProjectName != "" {
		req.Header.Set("Project-Name", c.ProjectName)
	}

	// Add subnet-id header if specified (required by volume API for provider lookup)
	if c.SubnetID != "" {
		req.Header.Set("subnet-id", c.SubnetID)
	}

	// Add availability zone header if specified
	if c.AvailabilityZone != "" {
		req.Header.Set("ce-availability-zone", c.AvailabilityZone)
	}

	// Log URL-encoded form request details in debug mode
	tflog.Debug(ctx, "Making URL-encoded form API request", map[string]interface{}{
		"method":       method,
		"url":          u.String(),
		"headers":      req.Header,
		"form_fields":  formData,
		"content_type": "application/x-www-form-urlencoded",
	})

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		tflog.Error(ctx, "URL-encoded form HTTP request failed", map[string]interface{}{
			"error": err.Error(),
			"url":   u.String(),
		})
		return nil, err
	}

	// Log response details in debug mode
	tflog.Debug(ctx, "Received URL-encoded form API response", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     resp.Header,
	})

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		// Read response body for logging and error parsing
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read URL-encoded form error response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		// Log error response body in debug mode
		tflog.Debug(ctx, "URL-encoded form error response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		var apiErr APIError
		if err := json.Unmarshal(bodyBytes, &apiErr); err != nil {
			tflog.Error(ctx, "Failed to parse URL-encoded form error response", map[string]interface{}{
				"error": err.Error(),
				"body":  string(bodyBytes),
			})
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}

	return resp, nil
}

// PostURLEncodedForm performs a POST request with application/x-www-form-urlencoded encoding
func (c *Client) PostURLEncodedForm(ctx context.Context, path string, formData map[string]interface{}, v interface{}) error {
	resp, err := c.doURLEncodedFormRequest(ctx, "POST", path, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read PostURLEncodedForm response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return readErr
		}

		tflog.Debug(ctx, "PostURLEncodedForm response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		return json.Unmarshal(bodyBytes, v)
	}

	return nil
}

// PutURLEncodedForm performs a PUT request with application/x-www-form-urlencoded encoding
func (c *Client) PutURLEncodedForm(ctx context.Context, path string, formData map[string]interface{}, v interface{}) error {
	resp, err := c.doURLEncodedFormRequest(ctx, "PUT", path, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			tflog.Error(ctx, "Failed to read PutURLEncodedForm response body", map[string]interface{}{
				"error": readErr.Error(),
			})
			return readErr
		}

		tflog.Debug(ctx, "PutURLEncodedForm response body", map[string]interface{}{
			"body": string(bodyBytes),
		})

		return json.Unmarshal(bodyBytes, v)
	}

	return nil
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) error {
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DeleteURLEncodedForm performs a DELETE request with application/x-www-form-urlencoded encoding
func (c *Client) DeleteURLEncodedForm(ctx context.Context, path string, formData map[string]interface{}) error {
	resp, err := c.doURLEncodedFormRequest(ctx, "DELETE", path, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// PostWithQueryParams performs a POST request with parameters in the URL query string (no body)
func (c *Client) PostWithQueryParams(ctx context.Context, path string, params url.Values, v interface{}) error {
	resp, err := c.doQueryParamRequest(ctx, "POST", path, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
		}
		tflog.Debug(ctx, "PostWithQueryParams response body", map[string]interface{}{
			"body": string(bodyBytes),
		})
		return json.Unmarshal(bodyBytes, v)
	}
	return nil
}

// PatchWithQueryParams performs a PATCH request with parameters in the URL query string (no body)
func (c *Client) PatchWithQueryParams(ctx context.Context, path string, params url.Values, v interface{}) error {
	resp, err := c.doQueryParamRequest(ctx, "PATCH", path, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
		}
		tflog.Debug(ctx, "PatchWithQueryParams response body", map[string]interface{}{
			"body": string(bodyBytes),
		})
		return json.Unmarshal(bodyBytes, v)
	}
	return nil
}

// doQueryParamRequest performs an HTTP request with parameters encoded in the URL query string
func (c *Client) doQueryParamRequest(ctx context.Context, method, path string, params url.Values) (*http.Response, error) {
	if !strings.HasPrefix(path, "/api") {
		path = "/api" + path
	}

	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Merge any existing query params with the provided ones
	existingParams := rel.Query()
	for k, vals := range params {
		for _, v := range vals {
			existingParams.Add(k, v)
		}
	}
	rel.RawQuery = existingParams.Encode()

	u := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Ce-Auth", c.generateHMACAuth())
	req.Header.Set("ce-region", c.Region)

	if c.Organization != "" {
		req.Header.Set("organisation-name", c.Organization)
	}
	if c.ProjectName != "" {
		req.Header.Set("Project-Name", c.ProjectName)
	}
	if c.SubnetID != "" {
		req.Header.Set("subnet-id", c.SubnetID)
	}
	if c.AvailabilityZone != "" {
		req.Header.Set("ce-availability-zone", c.AvailabilityZone)
	}

	tflog.Debug(ctx, "Making query param API request", map[string]interface{}{
		"method": method,
		"url":    u.String(),
	})

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		tflog.Debug(ctx, "Query param error response body", map[string]interface{}{
			"body": string(bodyBytes),
		})
		var apiErr APIError
		if err := json.Unmarshal(bodyBytes, &apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}

	return resp, nil
}

// WaitForOperation waits for a long-running operation to complete
func (c *Client) WaitForOperation(ctx context.Context, operationID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var operation struct {
			Status string `json:"status"`
			Error  string `json:"error,omitempty"`
		}

		err := c.Get(ctx, fmt.Sprintf("/api/v1/operations/%s", operationID), &operation)
		if err != nil {
			return err
		}

		switch operation.Status {
		case "completed":
			return nil
		case "failed":
			return fmt.Errorf("operation failed: %s", operation.Error)
		case "running", "pending":
			time.Sleep(5 * time.Second)
			continue
		default:
			return fmt.Errorf("unknown operation status: %s", operation.Status)
		}
	}

	return fmt.Errorf("operation timed out after %v", timeout)
}
