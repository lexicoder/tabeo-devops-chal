
variable "project_id" {
  type        = string
  description = "The project ID in which to create the network"
}

variable "name" {
  type        = string
  description = "The name of the network"
}

variable "description" {
  type        = string
  description = "The description of the network"
  default     = null
}
