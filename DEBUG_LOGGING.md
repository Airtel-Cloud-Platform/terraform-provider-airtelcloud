# Debug Logging in Airtel Cloud Provider

## Overview

The Airtel Cloud provider now includes comprehensive debug logging to help troubleshoot API interactions and authentication issues.

## Enabling Debug Logging

Set the Terraform log level to DEBUG:

```bash
export TF_LOG=DEBUG
terraform plan

# Or save logs to a file:
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform-debug.log
terraform apply
```

## What Gets Logged

### Request Logging
- **HTTP Method and URL**: Shows the exact API endpoint being called
- **Request Headers**: Including the Ce-Auth header with HMAC authentication
- **Request Body**: For JSON requests (POST/PUT with application/json)
- **Form Data**: For multipart form requests with field names and values
- **Content Type**: Shows whether request uses JSON or form-data encoding

### Response Logging
- **Status Code and Status**: HTTP response status
- **Response Headers**: All headers returned by the API
- **Response Body**: Full response body for both success and error cases
- **Error Details**: Detailed error parsing and formatting

### Authentication Logging
- **HMAC Generation**: Debug info about Ce-Auth header creation
- **API Endpoint**: Shows which endpoint is being used
- **Region Header**: ce-region header with region value
- **Organization Header**: organisation-name header (if configured)
- **Request Timing**: When requests are made and completed

## Log Message Patterns

Look for these patterns in the debug output:

```
Making API request - Standard JSON requests
Making form API request - Form-data requests
Received API response - Successful responses
Received form API response - Form request responses
GET/POST/PUT response body - Response content
Error response body - Error details
Form error response body - Form request errors
```

## Example Debug Output

```
[DEBUG] provider.terraform-provider-airtelcloud: Making API request: method=POST url=https://south.cloud.airtel.in/api/v1/volumes/
[DEBUG] provider.terraform-provider-airtelcloud: Request body: {"volume_name":"test","volume_size":10}
[DEBUG] provider.terraform-provider-airtelcloud: Received API response: status_code=201 status="201 Created"
[DEBUG] provider.terraform-provider-airtelcloud: POST response body: {"id":123,"status":"creating"}
```

## Troubleshooting

### Authentication Issues
- Look for `Ce-Auth` header in request logs
- Check if API endpoint URL is correct
- Verify HMAC signature generation

### API Errors
- Check response status codes
- Examine error response bodies
- Look for parsing errors

### Form Data Issues
- Verify form fields are being sent correctly
- Check Content-Type headers for multipart/form-data
- Review form field values and types

## Security Note

Debug logs may contain sensitive information including:
- API keys (in request headers)
- Request/response data
- Error details

**Never share debug logs publicly without removing sensitive data.**