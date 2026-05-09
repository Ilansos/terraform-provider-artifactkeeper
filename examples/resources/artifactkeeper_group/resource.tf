resource "artifactkeeper_group" "developers" {
  name        = "developers"
  description = "Developer group"
  user_ids    = [artifactkeeper_user.alice.id]
}

