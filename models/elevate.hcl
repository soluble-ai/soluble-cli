api_prefix = "/api/v1"

command "group" "elevate" {
  short = "Request temporary elevated access to kubernetes clusters"
  command "print_cluster" "list" {
    short  = "List elevated sessions"
    method = "GET"
    cluster_id_optional = true
    path   = "org/{org}/elevate"
    parameter "clusterId" {
        context_value = "clusterID"
    }
    result {
        path = [ "data" ]
        columns = [
            "requestId", "state", "solubleUser", "roleKind", "roleName", "namespace",
            "subjectName", "durationMinutes", "createTs", "sessionName"
        ]
        wide_columns = [ "subjectName", "roleKind", "durationMinutes" ]
    }
  }
  command "print_cluster" "request" {
    short  = "Request elevated credentials in a cluster"
    path   = "org/{org}/cluster/{clusterID}/elevate"
    method = "POST"
    parameter "subjectName" {
        context_value = "kubernetes_user"
    }
    parameter "subjectKind" {
        literal_value = "User"
    }
    parameter "namespace" {
        usage = "The specific namespace to request access to"
    }
    parameter "role" {
        usage = "The role to request access to.  By default either 'admin' for namespaces, or 'cluster-admin' for cluster-wide"
    }
    parameter "durationMintues" {
        usage = "The duration of the elevated session, in minutes."
    }
    parameter "sessionName" {
        usage = "The name of the elevated session"
        required = true
    }
  }
}