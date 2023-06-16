provider google { alias = "google6" }

resource "google_compute_network" "test6" {
  name = "test-network6"
}

resource "google_compute_firewall" "deny_all" {
  name          = "deny-all"
  network       = google_compute_network.test6.name
  source_ranges = ["0.0.0.0/0"]

  deny {
    protocol = "all"
  }
}

resource "google_compute_firewall" "allow_ports" {
  name          = "allow-ports"
  network       = google_compute_network.test6.name
  source_ranges = ["0.0.0.0/0"]
  priority      = 1

  allow {
    protocol = "tcp"
    ports    = ["22", "3389"]
  }
}
