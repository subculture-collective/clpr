# Clipper Backend AppRole Credentials

Place the Vault AppRole files generated from `vault read auth/approle/role/clpr-backend/role-id` and
`vault write -f auth/approle/role/clpr-backend/secret-id` in this directory:

- `role_id`
- `secret_id`

These files are mounted into the Vault agent container at runtime. Do **not** commit the actual values to git.
