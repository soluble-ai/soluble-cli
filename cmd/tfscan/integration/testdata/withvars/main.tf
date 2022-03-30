terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "3.75.1"
    }
  }
}

provider "aws" {
}

variable "acl" {
    default = "public-read"
}

resource "aws_s3_bucket" "test" {
    bucket = "test"
    acl = var.acl
}
