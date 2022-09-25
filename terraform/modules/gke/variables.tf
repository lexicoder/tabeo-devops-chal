variable "project_id" {
  type        = string
  description = "The project ID in which to create the cluster"
}

variable "name" {
  type        = string
  description = "The name of the cluster"
}

variable "description" {
  type        = string
  description = "The description of the cluster"
  default     = ""
}

variable "region" {
  type        = string
  description = "Region in which to host the cluster"
  default     = null
}

variable "network" {
  type        = string
  description = "The VPC network in which to host the cluster"
}

variable "network_cidr" {}

variable "kubernetes_version" {
  type        = string
  description = "The Kubernetes version of the masters. If set to 'latest' it will pull latest available version in the selected region."
  default     = "latest"
}

variable "instance_type" {
  type        = string
  description = "The instance type to be used for the cluster's node pool"
  default     = "e2-medium"
}

variable "node_pool_count" {
  type        = number
  description = "Number of nodes to create in this cluster's node pool."
  default     = 1
}

variable "service_account" {
  type        = string
  description = "The service account to run nodes"
  default     = null
}

variable "preemptible" {
  type        = bool
  description = "Use preemptible nodes in node pool?"
  default     = false
}

variable "cluster_autoscaling" {
  type = object({
    enabled       = bool
    min_cpu_cores = number
    max_cpu_cores = number
    min_memory_gb = number
    max_memory_gb = number
    gpu_resources = list(object({ resource_type = string, minimum = number, maximum = number }))
  })
  default = {
    enabled       = false
    max_cpu_cores = 0
    min_cpu_cores = 0
    max_memory_gb = 0
    min_memory_gb = 0
    gpu_resources = []
  }
  description = "Cluster autoscaling configuration. See [more details](https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters#clusterautoscaling)"
}

variable "grant_registry_access" {
  type        = bool
  description = "Grants created cluster-specific service account storage.objectViewer and artifactregistry.reader roles."
  default     = false
}
