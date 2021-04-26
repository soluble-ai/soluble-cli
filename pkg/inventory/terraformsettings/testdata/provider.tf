terraform {
  required_version = "~> 0.12.6 "
  required_providers {
    google      = "< 3.1"
    google-beta = "> 3.1"
  }
}

provider "google" {
  # credentials = file("gcp.json")
  #  project = var.automation_project_id
  # region = var.region
}

provider "google-beta" {
  # credentials = file("gcp.json")
  # project = var.automation_project_id
  # region = var.region
}


