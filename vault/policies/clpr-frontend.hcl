# Policy for Clpr Frontend AppRole
# Grants read-only access to the frontend secrets in KV v2

path "kv/data/clpr/frontend" {
  capabilities = ["read"]
}

path "kv/metadata/clpr/frontend" {
  capabilities = ["read", "list"]
}
