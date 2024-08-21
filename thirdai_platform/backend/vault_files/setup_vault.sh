# creates policy for nomad-server
vault policy write nomad-cluster thirdai_platform/backend/vault_files/nomad-cluster-policy.hcl

# creates policy for nomad-job
vault policy write nomad-job thirdai_platform/backend/vault_files/nomad-job-policy.hcl


# give auth to nomad-cluster policy to access nomad-job policy
vault write auth/token/roles/nomad-cluster \
  allowed_policies="nomad-cluster,nomad-job" \
  orphan=true \
  period="1h"
