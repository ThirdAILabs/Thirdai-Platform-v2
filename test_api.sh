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

# Function to make prediction
make_prediction() {
    local token=$1
    local text=$2
    local top_k=${3:-1}
    
    echo "Making prediction for text: ${text}"
    curl -X POST "${DEPLOYMENT_BASE_URL}/predict" \
        -H "Authorization: Bearer ${token}" \
        -H "Content-Type: application/json" \
        -d "{
            \"text\": \"${text}\",
            \"top_k\": ${top_k}
        }"
}

# Function to get labels
get_labels() {
    local token=$1
    
    echo "Getting labels..."
    curl -X GET "${DEPLOYMENT_BASE_URL}/get_labels" \
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

# Test the endpoints
echo -e "\n1. Testing get_labels endpoint:"
get_labels "$TOKEN"

echo -e "\n2. Testing predict endpoint with simple text:"
make_prediction "$TOKEN" "hi"

echo -e "\n3. Testing predict endpoint with longer text and top_k=3:"
make_prediction "$TOKEN" "This is a longer text to test the model's prediction capabilities" 3 