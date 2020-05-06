api_prefix = "/api/v1"

command "group" "org" {
  short = "Organization commands"
  command "print_client" "list" {
    short  = "List your organizations"
    method = "GET"
    path   = "org"
    result {
      path    = ["organizations"]
      columns = ["displayName", "name", "orgId", "createTs", "isCurrent"]
    }
  }
  command "print_client" "update" {
    short  = "Update an organization"
    method = "PATCH"
    path   = "org/{org}"
    parameter "name" {
      usage = "Set the organizaiton name"
    }
    parameter "displayName" {
      usage = "Set the organization's displayName"
    }
  }
  command "print_client" "list-users" {
    short  = "List the users in an organization"
    method = "GET"
    path   = "org/{org}/users"
    result {
      path    = ["data"]
      columns = ["displayName", "userId", "orgId", "status", "role", "lastLoginTs", "createTs"]
    }
  }
}