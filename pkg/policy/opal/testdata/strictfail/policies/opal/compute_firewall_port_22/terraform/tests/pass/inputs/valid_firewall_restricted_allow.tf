provider google { alias = "google6" }

resource "google_compute_network" "test6" {
  name = "test-network6"
}

resource "google_compute_firewall" "restricted_allow" {
  name = "restricted-allow"
  network = google_compute_network.test6.name
  source_ranges = [ "98.233.183.221/32" ]

  allow {
    protocol = "tcp"
    ports = [ "22", "3389" ]
  }
}
