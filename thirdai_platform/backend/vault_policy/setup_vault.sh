# make sure to start the independent vault service
# This repo uses HashiCorp Vault for storing secrets. You can use the following instructions to set it up:
# https://waytohksharma.medium.com/install-hashicorp-vault-on-mac-fdbd8cd9113b

# Make sure to setup the server in production and have addr and 
# token in the environment before running these commands

# creates policy for nomad-server
vault policy write nomad-cluster nomad-cluster-policy.hcl

# creates policy for nomad-job
vault policy write nomad-job nomad-job-policy.hcl

# creates policy for nomad-platform
vault policy write nomad-platform nomad-platform-policy.hcl

# give auth to nomad-cluster policy to access nomad-job policy
vault write auth/token/roles/nomad-cluster \
  allowed_policies="nomad-cluster,nomad-job,nomad-platform" \
  orphan=true \
  period="12h"
