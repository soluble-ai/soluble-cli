provider google {}

resource "google_compute_network" "test" {
  name = "test-network"
}

resource "google_compute_firewall" "allow_all" {
  name = "allow-all"
  network = google_compute_network.test.name
  source_ranges = [ "0.0.0.0/0" ]

  allow {
    protocol = "tcp"
    ports = [ "23-30", "3380-3388", "3390-3399" ]
  }
}
