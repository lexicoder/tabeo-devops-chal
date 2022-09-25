resource "google_compute_network" "this" {
  project                 = var.project_id
  name                    = var.name
  description             = var.description
  auto_create_subnetworks = true
}

data "google_compute_subnetwork" "this" {
  name = google_compute_network.this.name
}
