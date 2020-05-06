api_prefix = "/api/v1"

command "group" "cluster" {
  short = ""
  command "print_cluster" "get" {
    short  = "Get details of a single cluster"
    method = "GET"
    path   = "org/{org}/clusters/{clusterID}"
  }
  command "print_cluster" "set-name" {
    short = "Change the name of a cluster"
    method = "PATCH"
    path = "org/{org}/clusters/{clusterID}"
    parameter "displayName" {
      required = true
      usage = "The new display name for the cluster"
    }
  }
  command "print_cluster" "delete" {
    short = "Delete a cluster"
    path = "org/{org}/clusters/{clusterID}"
    method = "DELETE"
  }
}