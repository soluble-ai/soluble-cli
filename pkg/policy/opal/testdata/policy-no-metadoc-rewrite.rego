package policies.c_opl_test_policy

__rego__metadoc__ := {
  "description": "This is a great \"description\"",
  "id": "c-opl-test-policy",
  "title": "This is a \"great\" example"
}

default allow = false

resource_type := "aws_ebs_volume"

input_type := "tf"

allow {
    input.encrypted == true
}