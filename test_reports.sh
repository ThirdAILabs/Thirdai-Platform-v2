#!/bin/bash

# Configuration
BASE_URL="http://localhost:8000"
DEPLOYMENT_ID="105481fe-8719-45b0-9a60-fcf4cb8389d8"
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

# Function to get report status
get_report_status() {
    local token=$1
    local report_id=$2
    
    echo >&2 "Getting status for report: ${report_id}"
    curl -s -X GET "${DEPLOYMENT_BASE_URL}/report/${report_id}" \
        -H "Authorization: Bearer ${token}"
}

# Function to list reports
list_reports() {
    local token=$1
    
    echo >&2 "Listing all reports..."
    curl -s -X GET "${DEPLOYMENT_BASE_URL}/reports" \
        -H "Authorization: Bearer ${token}"
}

# Function to delete report
delete_report() {
    local token=$1
    local report_id=$2
    
    echo >&2 "Deleting report: ${report_id}"
    curl -s -X DELETE "${DEPLOYMENT_BASE_URL}/report/${report_id}" \
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
    # Create a simple PDF using pdftk or convert
    if command -v pdftk &> /dev/null; then
        echo "This is a test document for report creation." | pdftk - output "$TEST_FILE"
    elif command -v convert &> /dev/null; then
        convert -size 400x200 xc:white -pointsize 20 -draw "text 20,100 'This is a test document for report creation.'" "$TEST_FILE"
    else
        echo "Error: Neither pdftk nor ImageMagick's convert is installed. Please install one of them to create PDF files."
        exit 1
    fi
fi

# Test the endpoints
echo -e "\n1. Creating a new report..."
TOKEN=$(get_access_token)  # Refresh token
REPORT_RESPONSE=$(create_report "$TOKEN" "$TEST_FILE" "[\"NAME\", \"O\"]")
echo "Response:"
echo "$REPORT_RESPONSE" | jq '.'
REPORT_ID=$(echo "$REPORT_RESPONSE" | jq -r '.data.report_id // empty')
echo "Created report with ID: $REPORT_ID"

if [ -n "$REPORT_ID" ]; then
    echo -e "\n2. Getting report status..."
    TOKEN=$(get_access_token)  # Refresh token
    get_report_status "$TOKEN" "$REPORT_ID" | jq '.'

    echo -e "\n3. Listing all reports..."
    TOKEN=$(get_access_token)  # Refresh token
    list_reports "$TOKEN" | jq '.'

    echo -e "\n4. Deleting the report..."
    TOKEN=$(get_access_token)  # Refresh token
    delete_report "$TOKEN" "$REPORT_ID" | jq '.'
else
    echo "Failed to create report or extract report ID"
fi 