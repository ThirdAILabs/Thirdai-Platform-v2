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

# Function to create a report
create_report() {
    local token=$1
    local file_path=$2
    local tags=$3
    
    echo >&2 "Creating report with file: ${file_path}"
    curl -s -X POST "${DEPLOYMENT_BASE_URL}/report/create" \
        -H "Authorization: Bearer ${token}" \
        -F "documents={\"documents\":[{\"path\":\"$(basename ${file_path})\",\"location\":\"local\"}]}" \
        -F "tags=${tags:-[]}" \
        -F "files=@${file_path}"
}

# Function to list reports
list_reports() {
    local token=$1
    
    echo >&2 "Listing all reports..."
    curl -s -X GET "${DEPLOYMENT_BASE_URL}/reports" \
        -H "Authorization: Bearer ${token}"
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

# Create a test PDF file if it doesn't exist
TEST_FILE="test_report.pdf"
if [ ! -f "$TEST_FILE" ]; then
    echo "Creating test PDF file..."
    if command -v convert &> /dev/null; then
        convert -size 400x200 xc:white -pointsize 20 -draw "text 20,100 'This is a test document for report creation.'" "$TEST_FILE"
    else
        echo "Error: ImageMagick's convert is not installed. Please install it to create PDF files."
        exit 1
    fi
fi

# Create a new report
echo -e "\n1. Creating a new report..."
TOKEN=$(get_access_token)
REPORT_RESPONSE=$(create_report "$TOKEN" "$TEST_FILE" "[\"TEST\"]")
echo "Response:"
echo "$REPORT_RESPONSE" | jq '.'
REPORT_ID=$(echo "$REPORT_RESPONSE" | jq -r '.data.report_id // empty')

if [ -z "$REPORT_ID" ]; then
    echo "Failed to create report or extract report ID"
    exit 1
fi

echo "Created report with ID: $REPORT_ID"

# Immediately list reports (should show as queued)
echo -e "\n2. Listing reports immediately after creation..."
TOKEN=$(get_access_token)
echo "Reports list (should show as queued):"
list_reports "$TOKEN" | jq '.'

# Wait for report to complete (check status every 5 seconds)
echo -e "\n3. Waiting for report to complete..."
while true; do
    TOKEN=$(get_access_token)
    STATUS=$(get_report_status "$TOKEN" "$REPORT_ID" | jq -r '.data.status')
    echo "Current status: $STATUS"
    
    if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
        break
    fi
    
    sleep 5
done

# List reports again after completion
echo -e "\n4. Listing reports after completion..."
TOKEN=$(get_access_token)
echo "Reports list (after completion):"
list_reports "$TOKEN" | jq '.'

# Clean up
echo -e "\n5. Cleaning up..."
TOKEN=$(get_access_token)
curl -s -X DELETE "${DEPLOYMENT_BASE_URL}/report/${REPORT_ID}" \
    -H "Authorization: Bearer ${token}" 