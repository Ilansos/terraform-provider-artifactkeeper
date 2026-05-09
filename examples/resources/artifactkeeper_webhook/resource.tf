variable "webhook_secret" {
  type      = string
  sensitive = true
}

resource "artifactkeeper_webhook" "repo_events" {
  name          = "repository-events"
  url           = "https://hooks.example.com/artifactkeeper"
  events        = ["artifact_uploaded", "artifact_deleted"]
  repository_id = artifactkeeper_repository.docker_local.id
  secret        = var.webhook_secret
  enabled       = true
}

