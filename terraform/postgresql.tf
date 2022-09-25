resource "google_compute_global_address" "main" {
  name          = "private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = module.vpc.id
}

resource "google_service_networking_connection" "main" {
  network                 = module.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.main.name]
}

resource "google_sql_database" "main" {
  name     = var.postgres_db
  instance = google_sql_database_instance.main.name
}

resource "google_sql_database_instance" "main" {
  name                = var.postgres_db
  region              = var.region
  database_version    = "POSTGRES_14"
  deletion_protection = "false"
  depends_on          = [google_service_networking_connection.main]
  settings {
    tier = "db-f1-micro"
    #tfsec:ignore:google-sql-no-public-access
    #tfsec:ignore:google-sql-encrypt-in-transit-data
    ip_configuration {
      private_network = module.vpc.id
    }
    backup_configuration {
      enabled = true
    }
    database_flags {
      name  = "log_temp_files"
      value = "0"
    }
    database_flags {
      name  = "log_checkpoints"
      value = "on"
    }
    database_flags {
      name  = "log_connections"
      value = "on"
    }
    database_flags {
      name  = "log_disconnections"
      value = "on"
    }
    database_flags {
      name  = "log_lock_waits"
      value = "on"
    }
  }
}

resource "google_sql_user" "users" {
  name     = var.postgres_user
  instance = google_sql_database_instance.main.name
  password = var.postgres_password
}
