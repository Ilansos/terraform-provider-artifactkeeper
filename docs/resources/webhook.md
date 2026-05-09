# artifactkeeper_webhook

Manages an Artifact Keeper webhook.

## Example

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

## API

Uses `/api/v1/webhooks`, `/api/v1/webhooks/{id}`, `/api/v1/webhooks/{id}/enable`, and `/api/v1/webhooks/{id}/disable`.

## Arguments

- `name` (String, Required, Forces replacement): Webhook name.
- `url` (String, Required, Forces replacement): Destination URL.
- `events` (Set of String, Required, Forces replacement): Event names delivered to the webhook.
- `repository_id` (String, Optional, Forces replacement): Repository UUID for repository-scoped webhooks.
- `secret` (String, Optional, Sensitive, Forces replacement): Webhook signing secret. The API does not return this value.
- `enabled` (Boolean, Optional, Defaults to `true`): Whether the webhook is enabled.

## Computed Attributes

- `id`: Webhook UUID.
- `created_at`: Creation timestamp.
- `last_triggered_at`: Last trigger timestamp when returned by the API.

## Import

Import by webhook UUID:

```bash
terraform import artifactkeeper_webhook.repo_events 00000000-0000-0000-0000-000000000000
```

## Notes

The current API has no general webhook update endpoint. Most changes replace the webhook. Terraform never calls the webhook test endpoint automatically.
