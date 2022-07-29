package rules.foo

__rego__metadoc__ := {
  "id": "c-opl-test-rule"
}

default allow = false

resource_type := "aws_ebs_volume"

input_type := "tf"

allow {
    input.encrypted == true
}