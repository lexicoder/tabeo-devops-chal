data "google_client_config" "provider" {}

resource "kubernetes_namespace_v1" "monitoring" {
  metadata {
    annotations = {
      name = "monitoring"
    }
  }
}
