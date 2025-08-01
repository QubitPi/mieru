# Copyright 2025 Jiaqi Liu. All rights reserved.
variable "aws_region" {
  type        = string
  description = "The AWS region to build the AMI in."
}

variable "instance_type" {
  type        = string
  description = "The EC2 instance type to use for building the AMI."
}

variable "ami_name_prefix" {
  type        = string
  description = "Prefix for the name of the generated AMI."
}

variable "ubuntu_version" {
  type        = string
  default     = "24.04"
  description = "The Ubuntu OS version (e.g., '20.04', '22.04', '24.04'). Defaults to '24.04'"
}

variable "ubuntu_codename" {
  type        = string
  default     = "noble"
  description = "The corresponding Ubuntu OS codename for ubuntu_version (e.g., 'focal' for 20.04, 'jammy' for 22.04, 'noble' for 24.04). Defaults to 'noble'"
}

variable "mita_port_range" {
  type        = string
  description = "The portRange value of portBindings field inside mita JSON config file. The ranged value must be from 1025 to 65535. For example, '46234-46236'"
}

variable "mita_user_name" {
  type        = string
  description = "The username for connection authentication from mieru client to mita server"
}

variable "mita_user_password" {
  type        = string
  description = "The password for connection authentication from mieru client to mita server"
  sensitive   = true
}

source "amazon-ebs" "mita_server" {
  region        = var.aws_region
  instance_type = var.instance_type
  ami_name      = "${var.ami_name_prefix}-{{timestamp}}"
  ssh_username  = "ubuntu"

  source_ami_filter {
    filters = {
      name                = "ubuntu/images/hvm-ssd/ubuntu-${var.ubuntu_codename}-${var.ubuntu_version}-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    owners      = ["099720109477"] # Canonical's AWS account ID for Ubuntu AMIs
    most_recent = true
  }

  tags = {
    Name          = "${var.ami_name_prefix}"
    Description   = "AWS AMI with Mita (Mieru) SOCKS5 proxy server installed and configured."
    Project       = "MitaProxy"
    UbuntuVersion = var.ubuntu_version
  }
}

build {
  name = "mita-server-build"
  sources = [
    "source.amazon-ebs.mita_server"
  ]

  provisioner "shell" {
    inline = [
      "sudo apt update -y",
      "sudo apt install -y unzip curl",
    ]
  }

  provisioner "ansible" {
    playbook_file = "./ansible/playbook.yml"
    extra_arguments = [
      "--extra-vars",
      "mita_port_range=${var.mita_port_range}",
      "--extra-vars",
      "mita_user_name=${var.mita_user_name}",
      "--extra-vars",
      "mita_user_password=${var.mita_user_password}"
    ]
    ansible_folder = "./ansible"
  }
}
