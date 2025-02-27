# Minikube Setup Guide

This guide provides step-by-step instructions to set up Minikube, configure Docker registry credentials, mount a directory, enable ingress, and expose services.

## Prerequisites

- Installed [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/)
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
  --docker-server=thirdaiplatform.azurecr.io \
  --docker-username=thirdaiplatform-pull-local-dev \
  --docker-password='MBe9v3dhcRf1ZiroZEkBTz1foVJPHa9hPG6FcLzYWu+ACRDGHVcp' \
  -n default --dry-run=client -o yaml | kubectl apply -f -
```

### 3. Mount Local Directory to Minikube

Mount the local directory `/opt/model_bazaar/jobs/` to Minikube:

```sh
minikube mount /opt/model_bazaar/jobs/:/opt/model_bazaar/jobs/
```

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

## Run Tilt

Start your local development environment with Tilt:

```bash
tilt up
```

## Deploy the Helm Chart

Install the ModelBazaar Helm chart:

```bash
helm install platform-local ./helm-local
```

To upgrade the release after making changes:

```bash
helm upgrade platform-local ./helm-local
```

To render the templates without deploying:

```bash
helm template platform-local ./helm-local
```
