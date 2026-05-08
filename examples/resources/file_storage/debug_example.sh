#!/bin/bash

# Example: How to enable debug logging for the Airtel Cloud provider
#
# This will show detailed HTTP request/response logs including:
# - Request headers, body, and form data
# - Response headers and body
# - Error details and parsing information
# - Authentication header generation

echo "Running Terraform with debug logging enabled..."

# Set Terraform log level to DEBUG
export TF_LOG=DEBUG

# Optional: Save logs to a file
export TF_LOG_PATH=terraform-debug.log

# Run terraform plan with debug logging
terraform plan

terraform apply --auto-approve

echo "Debug logs written to terraform-debug.log"
echo "You can also view logs in real-time by running:"
echo "export TF_LOG=DEBUG && terraform plan"
