package policies.foo

default allow = false

resource_type := "aws_ebs_volume"

input_type := "tf"

allow {
    input.encrypted == true
}