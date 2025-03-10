# Minikube Setup Guide

This guide provides step-by-step instructions to set up Minikube, configure Docker registry credentials, mount a directory, enable ingress, and expose services.

## Prerequisites

- Install Minikube
- Install kubectl
- Install Tilt
- Install Helm
- Docker installed and running
- Sufficient system permissions for mounting directories

## Steps to Set Up Minikube

### 1. Start Minikube

Ensure Minikube is running:

```sh
minikube start
```

### 2. Create Docker Registry Secret

Create a Kubernetes secret for Azure Container Registry authentication:

```sh
kubectl create secret docker-registry docker-credentials-secret \
  --docker-server=<DOCKER_SERVER> \
  --docker-username=<DOCKER_USERNAME> \
  --docker-password=<DOCKER_PASSWORD> \
  -n default --dry-run=client -o yaml | kubectl apply -f -
```

### 3. Mount Local Directory to Minikube

Mount the local directory `/opt/model_bazaar/jobs/` to Minikube:
Create the following directory `/opt/model_bazaar/jobs` locally, with 777 permission for the users.

```sh
minikube mount /opt/model_bazaar/jobs/:/model_bazaar/jobs/
```

Add ndb_license to /opt/model_bazaar/jobs

### 4. Enable Ingress in Minikube

Enable the ingress controller to allow Ingress resources:

```sh
minikube addons enable ingress
```

### 5. Start Minikube Tunnel

Run the following command in a separate terminal to expose services:

```sh
minikube tunnel
```

## To Start the Application

### Using Tilt with automated docker building

Start your local development environment with Tilt(in folder local_setup/kubernetes):

```bash
tilt up
```

### Using push.py for Local Development

1. Run the script to build and push images:

```bash
python push.py --branch <branch-name> --version <version> --platform "linux/arm64"
```

This will:
- Build Docker images locally
- Push them to your local registry

