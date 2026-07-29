#!/bin/bash 
API_KEY="${API_KEY}" 
SECRET="${SECRET}" 
# Generate expiry timestamp (current time + 120 seconds) 
EXPIRY=$(($(date +%s) + 120)) 
# Create the message: apiKey.expiry 
DATA="${API_KEY}.${EXPIRY}" 
# Generate HMAC-SHA256 signature 
SIGNATURE=$(echo -n "$DATA" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
# Combine data and signature 
AUTH_HEADER="${DATA}.${SIGNATURE}" 
# Example: curl request with Ce-Auth header 
curl -X GET "https://north.cloud.airtel.in/api/v1/compute/" -H "Ce-Auth: ${AUTH_HEADER}" 
echo $AUTH_HEADER
