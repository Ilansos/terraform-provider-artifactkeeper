# artifactkeeper_permission

Manages repository permissions for a user or group.

## Example

```hcl
resource "artifactkeeper_permission" "developers_read" {
  repository_id = artifactkeeper_repository.docker_local.id
  group_id      = artifactkeeper_group.developers.id
  permissions   = ["read"]
}
```

## API

Uses `/api/v1/permissions` and `/api/v1/permissions/{id}`.

## Arguments

- `repository_id` (String, Required): Repository UUID. Maps to API `target_id` with `target_type = "repository"`.
- `user_id` (String, Optional): User UUID principal.
- `group_id` (String, Optional): Group UUID principal.
- `permissions` (Set of String, Required): Permission actions, for example `read`, `write`, or `admin`.

Set exactly one of `user_id` or `group_id`.

## Computed Attributes

- `id`: Permission UUID.
- `created_at`: Creation timestamp.
- `updated_at`: Update timestamp.

## Import

Import by permission UUID:

```bash
terraform import artifactkeeper_permission.developers_read 00000000-0000-0000-0000-000000000000
```
