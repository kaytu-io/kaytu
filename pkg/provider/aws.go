package provider

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstype "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWS struct {
	cfg aws.Config
}

func NewAWS(cfg aws.Config) (*AWS, error) {
	return &AWS{cfg: cfg}, nil
}

func (s *AWS) Identify() (map[string]string, error) {
	client := sts.NewFromConfig(s.cfg)
	out, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, errors.New("unable to retrieve AWS account details, please check your AWS cli and ensure that you are logged-in")
	}

	orgClient := organizations.NewFromConfig(s.cfg)
	orgOut, _ := orgClient.DescribeOrganization(context.Background(), &organizations.DescribeOrganizationInput{})

	identification := map[string]string{}
	identification["account"] = *out.Account
	identification["user_id"] = *out.UserId
	identification["sts_arn"] = *out.Arn

	if orgOut != nil && orgOut.Organization != nil {
		identification["org_id"] = *orgOut.Organization.Id
		identification["org_m_arn"] = *orgOut.Organization.MasterAccountArn
		identification["org_m_email"] = *orgOut.Organization.MasterAccountEmail
		identification["org_m_account"] = *orgOut.Organization.MasterAccountId
	}
	return identification, nil
}

func (s *AWS) ListAllRegions() ([]string, error) {
	regionClient := ec2.NewFromConfig(s.cfg)
	regions, err := regionClient.DescribeRegions(context.Background(), &ec2.DescribeRegionsInput{AllRegions: aws.Bool(false)})
	if err != nil {
		return nil, err
	}
	var regionCodes []string
	for _, r := range regions.Regions {
		regionCodes = append(regionCodes, *r.RegionName)
	}
	return regionCodes, nil
}

func (s *AWS) ListInstances(region string) ([]types.Instance, error) {
	localCfg := s.cfg
	localCfg.Region = region

	var vms []types.Instance
	ctx := context.Background()
	client := ec2.NewFromConfig(localCfg)
	paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, r := range page.Reservations {
			for _, v := range r.Instances {
				vms = append(vms, v)
			}
		}
	}
	return vms, nil
}

func (s *AWS) ListAttachedVolumes(region string, instance types.Instance) ([]types.Volume, error) {
	localCfg := s.cfg
	localCfg.Region = region

	var volumeIDs []string
	for _, v := range instance.BlockDeviceMappings {
		if v.Ebs != nil {
			volumeIDs = append(volumeIDs, *v.Ebs.VolumeId)
		}
	}

	client := ec2.NewFromConfig(localCfg)
	volumesResp, err := client.DescribeVolumes(context.Background(), &ec2.DescribeVolumesInput{
		VolumeIds: volumeIDs,
	})
	if err != nil {
		return nil, err
	}

	return volumesResp.Volumes, nil
}

func (s *AWS) ListRDSInstance(region string) ([]rdstype.DBInstance, error) {
	localCfg := s.cfg
	localCfg.Region = region

	var dbs []rdstype.DBInstance
	ctx := context.Background()
	client := rds.NewFromConfig(localCfg)
	paginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, r := range page.DBInstances {
			dbs = append(dbs, r)
		}
	}
	return dbs, nil
}
