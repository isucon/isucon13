package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Data struct {
	ExpectedAMIID string
	ExpectedAZID  string

	InstanceVPCID string

	DescribeInstances         []*ec2.DescribeInstancesOutput
	DescribeVolumes           []*ec2.DescribeVolumesOutput
	DescribeNetworkInterfaces []*ec2.DescribeNetworkInterfacesOutput

	DescribeAvailabilityZones *ec2.DescribeAvailabilityZonesOutput
}
