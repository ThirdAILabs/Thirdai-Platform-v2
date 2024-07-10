import requests
import json
from requests.auth import HTTPBasicAuth

# Define the job details
job_data = {
    "name": "test-job",
    "image": "busybox",
    "command": "echo Hello, Nomad!"
}

# Define the FastAPI endpoint URL
url = "http://localhost:8000/submit_job"

# Make a POST request to the FastAPI endpoint
response = requests.post(url, headers={"Content-Type": "application/json"}, data=json.dumps(job_data), auth=HTTPBasicAuth("yashuroyal", "password"))

# response = requests.get(url,params={"job":"job"} )

# Print the response
print(f"Status Code: {response.status_code}")
print(f"Response Body: {response.content}")
