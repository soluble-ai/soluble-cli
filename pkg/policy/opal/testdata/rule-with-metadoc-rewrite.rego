# comments
package rules.c_opl_test_rule

import data.d1.d2 as x

__rego__metadoc__ := {
  "custom": { "severity": "High" },
  "id": "c-opl-test-rule"
}

resource_type := "aws_security_group"

default deny = false

deny {
  rule = input.ingress[_]
}