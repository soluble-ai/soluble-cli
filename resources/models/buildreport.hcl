api_prefix = "/api/v1"

command "print_client" "build-report" {
  short  = "Display information about the assessments generated during the current CI build"
  method = "GET"
  path   = "org/{org}/assessments/build-report"
  options = [ "xcp_ci" ]
  parameter "detail" {
    literal_value = "true"
  }
  result {
    path = [ "findings" ]
    columns = [
      "module", "pass", "severity", "sid", "description"
    ]
  }
}
