terraform {
  backend "gcs" {
    bucket = "tabeo-devops-chal-terraform"
    prefix = "state"
  }
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    google = {
      source = "hashicorp/google"
    }
  }
}
