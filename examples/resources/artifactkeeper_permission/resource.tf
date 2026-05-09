resource "artifactkeeper_permission" "developers_read" {
  repository_id = artifactkeeper_repository.docker_local.id
  group_id      = artifactkeeper_group.developers.id
  permissions   = ["read"]
}

