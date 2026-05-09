resource "artifactkeeper_license_policy" "no_copyleft" {
  name        = "no-copyleft"
  description = "Block strong copyleft licenses"
  enabled     = true

  denied_licenses = [
    "GPL-2.0-only",
    "GPL-3.0-only",
    "AGPL-3.0-only",
    "LGPL-2.1-only",
    "LGPL-3.0-only",
  ]

  allowed_licenses = []
  allow_unknown    = false
  action           = "block"
}

resource "artifactkeeper_license_policy" "docker_local_allowlist" {
  repository_id = artifactkeeper_repository.docker_local.id
  name          = "docker-local-license-allowlist"
  description   = "Allow only approved permissive licenses"
  enabled       = true

  allowed_licenses = [
    "Apache-2.0",
    "BSD-2-Clause",
    "BSD-3-Clause",
    "MIT",
  ]

  denied_licenses = []
  allow_unknown   = false
  action          = "block"
}
