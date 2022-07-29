# comments
package rules.p1.p2

import data.d1.d2 as x

__rego__metadoc__ := {
  "custom": {
    "severity": "Medium"
  },
  "description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
  "id": "ID_12345",
  "title": "Excepteur sint occaecat cupidatat non proident"
}

resource_type := "aws_security_group"

default deny = false

deny {
  rule = input.ingress[_]
}