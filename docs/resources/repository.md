# artifactkeeper_repository

Manages an Artifact Keeper repository. The resource supports local, remote, staging, and virtual repository shapes exposed by the current API.

## Example

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

```hcl
resource "artifactkeeper_repository" "docker_hub" {
  key          = "docker-hub"
  name         = "Docker Hub"
  format       = "docker"
  repo_type    = "remote"
  upstream_url = "https://registry-1.docker.io"
}
```

```hcl
resource "artifactkeeper_repository" "docker_virtual" {
  key       = "docker-virtual"
  name      = "Docker Virtual"
  format    = "docker"
  repo_type = "virtual"

  repositories = [
    artifactkeeper_repository.docker_local.key,
    artifactkeeper_repository.docker_hub.key,
  ]
}
```

## API

Uses `/api/v1/repositories`, `/api/v1/repositories/{key}`, `/api/v1/repositories/{key}/members`, and `/api/v1/repositories/{key}/security`.

## Arguments

- `key` (String, Required, Forces replacement): Stable repository key used in API paths.
- `name` (String, Required): Display name.
- `format` (String, Required, Forces replacement): Repository format, such as `docker`, `npm`, `maven`, or `generic`.
- `repo_type` (String, Optional, Defaults to `local`, Forces replacement): Repository type.
- `description` (String, Optional): Repository description.
- `public` (Boolean, Optional, Defaults to `false`): Maps to API `is_public`.
- `quota_bytes` (Number, Optional): Repository quota in bytes.
- `repositories` (List of String, Optional): Ordered member repository keys for virtual repositories.
- `upstream_url` (String, Optional, Forces replacement): Upstream URL for remote repositories.
- `index_upstream_url` (String, Optional, Forces replacement): Separate index URL for formats such as Cargo.
- `upstream_auth_type` (String, Optional, Forces replacement): Upstream auth type such as `basic` or `bearer`.
- `upstream_username` (String, Optional, Forces replacement): Upstream username.
- `upstream_password` (String, Optional, Sensitive, Forces replacement): Upstream password or token.
- `scan_enabled` (Boolean, Optional): Enables repository scanning.
- `scan_on_upload` (Boolean, Optional): Scans uploaded artifacts.
- `scan_on_proxy` (Boolean, Optional): Scans proxied artifacts.
- `block_on_violation` (Boolean, Optional): Blocks operations when policy violations are found.
- `severity_threshold` (String, Optional): One of `Low`, `Medium`, `High`, `Critical`; lowercase is also accepted.

When any repository security setting is configured, all five security settings must be set.

## Computed Attributes

- `id`: Repository UUID.
- `created_at`: Creation timestamp.
- `updated_at`: Update timestamp.
- `size_bytes`: Storage usage in bytes.
- `upstream_auth_configured`: Whether upstream auth is configured.

## Import

Import by repository key:

```bash
terraform import artifactkeeper_repository.docker_local docker-local
```

## Notes

Remote upstream settings are create-time settings in this provider and require replacement when changed. Artifact count is intentionally omitted because repository responses do not expose it.
