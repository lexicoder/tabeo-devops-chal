module "vpc" {
  source     = "./modules/vpc"
  name       = "tabeo-devops-chal"
  project_id = var.project_id
}

module "gke_cluster" {
  source             = "./modules/gke"
  project_id         = var.project_id
  name               = "tabeo-devops-chal"
  region             = var.region
  network            = module.vpc.id
  instance_type      = "e2-micro"
  preemptible        = true
  kubernetes_version = "1.23"
  network_cidr       = module.vpc.cidr
}
