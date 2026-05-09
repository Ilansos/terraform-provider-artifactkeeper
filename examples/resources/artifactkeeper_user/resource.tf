variable "user_initial_password" {
  type      = string
  sensitive = true
}

resource "artifactkeeper_user" "alice" {
  username = "alice"
  email    = "alice@example.com"
  password = var.user_initial_password
  role     = "user"
}

