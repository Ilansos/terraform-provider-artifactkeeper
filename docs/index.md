# Artifact Keeper Provider

The Artifact Keeper provider manages Artifact Keeper repositories, access control, webhooks, lifecycle policies, security policies, and license policies through the Artifact Keeper REST API.

Use this provider to describe artifact-management infrastructure as Terraform configuration: create local and remote repositories, manage users and groups, attach repository permissions, configure security scanning behavior, and publish release automation hooks.

## Example Usage

```terraform
terraform {
  required_providers {
    artifactkeeper = {
      source  = "artifactkeeper/artifactkeeper"
      version = "~> 0.1"
    }
  }
}

provider "artifactkeeper" {
  url   = "https://artifactkeeper.example.com"
  token = var.artifactkeeper_token
}
```

```terraform
variable "artifactkeeper_token" {
  type      = string
  sensitive = true
}

resource "artifactkeeper_repository" "docker_local" {
  key         = "docker-local"
  name        = "Docker Local"
  format      = "docker"
  repo_type   = "local"
  description = "Private Docker image repository"
  public      = false

  scan_enabled       = true
  scan_on_upload     = true
  scan_on_proxy      = false
  block_on_violation = true
  severity_threshold = "High"
}

resource "artifactkeeper_user" "alice" {
  username = "alice"
  email    = "alice@example.com"
  role     = "user"
}

resource "artifactkeeper_group" "developers" {
  name        = "developers"
  description = "Developer access"
  user_ids    = [artifactkeeper_user.alice.id]
}

resource "artifactkeeper_permission" "developers_read" {
  repository_id = artifactkeeper_repository.docker_local.id
  group_id      = artifactkeeper_group.developers.id
  permissions   = ["read"]
}
```

## Authentication

The provider supports bearer tokens and username/password login. Token authentication is recommended for automation.

### Bearer Token

```terraform
provider "artifactkeeper" {
  url   = "https://artifactkeeper.example.com"
  token = var.artifactkeeper_token
}
```

### Username And Password

```terraform
provider "artifactkeeper" {
  url      = "https://artifactkeeper.example.com"
  username = var.artifactkeeper_username
  password = var.artifactkeeper_password
}
```

### Environment Variables

Provider configuration can also be supplied through environment variables:

```shell
export ARTIFACTKEEPER_URL="https://artifactkeeper.example.com"
export ARTIFACTKEEPER_TOKEN="..."
```

Or, for username/password login:

```shell
export ARTIFACTKEEPER_URL="https://artifactkeeper.example.com"
export ARTIFACTKEEPER_USERNAME="admin"
export ARTIFACTKEEPER_PASSWORD="..."
```

The `url` value can be either the Artifact Keeper server root or a URL already ending in `/api/v1`. The provider normalizes the API path automatically.

## Common Workflows

### Remote Repository

```terraform
resource "artifactkeeper_repository" "docker_hub" {
  key          = "docker-hub"
  name         = "Docker Hub"
  format       = "docker"
  repo_type    = "remote"
  description  = "Remote proxy for Docker Hub"
  public       = false
  upstream_url = "https://registry-1.docker.io"
}
```

### Webhook

```terraform
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
```

### Lifecycle Policy

```terraform
resource "artifactkeeper_lifecycle_policy" "docker_keep_recent" {
  repository_id = artifactkeeper_repository.docker_local.id
  name          = "Docker: keep latest 10"
  description   = "Keep only the latest 10 Docker package versions"
  enabled       = true
  policy_type   = "max_versions"
  config        = jsonencode({ keep = 10 })
  priority      = 10
}
```

### Security Policy

```terraform
resource "artifactkeeper_security_policy" "strict_production" {
  name              = "strict-production"
  enabled           = true
  max_severity      = "high"
  block_unscanned   = true
  block_on_fail     = true
  require_signature = true
}
```

### License Policy

```terraform
resource "artifactkeeper_license_policy" "no_copyleft" {
  name        = "no-copyleft"
  description = "Block strong copyleft licenses"
  enabled     = true

  denied_licenses = [
    "GPL-2.0-only",
    "GPL-3.0-only",
    "AGPL-3.0-only",
  ]

  allowed_licenses = []
  allow_unknown    = false
  action           = "block"
}
```

### Look Up Existing Objects

```terraform
data "artifactkeeper_repository" "docker_local" {
  key = "docker-local"
}

data "artifactkeeper_repositories" "docker" {
  format = "docker"
}

data "artifactkeeper_packages" "docker" {
  repository_key = "docker-local"
  format         = "docker"
}
```

## Available Resources

- [`artifactkeeper_repository`](resources/repository.md)
- [`artifactkeeper_user`](resources/user.md)
- [`artifactkeeper_group`](resources/group.md)
- [`artifactkeeper_permission`](resources/permission.md)
- [`artifactkeeper_webhook`](resources/webhook.md)
- [`artifactkeeper_lifecycle_policy`](resources/lifecycle_policy.md)
- [`artifactkeeper_security_policy`](resources/security_policy.md)
- [`artifactkeeper_license_policy`](resources/license_policy.md)

## Available Data Sources

- `artifactkeeper_repository`
- `artifactkeeper_repositories`
- `artifactkeeper_user`
- `artifactkeeper_users`
- `artifactkeeper_package`
- `artifactkeeper_packages`

## Configuration Reference

### Required

One URL value and one authentication method must be provided.

- `url` - Artifact Keeper base URL. Accepts either the server root or a URL already ending in `/api/v1`. Can also be set with `ARTIFACTKEEPER_URL`.
- `base_url` - Alias for `url`. Do not set both `url` and `base_url`.

### Authentication

- `token` - Bearer token. Can also be set with `ARTIFACTKEEPER_TOKEN`.
- `username` - Username for password login. Can also be set with `ARTIFACTKEEPER_USERNAME`.
- `password` - Password for password login. Can also be set with `ARTIFACTKEEPER_PASSWORD`.

### Optional

- `insecure_skip_verify` - Skip TLS certificate verification for lab or development instances. Can also be set with `ARTIFACTKEEPER_INSECURE_SKIP_VERIFY`.

## Security Notes

Terraform stores resource state after apply. Sensitive provider attributes such as `token` and `password` are marked sensitive, but resource values such as initial user passwords, webhook secrets, and upstream repository passwords can still exist in Terraform state when configured. Protect local and remote Terraform state as sensitive material.

Use `insecure_skip_verify` only for local development or isolated lab environments.
