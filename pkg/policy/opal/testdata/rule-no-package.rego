default allow = false

resource_type := "aws_ebs_volume"

allow {
    input.encrypted == true
}