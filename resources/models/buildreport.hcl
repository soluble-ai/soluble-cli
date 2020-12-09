api_prefix = "/api/v1"

command "group" "iac-scan" {
  command "print_client" "build-report" {
    short   = "Display information about the assessments generated during the current CI build"
    method  = "GET"
    path    = "org/{org}/assessments/build-report"
    options = ["xcp_ci"]
    parameter "detail" {
      literal_value = "true"
    }
    parameter "fail" {
      usage = "Exit with failure if there are more than `level=count` failed findings"
      disposition = "noop"
      map = true
    }
    result {
      local_action = "exit_on_failures"
      path = ["findings"]
      columns = [
        "module", "pass", "severity", "sid", "description"
      ]
    }
  }
}
