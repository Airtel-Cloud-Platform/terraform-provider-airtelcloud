# Airtel Cloud Provider Authentication

## Overview

The Airtel Cloud provider uses a multi-header authentication scheme including HMAC signature and additional context headers.

## Required Headers

### 1. Ce-Auth (HMAC Authentication)
- **Format**: `{apiKey}.{expiry}.{signature}`
- **Example**: `4a15ebf5-68ee-417f-99cf-482f4f5273a9.1730030000.a1b2c3d4...`
- **Generation**:
  - `expiry` = current Unix timestamp + 120 seconds
  - `signature` = HMAC-SHA256(`{apiKey}.{expiry}`, apiSecret)
- **Purpose**: Primary authentication and request validation

### 2. ce-region
- **Format**: String value of the region
- **Example**: `south-1`
- **Default**: `south-1` (if not specified)
- **Purpose**: Identifies the target region for API calls

### 3. organisation-name (Optional)
- **Format**: String value of organization name or domain
- **Example**: `my-company` or `company.com`
- **Default**: Not sent if not configured
- **Purpose**: Organization context for multi-tenant environments

### 4. Project-Name (Optional)
- **Format**: String value of project name
- **Example**: `my-project` or `webapp-prod`
- **Default**: Not sent if not configured
- **Purpose**: Project context for API calls

## Provider Configuration

```hcl
provider "airtelcloud" {
  api_endpoint = "https://south.cloud.airtel.in"
  api_key      = var.airtel_api_key      # Required
  api_secret   = var.airtel_api_secret   # Required
  region       = "south-1"               # Optional, defaults to "south-1"
  organization = var.airtel_organization # Optional
  project_name = var.airtel_project_name # Optional
}
```

## Environment Variables

```bash
# Required
AIRTEL_API_KEY=your-api-key-here
AIRTEL_API_SECRET=your-api-secret-here

# Optional
AIRTEL_API_ENDPOINT=https://south.cloud.airtel.in
AIRTEL_REGION=south-1
AIRTEL_ORGANIZATION=your-organization-name
AIRTEL_PROJECT_NAME=your-project-name
```

## Example HTTP Request

```http
POST /api/v1/volumes/ HTTP/1.1
Host: south.cloud.airtel.in
Content-Type: multipart/form-data; boundary=...
Accept: application/json
User-Agent: terraform-provider-airtelcloud/0.2.0
Ce-Auth: 4a15ebf5-68ee-417f-99cf-482f4f5273a9.1730030000.a1b2c3d4e5f6...
ce-region: south-1
organisation-name: my-company
Project-Name: my-project
```

## Troubleshooting

### 401 Unauthorized
- Verify API key and secret are correct
- Check HMAC signature generation
- Ensure expiry timestamp is not too old/future
- Verify all required headers are present

### Region Issues
- Confirm `ce-region` header matches your account region
- Check if resources exist in the specified region

### Organization Issues
- Verify organization name matches your account
- Check if organization header is required for your account type

## Debug Logging

Enable debug logging to see all authentication headers:

```bash
export TF_LOG=DEBUG
terraform plan
```

Look for log entries showing request headers including Ce-Auth, ce-region, and organisation-name values.