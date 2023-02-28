package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func InstanceInputMapperGet(scope, query string) (*ec2.DescribeInstancesInput, error) {
	return &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			query,
		},
	}, nil
}

func InstanceInputMapperList(scope string) (*ec2.DescribeInstancesInput, error) {
	return &ec2.DescribeInstancesInput{}, nil
}

func InstanceOutputMapper(scope string, _ *ec2.DescribeInstancesInput, output *ec2.DescribeInstancesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			attrs, err := sources.ToAttributesCase(instance)

			if err != nil {
				return nil, &sdp.ItemRequestError{
					ErrorType:   sdp.ItemRequestError_OTHER,
					ErrorString: err.Error(),
					Scope:       scope,
				}
			}

			item := sdp.Item{
				Type:            "ec2-instance",
				UniqueAttribute: "instanceId",
				Scope:           scope,
				Attributes:      attrs,
				LinkedItemRequests: []*sdp.ItemRequest{
					{
						// Always get the status
						Type:   "ec2-instance-status",
						Method: sdp.RequestMethod_GET,
						Query:  *instance.InstanceId,
						Scope:  scope,
					},
				},
			}

			if instance.ImageId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-image",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.ImageId,
					Scope:  scope,
				})
			}

			if instance.KeyName != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-key-pair",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.KeyName,
					Scope:  scope,
				})
			}

			if instance.Placement != nil {
				if instance.Placement.AvailabilityZone != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-availability-zone",
						Method: sdp.RequestMethod_GET,
						Query:  *instance.Placement.AvailabilityZone,
						Scope:  scope,
					})
				}

				if instance.Placement.GroupId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-placement-group",
						Method: sdp.RequestMethod_GET,
						Query:  *instance.Placement.GroupId,
						Scope:  scope,
					})
				}
			}

			for _, nic := range instance.NetworkInterfaces {
				// IPs
				for _, ip := range nic.Ipv6Addresses {
					if ip.Ipv6Address != nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "ip",
							Method: sdp.RequestMethod_GET,
							Query:  *ip.Ipv6Address,
							Scope:  "global",
						})
					}
				}

				for _, ip := range nic.PrivateIpAddresses {
					if ip.PrivateIpAddress != nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "ip",
							Method: sdp.RequestMethod_GET,
							Query:  *ip.PrivateIpAddress,
							Scope:  "global",
						})
					}
				}

				// Subnet
				if nic.SubnetId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-subnet",
						Method: sdp.RequestMethod_GET,
						Query:  *nic.SubnetId,
						Scope:  scope,
					})
				}

				// VPC
				if nic.VpcId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-vpc",
						Method: sdp.RequestMethod_GET,
						Query:  *nic.VpcId,
						Scope:  scope,
					})
				}
			}

			if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "dns",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.PublicDnsName,
					Scope:  "global",
				})
			}

			if instance.PublicIpAddress != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.PublicIpAddress,
					Scope:  "global",
				})
			}

			// Security groups
			for _, group := range instance.SecurityGroups {
				if group.GroupId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-security-group",
						Method: sdp.RequestMethod_GET,
						Query:  *group.GroupId,
						Scope:  scope,
					})
				}
			}

			for _, mapping := range instance.BlockDeviceMappings {
				if mapping.Ebs != nil && mapping.Ebs.VolumeId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-volume",
						Method: sdp.RequestMethod_GET,
						Query:  *mapping.Ebs.VolumeId,
						Scope:  scope,
					})
				}
			}

			items = append(items, &item)
		}
	}

	return items, nil
}

func NewInstanceSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeInstancesInput, *ec2.DescribeInstancesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeInstancesInput, *ec2.DescribeInstancesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-instance",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeInstances(ctx, input)
		},
		InputMapperGet:  InstanceInputMapperGet,
		InputMapperList: InstanceInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeInstancesInput) sources.Paginator[*ec2.DescribeInstancesOutput, *ec2.Options] {
			return ec2.NewDescribeInstancesPaginator(client, params)
		},
		OutputMapper: InstanceOutputMapper,
	}
}
