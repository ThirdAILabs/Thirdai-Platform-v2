# Allow reading of all secrets
path "secret/data/*" {
  capabilities = ["read", "list"]
}

# Allow listing of all secrets
path "secret/metadata/*" {
  capabilities = ["list"]
}