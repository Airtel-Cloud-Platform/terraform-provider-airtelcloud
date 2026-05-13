# ACP Terraform Provider — Developer Guide

A comprehensive guide for adding new resources to the `terraform-provider-airtelcloud` provider.

**Framework**: Terraform Plugin Framework v1.16.1 (not SDKv2)
**Go version**: 1.24+
**Module**: `github.com/terraform-providers/terraform-provider-airtelcloud`

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Decision Matrix — Before You Start](#2-decision-matrix--before-you-start)
3. [Step-by-Step Implementation Guide](#3-step-by-step-implementation-guide)
   - [3.1 Define Models](#31-define-models)
   - [3.2 Implement Client Methods](#32-implement-client-methods)
   - [3.3 Implement the Resource](#33-implement-the-resource)
   - [3.4 Register in Provider](#34-register-in-provider)
   - [3.5 Add Example Configuration](#35-add-example-configuration)
4. [Testing Guide](#4-testing-guide)
   - [4.1 Unit Tests](#41-unit-tests)
   - [4.2 Integration Tests](#42-integration-tests)
   - [4.3 Acceptance Tests](#43-acceptance-tests)
5. [Conventions, Gotchas & Common Mistakes](#5-conventions-gotchas--common-mistakes)
6. [Complete Worked Example — Load Balancer](#6-complete-worked-example--load-balancer)
7. [Implementation Checklist](#7-implementation-checklist)
8. [Resource Pattern Reference Table](#8-resource-pattern-reference-table)

---

## 1. Architecture Overview

```
┌──────────────────┐
│  Terraform Core  │
└────────┬─────────┘
         │ gRPC (Plugin Framework)
┌────────▼─────────────────────────────────────────────┐
│  Provider Layer   (internal/provider/*_resource.go)   │
│  - Schema definitions (tfsdk tags)                    │
│  - CRUD lifecycle methods                             │
│  - Terraform ↔ Go type mapping                        │
└────────┬─────────────────────────────────────────────┘
         │ Go function calls
┌────────▼─────────────────────────────────────────────┐
│  Client Layer     (internal/client/*.go)              │
│  - HTTP requests (JSON / form-encoded)                │
│  - HMAC auth, path construction                       │
│  - Delete polling, async operations                   │
└────────┬─────────────────────────────────────────────┘
         │ HTTP (JSON / form-data)
┌────────▼─────────────────────────────────────────────┐
│  Airtel Cloud API                                     │
└──────────────────────────────────────────────────────┘

  Models (internal/models/*.go) — shared between Provider and Client layers
```

### Directory Structure

```
internal/
├── provider/
│   ├── provider.go                          # Provider definition & resource registration
│   ├── dns_zone_resource.go                 # One file per resource
│   ├── security_group_resource.go
│   ├── vm_resource.go
│   └── ...
├── client/
│   ├── client.go                            # Core HTTP client, auth, APIError
│   ├── dns_zone.go                          # Client methods per resource
│   ├── security_group.go
│   ├── vm.go                                # Also contains structToFormData()
│   ├── testutil/
│   │   └── mock_server.go                   # Shared mock server for unit tests
│   ├── security_group_test.go               # Unit tests
│   └── security_group_integration_test.go   # Integration tests
├── models/
│   ├── dns_zone.go                          # Request/response structs
│   ├── security_group.go
│   ├── vm.go
│   └── ...
tests/
└── resources/
    └── security_group_test.go               # Acceptance tests (Terraform lifecycle)
```

### File Naming Conventions

| Layer | Pattern | Example |
|-------|---------|---------|
| Models | `internal/models/{name}.go` | `internal/models/dns_zone.go` |
| Client | `internal/client/{name}.go` | `internal/client/dns_zone.go` |
| Provider | `internal/provider/{name}_resource.go` | `internal/provider/dns_zone_resource.go` |
| Unit tests | `internal/client/{name}_test.go` | `internal/client/security_group_test.go` |
| Integration tests | `internal/client/{name}_integration_test.go` | `internal/client/security_group_integration_test.go` |
| Acceptance tests | `tests/resources/{name}_test.go` | `tests/resources/security_group_test.go` |

---

## 2. Decision Matrix — Before You Start

Answer these questions before writing any code:

| Question | Options | Impact |
|----------|---------|--------|
| **API content-type** | JSON / form-encoded | Determines struct tags (`json` vs `form`) and which client method to call (`Post` vs `PostURLEncodedForm`) |
| **Resource ID type** | String UUID / Integer | Determines Terraform attribute type (`types.String` vs `types.Int64`) and import pattern |
| **Has update API?** | Yes / No | If no, add `RequiresReplace()` plan modifier to all user-settable attributes |
| **Create returns full resource?** | Yes / No | If no, perform a Read-after-Create to populate state |
| **Is child resource?** | Yes / No | Determines import format (single ID vs composite `parent_id/child_id`) |
| **Async delete?** | Yes / No | Determines if delete-with-polling is needed in client |

### Pattern Reference — Existing Resources

| Resource | Encoding | ID Type | Has Update | Create Returns Full | Child | Import Format |
|----------|----------|---------|------------|-------------------|-------|---------------|
| DNS Zone | JSON | String (UUID) | Yes | Yes* | No | `{uuid}` |
| DNS Record | JSON | String (UUID) | Yes | Yes* | Yes | `{zone_id}/{record_uuid}` |
| VPC | JSON | String (UUID) | Yes | Yes | No | `{id}` |
| Subnet | JSON | String | Yes | Yes | Yes | `{network_id}/{subnet_id}` |
| VM | Form | String (UUID) | Yes | Yes | No | `{id}` |
| Volume | Form | Integer | Yes | Yes | No | `{id}` (int) |
| Object Storage Bucket | JSON | String (name) | Yes | Yes* | No | `{name}` |
| Object Storage Access Key | JSON | String | No | Yes | Yes | `{access_key_id}` |
| File Storage | JSON | String (UUID) | Yes | Yes | No | `{uuid}` |
| File Storage Export | JSON | String | Yes | Yes | Yes | `{volume_id}/{path_id}` |
| Security Group | Form | Integer | No | Yes | No | `{id}` (int, parsed) |
| Security Group Rule | JSON† | Integer | No | Yes | Yes | `{sg_id}/{rule_id}` |
| VPC Peering | JSON | String (UUID) | No | Yes | No | `{id}` |
| LB Service | URL-encoded | String | No | Yes | No | `{id}` |
| LB VIP | JSON | String | No | Yes | Yes | `{lb_service_id}/{vip}` |
| LB Certificate | URL-encoded | String | No | Yes | Yes | `{lb_service_id}/{cert_name}` |
| LB Virtual Server | URL-encoded | String | No | Yes | Yes | `{lb_service_id}/{vs_name}` |
| Compute Snapshot | URL-encoded | String (UUID) | No | Yes (poll) | No | `{uuid}` |
| Protection | URL-encoded | String (int) | Yes | Yes | No | `{id}` |
| Protection Plan | URL-encoded | String (int) | No | Yes | No | `{id}` |
| Public IP | JSON | String (UUID) | No | Yes (poll) | No | `{uuid}` |
| Public IP Policy Rule | JSON | String (UUID) | No | Yes | Yes | composite‡ |

\* Falls back to List+filter if create response is empty.
† Rule creation uses JSON serialized into a form field (`security_group_data`).
‡ Public IP Policy Rule uses a composite lookup via `public_ip_id/target_vip/public_ip/rule_uuid`.

---

## 3. Step-by-Step Implementation Guide

### 3.1 Define Models

Create `internal/models/{name}.go` with request and response structs.

#### JSON API Pattern (most resources)

Response structs always use `json` tags. Request structs also use `json` tags when the API accepts JSON.

```go
package models

// {Name} represents the API response for a {name}
type {Name} struct {
    UUID        string  `json:"uuid"`
    Name        string  `json:"name"`
    Description *string `json:"description"`     // nullable → use pointer
    Status      string  `json:"status"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

// Create{Name}Request represents the request to create a {name}
type Create{Name}Request struct {
    Name        string  `json:"name"`
    Description *string `json:"description,omitempty"` // optional → omitempty
}

// Update{Name}Request represents the request to update a {name}
type Update{Name}Request struct {
    Description *string `json:"description,omitempty"`
}
```

**Reference**: `internal/models/dns_zone.go` — simplest JSON model.

#### Form-Encoded API Pattern (compute, volume, security group create)

Request structs use `form` tags. Response structs still use `json` tags.

```go
package models

// {Name} represents the API response (always JSON)
type {Name} struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Status string `json:"status"`
}

// Create{Name}Request — form-encoded
type Create{Name}Request struct {
    Name       string `form:"name"`
    Size       int    `form:"size,omitempty"`
    EnableFlag bool   `form:"enable_flag"`
}
```

**Reference**: `internal/models/volume.go` — form-encoded create with `json` response.

#### Key Gotchas

- **`json.RawMessage`** — Use when the API returns inconsistent types for the same field (e.g., `volume_type_id` can be a string or integer). See `internal/models/volume.go`.
- **`*string`** — Use for nullable optional fields (e.g., `Description *string`). In the provider layer, map `nil` to `types.StringNull()`.
- **Tags are per-direction**: Responses always decode with `json` tags. Requests use `json` for JSON APIs and `form` for form-encoded APIs.

---

### 3.2 Implement Client Methods

Create `internal/client/{name}.go`.

#### Base Path Patterns

Different API services use different URL structures:

| Service | Path Pattern | Example |
|---------|-------------|---------|
| Compute (v2.1) | `/api/v2.1/computes/domain/{org}/project/{project}/computes` | VM, flavors, images, snapshots |
| Volume (v2.1) | `/api/v2.1/volumes/domain/{org}/project/{project}/volumes` | Block storage |
| Backup (v2.1) | `/api/v2.1/backups/domain/{org}/project/{project}/backups` | Protection, protection plans |
| Load Balancer (v2.1) | `/api/v2.1/load-balancers/domain/{org}/project/{project}/load-balancers` | LB Service, VIP, Cert, VS |
| Network Manager | `/api/network-manager/v1/domain/{org}/project/{project}/...` | VPC, Subnet, VPC Peering |
| Network (v1) | `/api/v1/networks/securitygroup` | Security groups |
| DNS (v1) | `/api/v1/zones` | DNS zones, records |
| Storage Plugin | `/api/storage-plugin/v1/domain/{org}/project/{project}/...` | Object/File storage |
| IPAM | `/ipam/domain/{org}/project/{project}` | Public IP allocation |
| IPAM Admin | `/api/v1/admin/ipam_vip` | Public IP policy rules |

> **IMPORTANT — Path Prefix Rule**: The `doRequest()` and `doFormRequest()` methods in `client.go` automatically prepend `/api` if the path doesn't already start with `/api`. This means:
> - Path `/v1/zones` → becomes `/api/v1/zones` ✓
> - Path `/api/v2.1/computes/...` → stays as-is ✓
> - Do NOT double-prefix: `/api/api/v1/zones` ✗

#### Client Method Templates

**JSON API — CRUD methods:**

```go
package client

import (
    "context"
    "fmt"
    "time"

    "github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// basePath returns the base URL path for {name} endpoints
func (c *Client) {name}BasePath() string {
    return fmt.Sprintf("/api/v1/service/domain/%s/project/%s/{names}",
        c.Organization, c.ProjectName)
}

// Get{Name} retrieves a {name} by ID
func (c *Client) Get{Name}(ctx context.Context, id string) (*models.{Name}, error) {
    var result models.{Name}
    err := c.Get(ctx, fmt.Sprintf("%s/%s", c.{name}BasePath(), id), &result)
    if err != nil {
        return nil, err
    }
    return &result, nil
}

// Create{Name} creates a new {name}
func (c *Client) Create{Name}(ctx context.Context, req *models.Create{Name}Request) (*models.{Name}, error) {
    var result models.{Name}
    err := c.Post(ctx, c.{name}BasePath(), req, &result)
    if err != nil {
        return nil, err
    }

    // Fallback: if API returns empty body, fetch by listing
    if result.UUID == "" {
        return c.Get{Name}(ctx, req.Name) // or list and filter
    }

    return &result, nil
}

// Update{Name} updates an existing {name}
func (c *Client) Update{Name}(ctx context.Context, id string, req *models.Update{Name}Request) (*models.{Name}, error) {
    var result models.{Name}
    err := c.Put(ctx, fmt.Sprintf("%s/%s", c.{name}BasePath(), id), req, &result)
    if err != nil {
        return nil, err
    }

    if result.UUID == "" {
        return c.Get{Name}(ctx, id)
    }
    return &result, nil
}

// Delete{Name} deletes a {name} and polls until confirmed
func (c *Client) Delete{Name}(ctx context.Context, id string) error {
    err := c.Delete(ctx, fmt.Sprintf("%s/%s", c.{name}BasePath(), id))
    if err != nil {
        return err
    }

    // Poll until 404 confirms deletion
    for i := 0; i < 60; i++ {
        _, err := c.Get{Name}(ctx, id)
        if err != nil {
            if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
                return nil
            }
            return err
        }
        time.Sleep(5 * time.Second)
    }

    return fmt.Errorf("{name} deletion timed out")
}
```

**Form-Encoded API — Create method:**

```go
// Create{Name} creates a new {name} using form-encoded data
func (c *Client) Create{Name}(ctx context.Context, req *models.Create{Name}Request) (*models.{Name}, error) {
    // structToFormData is defined in internal/client/vm.go
    formData := structToFormData(req)

    var result models.{Name}
    err := c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/", c.{name}BasePath()), formData, &result)
    if err != nil {
        return nil, err
    }

    return &result, nil
}
```

#### Available Client Methods

| Method | Encoding | Usage |
|--------|----------|-------|
| `c.Get(ctx, path, &result)` | — | GET request, decode JSON response |
| `c.Post(ctx, path, body, &result)` | JSON | POST with JSON body |
| `c.Put(ctx, path, body, &result)` | JSON | PUT with JSON body |
| `c.Patch(ctx, path, body, &result)` | JSON | PATCH with JSON body |
| `c.Delete(ctx, path)` | — | DELETE request |
| `c.DeleteURLEncodedForm(ctx, path, formData)` | x-www-form-urlencoded | DELETE with URL-encoded body |
| `c.PostForm(ctx, path, formData, &result)` | multipart/form-data | POST with multipart body |
| `c.PutForm(ctx, path, formData, &result)` | multipart/form-data | PUT with multipart body |
| `c.PostURLEncodedForm(ctx, path, formData, &result)` | x-www-form-urlencoded | POST with URL-encoded body |
| `c.PutURLEncodedForm(ctx, path, formData, &result)` | x-www-form-urlencoded | PUT with URL-encoded body |
| `c.PostWithQueryParams(ctx, path, params, &result)` | query string | POST with params in URL (no body) |
| `c.PatchWithQueryParams(ctx, path, params, &result)` | query string | PATCH with params in URL (no body) |

#### `structToFormData()` Helper

Located at `internal/client/vm.go:143`. Converts a struct with `form` tags to `map[string]interface{}` for use with `PostURLEncodedForm` / `PutURLEncodedForm`.

```go
formData := structToFormData(req) // reads `form` struct tags
```

Supports: `string`, `int`, `int64`, `bool`, `[]string`. Respects `omitempty`.

#### `WithAvailabilityZone()` and `WithSubnetID()` Scoping

Some APIs require an availability zone or subnet ID header. Instead of modifying the shared client, create a scoped copy:

```go
func (c *Client) Create{Name}(ctx context.Context, req *models.Create{Name}Request, az string) (*models.{Name}, error) {
    scopedClient := c.WithAvailabilityZone(az)  // returns a shallow copy with header set
    var result models.{Name}
    err := scopedClient.Post(ctx, scopedClient.basePath(), req, &result)
    // ...
}
```

**Reference**: `internal/client/public_ip.go:17` (availability zone), `internal/client/volume.go` (subnet ID).

#### Query-Param Requests

For APIs where parameters go in the URL query string (no request body):

```go
import "net/url"

params := url.Values{}
params.Set("compute_id", computeID)
params.Set("name", name)

err := c.PostWithQueryParams(ctx, path, params, &result)
```

**Reference**: `internal/client/compute.go` (snapshot and protection operations).

#### `WaitFor{Name}Ready()` — Async Create Polling

When the API returns immediately but the resource takes time to reach a stable state:

```go
func (c *Client) WaitFor{Name}Ready(ctx context.Context, id string, timeout time.Duration) (*models.{Name}, error) {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        result, err := c.Get{Name}(ctx, id)
        if err != nil {
            return nil, err
        }
        switch result.Status {
        case "active", "available", "Created":
            return result, nil
        case "error":
            return nil, fmt.Errorf("{name} entered error state")
        }
        time.Sleep(10 * time.Second)
    }
    return nil, fmt.Errorf("{name} did not become ready within %v", timeout)
}
```

**Reference**: `internal/client/snapshot.go:64` (`WaitForSnapshotReady`), `internal/client/public_ip.go:118` (`WaitForPublicIPReady`).

#### Delete Polling Pattern

All delete methods follow the same pattern: issue DELETE, then poll GET until 404.

```go
// Standard: 60 attempts, 5-second intervals (5 minutes max)
for i := 0; i < 60; i++ {
    _, err := c.Get{Name}(ctx, id)
    if err != nil {
        if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
            return nil // confirmed deleted
        }
        return err // unexpected error
    }
    time.Sleep(5 * time.Second)
}
return fmt.Errorf("{name} deletion timed out")
```

**Exception**: Compute uses 10-second intervals (`internal/client/vm.go:95`).

#### Trailing Slashes

Many API endpoints require a trailing slash. Check the API docs, but when in doubt, include one:
```go
fmt.Sprintf("%s/%d/", c.basePath(), id)  // note trailing /
```

---

### 3.3 Implement the Resource

Create `internal/provider/{name}_resource.go`.

#### Compile-Time Interface Checks

Always start with these to catch missing methods at compile time:

```go
var _ resource.Resource = &{Name}Resource{}
var _ resource.ResourceWithImportState = &{Name}Resource{}
// Optional:
var _ resource.ResourceWithValidateConfig = &{Name}Resource{}
```

#### Resource Struct and Constructor

```go
func New{Name}Resource() resource.Resource {
    return &{Name}Resource{}
}

type {Name}Resource struct {
    client *client.Client
}
```

#### Terraform Model Struct

Maps between Terraform state and Go types using `tfsdk` tags:

```go
type {Name}ResourceModel struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    Description types.String `tfsdk:"description"`
    Status      types.String `tfsdk:"status"`
    CreatedAt   types.String `tfsdk:"created_at"`
}
```

**Type Mapping:**

| Go / API Type | Terraform Type | Notes |
|---------------|---------------|-------|
| `string` | `types.String` | Most common |
| `int`, `int64` | `types.Int64` | Used for integer IDs (volume, security group) |
| `float64` | `types.Float64` | Used for timestamps (DNS zone) |
| `bool` | `types.Bool` | |
| `*string` (nullable) | `types.String` | Map `nil` → `types.StringNull()` |
| `map[string]string` | `types.Map` | Use `types.MapValueFrom()` / `ElementsAs()` |

#### Metadata Method

```go
func (r *{Name}Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_{resource_name}"
}
```

The `ProviderTypeName` is `"airtelcloud"`, so the full resource name becomes `airtelcloud_{resource_name}`.

#### Schema Method

```go
func (r *{Name}Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages an Airtel Cloud {name}.",

        Attributes: map[string]schema.Attribute{
            // Computed ID — always include UseStateForUnknown
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "The unique identifier.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },

            // Required attribute
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "The name.",
                // Add RequiresReplace if changing requires destroy+create
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },

            // Optional attribute with default
            "zone_type": schema.StringAttribute{
                Optional: true,
                Computed: true,  // required when using Default
                Default:  stringdefault.StaticString("forward"),
            },

            // Optional attribute with validator
            "replication_type": schema.StringAttribute{
                Required: true,
                Validators: []validator.String{
                    stringvalidator.OneOf("Local", "Replicated"),
                },
            },

            // Computed-only attribute (read from API)
            "status": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Current status.",
            },
        },
    }
}
```

**Schema Attribute Patterns:**

| Pattern | Required | Optional | Computed | Default | PlanModifier |
|---------|----------|----------|----------|---------|-------------|
| User must provide | ✓ | | | | |
| User can provide, has default | | ✓ | ✓ | ✓ | |
| User can provide, no default | | ✓ | | | |
| API-only, stable after create | | | ✓ | | `UseStateForUnknown()` |
| API-only, can change | | | ✓ | | |
| Immutable after create | ✓ or ✓ | | | | `RequiresReplace()` |

**Common PlanModifiers:**

```go
// String
stringplanmodifier.UseStateForUnknown()  // Don't show diff for computed fields
stringplanmodifier.RequiresReplace()     // Force destroy+create on change

// Int64
int64planmodifier.UseStateForUnknown()
int64planmodifier.RequiresReplace()

// Bool
boolplanmodifier.UseStateForUnknown()
```

**Common Validators** (from `terraform-plugin-framework-validators`):

```go
stringvalidator.OneOf("value1", "value2")
int64validator.Between(1, 10)
```

**Common Defaults:**

```go
stringdefault.StaticString("forward")
booldefault.StaticBool(false)
int64default.StaticInt64(1)
```

#### Timeouts Block (Optional)

For resources with async operations, add configurable timeouts using `terraform-plugin-framework-timeouts`:

```go
import "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

// In the model struct:
type {Name}ResourceModel struct {
    // ... other fields ...
    Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// In the Schema method — add as a Block, not Attribute:
Blocks: map[string]schema.Block{
    "timeouts": timeouts.Block(ctx, timeouts.Opts{
        Create: true,
        Delete: true,
    }),
},

// In the Create method — read timeout with default:
createTimeout, diags := data.Timeouts.Create(ctx, 10*time.Minute)
resp.Diagnostics.Append(diags...)
if resp.Diagnostics.HasError() {
    return
}
// Pass createTimeout to WaitFor{Name}Ready()
```

**Reference**: `internal/provider/compute_snapshot_resource.go:87` (timeouts block), `internal/provider/public_ip_resource.go:111`.

#### Configure Method (Boilerplate)

Identical for every resource — just copy:

```go
func (r *{Name}Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    c, ok := req.ProviderData.(*client.Client)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    r.client = c
}
```

#### Create Method

```go
func (r *{Name}Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data {Name}ResourceModel

    // 1. Read the plan
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Build the API request
    createReq := &models.Create{Name}Request{
        Name: data.Name.ValueString(),
    }

    // Handle optional pointer fields
    if !data.Description.IsNull() && !data.Description.IsUnknown() {
        desc := data.Description.ValueString()
        createReq.Description = &desc
    }

    // 3. Call the client
    result, err := r.client.Create{Name}(ctx, createReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error",
            fmt.Sprintf("Unable to create {name}, got error: %s", err))
        return
    }

    // 4. Map response to Terraform model
    data.ID = types.StringValue(result.UUID)
    data.Status = types.StringValue(result.Status)
    // ... map all computed fields

    // Handle nullable fields
    if result.Description != nil {
        data.Description = types.StringValue(*result.Description)
    } else {
        data.Description = types.StringNull()
    }

    tflog.Trace(ctx, "created a {name} resource")

    // 5. Save state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

#### Read Method

```go
func (r *{Name}Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data {Name}ResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    result, err := r.client.Get{Name}(ctx, data.ID.ValueString())
    if err != nil {
        // CRITICAL: Handle 404 by removing from state
        if client.IsNotFoundError(err) {
            resp.State.RemoveResource(ctx)
            return
        }
        resp.Diagnostics.AddError("Client Error",
            fmt.Sprintf("Unable to read {name}, got error: %s", err))
        return
    }

    // Map all fields from API response to Terraform model
    data.ID = types.StringValue(result.UUID)
    data.Name = types.StringValue(result.Name)
    data.Status = types.StringValue(result.Status)
    // ... map all fields

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

#### Update Method — Normal

```go
func (r *{Name}Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data {Name}ResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    updateReq := &models.Update{Name}Request{
        // Only include updatable fields
    }

    result, err := r.client.Update{Name}(ctx, data.ID.ValueString(), updateReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error",
            fmt.Sprintf("Unable to update {name}, got error: %s", err))
        return
    }

    // Map updated fields
    data.Status = types.StringValue(result.Status)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

#### Update Method — No Update API

When the API doesn't support updates, add `RequiresReplace()` to all user-settable attributes in the schema. The Update method should return an error as a safety net:

```go
func (r *{Name}Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    resp.Diagnostics.AddError("Update Not Supported",
        "{Name} cannot be updated. All changes require replacement.")
}
```

**Reference**: `internal/provider/security_group_resource.go:176`

#### Delete Method

```go
func (r *{Name}Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data {Name}ResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.client.Delete{Name}(ctx, data.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Client Error",
            fmt.Sprintf("Unable to delete {name}, got error: %s", err))
        return
    }
}
```

Note: Delete polling lives in the **client layer**, not the provider layer.

#### ImportState Patterns

**Pattern 1: String ID passthrough** (DNS Zone, VPC, VM, Object Storage)

```go
func (r *{Name}Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

Usage: `terraform import airtelcloud_{name}.example <uuid>`

**Pattern 2: Integer ID** (Security Group, Volume)

`ImportStatePassthroughID` sets the value as a string, which fails for `types.Int64` attributes. You must manually parse:

```go
func (r *{Name}Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    id, err := strconv.ParseInt(req.ID, 10, 64)
    if err != nil {
        resp.Diagnostics.AddError(
            "Invalid Import ID",
            fmt.Sprintf("Expected a numeric ID, got: %q", req.ID),
        )
        return
    }

    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
```

Usage: `terraform import airtelcloud_{name}.example 123`

**Reference**: `internal/provider/security_group_resource.go:196`

**Pattern 3: Composite ID for child resources** (Security Group Rule, DNS Record, Subnet)

```go
func (r *{Name}Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    idParts := strings.Split(req.ID, "/")
    if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
        resp.Diagnostics.AddError(
            "Unexpected Import Identifier",
            fmt.Sprintf("Expected import identifier with format: parent_id/child_id. Got: %q", req.ID),
        )
        return
    }

    parentID, err := strconv.ParseInt(idParts[0], 10, 64)
    if err != nil {
        resp.Diagnostics.AddError("Invalid Import ID",
            fmt.Sprintf("Expected numeric parent_id, got: %q", idParts[0]))
        return
    }

    childID, err := strconv.ParseInt(idParts[1], 10, 64)
    if err != nil {
        resp.Diagnostics.AddError("Invalid Import ID",
            fmt.Sprintf("Expected numeric child_id, got: %q", idParts[1]))
        return
    }

    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_id"), parentID)...)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), childID)...)
}
```

Usage: `terraform import airtelcloud_{name}.example 42/10`

**Reference**: `internal/provider/security_group_rule_resource.go:280`

#### ValidateConfig (Optional)

Use for cross-attribute validation (e.g., mutual exclusion):

```go
var _ resource.ResourceWithValidateConfig = &{Name}Resource{}

func (r *{Name}Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
    var data {Name}ResourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Skip if values are unknown (during plan with variables)
    if data.FieldA.IsUnknown() || data.FieldB.IsUnknown() {
        return
    }

    // Example: mutual exclusion
    if !data.FieldA.IsNull() && !data.FieldB.IsNull() {
        resp.Diagnostics.AddAttributeError(
            path.Root("field_a"),
            "Conflicting Attributes",
            "field_a and field_b are mutually exclusive.",
        )
    }
}
```

**Reference**: `internal/provider/vm_resource.go:261` (flavor_id vs flavor_name), `internal/provider/object_storage_bucket_resource.go` (replication_type vs replication_tag).

---

### 3.4 Register in Provider

Add a single line to the `Resources()` function in `internal/provider/provider.go:176`:

```go
func (p *AirtelCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewVMResource,
        NewVolumeResource,
        // ... existing resources ...
        New{Name}Resource,  // ← Add this line
    }
}
```

---

### 3.5 Add Example Configuration

Create `examples/resources/{name}/main.tf`:

```hcl
resource "airtelcloud_{name}" "example" {
  name        = "my-resource"
  description = "Created by Terraform"
}
```

---

## 4. Testing Guide

### 4.1 Unit Tests

**File**: `internal/client/{name}_test.go`

Unit tests use the mock server from `internal/client/testutil/mock_server.go` to test client methods without hitting a real API.

#### Setup

```go
package client

import (
    "context"
    "strings"
    "testing"

    "github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
    "github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestCreate{Name}(t *testing.T) {
    mockServer := testutil.NewMockServer()
    defer mockServer.Close()

    baseURL := strings.TrimSuffix(mockServer.URL, "/")
    client, _ := NewClient(baseURL, "test-api-key", "test-api-secret",
        "south-1", "test-org", "test-project", "", "")

    // ... test cases
}
```

#### Adding Mock Handlers

Add handlers to `internal/client/testutil/mock_server.go` for your new resource. The handler key format is `"METHOD /full/path"`:

```go
// In setupDefaultHandlers():
ms.Handlers["POST /api/v1/service/resource/"] = ms.create{Name}Handler
ms.Handlers["GET /api/v1/service/resource/1/"] = ms.get{Name}Handler
ms.Handlers["DELETE /api/v1/service/resource/1/"] = ms.delete{Name}Handler

// Handler implementation:
func (ms *MockServer) create{Name}Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)

    result := models.{Name}{
        UUID:   "test-uuid-1234",
        Name:   "test-resource",
        Status: "ACTIVE",
    }
    json.NewEncoder(w).Encode(result)
}

// Delete handler — flip GET to return 404 after delete:
func (ms *MockServer) delete{Name}Handler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNoContent)
    ms.Handlers["GET /api/v1/service/resource/1/"] = func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
    }
}
```

#### Table-Driven Test Pattern

```go
func TestGet{Name}(t *testing.T) {
    mockServer := testutil.NewMockServer()
    defer mockServer.Close()

    baseURL := strings.TrimSuffix(mockServer.URL, "/")
    client, _ := NewClient(baseURL, "test-api-key", "test-api-secret",
        "south-1", "test-org", "test-project", "", "")

    tests := []struct {
        name     string
        id       int
        wantErr  bool
        wantName string
    }{
        {
            name:     "successful retrieval",
            id:       1,
            wantErr:  false,
            wantName: "test-resource",
        },
        {
            name:    "not found",
            id:      999,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.id == 999 {
                mockServer.SetErrorResponse("GET",
                    "/api/v1/service/resource/999/", 404, "Not found")
            }

            result, err := client.Get{Name}(context.Background(), tt.id)

            if (err != nil) != tt.wantErr {
                t.Errorf("Get{Name}() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr {
                if result == nil {
                    t.Error("Get{Name}() returned nil")
                    return
                }
                if result.Name != tt.wantName {
                    t.Errorf("Get{Name}() name = %v, want %v", result.Name, tt.wantName)
                }
            }
        })
    }
}
```

#### Custom Setup Per Test Case

For tests that need different mock behavior:

```go
tests := []struct {
    name      string
    setup     func(ms *testutil.MockServer)
    wantCount int
    wantErr   bool
}{
    {
        name:      "successful list",
        wantCount: 2,
    },
    {
        name: "empty list",
        setup: func(ms *testutil.MockServer) {
            ms.AddHandler("GET", "/api/v1/service/resource/",
                func(w http.ResponseWriter, r *http.Request) {
                    w.Header().Set("Content-Type", "application/json")
                    json.NewEncoder(w).Encode([]models.{Name}{})
                })
        },
        wantCount: 0,
    },
    {
        name: "server error",
        setup: func(ms *testutil.MockServer) {
            ms.SetErrorResponse("GET", "/api/v1/service/resource/",
                500, "Internal server error")
        },
        wantErr: true,
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        mockServer := testutil.NewMockServer()
        defer mockServer.Close()

        if tt.setup != nil {
            tt.setup(mockServer)
        }

        // ... create client and test
    })
}
```

**Run unit tests:**
```bash
go test ./internal/client/ -v -run TestCreate{Name}
```

**Reference**: `internal/client/security_group_test.go`

---

### 4.2 Integration Tests

**File**: `internal/client/{name}_integration_test.go`

Integration tests run against the real API. They are gated by a build tag.

```go
//go:build integration

package client

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func Test{Name}Integration_CreateGetDelete(t *testing.T) {
    config := getVPCTestConfig(t)         // loads from env vars / .env file
    client := createVPCTestClient(t, config)
    ctx := context.Background()

    // Use timestamped names for isolation
    name := fmt.Sprintf("test-{name}-%d", time.Now().Unix())

    // Create
    createReq := &models.Create{Name}Request{
        Name: name,
    }

    t.Logf("Creating {name}: %s", name)
    result, err := client.Create{Name}(ctx, createReq)
    if err != nil {
        t.Fatalf("Create{Name} failed: %v", err)
    }

    t.Logf("{Name} created with ID: %s", result.UUID)

    // Deferred cleanup — always runs, even on test failure
    defer func() {
        t.Logf("Deleting {name}: %s", result.UUID)
        err := client.Delete{Name}(ctx, result.UUID)
        if err != nil {
            t.Errorf("Delete{Name} failed: %v", err)
        }
    }()

    // Read
    fetched, err := client.Get{Name}(ctx, result.UUID)
    if err != nil {
        t.Fatalf("Get{Name} failed: %v", err)
    }

    if fetched.Name != name {
        t.Errorf("Expected name %s, got %s", name, fetched.Name)
    }
}
```

**Run integration tests:**
```bash
go test ./internal/client/ -v -tags=integration -run Test{Name}Integration
```

---

### 4.3 Acceptance Tests

**File**: `tests/resources/{name}_test.go`

Acceptance tests exercise the full Terraform lifecycle (plan → apply → import → destroy).

```go
package tests

import (
    "fmt"
    "testing"

    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc{Name}Resource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create and Read
            {
                Config: testAcc{Name}ResourceConfig("test-resource"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr(
                        "airtelcloud_{name}.test", "name", "test-resource"),
                    resource.TestCheckResourceAttrSet(
                        "airtelcloud_{name}.test", "id"),
                    resource.TestCheckResourceAttrSet(
                        "airtelcloud_{name}.test", "status"),
                ),
            },
            // Step 2: Import
            {
                ResourceName:      "airtelcloud_{name}.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Step 3 (optional): Update
            {
                Config: testAcc{Name}ResourceConfigUpdated("test-resource-updated"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr(
                        "airtelcloud_{name}.test", "name", "test-resource-updated"),
                ),
            },
            // Delete is automatic at end of TestCase
        },
    })
}

