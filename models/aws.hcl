api_prefix = "/api/v1"

command "group" "aws" {
  short = "Manage AWS integration"

  command "print_client" "list-accounts" {
    short  = "List AWS accounts currently configured"
    path   = "org/{organizationID}/aws/accounts"
    method = "GET"
    result {
      path = ["data"]
      columns = [
        "account",
        "name",
        "orgAccount",
        "lastScanSuccessTs+",
        "updateTs+"
      ]
      wide_columns = [
        "nameFromOrg",
        "orgAccountArn",
        "orgEmail", "orgJoinedMethod",
        "orgJoinedTs", "orgStatus",
        "lastScanEndTs",
        "lastScanStartTs",
      ]
    }
  }
  command "print_client" "check-access" {
    short  = "Verify AWS access"
    method = "GET"
    path   = "org/{organizationID}/aws/accounts/{account}/status"
    parameter "account" {
      usage       = "AWS account id"
      disposition = "context"
      required    = true
    }
  }
  command "print_client" "add-account" {
    short  = "Add an AWS account"
    method = "POST"
    path   = "org/{organizationID}/aws/accounts"
    parameter "account" {
      usage    = "AWS account id"
      required = true
    }
  }
  command "print_client" "remove-account" {
    short  = "Remove an AWS account"
    path   = "org/{organizationID}/aws/accounts/{account}"
    method = "DELETE"
    parameter "account" {
      usage       = "AWS account id"
      disposition = "context"
      required    = true
    }
  }
  command "print_client" "schedule-scan" {
    short  = "Schedule a scan of all AWS accounts"
    method = "POST"
    path   = "org/{organizationID}/aws/accounts/schedule-scan"
  }
}
