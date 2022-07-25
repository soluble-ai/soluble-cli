resource "aws_s3_bucket" "test" {
    bucket_prefix = "test"  
    acl = "public-read" 
}