// Config helper — returns raw HCL
func testAcc{Name}ResourceConfig(name string) string {
    return fmt.Sprintf(`
resource "airtelcloud_{name}" "test" {
  name = %[1]q
}
`, name)
}
```

**Run acceptance tests:**
```bash
TF_ACC=1 go test ./tests/resources/ -v -run TestAcc{Name}Resource -timeout 30m
```

**Reference**: `tests/resources/security_group_test.go`

---

## 5. Conventions, Gotchas & Common Mistakes

### 1. Path Prefix Auto-Prepend

`doRequest()` in `client.go:98` prepends `/api` if the path doesn't start with `/api`:

```go
if !strings.HasPrefix(path, "/api") {
    path = "/api" + path
}
```

- DNS zone uses `/v1/zones` → auto-becomes `/api/v1/zones` ✓
- Compute uses `/api/v2.1/computes/...` → stays as-is ✓
- **Never** pass `/api/api/...` — it won't be caught ✗

### 2. Trailing Slashes

Many endpoints require trailing slashes. Missing one typically returns 301 redirect or 404. Compare:
- `GET /api/v1/networks/securitygroup/1/` ✓
- `GET /api/v1/networks/securitygroup/1` ✗ (may fail)

### 3. `form` vs `json` Tags

- **Response** structs always use `json` tags (API always returns JSON)
- **Request** structs use `json` tags for JSON APIs, `form` tags for form-encoded APIs
- Some resources mix both: Security Group Rule creation marshals JSON into a form field
- The `structToFormData()` helper in `vm.go:143` reads `form` tags — it ignores `json` tags

### 4. `structToFormData()` Location

This helper is in `internal/client/vm.go:143`, not in a separate utils file. It's package-level (lowercase `s`) so it's accessible from any file in the `client` package.

### 5. Null vs Empty String

```go
// API field is nil/missing → use types.StringNull()
if result.Description != nil {
    data.Description = types.StringValue(*result.Description)
} else {
    data.Description = types.StringNull()
}
```

If you set `types.StringValue("")` instead of `types.StringNull()`, Terraform will show a diff on every plan.

### 6. 404 Handling in Read

Always check for 404 in the Read method and remove from state. Use the `client.IsNotFoundError()` helper (defined in `internal/client/subnet.go:127`):

```go
if client.IsNotFoundError(err) {
    resp.State.RemoveResource(ctx)
    return
}
```

This helper checks for `*APIError` with status 404 and also handles fallback string matching for `"HTTP 404"`. This tells Terraform the resource was deleted out-of-band. Without this, a deleted resource causes permanent errors.

### 7. `UseStateForUnknown()` on ID Fields

Always add `UseStateForUnknown()` to computed fields that don't change after creation (like `id`, `uuid`). Without it, Terraform shows `(known after apply)` on every plan.

### 8. Integer ID Import

You cannot use `resource.ImportStatePassthroughID()` for integer IDs because it sets the value as a `types.String`, which fails type checking against `types.Int64`. You must manually parse with `strconv.ParseInt`.

### 9. Delete Polling Lives in Client

The polling loop (GET until 404) belongs in the client layer, not the provider layer. The provider's Delete method is a simple call:

```go
// Provider (simple)
err := r.client.Delete{Name}(ctx, id)

