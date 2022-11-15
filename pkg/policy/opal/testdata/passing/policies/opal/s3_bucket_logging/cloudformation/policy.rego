
input_type := "cfn"

resource_type = "AWS::S3::Bucket"

default allow = false

allow {
  input.LoggingConfiguration.DestinationBucketName = "acme-corp-logs"
}