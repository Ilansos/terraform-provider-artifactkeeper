terraform {
  required_providers {
    artifactkeeper = {
      source = "artifactkeeper/artifactkeeper"
    }
  }
}

variable "artifactkeeper_username" {
  type = string
}

variable "artifactkeeper_password" {
  type      = string
  sensitive = true
}

provider "artifactkeeper" {
  url      = "https://artifactkeeper.example.com"
  username = var.artifactkeeper_username
  password = var.artifactkeeper_password
}

