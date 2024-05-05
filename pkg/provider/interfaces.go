package provider

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	rds "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

type Provider interface {
	Identify() (map[string]string, error)
	ListAllRegions() ([]string, error)

	ListInstances(region string) ([]types.Instance, error)
	ListAttachedVolumes(region string, instance types.Instance) ([]types.Volume, error)

	ListRDSInstance(region string) ([]rds.DBInstance, error)
}