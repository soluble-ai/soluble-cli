api_prefix = "/api/v1"

command "group" "agent" {
  short = ""
  command "print_cluster" "list-tokens" {
    short  = "List issued agent tokens"
    method = "GET"
    path   = "org/{org}/agent-tokens"
    result {
      path = ["data"]
      columns = [
        "id", "status", "expirationTs", "issuedTs", "lastUseTs",
      ]
      formatters = {
        lastUseTs : "relative_ts"
        expirationTs : "relative_ts"
      }
      sort_by = [
        "-lastUseTs"
      ]
    }
  }
  command "print_client" "get" {
    short  = "Show details of an agent instance"
    path   = "org/{org}/agents/{agentID}"
    method = "GET"
    parameter "agentID" {
      usage       = "The agent instance ID"
      required    = true
      disposition = "context"
    }
  }
  command "print_cluster" "restart" {
    short  = "Request an agent to restart"
    path   = "org/{org}/clusters/{clusterID}/restart"
    method = "POST"
  }
  command "print_cluster" "update" {
    short  = "Update the version of an agent"
    path   = "org/{org}/clusters/{clusterID}/update-agent"
    method = "POST"
    parameter "tag" {
      usage = "The specific version to upgrade to.  If unset, stable will be used."
    }
    parameter "repository" {
      usage = "The base repository to upgrade to"
    }
  }
  command "print_cluster" "list" {
    short               = "List agents"
    cluster_id_optional = true
    path                = "org/{org}/agents"
    method              = "GET"
    parameter "clusterId" {
      context_value = "clusterID"
    }
    result {
      path = ["agents"]
      columns = [
        "agentInstanceId", "clusterId", "agentVersion", "agentStartTs", "updateTs",
        "agentRemoteIp", "memAlloc", "message"
      ]
      formatters = {
        "memAlloc" : "bytes"
        "updateTs" : "relative_ts"
      }
      computed_columns = {
        message : "first_message"
      }
    }
  }
}