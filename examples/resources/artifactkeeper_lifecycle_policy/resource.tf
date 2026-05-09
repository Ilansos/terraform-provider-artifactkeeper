resource "artifactkeeper_lifecycle_policy" "docker_keep_recent" {
  repository_id = artifactkeeper_repository.docker_local.id
  name          = "Docker: keep latest 10"
  description   = "Keep only the latest 10 Docker package versions"
  enabled       = true
  policy_type   = "max_versions"
  config        = jsonencode({ keep = 10 })
  priority      = 10
}

resource "artifactkeeper_lifecycle_policy" "global_unused" {
  name        = "Global: remove unused after 180 days"
  enabled     = true
  policy_type = "no_downloads_days"
  config      = jsonencode({ days = 180 })
  priority    = 1
}
