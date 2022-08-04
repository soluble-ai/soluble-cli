package rules.c_opl_test_no_package

__rego__metadoc__ := {
  "id": "c-opl-test-no-package"
}

default allow = false

resource_type := "aws_ebs_volume"

allow {
    input.encrypted == true
}