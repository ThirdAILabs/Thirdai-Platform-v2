#!/bin/bash

# Ensure Nomad is in your PATH
NOMAD_PATH=$(which nomad)
if [ -z "$NOMAD_PATH" ]; then
  echo "Nomad is not installed or not in your PATH. Please install it first."
  exit 1
fi

# Fetch all job names
echo "Fetching all Nomad jobs..."
NOMAD_JOBS=$($NOMAD_PATH job status | awk 'NR>1 {print $1}')

# Check if any jobs were found
if [ -z "$NOMAD_JOBS" ]; then
  echo "No Nomad jobs found."
  exit 0
fi

# Loop through all jobs to fetch their logs
for job in $NOMAD_JOBS; do
  echo "Fetching allocations for job: $job"

  # Fetch stdout logs
  $NOMAD_PATH alloc logs -job "$job"  > "${alloc_id}_stdout.log"
  echo "STDOUT Logs for allocation $alloc_id:"
  cat "${alloc_id}_stdout.log"

  # Fetch stderr logs
  $NOMAD_PATH alloc logs -job --stderr "$job"> "${alloc_id}_stderr.log"
  echo "STDERR Logs for allocation $alloc_id:"
  cat "${alloc_id}_stderr.log"
done
