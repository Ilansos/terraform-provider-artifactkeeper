# artifactkeeper_lifecycle_policy

Manages an Artifact Keeper lifecycle policy. Policies can be global or scoped to a repository.

## Example

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

```hcl
resource "artifactkeeper_lifecycle_policy" "global_unused" {
  name        = "Delete unused artifacts"
  enabled     = true
  policy_type = "no_downloads_days"
  config      = jsonencode({ days = 180 })
}
```

## API

Uses `/api/v1/admin/lifecycle` and `/api/v1/admin/lifecycle/{id}`.

## Arguments

- `repository_id` (String, Optional, Forces replacement): Repository UUID to scope the policy. Omit for a global policy.
- `name` (String, Required): Policy name.
- `description` (String, Optional): Policy description.
- `enabled` (Boolean, Optional, Defaults to `true`): Whether the policy is enabled.
- `policy_type` (String, Required, Forces replacement): One of `max_age_days`, `max_versions`, `no_downloads_days`, `tag_pattern_keep`, `tag_pattern_delete`, `size_quota_bytes`.
- `config` (String, Required): JSON object containing policy-specific configuration.
- `priority` (Number, Optional, Defaults to `0`): Policy execution priority.

`max_versions` and `size_quota_bytes` must be scoped to a repository.

## Config Examples

```hcl
config = jsonencode({ days = 90 })
config = jsonencode({ keep = 5 })
config = jsonencode({ pattern = "^(release|stable)-" })
config = jsonencode({ quota_bytes = 10737418240 })
```

## Computed Attributes

- `id`: Lifecycle policy UUID.
- `last_run_at`: Last execution timestamp.
- `last_run_items_removed`: Number of items removed by the last run.
- `created_at`: Creation timestamp.
- `updated_at`: Update timestamp.

## Import

Import by lifecycle policy UUID:

```bash
terraform import artifactkeeper_lifecycle_policy.docker_keep_recent 00000000-0000-0000-0000-000000000000
```

## Notes

The provider canonicalizes JSON for stable plans and preserves equivalent user formatting where possible.
