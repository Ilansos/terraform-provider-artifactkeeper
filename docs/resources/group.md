# artifactkeeper_group

Manages an Artifact Keeper group and reconciles its membership.

## Example

```hcl
resource "artifactkeeper_group" "developers" {
  name        = "developers"
  description = "Developer access"
  user_ids    = [artifactkeeper_user.alice.id]
}
```

## API

Uses `/api/v1/groups`, `/api/v1/groups/{id}`, and `/api/v1/groups/{id}/members`.

## Arguments

- `name` (String, Required): Group name.
- `description` (String, Optional): Group description.
- `user_ids` (Set of String, Optional): User UUIDs that should be group members.

## Computed Attributes

- `id`: Group UUID.
- `created_at`: Creation timestamp.
- `updated_at`: Update timestamp.
- `member_count`: Number of group members reported by the API.

## Import

Import by group UUID:

```bash
terraform import artifactkeeper_group.developers 00000000-0000-0000-0000-000000000000
```

## Notes

Membership is reconciled by adding and removing users through the group membership endpoints.
