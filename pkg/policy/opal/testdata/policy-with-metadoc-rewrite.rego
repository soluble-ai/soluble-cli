# comments
package policies.c_opl_test_policy_terraform

import data.d1.d2 as x

__rego__metadoc__ := {
  "custom": { "severity": "High" },
  "id": "c-opl-test-policy"
}

resource_type := "aws_security_group"

default deny = false

deny {
  policy = input.ingress[_]
}