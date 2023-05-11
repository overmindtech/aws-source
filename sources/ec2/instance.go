package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func instanceInputMapperGet(scope, query string) (*ec2.DescribeInstancesInput, error) {
	return &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			query,
		},
	}, nil
}

func instanceInputMapperList(scope string) (*ec2.DescribeInstancesInput, error) {
	return &ec2.DescribeInstancesInput{}, nil
}

func instanceOutputMapper(scope string, _ *ec2.DescribeInstancesInput, output *ec2.DescribeInstancesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			attrs, err := sources.ToAttributesCase(instance)

			if err != nil {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_OTHER,
					ErrorString: err.Error(),
					Scope:       scope,
				}
			}

			item := sdp.Item{
				Type:            "ec2-instance",
				UniqueAttribute: "instanceId",
				Scope:           scope,
				Attributes:      attrs,
				LinkedItemQueries: []*sdp.LinkedItemQuery{
					{
						Query: &sdp.Query{
							// +overmind:link ec2-instance-status
							// Always get the status
							Type:   "ec2-instance-status",
							Method: sdp.QueryMethod_GET,
							Query:  *instance.InstanceId,
							Scope:  scope,
						},
					},
				},
			}

			if instance.ImageId != nil {
				// +overmind:link ec2-image
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "ec2-image",
					Method: sdp.QueryMethod_GET,
					Query:  *instance.ImageId,
					Scope:  scope,
				}})
			}

			if instance.KeyName != nil {
				// +overmind:link ec2-key-pair
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "ec2-key-pair",
					Method: sdp.QueryMethod_GET,
					Query:  *instance.KeyName,
					Scope:  scope,
				}})
			}

			if instance.Placement != nil {
				if instance.Placement.AvailabilityZone != nil {
					// +overmind:link ec2-availability-zone
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "ec2-availability-zone",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.Placement.AvailabilityZone,
						Scope:  scope,
					}})
				}

				if instance.Placement.GroupId != nil {
					// +overmind:link ec2-placement-group
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "ec2-placement-group",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.Placement.GroupId,
						Scope:  scope,
					}})
				}
			}

			for _, nic := range instance.NetworkInterfaces {
				// IPs
				for _, ip := range nic.Ipv6Addresses {
					if ip.Ipv6Address != nil {
						// +overmind:link ip
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *ip.Ipv6Address,
							Scope:  "global",
						}})
					}
				}

				for _, ip := range nic.PrivateIpAddresses {
					if ip.PrivateIpAddress != nil {
						// +overmind:link ip
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *ip.PrivateIpAddress,
							Scope:  "global",
						}})
					}
				}

				// Subnet
				if nic.SubnetId != nil {
					// +overmind:link ec2-subnet
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "ec2-subnet",
						Method: sdp.QueryMethod_GET,
						Query:  *nic.SubnetId,
						Scope:  scope,
					}})
				}

				// VPC
				if nic.VpcId != nil {
					// +overmind:link ec2-vpc
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "ec2-vpc",
						Method: sdp.QueryMethod_GET,
						Query:  *nic.VpcId,
						Scope:  scope,
					}})
				}
			}

			if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
				// +overmind:link dns
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.PublicDnsName,
					Scope:  "global",
				}})
			}

			if instance.PublicIpAddress != nil {
				// +overmind:link ip
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "ip",
					Method: sdp.QueryMethod_GET,
					Query:  *instance.PublicIpAddress,
					Scope:  "global",
				}})
			}

			// Security groups
			for _, group := range instance.SecurityGroups {
				if group.GroupId != nil {
					// +overmind:link ec2-security-group
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "ec2-security-group",
						Method: sdp.QueryMethod_GET,
						Query:  *group.GroupId,
						Scope:  scope,
					}})
				}
			}

			for _, mapping := range instance.BlockDeviceMappings {
				if mapping.Ebs != nil && mapping.Ebs.VolumeId != nil {
					// +overmind:link ec2-volume
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "ec2-volume",
						Method: sdp.QueryMethod_GET,
						Query:  *mapping.Ebs.VolumeId,
						Scope:  scope,
					}})
				}
			}

			items = append(items, &item)
		}
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-instance
// +overmind:descriptiveType EC2 Instance
// +overmind:get Get an EC2 instance by ID
// +overmind:list List all EC2 instances
// +overmind:search Search EC2 instances by ARN
// +overmind:group AWS

func NewInstanceSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeInstancesInput, *ec2.DescribeInstancesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeInstancesInput, *ec2.DescribeInstancesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-instance",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeInstances(ctx, input)
		},
		InputMapperGet:  instanceInputMapperGet,
		InputMapperList: instanceInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeInstancesInput) sources.Paginator[*ec2.DescribeInstancesOutput, *ec2.Options] {
			return ec2.NewDescribeInstancesPaginator(client, params)
		},
		OutputMapper: instanceOutputMapper,
	}
}
