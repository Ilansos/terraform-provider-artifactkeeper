# artifactkeeper_license_policy

Manages an Artifact Keeper SBOM license compliance policy.

## Example

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
```

```hcl
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

## API

Uses `/api/v1/sbom/license-policies` and `/api/v1/sbom/license-policies/{id}`.

## Arguments

- `repository_id` (String, Optional, Forces replacement): Repository UUID to scope this policy. Omit for a global policy.
- `name` (String, Required, Forces replacement): License policy name. The API upserts by name/scope.
- `description` (String, Optional): License policy description.
- `allowed_licenses` (Set of String, Required): SPDX license identifiers explicitly allowed.
- `denied_licenses` (Set of String, Required): SPDX license identifiers explicitly denied.
- `allow_unknown` (Boolean, Optional, Defaults to `false`): Allows unknown or unrecognized licenses.
- `action` (String, Optional, Defaults to `block`): One of `block`, `warn`, or `allow`.
- `enabled` (Boolean, Optional, Defaults to `true`): Maps to API `is_enabled`.

Use an empty `allowed_licenses` set for deny-list style policies. Use an empty `denied_licenses` set for allow-list style policies.

## Computed Attributes

- `id`: License policy UUID.
- `created_at`: Creation timestamp.
- `updated_at`: Update timestamp when returned by the API.

## Import

Import by license policy UUID:

```bash
terraform import artifactkeeper_license_policy.no_copyleft 00000000-0000-0000-0000-000000000000
```

## Notes

Updates use the API upsert endpoint. Changing `name` or `repository_id` replaces the Terraform resource to avoid accidentally creating a second policy under the new scope.
