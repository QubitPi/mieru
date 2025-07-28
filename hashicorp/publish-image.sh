#!/bin/bash
set -x
set -e

# Copyright 2025 Paion Data. All rights reserved.

cd images
packer init .
packer validate .
packer build .
cd ../