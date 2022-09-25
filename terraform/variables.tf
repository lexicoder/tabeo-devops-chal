variable "project_id" {
  type        = string
  description = "The project ID in which to create the cluster"
  default     = "tabeo-devops-chal"
}

variable "region" {
  type        = string
  description = "Region in which to host the cluster"
  default     = "europe-west3"
}
