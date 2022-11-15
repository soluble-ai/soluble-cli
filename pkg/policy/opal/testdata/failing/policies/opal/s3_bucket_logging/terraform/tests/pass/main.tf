resource "aws_s3_bucket" "test" {
  bucket = "test-bucket"
  logging {
    target_bucket = "acme-corp-logs"
  }
}

resource "aws_s3_bucket" "test2" {
  bucket = "test-bucket"
  logging {
    target_bucket = "invalid"
  }
}
