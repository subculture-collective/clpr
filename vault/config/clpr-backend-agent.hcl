pid_file = "/vault-agent/pid/clpr-backend-agent.pid"

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
  address               = "http://vault:8200"
  retry {
    num_retries = 5
  }
}

template {
  source      = "/vault-agent/templates/backend.env.ctmpl"
  destination = "/vault-agent/rendered/backend.env"
  perms       = "0640"
  left_delimiter  = "{{"
  right_delimiter = "}}"
}

template {
  source      = "/vault-agent/templates/postgres.env.ctmpl"
  destination = "/vault-agent/rendered/postgres.env"
  perms       = "0640"
  left_delimiter  = "{{"
  right_delimiter = "}}"
}

template {
  source      = "/vault-agent/templates/frontend.env.ctmpl"
  destination = "/vault-agent/rendered/frontend.env"
  perms       = "0640"
  left_delimiter  = "{{"
  right_delimiter = "}}"
}
