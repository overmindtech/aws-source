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
						BlastPropagation: &sdp.BlastPropagation{
							// The status and the instance are closely linked and
							// affect each other
							In:  true,
							Out: true,
						},
					},
				},
			}

			if instance.IamInstanceProfile != nil {
				// Prefer the ARN
				if instance.IamInstanceProfile.Arn != nil {
					if arn, err := sources.ParseARN(*instance.IamInstanceProfile.Arn); err == nil {
						// +overmind:link iam-instance-profile
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "iam-instance-profile",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *instance.IamInstanceProfile.Arn,
								Scope:  sources.FormatScope(arn.AccountID, arn.Region),
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changes to the profile will affect this instance
								In: true,
								// We can't affect the profile
								Out: false,
							},
						})
					}
				} else if instance.IamInstanceProfile.Id != nil {
					// +overmind:link iam-instance-profile
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "iam-instance-profile",
							Method: sdp.QueryMethod_GET,
							Query:  *instance.IamInstanceProfile.Id,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changes to the profile will affect this instance
							In: true,
							// We can't affect the profile
							Out: false,
						},
					})
				}
			}

			if instance.CapacityReservationId != nil {
				// +overmind:link ec2-capacity-reservation
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-capacity-reservation",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.CapacityReservationId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the reservation will affect the instance
						In: true,
						// Changing the instance won't affect the reservation
						Out: false,
					},
				})
			}

			for _, assoc := range instance.ElasticGpuAssociations {
				if assoc.ElasticGpuId != nil {
					// +overmind:link ec2-elastic-gpu
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-elastic-gpu",
							Method: sdp.QueryMethod_GET,
							Query:  *assoc.ElasticGpuId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the GPU will affect the instance
							In: true,
							// Changing the instance won't affect the GPU
							Out: false,
						},
					})
				}
			}

			for _, assoc := range instance.ElasticInferenceAcceleratorAssociations {
				if assoc.ElasticInferenceAcceleratorArn != nil {
					if arn, err := sources.ParseARN(*assoc.ElasticInferenceAcceleratorArn); err == nil {
						// +overmind:link elastic-inference-accelerator
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "elastic-inference-accelerator",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *assoc.ElasticInferenceAcceleratorArn,
								Scope:  sources.FormatScope(arn.AccountID, arn.Region),
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changing the accelerator will affect the instance
								In: true,
								// Changing the instance won't affect the accelerator
								Out: false,
							},
						})
					}
				}
			}

			for _, license := range instance.Licenses {
				if license.LicenseConfigurationArn != nil {
					if arn, err := sources.ParseARN(*license.LicenseConfigurationArn); err == nil {
						// +overmind:link license-manager-license-configuration
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "license-manager-license-configuration",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *license.LicenseConfigurationArn,
								Scope:  sources.FormatScope(arn.AccountID, arn.Region),
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changing the license will affect the instance
								In: true,
								// Changing the instance won't affect the license
								Out: false,
							},
						})
					}
				}
			}

			if instance.OutpostArn != nil {
				if arn, err := sources.ParseARN(*instance.OutpostArn); err == nil {
					// +overmind:link outposts-outpost
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "outposts-outpost",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *instance.OutpostArn,
							Scope:  sources.FormatScope(arn.AccountID, arn.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the outpost will affect the instance
							In: true,
							// Changing the instance won't affect the outpost
							Out: false,
						},
					})
				}
			}

			if instance.SpotInstanceRequestId != nil {
				// +overmind:link ec2-spot-instance-request
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-spot-instance-request",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.SpotInstanceRequestId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the spot request will affect the instance
						In: true,
						// Changing the instance won't affect the spot request
						Out: false,
					},
				})
			}

			if instance.ImageId != nil {
				// +overmind:link ec2-image
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-image",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.ImageId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the image can't affect the instance once it
						// has been created
						In:  false,
						Out: false,
					},
				})
			}

			if instance.KeyName != nil {
				// +overmind:link ec2-key-pair
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-key-pair",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.KeyName,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the key pair will affect your ability to
						// connect to the instance
						In: true,
						// Changing the instance won't affect the key pair
						Out: false,
					},
				})
			}

			if instance.Placement != nil {
				if instance.Placement.AvailabilityZone != nil {
					// +overmind:link ec2-availability-zone
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-availability-zone",
							Method: sdp.QueryMethod_GET,
							Query:  *instance.Placement.AvailabilityZone,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// AZs don't change
							In:  false,
							Out: false,
						},
					})
				}

				if instance.Placement.GroupId != nil {
					// +overmind:link ec2-placement-group
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-placement-group",
							Method: sdp.QueryMethod_GET,
							Query:  *instance.Placement.GroupId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing a placement group will affect instances
							In: true,
							// Changing an instance won't affect the group
							Out: false,
						},
					})
				}
			}

			if instance.Ipv6Address != nil {
				// +overmind:link ip
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.Ipv6Address,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// IPs are always linked
						In:  true,
						Out: true,
					},
				})
			}

			for _, nic := range instance.NetworkInterfaces {
				// IPs
				for _, ip := range nic.Ipv6Addresses {
					if ip.Ipv6Address != nil {
						// +overmind:link ip
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "ip",
								Method: sdp.QueryMethod_GET,
								Query:  *ip.Ipv6Address,
								Scope:  "global",
							},
							BlastPropagation: &sdp.BlastPropagation{
								// IPs are always linked
								In:  true,
								Out: true,
							},
						})
					}
				}

				for _, ip := range nic.PrivateIpAddresses {
					if ip.PrivateIpAddress != nil {
						// +overmind:link ip
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "ip",
								Method: sdp.QueryMethod_GET,
								Query:  *ip.PrivateIpAddress,
								Scope:  "global",
							},
							BlastPropagation: &sdp.BlastPropagation{
								// IPs are always linked
								In:  true,
								Out: true,
							},
						})
					}
				}

				// Subnet
				if nic.SubnetId != nil {
					// +overmind:link ec2-subnet
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-subnet",
							Method: sdp.QueryMethod_GET,
							Query:  *nic.SubnetId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the subnet will affect the instance
							In: true,
							// Changing the instance won't affect the subnet
							Out: false,
						},
					})
				}

				// VPC
				if nic.VpcId != nil {
					// +overmind:link ec2-vpc
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-vpc",
							Method: sdp.QueryMethod_GET,
							Query:  *nic.VpcId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the VPC will affect the instance
							In: true,
							// Changing the instance won't affect the VPC
							Out: false,
						},
					})
				}
			}

			if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
				// +overmind:link dns
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "dns",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *instance.PublicDnsName,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// DNS records are always linked
						In:  true,
						Out: true,
					},
				})
			}

			if instance.PublicIpAddress != nil {
				// +overmind:link ip
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.PublicIpAddress,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// IPs are always propagating
						In:  true,
						Out: true,
					},
				})
			}

			// Security groups
			for _, group := range instance.SecurityGroups {
				if group.GroupId != nil {
					// +overmind:link ec2-security-group
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-security-group",
							Method: sdp.QueryMethod_GET,
							Query:  *group.GroupId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the security group will affect the instance
							In: true,
							// Changing the instance won't affect the security group
							Out: false,
						},
					})
				}
			}

			for _, mapping := range instance.BlockDeviceMappings {
				if mapping.Ebs != nil && mapping.Ebs.VolumeId != nil {
					// +overmind:link ec2-volume
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-volume",
							Method: sdp.QueryMethod_GET,
							Query:  *mapping.Ebs.VolumeId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the volume will affect the instance
							In: true,
							// Changing the instance could also affect the
							// volume since it's writing to it
							Out: true,
						},
					})
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
// +overmind:terraform:queryMap aws_instance.id

func NewInstanceSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeInstancesInput, *ec2.DescribeInstancesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeInstancesInput, *ec2.DescribeInstancesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-instance",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
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
