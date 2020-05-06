# Running Queries


`query list` returns a list of the available queries.  For example:

```
% soluble query list
[ Info] Using profile demo - https://api.demo.soluble.cloud
[ Info] GET https://api.demo.soluble.cloud/api/v1/org/900000000000/queries returned 200
NAME          DESCRIPTION
image         Images in use across your clusters
kube-resource Query kubernetes resources (Pod, Deployment, etc.) across clusters
pod-image     no description
```

`query list-parameters` lists the parameters of a query:

```
% soluble query list-parameters --query-name kube-resource [ Info] Using profile demo - https://api.demo.soluble.cloud
[ Info] GET https://api.demo.soluble.cloud/api/v1/org/900000000000/queries returned 200
NAME      REQUIRED DESCRIPTION
kind      true     kubernetes resource type (Pod, Deployment, etc.)
name      false    kubernetes resource name
namespace false    kubernetes namespace
clusterId false    cluster identifier
urn       false    globally unique name
uid       false    cluster unique identifier
```

`query run` runs a query.  For example:

```
% soluble query run --query-name kube-resource -p kind=Deployment -p name=soluble-agent
[ Info] Using profile demo - https://api.demo.soluble.cloud
[ Info] GET https://api.demo.soluble.cloud/api/v1/org/900000000000/queries/kube-resource?kind=Deployment&name=soluble-agent returned 200
KIND       NAME          NAMESPACE    CLUSTER-ID                      CREATION-TIMESTAMP        UPDATE-TS
Deployment soluble-agent soluble      900000000000:c-2e73d1d7ff145159 2020-03-20T08:48:55-07:00 43h23m21s
Deployment soluble-agent soluble-demo c-2e73d1d7ff454a1a              2020-04-07T20:24:04-07:00 291h19m35s
Deployment soluble-agent soluble-demo c-2e73d1d7ff487162              2020-04-03T18:47:14-07:00 343h49m38s
Deployment soluble-agent soluble-demo c-2e73d1d7ff4ca2ef              2020-04-14T09:28:12-07:00 311h15m31s
Deployment soluble-agent soluble      c-2e73d1d7ff4ca2ef              2020-04-15T08:20:39-07:00 296h4m6s
Deployment soluble-agent soluble-demo c-2e73d1d7ff52d6c5              2020-04-16T00:42:40-07:00 279h41m43s
Deployment soluble-agent soluble-demo c-2e73d1d7ff5e6858              2020-04-14T16:29:02-07:00 18m22s
Deployment soluble-agent soluble      c-2e73d1d7ff5e6858              2020-04-14T16:27:58-07:00 18m23s
```

Only some columns are displayed.  Add the `--wide` flag to display all of them.  Use `--full` to display the results in YAML format.
