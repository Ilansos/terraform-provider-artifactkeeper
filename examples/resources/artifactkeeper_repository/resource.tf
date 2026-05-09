resource "artifactkeeper_repository" "docker_local" {
  key         = "docker-local"
  name        = "Docker Local"
  format      = "docker"
  repo_type   = "local"
  description = "Local Docker repository"
  public      = false

  scan_enabled       = true
  scan_on_upload     = true
  scan_on_proxy      = false
  block_on_violation = true
  severity_threshold = "High"
}

resource "artifactkeeper_repository" "docker_hub" {
  key          = "docker-hub"
  name         = "Docker Hub"
  format       = "docker"
  repo_type    = "remote"
  description  = "Remote proxy for Docker Hub"
  public       = false
  upstream_url = "https://registry-1.docker.io"
}