// Client (handles polling)
func (c *Client) Delete{Name}(ctx context.Context, id string) error {
    c.Delete(ctx, path)
    // ... poll until 404 ...
}
```

### 10. API Version Differences

Different services use different API versions:
- Compute, Volume: `/api/v2.1/...`
- Network Manager (VPC, Subnet): `/api/network-manager/v1/...`
- Security Groups: `/api/v1/networks/...`
- DNS: `/api/v1/zones/...`
- Storage Plugin (Object/File): `/api/storage-plugin/v1/...`

Check the API docs for the correct version prefix.

### 11. Organization and Project in URLs

Most API paths include `domain/{org}/project/{project}`. These come from `c.Organization` and `c.ProjectName` on the client. DNS and Security Group APIs are exceptions — they use headers instead of URL segments.

### 12. `json.RawMessage` for Inconsistent API Types

When an API field can be either a string or an integer (e.g., `volume_type_id`), use `json.RawMessage` to defer parsing:

```go
VolumeTypeID json.RawMessage `json:"volume_type_id"`

// Later, to read the value:
strings.Trim(string(volume.VolumeTypeID), `"`)
```

---

## 6. Complete Worked Example — Load Balancer

This section walks through adding a hypothetical `airtelcloud_load_balancer` resource end-to-end. The Load Balancer:
- Uses JSON encoding
- Has a string UUID ID
- Supports Create, Read, Update, Delete
- Is not a child resource
- Delete requires polling

### 6.1 Model — `internal/models/load_balancer.go`

```go
package models

