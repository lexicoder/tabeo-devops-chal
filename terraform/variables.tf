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

variable "postgres_db" {
  type        = string
  description = "Name for postgres db"
}

variable "postgres_user" {
  type        = string
  description = "Username for postgres db"
}

variable "postgres_password" {
  type        = string
  description = "Password for postgres db"
  sensitive   = true
}
