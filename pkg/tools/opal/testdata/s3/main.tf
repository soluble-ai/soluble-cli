resource "aws_s3_bucket" "test" {
    bucket_prefix = "test"
    aws_s3_bucket_acl = "public-read"
}