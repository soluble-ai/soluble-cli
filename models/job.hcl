api_prefix = "/api/v1"

command "group" "job" {
  short = "Manage cluster jobs"
  command "print_client" "list-templates" {
    short  = "List job templates"
    method = "GET"
    path   = "org/{org}/job-templates"
    result {
      path    = ["templates"]
      columns = ["name", "description"]
    }
  }
  command "print_client" "list" {
    short  = "List jobs"
    method = "GET"
    path   = "org/{org}/job"
    parameter "limit" {
      default_value = 10
      usage         = "Return no more than this number of jobs"
    }
    result {
      path    = ["jobs"]
      columns = ["clusterId", "jobId", "jobName", "createTs", "createTs+", "updateTs+", "status", ]
    }
  }
  command "print_cluster" "list-cluster" {
    short  = "List jobs on a particular cluster"
    method = "GET"
    path   = "org/{org}/clusters/{clusterID}/job"
    parameter "limit" {
      default_value = 10
      usage         = "Return no more than this number of jobs"
    }
    result {
      path    = ["jobs"]
      columns = ["jobId", "jobName", "createTs", "createTs+", "updateTs+", "status", ]
    }
  }
  command "print_client" "get" {
    short  = "Get details of a single job"
    method = "GET"
    path   = "org/{org}/job/{jobID}"
    parameter "jobID" {
      usage       = "The job id"
      disposition = "context"
    }
  }
  command "print_client" "cancel" {
    short  = "Attempt to cancel a running job"
    method = "DELETE"
    path   = "org/{org}/job/{jobID}"
    parameter "jobID" {
      usage       = "The job id"
      disposition = "context"
    }
  }
  command "print_cluster" "start" {
    short  = "Start a new job"
    method = "POST"
    path   = "org/{org}/clusters/{clusterID}/job"
    parameter "name" {
      usage = "The name of the job template (see job list-templates)"
    }
  }
}