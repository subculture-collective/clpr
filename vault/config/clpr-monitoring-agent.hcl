pid_file = "/vault-agent/pid/clpr-monitoring-agent.pid"

auto_auth {
  method "approle" {
    config = {
      role_id_file_path   = "/run/secrets/clpr_backend_role_id"
      secret_id_file_path = "/run/secrets/clpr_backend_secret_id"
    }
  }

  sink "file" {
    config = {
      path = "/vault-agent/rendered/token"
      mode = 0640
    }
  }
}

vault {
  address      = "https://vault.subcult.tv"
  unwrap_token = true
  retry {
    num_retries = 5
  }
}

template {
  source      = "/vault-agent/templates/postgres_exporter.env.ctmpl"
  destination = "/vault-agent/rendered/postgres_exporter.env"
  perms       = "0644"
  left_delimiter  = "{{"
  right_delimiter = "}}"
}
