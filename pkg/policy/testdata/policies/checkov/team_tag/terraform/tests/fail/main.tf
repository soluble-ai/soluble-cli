resource "aws_s3_bucket" "test" {
    tags = {
        Project = "Bucket"
    }
}