// LoadBalancer represents an Airtel Cloud Load Balancer (API response)
type LoadBalancer struct {
	UUID             string `json:"uuid"`
	Name             string `json:"name"`
	Algorithm        string `json:"algorithm"`
	Port             int    `json:"port"`
	HealthCheckPath  string `json:"health_check_path"`
	Status           string `json:"status"`
	VIP              string `json:"vip"`
	VPCID            string `json:"vpc_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// CreateLoadBalancerRequest represents the request to create a load balancer
type CreateLoadBalancerRequest struct {
	Name            string `json:"name"`
	Algorithm       string `json:"algorithm"`
	Port            int    `json:"port"`
	HealthCheckPath string `json:"health_check_path,omitempty"`
	VPCID           string `json:"vpc_id"`
}

// UpdateLoadBalancerRequest represents the request to update a load balancer
type UpdateLoadBalancerRequest struct {
	Algorithm       string `json:"algorithm,omitempty"`
	HealthCheckPath string `json:"health_check_path,omitempty"`
}
```

### 6.2 Client — `internal/client/load_balancer.go`

```go
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func (c *Client) loadBalancerBasePath() string {
	return fmt.Sprintf("/api/v1/lbaas/domain/%s/project/%s/loadbalancers",
		c.Organization, c.ProjectName)
}

func (c *Client) GetLoadBalancer(ctx context.Context, id string) (*models.LoadBalancer, error) {
	var lb models.LoadBalancer
	err := c.Get(ctx, fmt.Sprintf("%s/%s/", c.loadBalancerBasePath(), id), &lb)
	if err != nil {
		return nil, err
	}
	return &lb, nil
}

func (c *Client) CreateLoadBalancer(ctx context.Context, req *models.CreateLoadBalancerRequest) (*models.LoadBalancer, error) {
	var lb models.LoadBalancer
	err := c.Post(ctx, fmt.Sprintf("%s/", c.loadBalancerBasePath()), req, &lb)
	if err != nil {
		return nil, err
	}

	if lb.UUID == "" {
		return nil, fmt.Errorf("create returned empty response")
	}

	return &lb, nil
}

func (c *Client) UpdateLoadBalancer(ctx context.Context, id string, req *models.UpdateLoadBalancerRequest) (*models.LoadBalancer, error) {
	var lb models.LoadBalancer
	err := c.Put(ctx, fmt.Sprintf("%s/%s/", c.loadBalancerBasePath(), id), req, &lb)
	if err != nil {
		return nil, err
	}

	if lb.UUID == "" {
		return c.GetLoadBalancer(ctx, id)
	}
	return &lb, nil
}

func (c *Client) DeleteLoadBalancer(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s/", c.loadBalancerBasePath(), id))
	if err != nil {
		return err
	}

	for i := 0; i < 60; i++ {
		_, err := c.GetLoadBalancer(ctx, id)
		if err != nil {
			if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
				return nil
			}
			return err
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("load balancer deletion timed out")
}
```

### 6.3 Resource — `internal/provider/load_balancer_resource.go`

```go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &LoadBalancerResource{}
var _ resource.ResourceWithImportState = &LoadBalancerResource{}

func NewLoadBalancerResource() resource.Resource {
	return &LoadBalancerResource{}
}

type LoadBalancerResource struct {
	client *client.Client
}

type LoadBalancerResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Algorithm       types.String `tfsdk:"algorithm"`
	Port            types.Int64  `tfsdk:"port"`
	HealthCheckPath types.String `tfsdk:"health_check_path"`
	VPCID           types.String `tfsdk:"vpc_id"`
	Status          types.String `tfsdk:"status"`
	VIP             types.String `tfsdk:"vip"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *LoadBalancerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (r *LoadBalancerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Load Balancer.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Load balancing algorithm. Valid values: `round_robin`, `least_connections`, `source_ip`.",
				Default:             stringdefault.StaticString("round_robin"),
				Validators: []validator.String{
					stringvalidator.OneOf("round_robin", "least_connections", "source_ip"),
				},
			},
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The port the load balancer listens on.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"health_check_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "HTTP path for health checks.",
				Default:             stringdefault.StaticString("/"),
			},
			"vpc_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The VPC ID to place the load balancer in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current status of the load balancer.",
			},
			"vip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Virtual IP address assigned to the load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp.",
			},
		},
	}
}

