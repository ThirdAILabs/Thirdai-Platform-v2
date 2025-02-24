#!/bin/bash
set -x

echo "Checking for required deployments..."

# List of deployments to update
deployments=(on-prem-generation llm-dispatch llm-cache)

for dep in "${deployments[@]}"; do
  if kubectl get deployment "$dep" &>/dev/null; then
    echo "Deployment '$dep' found."
  else
    echo "Deployment '$dep' not found. Skipping update for '$dep'."
  fi
done

echo ""
echo "Updating deployments with new image..."

if kubectl get deployment on-prem-generation &>/dev/null; then
  kubectl set image deployment/on-prem-generation backend=thirdaiplatform.azurecr.io/thirdai_platform_jobs_local-dev:latest --record
  echo "Updated 'on-prem-generation'."
fi

if kubectl get deployment llm-dispatch &>/dev/null; then
  kubectl set image deployment/llm-dispatch backend=thirdaiplatform.azurecr.io/thirdai_platform_jobs_local-dev:latest --record
  echo "Updated 'llm-dispatch'."
fi

if kubectl get deployment llm-cache &>/dev/null; then
  kubectl set image deployment/llm-cache backend=thirdaiplatform.azurecr.io/thirdai_platform_jobs_local-dev:latest --record
  echo "Updated 'llm-cache'."
fi

echo ""
echo "Updating all jobs starting with 'train-'..."

# Get jobs with names starting with "train-"
job_names=$(kubectl get jobs -o jsonpath='{.items[*].metadata.name}' | tr ' ' '\n' | grep '^train-')
if [ -z "$job_names" ]; then
  echo "No jobs starting with 'train-' found."
else
  for job in $job_names; do
    echo "Processing job '$job'..."
    # Retrieve the current job manifest
    kubectl get job "$job" -o yaml > /tmp/job.yaml

    # Remove immutable/auto-generated fields one by one:
    yq eval 'del(.metadata.uid)' -i /tmp/job.yaml
    yq eval 'del(.metadata.resourceVersion)' -i /tmp/job.yaml
    yq eval 'del(.metadata.creationTimestamp)' -i /tmp/job.yaml
    yq eval 'del(.metadata.managedFields)' -i /tmp/job.yaml
    yq eval 'del(.status)' -i /tmp/job.yaml
    yq eval 'del(.spec.selector)' -i /tmp/job.yaml
    yq eval 'del(.spec.template.metadata.labels."controller-uid")' -i /tmp/job.yaml
    yq eval 'del(.spec.template.metadata.labels."batch.kubernetes.io/controller-uid")' -i /tmp/job.yaml
    yq eval 'del(.spec.template.metadata.labels."job-name")' -i /tmp/job.yaml

    # Update the container image for the container named "backend"
    yq eval '.spec.template.spec.containers[] |= (select(.name == "backend") .image = "thirdaiplatform.azurecr.io/thirdai_platform_jobs_local-dev:latest")' -i /tmp/job.yaml

    echo "Deleting job '$job'..."
    kubectl delete job "$job"

    echo "Recreating job '$job' with updated manifest..."
    kubectl create -f /tmp/job.yaml
  done
fi

echo ""
echo "Done updating deployments/jobs."
