data "google_container_engine_versions" "this" {
  count          = var.kubernetes_version != "latest" ? 1 : 0
  location       = var.region
  version_prefix = "${var.kubernetes_version}."
}

resource "google_service_account" "this" {
  count      = can(var.service_account) ? 0 : 1
  account_id = var.name
}

#tfsec:ignore:google-gke-enforce-pod-security-policy
#tfsec:ignore:google-gke-enable-private-cluster
#tfsec:ignore:google-gke-enable-master-networks
resource "google_container_cluster" "this" {
  name                     = var.name
  location                 = var.region
  remove_default_node_pool = true
  initial_node_count       = 1
  network                  = try(var.network, null)
  min_master_version       = try(data.google_container_engine_versions.this[0].latest_master_version, null)
  ip_allocation_policy {}
  resource_labels = {
    "application" = "app"
  }
  cluster_autoscaling {
    enabled = true
    resource_limits {
      resource_type = "cpu"
      maximum       = 4
    }
    resource_limits {
      resource_type = "memory"
      maximum       = 64
    }
  }
  network_policy {
    enabled = true
  }
  #tfsec:ignore:google-gke-node-metadata-security
  node_config {
    disk_size_gb    = 10
    local_ssd_count = 0
    metadata = {
      disable-legacy-endpoints = true
    }
  }
  lifecycle {
    ignore_changes = [
      node_config
    ]
  }
}

# resource "google_container_node_pool" "this" {
#   name       = var.name
#   location   = var.region
#   cluster    = google_container_cluster.this.name
#   node_count = 1
#   #tfsec:ignore:google-gke-node-metadata-security
#   node_config {
#     preemptible     = var.preemptible
#     disk_size_gb    = 10
#     local_ssd_count = 0
#     machine_type    = var.instance_type
#     service_account = can(var.service_account) ? var.service_account : google_service_account.this[0].account_id
#     image_type      = "COS_CONTAINERD"
#     oauth_scopes = [
#       "https://www.googleapis.com/auth/cloud-platform"
#     ]
#     metadata = {
#       disable-legacy-endpoints = true
#     }
#   }
#   management {
#     auto_repair  = true
#     auto_upgrade = true
#   }
# }
