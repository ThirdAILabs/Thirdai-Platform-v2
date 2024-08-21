# Allow reading of all secrets
path "secret/data/*" {
  capabilities = ["read", "create", "update", "list"]
}

# Allow listing of all secrets
path "secret/metadata/*" {
  capabilities = ["read", "create", "update", "list"]
}