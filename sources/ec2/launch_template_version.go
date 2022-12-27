package ec2

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func LaunchTemplateVersionInputMapperGet(scope string, query string) (*ec2.DescribeLaunchTemplateVersionsInput, error) {
	// We are expecting the query to be {id}.{version}
	sections := strings.Split(query, ".")

	if len(sections) != 2 {
		return nil, errors.New("input did not have 2 sections")
	}

	return &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: &sections[0],
		Versions: []string{
			sections[1],
		},
	}, nil
}

func LaunchTemplateVersionInputMapperList(scope string) (*ec2.DescribeLaunchTemplateVersionsInput, error) {
	return &ec2.DescribeLaunchTemplateVersionsInput{}, nil
}

func LaunchTemplateVersionOutputMapper(scope string, output *ec2.DescribeLaunchTemplateVersionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ltv := range output.LaunchTemplateVersions {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ltv)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		if ltv.LaunchTemplateId != nil && ltv.VersionNumber != nil {
			// Create a custom UAV here since there is no one unique attribute.
			// The new UAV will be {templateId}.{version}
			attrs.Set("versionIdCombo", fmt.Sprintf("%v.%v", ltv.LaunchTemplateId, ltv.VersionNumber))
		} else {
			return nil, errors.New("ec2-launch-template-version must have LaunchTemplateId and VersionNumber populated")
		}

		item := sdp.Item{
			Type:            "ec2-launch-template-version",
			UniqueAttribute: "versionIdCombo",
			Scope:           scope,
			Attributes:      attrs,
		}

		if lt := ltv.LaunchTemplateData; lt != nil {
			for _, ni := range lt.NetworkInterfaces {
				for _, ip := range ni.Ipv6Addresses {
					if ip.Ipv6Address != nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "ip",
							Method: sdp.RequestMethod_GET,
							Query:  *ip.Ipv6Address,
							Scope:  "global",
						})
					}
				}

				if ni.NetworkInterfaceId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-network-interface",
						Method: sdp.RequestMethod_GET,
						Query:  *ni.NetworkInterfaceId,
						Scope:  scope,
					})
				}

				for _, ip := range ni.PrivateIpAddresses {
					if ip.PrivateIpAddress != nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "ip",
							Method: sdp.RequestMethod_GET,
							Query:  *ip.PrivateIpAddress,
							Scope:  "global",
						})
					}
				}

				if ni.SubnetId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-subnet",
						Method: sdp.RequestMethod_GET,
						Query:  *ni.SubnetId,
						Scope:  scope,
					})
				}

				for _, group := range ni.Groups {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-security-group",
						Method: sdp.RequestMethod_GET,
						Query:  group,
						Scope:  scope,
					})
				}
			}

			if lt.ImageId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-image",
					Method: sdp.RequestMethod_GET,
					Query:  *lt.ImageId,
					Scope:  scope,
				})
			}

			if lt.KeyName != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-key-pair",
					Method: sdp.RequestMethod_GET,
					Query:  *lt.KeyName,
					Scope:  scope,
				})
			}

			for _, mapping := range lt.BlockDeviceMappings {
				if mapping.Ebs != nil && mapping.Ebs.SnapshotId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-snapshot",
						Method: sdp.RequestMethod_GET,
						Query:  *mapping.Ebs.SnapshotId,
						Scope:  scope,
					})
				}
			}

			if spec := lt.CapacityReservationSpecification; spec != nil {
				if target := spec.CapacityReservationTarget; target != nil {
					if target.CapacityReservationId != nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "ec2-capacity-reservation",
							Method: sdp.RequestMethod_GET,
							Query:  *target.CapacityReservationId,
							Scope:  scope,
						})
					}
				}
			}

			if lt.Placement != nil {
				if lt.Placement.AvailabilityZone != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-availability-zone",
						Method: sdp.RequestMethod_GET,
						Query:  *lt.Placement.AvailabilityZone,
						Scope:  scope,
					})
				}

				if lt.Placement.GroupId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-placement-group",
						Method: sdp.RequestMethod_GET,
						Query:  *lt.Placement.GroupId,
						Scope:  scope,
					})
				}

				if lt.Placement.HostId != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-host",
						Method: sdp.RequestMethod_GET,
						Query:  *lt.Placement.HostId,
						Scope:  scope,
					})
				}
			}

			for _, id := range lt.SecurityGroupIds {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-security-group",
					Method: sdp.RequestMethod_GET,
					Query:  id,
					Scope:  scope,
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewLaunchTemplateVersionSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeLaunchTemplateVersionsInput, *ec2.DescribeLaunchTemplateVersionsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeLaunchTemplateVersionsInput, *ec2.DescribeLaunchTemplateVersionsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-launch-template-version",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeLaunchTemplateVersions(ctx, input)
		},
		InputMapperGet:  LaunchTemplateVersionInputMapperGet,
		InputMapperList: LaunchTemplateVersionInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeLaunchTemplateVersionsInput) sources.Paginator[*ec2.DescribeLaunchTemplateVersionsOutput, *ec2.Options] {
			return ec2.NewDescribeLaunchTemplateVersionsPaginator(client, params)
		},
		OutputMapper: LaunchTemplateVersionOutputMapper,
	}
}
