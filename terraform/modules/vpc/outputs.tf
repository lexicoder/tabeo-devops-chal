output "id" {
  value = google_compute_network.this.id
}

output "cidr" {
  value = data.google_compute_subnetwork.this.ip_cidr_range
}
