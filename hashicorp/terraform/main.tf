# Copyright 2025 Jiaqi Liu. All rights reserved.
provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "The AWS region to deploy the EC2 instance in."
  type        = string
}

variable "ami_name_prefix" {
  description = "The prefix of the AMI name created by Packer (e.g., 'mita-server-ami')."
  type        = string
  default     = "mita-server-ami" # Must match the prefix used in Packer
}

variable "instance_type" {
  description = "The EC2 instance type to deploy."
  type        = string
  default     = "t2.micro"
}

variable "key_pair_name" {
  description = "The name of an existing EC2 Key Pair to use for SSH access."
  type        = string
  # IMPORTANT: Replace 'your-ssh-key' with the actual name of your SSH key pair in AWS
  default     = "your-ssh-key"
}

variable "mita_proxy_port_range" {
  description = "The Mita proxy port range (e.g., '2012-2022'). This is for security group rules."
  type        = string
  default     = "2012-2022" # Must match the portRange configured in Mita
}

data "aws_ami" "mita_server_ami" {
  most_recent = true
  owners      = ["self"]

  filter {
    name   = "name"
    values = ["${var.ami_name_prefix}-*"]
  }

  filter {
    name   = "tag:Project"
    values = ["MitaProxy"]
  }
}

resource "aws_security_group" "mita_sg" {
  name        = "${var.ami_name_prefix}-sg"
  description = "Allow SSH and Mita proxy traffic"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # TODO: 0.0.0.0/0 allows access from anywhere. Restrict this in production!
    description = "Allow SSH access"
  }

  # Ingress rule for Mita proxy ports
  # This assumes the mita_proxy_port_range is in "START-END" format or a single port.
  # Terraform does not directly support ranges in from_port/to_port, so we parse it.
  dynamic "ingress" {
    for_each = split("-", var.mita_proxy_port_range)
    content {
      from_port   = tonumber(ingress.value)
      to_port     = length(split("-", var.mita_proxy_port_range)) == 2 ? tonumber(split("-", var.mita_proxy_port_range)[1]) : tonumber(ingress.value)
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"] # TODO: 0.0.0.0/0 allows access from anywhere. Restrict this in production!
      description = "Allow Mita Proxy TCP traffic on port ${ingress.value}"
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = {
    Name    = "${var.ami_name_prefix}-sg"
    Project = "MitaProxy"
  }
}

data "aws_vpc" "default" {
  default = true
}

resource "aws_instance" "mita_server_instance" {
  ami           = data.aws_ami.mita_server_ami.id
  instance_type = var.instance_type
  key_name      = var.key_pair_name
  vpc_security_group_ids = [aws_security_group.mita_sg.id]

  tags = {
    Name    = "${var.ami_name_prefix}-instance"
    Project = "MitaProxy"
  }
}

output "mita_server_public_ip" {
  description = "The public IP address of the Mita SOCKS5 proxy server instance."
  value       = aws_instance.mita_server_instance.public_ip
}

output "ami_id_used" {
  description = "The ID of the AMI used to launch the instance."
  value       = data.aws_ami.mita_server_ami.id
}
