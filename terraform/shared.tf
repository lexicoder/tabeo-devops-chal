data "google_client_config" "provider" {}

resource "google_artifact_registry_repository" "registry" {
  location      = var.region
  repository_id = "tabeo-devops-chal"
  format        = "DOCKER"
}

resource "kubernetes_namespace_v1" "monitoring" {
  metadata {
    annotations = {
      name = "monitoring"
    }
  }
}
