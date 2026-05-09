# artifactkeeper_user

Manages an Artifact Keeper user.

## Example

```hcl
resource "artifactkeeper_user" "alice" {
  username = "alice"
  email    = "alice@example.com"
  password = var.alice_initial_password
  role     = "user"
}
```

## API

Uses `/api/v1/users` and `/api/v1/users/{id}`. Updates use the API patch behavior for email and admin state.

## Arguments

- `username` (String, Required, Forces replacement): Login username.
- `email` (String, Required): User email address.
- `password` (String, Optional, Sensitive, Forces replacement): Initial password. Artifact Keeper does not return passwords.
- `role` (String, Optional, Defaults to `user`): Either `user` or `admin`; maps to API `is_admin`.

## Computed Attributes

- `id`: User UUID.
- `created_at`: Creation timestamp.

## Import

Import by user UUID:

```bash
terraform import artifactkeeper_user.alice 00000000-0000-0000-0000-000000000000
```

## Notes

Password changes are not managed after creation because the API does not read passwords back. Changing `password` recreates the user.
