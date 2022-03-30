data "aws_ami" "example" {}
locals {
  foo = 1
}
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "4.8.0"
    }
  }
}

provider "aws" {
  # Configuration options
}
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.14.0"
  # insert the 23 required variables here
}
module "security-group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "4.9.0"
  # insert the 3 required variables here
}
resource "aws_security_group" "test1" {}
resource "aws_security_group" "test2" {}
resource "aws_instance" "test" {}
terraform {
  required_version = "~> 1.0.0"
  backend "s3" {}
}