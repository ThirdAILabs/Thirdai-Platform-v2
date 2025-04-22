#!/bin/bash

# Configuration
BASE_URL="http://localhost:8000"
DEPLOYMENT_ID="c44c1b14-7b39-4081-992a-81363c7dc8c1"
DEPLOYMENT_BASE_URL="http://localhost/${DEPLOYMENT_ID}"
EMAIL="peter@thirdai.com"
PASSWORD="pass"

# Function to get access token
get_access_token() {
    TOKEN=$(curl -s -X GET "${BASE_URL}/api/v2/user/login" \
        -H "Accept: application/json" \
        -H "Authorization: Basic $(echo -n "${EMAIL}:${PASSWORD}" | base64)" | \
        jq -r '.access_token')
    echo "$TOKEN"
}

# Function to list reports
list_reports() {
    local token=$1
    
    echo >&2 "Listing all reports..."
    curl -s -X GET "${DEPLOYMENT_BASE_URL}/reports" \
        -H "Authorization: Bearer ${token}"
}

# Main execution
echo "Getting access token..."
TOKEN=$(get_access_token)

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "Failed to get access token"
    exit 1
fi

echo "Access token obtained successfully"

# List reports
echo -e "\nListing all reports..."
TOKEN=$(get_access_token)
list_reports "$TOKEN" | jq '.' 