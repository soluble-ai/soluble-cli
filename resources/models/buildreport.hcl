api_prefix = "/api/v1"

command "print_client" "build-report" {
  short  = "Display information about the assessments generated during the current CI build"
  method = "GET"
  path   = "org/{org}/assessments"
  options = [ "xcp_ci" ]
  parameter "searchType" {
    literal_value = "ci"
  }
  parameter "detail" {
    usage = "Display details about each assessment"
    boolean = "true"
  }
}
