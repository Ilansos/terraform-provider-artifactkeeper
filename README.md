# Terraform Provider for Artifact Keeper

This provider manages Artifact Keeper resources through the REST API under `/api/v1`.

The implementation targets the Artifact Keeper OpenAPI reference `1.0.0-rc.3`. Where the rendered docs and OpenAPI reference differ, this provider follows the OpenAPI shape.

## Build

```bash
go build -o terraform-provider-artifactkeeper
go test ./...
```

## CI And Releases

Pull requests to `main` run formatting checks, unit tests, vulnerability checks, and a provider build through GitHub Actions. Releases are intentionally manual for public repository safety: push a signed or reviewed version tag such as `v0.1.0`, then the release workflow publishes with GoReleaser, signed SHA256 checksums, and GitHub Releases so the Terraform Registry can discover them.

See [PUBLISHING.md](PUBLISHING.md) for the required GitHub secrets and release process.

## Local Installation

Terraform development overrides are the simplest local workflow:

```hcl
provider_installation {
  dev_overrides {
    "artifactkeeper/artifactkeeper" = "/absolute/path/to/terraform-provider-artifactkeeper"
  }

  direct {}
}
```

Place that in `~/.terraformrc`, then run Terraform from an example directory:

```bash
terraform init
terraform plan
```

You can also install into Terraform's local plugin mirror layout:

```bash
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/artifactkeeper/artifactkeeper/0.1.0/linux_amd64
go build -o ~/.terraform.d/plugins/registry.terraform.io/artifactkeeper/artifactkeeper/0.1.0/linux_amd64/terraform-provider-artifactkeeper
```

Adjust `linux_amd64` for your OS and architecture.

## Provider Configuration

```hcl
provider "artifactkeeper" {
  url      = "https://artifactkeeper.example.com"
  username = var.artifactkeeper_username
  password = var.artifactkeeper_password
}
```

The `url` can be either a server root or a URL already ending in `/api/v1`; the provider normalizes it and does not double-append the API path.

Direct bearer token authentication is also supported:

```hcl
provider "artifactkeeper" {
  url   = "https://artifactkeeper.example.com/api/v1"
  token = var.artifactkeeper_token
}
```

Environment variables:

```bash
export ARTIFACTKEEPER_URL="https://artifactkeeper.example.com"
export ARTIFACTKEEPER_USERNAME="admin"
export ARTIFACTKEEPER_PASSWORD="..."
# or
export ARTIFACTKEEPER_TOKEN="..."
export ARTIFACTKEEPER_INSECURE_SKIP_VERIFY=false
```

## Resources

```hcl
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
```

Remote repositories can set an upstream URL:

```hcl
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

For authenticated upstreams:

```hcl
resource "artifactkeeper_repository" "private_remote" {
  key                = "private-remote"
  name               = "Private Remote"
  format             = "docker"
  repo_type          = "remote"
  upstream_url       = "https://registry.example.com"
  upstream_auth_type = "basic"
  upstream_username  = var.upstream_username
  upstream_password  = var.upstream_password
}
```

```hcl
resource "artifactkeeper_user" "alice" {
  username = "alice"
  email    = "alice@example.com"
  password = var.alice_initial_password
  role     = "user"
}
```

```hcl
resource "artifactkeeper_group" "developers" {
  name        = "developers"
  description = "Developer access"
  user_ids    = [artifactkeeper_user.alice.id]
}
```

```hcl
resource "artifactkeeper_permission" "developers_read" {
  repository_id = artifactkeeper_repository.docker_local.id
  group_id      = artifactkeeper_group.developers.id
  permissions   = ["read"]
}
```

```hcl
resource "artifactkeeper_webhook" "repo_events" {
  name          = "repository-events"
  url           = "https://hooks.example.com/artifactkeeper"
  events        = ["artifact_uploaded", "artifact_deleted"]
  repository_id = artifactkeeper_repository.docker_local.id
  secret        = var.webhook_secret
  enabled       = true
}
```

```hcl
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

Lifecycle policy types:

```text
max_age_days
max_versions
no_downloads_days
tag_pattern_keep
tag_pattern_delete
size_quota_bytes
```

The `config` value is a JSON object and depends on `policy_type`, for example:

```hcl
config = jsonencode({ days = 90 })
config = jsonencode({ keep = 5 })
config = jsonencode({ pattern = "^(release|stable)-" })
config = jsonencode({ quota_bytes = 10737418240 })
```

```hcl
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
```

```hcl
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

resource "artifactkeeper_license_policy" "docker_allowlist" {
  repository_id = artifactkeeper_repository.docker_local.id
  name          = "docker-license-allowlist"
  enabled       = true

  allowed_licenses = [
    "Apache-2.0",
    "BSD-3-Clause",
    "MIT",
  ]

  denied_licenses = []
  allow_unknown   = false
  action          = "block"
}
```

## Data Sources

```hcl
data "artifactkeeper_repository" "docker_local" {
  key = "docker-local"
}

data "artifactkeeper_repositories" "docker" {
  format = "docker"
}

data "artifactkeeper_user" "alice" {
  id = artifactkeeper_user.alice.id
}

data "artifactkeeper_users" "matches" {
  search = "alice"
}

data "artifactkeeper_package" "package" {
  id = "00000000-0000-0000-0000-000000000000"
}

data "artifactkeeper_packages" "docker" {
  repository_key = "docker-local"
  format         = "docker"
}
```

## Acceptance Tests

Acceptance tests are opt-in:

```bash
export ARTIFACTKEEPER_ACC=1
export ARTIFACTKEEPER_URL="https://artifactkeeper.example.com"
export ARTIFACTKEEPER_TOKEN="..."
go test ./...
```

Instead of `ARTIFACTKEEPER_TOKEN`, set `ARTIFACTKEEPER_USERNAME` and `ARTIFACTKEEPER_PASSWORD`.

## Known Limitations

- Repository `artifact_count` is not exposed because the current OpenAPI repository response does not include it.
- Webhooks cannot be updated through the current OpenAPI endpoint set except for enable/disable; most webhook changes require replacement.
- Webhook secrets, user passwords, and upstream repository passwords are write-only from the API perspective and cannot be read back.
- Remote repository upstream settings are create-time settings in the current provider and require replacement if changed.
- The current OpenAPI-backed security policy endpoint exposes threshold-style policies (`max_severity`, `block_unscanned`, `block_on_fail`, `require_signature`) rather than the prose documentation's free-form `rules` examples.
- Security policy CVE allowlisting is not exposed by the current security policy OpenAPI schema. CVE triage is currently available through finding acknowledgement workflows outside this resource.
- License policies use the SBOM API (`/api/v1/sbom/license-policies`) and upsert by name/scope. Changing `name` or `repository_id` replaces the Terraform resource.
- Package data sources use package IDs. Name-based package lookup is not implemented because the current OpenAPI reference exposes `/packages/{id}`.
- User `role` maps only to `user` and `admin` through the API `is_admin` flag. Full role assignment endpoints are outside this initial scope.
- `insecure_skip_verify` is intended only for lab and development environments.
