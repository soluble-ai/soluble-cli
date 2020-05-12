api_prefix = "/api/v1"

command "group" "kubernetes" {
  short   = "Display kubernetes resources"
  aliases = ["k8s"]

  command "print_cluster" "get" {
    short  = "Display yaml of a kubernetes object"
    path   = "org/{organizationID}/kubernetes/resources/urn:kubernetes:{organizationID}:{clusterID}:{kind}:{namespace}:{name}/json"
    method = "GET"
    parameter "namespace" {
      usage         = "Namespace of the object"
      default_value = "default"
      disposition   = "context"
    }
    parameter "name" {
      usage       = "Name of the object"
      disposition = "context"
      required    = true
    }
    parameter "kind" {
      usage       = "The kind of object"
      disposition = "context"
      required    = true
    }
  }
  command "print_cluster" "list" {
    short               = "Display kubernetes resources"
    path                = "org/{org}/kubernetes/entities"
    method              = "GET"
    cluster_id_optional = true
    parameter "kind" {
      usage    = "The kind to list"
      required = true
    }
    result {
      path    = ["data"]
      columns = ["kind", "name", "namespace", "creationTimestamp", "clusterId", "clusterDisplayName", "updateTs+", ]
    }
  }
}