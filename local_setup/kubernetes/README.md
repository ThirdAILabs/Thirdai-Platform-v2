# Quick Start

This README provides the basic commands to run Tilt and deploy the ModelBazaar application using the Helm chart.

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
