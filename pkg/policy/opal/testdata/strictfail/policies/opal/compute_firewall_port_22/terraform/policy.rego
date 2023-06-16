package policies.tf_google_compute_firewall_port_22

import data.google.compute.firewall_library as lib
import data.lacework


firewalls = lacework.resources("google_compute_firewall")

resource_type := "MULTIPLE"

port = "22"

policy[j] {
  firewall = firewalls[_]
  network = lib.network_for_firewall(firewall)
  lib.is_network_vulnerable(network, port)
  j = lacework.allow_resource(firewall)
} {
  firewall = firewalls[_]
  network = lib.network_for_firewall(firewall)
  not lib.is_network_vulnerable(network, port)
  p = lib.lowest_allow_ingress_zero_cidr_by_port(network, port)
  f = lib.firewalls_by_priority_and_port(network, p, port)[_]
  j = lacework.deny_resource(f)
}
