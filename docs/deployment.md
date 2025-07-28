Deployment
==========

HashiCorp
---------

Opinionated assumptions:

- [AWS instance type](https://aws.amazon.com/ec2/instance-types/) is set to `t2.small`

![Packer + Terraform "Error loading hashicorp.png from QubitPi repo"](https://github.com/QubitPi/QubitPi/blob/master/img/hashicorp.png?raw=true)

```console
git clone git@github.com:QubitPi/mieru.git
cd mieru/hashicorp
```

```console
export PKR_VAR_ec2_region=us-east-2
```

- `PKR_VAR_ec2_region`: is the [image region](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html#Concepts.RegionsAndAvailabilityZones.Availability) where mita [AMI](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) will be published to. The published image will be private
