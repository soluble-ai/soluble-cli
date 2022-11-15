package policies.acme

input_type = "tf"

resource_type = "aws_s3_bucket"

default allow = false

allow {
  input.logging[_].target_bucket = "acme-corp-logs"
}

