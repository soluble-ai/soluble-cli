package policies.c_opl_test_no_package_terraform

__rego__metadoc__ := {
  "id": "c-opl-test-no-package"
}

default allow = false

resource_type := "aws_ebs_volume"

allow {
    input.encrypted == true
}