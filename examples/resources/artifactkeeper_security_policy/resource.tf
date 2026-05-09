resource "artifactkeeper_security_policy" "strict_production" {
  name              = "strict-production"
  enabled           = true
  max_severity      = "high"
  block_unscanned   = true
  block_on_fail     = true
  require_signature = true
}

resource "artifactkeeper_security_policy" "docker_local" {
  repository_id     = artifactkeeper_repository.docker_local.id
  name              = "docker-local-security"
  enabled           = true
  max_severity      = "medium"
  block_unscanned   = false
  block_on_fail     = true
  require_signature = false
}
