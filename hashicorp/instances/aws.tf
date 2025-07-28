/**
 * Copyright 2025 Paion Data. All rights reserved.
 */
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.42.0"
    }
  }
  required_version = ">= 0.14.5"
}

provider "aws" {
  region = "us-west-1"
}

data "template_file" "ec2-init" {
  template = file("aws-tf-init.sh")
}

data "aws_ami" "latest-ami" {
  most_recent = true
  owners      = ["899075777617"]

  filter {
    name   = "name"
    values = ["mail.paion-data.dev"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_instance" "ec2" {
  ami           = data.aws_ami.latest-ami.id
  instance_type = "t2.small"
  tags = {
    Name = "Paion Data Mail Server"
  }

  key_name = "testKey"
  security_groups = ["Paion Data Mail Server", "testKey SSH"]

  user_data = data.template_file.ec2-init.rendered
}

resource "aws_route53_record" "ec2" {
  zone_id         = "Z02600613NNEBWDLMOCCJ"
  name            = "mail.paion-data.dev"
  type            = "A"
  ttl             = 300
  records         = [aws_instance.ec2.public_ip]
  allow_overwrite = true
}
