api_prefix = "/api/v1"

command "group" "auth" {
  short = ""
  command "print_client" "list-ssh-keys" {
    short  = "List ssh keys"
    method = "GET"
    path   = "users/profile/ssh-keys"
    result {
      path         = []
      columns      = ["id", "description"]
      wide_columns = ["key"]
    }
  }
  command "print_client" "add-ssh-key" {
    short  = "Add an ssh key"
    path   = "users/profile/ssh-keys"
    method = "POST"
    parameter "githubUser" {
      usage = "Add ssh keys from a github.com user"
    }
    parameter "key" {
      usage = "Add an ssh key"
    }
  }
  command "print_client" "delete-ssh-key" {
    short  = "Delete an ssh key"
    path   = "users/profile/ssh-keys/{id}"
    method = "DELETE"
    parameter "id" {
      usage       = "The ssh key ID to delete"
      required    = true
      disposition = "context"
    }
  }
}