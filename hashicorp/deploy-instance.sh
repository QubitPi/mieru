#!/bin/bash
set -x
set -e

# Copyright 2025 Paion Data. All rights reserved.

cd instances
rm -rf .terraform terraform.tfstate .terraform.lock.hcl terraform.tfstate.backup
terraform init
terraform validate
terraform apply -auto-approve
rm -rf .terraform terraform.tfstate .terraform.lock.hcl terraform.tfstate.backup
cd ../
