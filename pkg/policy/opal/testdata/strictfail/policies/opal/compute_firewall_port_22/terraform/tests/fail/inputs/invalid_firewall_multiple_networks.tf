# This test case has repeated resource names and opal will not be able to load it,=.
# This case is skipped by opal and does not appear in the results; unless ---strict-loading is
# set, in which case the test fails.

provider google { alias = "google9" }

resource "google_compute_network" "test1" {
  name = "test-network1"
}

resource "google_compute_firewall" "deny_all" {
  name          = "deny-all"
  network       = google_compute_network.test1.name
  source_ranges = ["0.0.0.0/0"]
  priority      = 1

  deny {
    protocol = "all"
  }
}

resource "google_compute_network" "test2" {
  name = "test-network2"
}

resource "google_compute_firewall" "allow_ports1" {
  name          = "allow-ports1"
  network       = google_compute_network.test2.name
  source_ranges = ["0.0.0.0/0"]
  priority      = 1000

  allow {
    protocol = "tcp"
    ports    = ["22", "3389"]
  }
}
