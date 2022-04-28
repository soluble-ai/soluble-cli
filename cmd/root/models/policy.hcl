api_prefix = "/api/v1"

command "group" "early-access" {
  short = ""
  command "group" "policy" {
      short = ""
      command "print_client" "list" {
            short = "List organization custom policies"
            method = "GET"
            path   = "org/{org}/custom/policy/bundles"
            result {
                path = ["data"]
                columns = [
                    "isActive", "id", "gitRepo", "gitBranch", "gitCommit", "updateTs"
                ]
                formatters = {
                    "gitCommit": "commit"
                }
            }
      }
      command "print_client" "activate" {
          short = "Activate a custom policy bundle"
          method = "POST"
          path = "org/{org}/custom/policy/bundles/{id}/activate"
          parameter "id" {
              usage = "The bundle ID to activate"
              required = true
              disposition = "context"
          }
      }
  }
}