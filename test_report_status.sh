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

# Function to get report status
get_report_status() {
    local token=$1
    local report_id=$2
    
    echo >&2 "Getting status for report: ${report_id}"
    curl -s -X GET "${DEPLOYMENT_BASE_URL}/report/${report_id}" \
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

# Check if report ID is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <report_id>"
    exit 1
fi

REPORT_ID=$1

# Get and print report status
echo -e "\nGetting report status..."
TOKEN=$(get_access_token)
STATUS_RESPONSE=$(get_report_status "$TOKEN" "$REPORT_ID")
echo "Report Status:"
echo "$STATUS_RESPONSE" | jq '.'

# Extract and print specific fields
STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.data.status')
SUBMITTED_AT=$(echo "$STATUS_RESPONSE" | jq -r '.data.submitted_at')
UPDATED_AT=$(echo "$STATUS_RESPONSE" | jq -r '.data.updated_at')
MSG=$(echo "$STATUS_RESPONSE" | jq -r '.data.msg // "No message"')

echo -e "\nReport Details:"
echo "Status: $STATUS"
echo "Submitted At: $SUBMITTED_AT"
echo "Updated At: $UPDATED_AT"
echo "Message: $MSG"

# If report is complete, print content summary
if [ "$STATUS" = "complete" ]; then
    echo -e "\nReport Content Summary:"
    echo "$STATUS_RESPONSE" | jq '.data.content.results[0] | keys'
fi 