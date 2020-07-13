api_prefix = "/api/v1"

command "group" "cinderella" {
  short = "Manage temporary k3s clusters"
  command "print_cluster" "create" {
    short  = "Create a new temporary k3s cluster"
    method = "POST"
    path   = "org/{org}/cinderella/k3s"
  }
  command "print_client" "list" {
    short  = "List cinderella clusters"
    path   = "org/{org}/cinderella/requests"
    method = "GET"
    result {
      path = [ "data" ]
      columns = [
        "requestId", "ipAddress", "status", "userDisplayName", "createTs", "expirationTs", "message",
      ]
      formatters = {
        "expirationTs": "relative_ts"
      }
    }
  }
}