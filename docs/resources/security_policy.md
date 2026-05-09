# artifactkeeper_security_policy

Manages an Artifact Keeper security policy using the OpenAPI-backed threshold policy endpoint.

## Example

```hcl
resource "artifactkeeper_security_policy" "strict_production" {
  name              = "strict-production"
  enabled           = true
  max_severity      = "high"
  block_unscanned   = true
  block_on_fail     = true
  require_signature = true
}
```

```hcl
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

## API

Uses `/api/v1/security/policies` and `/api/v1/security/policies/{id}`.

## Arguments

- `repository_id` (String, Optional, Forces replacement): Repository UUID to scope this policy. Omit for a global policy.
- `name` (String, Required): Policy name.
- `enabled` (Boolean, Optional, Defaults to `true`): Maps to API `is_enabled`.
- `max_severity` (String, Optional, Defaults to `high`): One of `low`, `medium`, `high`, `critical`.
- `block_unscanned` (Boolean, Optional, Defaults to `false`): Blocks artifacts without scan results.
- `block_on_fail` (Boolean, Optional, Defaults to `true`): Blocks when scanning or policy evaluation fails.
- `require_signature` (Boolean, Optional, Defaults to `false`): Requires a valid artifact signature.
- `max_artifact_age_days` (Number, Optional): Maximum allowed artifact age.
- `min_staging_hours` (Number, Optional): Minimum staging duration before passing policy.
- `description` (String, Optional, Deprecated): Kept in Terraform state only. The current API does not persist descriptions.
- `rules` (String, Optional, Deprecated): Kept in Terraform state only. The current API does not accept free-form rule JSON.

## Computed Attributes

- `id`: Security policy UUID.
- `created_at`: Creation timestamp.
- `updated_at`: Update timestamp.

## Import

Import by security policy UUID:

```bash
terraform import artifactkeeper_security_policy.strict_production 00000000-0000-0000-0000-000000000000
```

## Notes

The prose security policy documentation includes free-form rules and CVE exception examples, but the current OpenAPI endpoint exposes threshold-style fields only. CVE allowlisting is not managed by this resource.