func (r *LoadBalancerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *LoadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &models.CreateLoadBalancerRequest{
		Name:            data.Name.ValueString(),
		Algorithm:       data.Algorithm.ValueString(),
		Port:            int(data.Port.ValueInt64()),
		HealthCheckPath: data.HealthCheckPath.ValueString(),
		VPCID:           data.VPCID.ValueString(),
	}

	lb, err := r.client.CreateLoadBalancer(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create load balancer, got error: %s", err))
		return
	}

	data.ID = types.StringValue(lb.UUID)
	data.Name = types.StringValue(lb.Name)
	data.Algorithm = types.StringValue(lb.Algorithm)
	data.Port = types.Int64Value(int64(lb.Port))
	data.HealthCheckPath = types.StringValue(lb.HealthCheckPath)
	data.VPCID = types.StringValue(lb.VPCID)
	data.Status = types.StringValue(lb.Status)
	data.VIP = types.StringValue(lb.VIP)
	data.CreatedAt = types.StringValue(lb.CreatedAt)
	data.UpdatedAt = types.StringValue(lb.UpdatedAt)

	tflog.Trace(ctx, "created a load balancer resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lb, err := r.client.GetLoadBalancer(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read load balancer, got error: %s", err))
		return
	}

	data.ID = types.StringValue(lb.UUID)
	data.Name = types.StringValue(lb.Name)
	data.Algorithm = types.StringValue(lb.Algorithm)
	data.Port = types.Int64Value(int64(lb.Port))
	data.HealthCheckPath = types.StringValue(lb.HealthCheckPath)
	data.VPCID = types.StringValue(lb.VPCID)
	data.Status = types.StringValue(lb.Status)
	data.VIP = types.StringValue(lb.VIP)
	data.CreatedAt = types.StringValue(lb.CreatedAt)
	data.UpdatedAt = types.StringValue(lb.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &models.UpdateLoadBalancerRequest{
		Algorithm:       data.Algorithm.ValueString(),
		HealthCheckPath: data.HealthCheckPath.ValueString(),
	}

	lb, err := r.client.UpdateLoadBalancer(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update load balancer, got error: %s", err))
		return
	}

	data.Algorithm = types.StringValue(lb.Algorithm)
	data.HealthCheckPath = types.StringValue(lb.HealthCheckPath)
	data.Status = types.StringValue(lb.Status)
	data.UpdatedAt = types.StringValue(lb.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLoadBalancer(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete load balancer, got error: %s", err))
		return
	}
}

func (r *LoadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

### 6.4 Registration Diff — `internal/provider/provider.go`

```diff
 func (p *AirtelCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
     return []func() resource.Resource{
         NewVMResource,
         NewVolumeResource,
         // ... existing ...
         NewSecurityGroupRuleResource,
+        NewLoadBalancerResource,
     }
 }
```

### 6.5 Mock Handler — `internal/client/testutil/mock_server.go`

Add to `setupDefaultHandlers()`:

```go
// Load Balancer handlers
ms.Handlers["POST /api/v1/lbaas/domain/test-org/project/test-project/loadbalancers/"] = ms.createLoadBalancerHandler
ms.Handlers["GET /api/v1/lbaas/domain/test-org/project/test-project/loadbalancers/lb-uuid-1/"] = ms.getLoadBalancerHandler
ms.Handlers["PUT /api/v1/lbaas/domain/test-org/project/test-project/loadbalancers/lb-uuid-1/"] = ms.updateLoadBalancerHandler
ms.Handlers["DELETE /api/v1/lbaas/domain/test-org/project/test-project/loadbalancers/lb-uuid-1/"] = ms.deleteLoadBalancerHandler
```

Add handler methods:

```go
func (ms *MockServer) createLoadBalancerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	lb := models.LoadBalancer{
		UUID:            "lb-uuid-1",
		Name:            "test-lb",
		Algorithm:       "round_robin",
		Port:            80,
		HealthCheckPath: "/",
		Status:          "ACTIVE",
		VIP:             "10.0.1.100",
		VPCID:           "vpc-1",
	}
	json.NewEncoder(w).Encode(lb)
}

func (ms *MockServer) getLoadBalancerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	lb := models.LoadBalancer{
		UUID:            "lb-uuid-1",
		Name:            "test-lb",
		Algorithm:       "round_robin",
		Port:            80,
		HealthCheckPath: "/",
		Status:          "ACTIVE",
		VIP:             "10.0.1.100",
		VPCID:           "vpc-1",
	}
	json.NewEncoder(w).Encode(lb)
}

func (ms *MockServer) updateLoadBalancerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	lb := models.LoadBalancer{
		UUID:            "lb-uuid-1",
		Name:            "test-lb",
		Algorithm:       "least_connections",
		Port:            80,
		HealthCheckPath: "/health",
		Status:          "ACTIVE",
		VIP:             "10.0.1.100",
		VPCID:           "vpc-1",
	}
	json.NewEncoder(w).Encode(lb)
}

func (ms *MockServer) deleteLoadBalancerHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	ms.Handlers["GET /api/v1/lbaas/domain/test-org/project/test-project/loadbalancers/lb-uuid-1/"] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}
```

### 6.6 Unit Test — `internal/client/load_balancer_test.go`

```go
package client

import (
	"context"
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestCreateLoadBalancer(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret",
		"south-1", "test-org", "test-project", "", "")

	tests := []struct {
		name    string
		request *models.CreateLoadBalancerRequest
		wantErr bool
	}{
		{
			name: "successful creation",
			request: &models.CreateLoadBalancerRequest{
				Name:      "test-lb",
				Algorithm: "round_robin",
				Port:      80,
				VPCID:     "vpc-1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb, err := client.CreateLoadBalancer(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLoadBalancer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if lb == nil {
					t.Fatal("CreateLoadBalancer() returned nil")
				}
				if lb.UUID == "" {
					t.Error("CreateLoadBalancer() returned empty UUID")
				}
				if lb.Name != "test-lb" {
					t.Errorf("CreateLoadBalancer() name = %v, want test-lb", lb.Name)
				}
				if lb.Status != "ACTIVE" {
					t.Errorf("CreateLoadBalancer() status = %v, want ACTIVE", lb.Status)
				}
			}
		})
	}
}

func TestGetLoadBalancer(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret",
		"south-1", "test-org", "test-project", "", "")

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "successful retrieval",
			id:      "lb-uuid-1",
			wantErr: false,
		},
		{
			name:    "not found",
			id:      "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb, err := client.GetLoadBalancer(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLoadBalancer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if lb == nil {
					t.Fatal("GetLoadBalancer() returned nil")
				}
				if lb.UUID != tt.id {
					t.Errorf("GetLoadBalancer() UUID = %v, want %v", lb.UUID, tt.id)
				}
			}
		})
	}
}

func TestDeleteLoadBalancer(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret",
		"south-1", "test-org", "test-project", "", "")

	err := client.DeleteLoadBalancer(context.Background(), "lb-uuid-1")
	if err != nil {
		t.Errorf("DeleteLoadBalancer() error = %v", err)
	}
}
```

### 6.7 Acceptance Test — `tests/resources/load_balancer_test.go`

```go
package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoadBalancerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig("test-lb", 80),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"airtelcloud_load_balancer.test", "name", "test-lb"),
					resource.TestCheckResourceAttr(
						"airtelcloud_load_balancer.test", "port", "80"),
					resource.TestCheckResourceAttr(
						"airtelcloud_load_balancer.test", "algorithm", "round_robin"),
					resource.TestCheckResourceAttrSet(
						"airtelcloud_load_balancer.test", "id"),
					resource.TestCheckResourceAttrSet(
						"airtelcloud_load_balancer.test", "status"),
					resource.TestCheckResourceAttrSet(
						"airtelcloud_load_balancer.test", "vip"),
				),
			},
			{
				ResourceName:      "airtelcloud_load_balancer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfigUpdated("test-lb", 80),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"airtelcloud_load_balancer.test", "algorithm", "least_connections"),
					resource.TestCheckResourceAttr(
						"airtelcloud_load_balancer.test", "health_check_path", "/health"),
				),
			},
		},
	})
}

func testAccLoadBalancerConfig(name string, port int) string {
	return fmt.Sprintf(`
resource "airtelcloud_load_balancer" "test" {
  name   = %[1]q
  port   = %[2]d
  vpc_id = "vpc-1"
}
`, name, port)
}

func testAccLoadBalancerConfigUpdated(name string, port int) string {
	return fmt.Sprintf(`
resource "airtelcloud_load_balancer" "test" {
  name              = %[1]q
  port              = %[2]d
  vpc_id            = "vpc-1"
  algorithm         = "least_connections"
  health_check_path = "/health"
}
`, name, port)
}
```

### 6.8 Example Config — `examples/resources/load_balancer/main.tf`

```hcl
resource "airtelcloud_load_balancer" "web" {
  name              = "web-lb"
  port              = 80
  algorithm         = "round_robin"
  health_check_path = "/health"
  vpc_id            = airtelcloud_vpc.main.id
}
```

---

## 7. Implementation Checklist

### Pre-Implementation
- [ ] Identified API content-type (JSON or form-encoded)
- [ ] Identified resource ID type (string UUID or integer)
- [ ] Confirmed whether update API exists
- [ ] Confirmed whether create returns the full resource
- [ ] Determined if this is a child resource
- [ ] Checked if delete is asynchronous

### Models (`internal/models/{name}.go`)
- [ ] Response struct with `json` tags
- [ ] Create request struct with appropriate tags (`json` or `form`)
- [ ] Update request struct (if applicable)
- [ ] Used `*string` for nullable optional fields
- [ ] Used `json.RawMessage` for inconsistent types (if needed)

### Client (`internal/client/{name}.go`)
- [ ] Base path function defined
- [ ] `Get{Name}` method
- [ ] `Create{Name}` method (with read-after-create fallback if needed)
- [ ] `Update{Name}` method (if applicable)
- [ ] `Delete{Name}` method with polling until 404
- [ ] Trailing slashes on API paths (if required)
- [ ] Path prefix is correct (not double-prefixed)

### Provider (`internal/provider/{name}_resource.go`)
- [ ] Compile-time interface checks (`var _ resource.Resource = ...`)
- [ ] Constructor function (`New{Name}Resource`)
- [ ] Terraform model struct with `tfsdk` tags
- [ ] `Metadata` method (correct type name)
- [ ] `Schema` method with all attributes
- [ ] `UseStateForUnknown()` on computed stable fields (id, uuid)
- [ ] `RequiresReplace()` on immutable fields
- [ ] Defaults on Optional+Computed fields
- [ ] Validators where appropriate
- [ ] `Configure` method (boilerplate)
- [ ] `Create` method (plan → request → client → map → state)
- [ ] `Read` method (state → client → 404 check → map → state)
- [ ] `Update` method (plan → request → client → map → state) or error
- [ ] `Delete` method (state → client)
- [ ] `ImportState` method (correct pattern for ID type)
- [ ] `ValidateConfig` (if cross-attribute validation needed)
- [ ] Null handling for optional pointer fields

### Registration
- [ ] Added `New{Name}Resource` to `Resources()` in `provider.go`

### Testing
- [ ] Mock handlers added to `mock_server.go`
- [ ] Unit tests in `internal/client/{name}_test.go`
- [ ] Integration tests in `internal/client/{name}_integration_test.go`
- [ ] Acceptance tests in `tests/resources/{name}_test.go`
- [ ] All tests pass: `go test ./internal/client/ -v -run Test{Name}`

### Documentation
- [ ] Example config in `examples/resources/{name}/main.tf`

---

## 8. Resource Pattern Reference Table

| Resource | Encoding | ID Type | Has Update | Async Delete | Child | Import Format | Model File | Client File | Resource File |
|----------|----------|---------|------------|-------------|-------|---------------|------------|-------------|---------------|
| VM | Form | String | Yes | Poll (10s) | No | `{id}` | `models/vm.go` | `client/vm.go` | `provider/vm_resource.go` |
| Volume | Form | Integer | Yes (size) | Poll (5s) | No | `{id}` (int) | `models/volume.go` | `client/volume.go` | `provider/volume_resource.go` |
| VPC | JSON | String | Yes | No | No | `{id}` | `models/vpc.go` | `client/vpc.go` | `provider/vpc_resource.go` |
| Subnet | JSON | String | Yes | Poll | Yes | `{network_id}/{subnet_id}` | `models/subnet.go` | `client/subnet.go` | `provider/subnet_resource.go` |
| Object Storage Bucket | JSON | String (name) | Yes | Poll (5s) | No | `{name}` | `models/object_storage.go` | `client/object_storage.go` | `provider/object_storage_bucket_resource.go` |
| Object Storage Access Key | JSON | String | No | No | Yes | `{access_key_id}` | `models/object_storage.go` | `client/object_storage.go` | `provider/object_storage_access_key_resource.go` |
| File Storage | JSON | String (UUID) | Yes | Poll | No | `{uuid}` | `models/nfs.go` | `client/nfs.go` | `provider/file_storage_resource.go` |
| File Storage Export | JSON | String | Yes | No | Yes | `{volume_id}/{path_id}` | `models/nfs.go` | `client/nfs.go` | `provider/file_storage_export_path_resource.go` |
| DNS Zone | JSON | String (UUID) | Yes (desc) | Poll (5s) | No | `{uuid}` | `models/dns_zone.go` | `client/dns_zone.go` | `provider/dns_zone_resource.go` |
| DNS Record | JSON | String (UUID) | Yes | Poll | Yes | `{zone_id}/{record_uuid}` | `models/dns_record.go` | `client/dns_record.go` | `provider/dns_record_resource.go` |
| Security Group | Form | Integer | No | Poll (5s) | No | `{id}` (int, parsed) | `models/security_group.go` | `client/security_group.go` | `provider/security_group_resource.go` |
| Security Group Rule | JSON* | Integer | No | Poll (5s) | Yes | `{sg_id}/{rule_id}` | `models/security_group.go` | `client/security_group.go` | `provider/security_group_rule_resource.go` |
| VPC Peering | JSON | String | No | No | No | `{id}` | `models/vpc_peering.go` | `client/vpc_peering.go` | `provider/vpc_peering_resource.go` |
| LB Service | URL-encoded | String | No | No | No | `{id}` | `models/lb_service.go` | `client/lb_service.go` | `provider/lb_service_resource.go` |
| LB VIP | JSON | String | No | No | Yes | `{lb_service_id}/{vip}` | `models/lb_service.go` | `client/lb_service.go` | `provider/lb_vip_resource.go` |
| LB Certificate | URL-encoded | String | No | No | Yes | `{lb_service_id}/{name}` | `models/lb_service.go` | `client/lb_service.go` | `provider/lb_certificate_resource.go` |
| LB Virtual Server | URL-encoded | String | No | No | Yes | `{lb_service_id}/{name}` | `models/lb_service.go` | `client/lb_virtual_server.go` | `provider/lb_virtual_server_resource.go` |
| Compute Snapshot | URL-encoded | String (UUID) | No | Poll (10s) | No | `{uuid}` | `models/snapshot.go` | `client/snapshot.go` | `provider/compute_snapshot_resource.go` |
| Protection | URL-encoded | String (int) | Yes | No | No | `{id}` | `models/backup.go` | `client/backup.go` | `provider/protection_resource.go` |
| Protection Plan | URL-encoded | String (int) | No | No | No | `{id}` | `models/backup.go` | `client/backup.go` | `provider/protection_plan_resource.go` |
| Public IP | JSON | String (UUID) | No | No | No | `{uuid}` | `models/public_ip.go` | `client/public_ip.go` | `provider/public_ip_resource.go` |
| Public IP Policy Rule | JSON | String (UUID) | No | No | Yes | composite† | `models/public_ip.go` | `client/public_ip.go` | `provider/public_ip_policy_rule_resource.go` |

\* Security Group Rule creation serializes JSON into a form field (`security_group_data`).
† Public IP Policy Rule lookup requires `public_ip_id`, `target_vip`, `public_ip`, and `rule_uuid`.

All model files are in `internal/models/`. All client files are in `internal/client/`. All resource files are in `internal/provider/`.

**Best starting templates by pattern:**
- **Simplest JSON resource**: `dns_zone_resource.go` + `dns_zone.go` (client) + `dns_zone.go` (models)
- **Simplest immutable resource**: `compute_snapshot_resource.go` (no update, minimal schema)
- **Form-encoded resource**: `security_group_resource.go` + `security_group.go` (client)
- **No-update pattern**: `compute_snapshot_resource.go` or `public_ip_resource.go`
- **Child resource with composite import**: `security_group_rule_resource.go`
- **Complex resource with validators**: `vm_resource.go`
- **Integer ID import**: `security_group_resource.go`
- **Resource with timeouts block**: `compute_snapshot_resource.go` or `public_ip_resource.go`
- **Resource with availability zone scoping**: `public_ip_resource.go` (uses `WithAvailabilityZone()`)
- **Async create with polling**: `public_ip_resource.go` (`WaitForPublicIPReady`)
- **Shared model file (multiple resources)**: `lb_service.go` (models for LB Service, VIP, Certificate, Virtual Server)
