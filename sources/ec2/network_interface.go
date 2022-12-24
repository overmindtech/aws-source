package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func NetworkInterfaceInputMapperGet(scope string, query string) (*ec2.DescribeNetworkInterfacesInput, error) {
	return &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{
			query,
		},
	}, nil
}

func NetworkInterfaceInputMapperList(scope string) (*ec2.DescribeNetworkInterfacesInput, error) {
	return &ec2.DescribeNetworkInterfacesInput{}, nil
}

func NetworkInterfaceOutputMapper(scope string, output *ec2.DescribeNetworkInterfacesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ni := range output.NetworkInterfaces {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ni)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-network-interface",
			UniqueAttribute: "networkInterfaceId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if ni.Attachment != nil {
			if ni.Attachment.InstanceId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-instance",
					Method: sdp.RequestMethod_GET,
					Query:  *ni.Attachment.InstanceId,
					Scope:  scope,
				})
			}
		}

		if ni.AvailabilityZone != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-availability-zone",
				Method: sdp.RequestMethod_GET,
				Query:  *ni.AvailabilityZone,
				Scope:  scope,
			})
		}

		for _, sg := range ni.Groups {
			if sg.GroupId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-security-group",
					Method: sdp.RequestMethod_GET,
					Query:  *sg.GroupId,
					Scope:  scope,
				})
			}
		}

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

		for _, ip := range ni.PrivateIpAddresses {
			if assoc := ip.Association; assoc != nil {
				if assoc.PublicDnsName != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "dns",
						Method: sdp.RequestMethod_GET,
						Query:  *assoc.PublicDnsName,
						Scope:  "global",
					})
				}

				if assoc.PublicIp != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ip",
						Method: sdp.RequestMethod_GET,
						Query:  *assoc.PublicIp,
						Scope:  "global",
					})
				}

				if assoc.CarrierIp != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ip",
						Method: sdp.RequestMethod_GET,
						Query:  *assoc.CarrierIp,
						Scope:  "global",
					})
				}

				if assoc.CustomerOwnedIp != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ip",
						Method: sdp.RequestMethod_GET,
						Query:  *assoc.CustomerOwnedIp,
						Scope:  "global",
					})
				}
			}

			if ip.PrivateDnsName != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "dns",
					Method: sdp.RequestMethod_GET,
					Query:  *ip.PrivateDnsName,
					Scope:  "global",
				})
			}

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

		if ni.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-vpc",
				Method: sdp.RequestMethod_GET,
				Query:  *ni.VpcId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewNetworkInterfaceSource(config aws.Config, accountID string) *sources.AWSSource[*ec2.DescribeNetworkInterfacesInput, *ec2.DescribeNetworkInterfacesOutput, *ec2.Client, *ec2.Options] {
	return &sources.AWSSource[*ec2.DescribeNetworkInterfacesInput, *ec2.DescribeNetworkInterfacesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-network-interface",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
			return client.DescribeNetworkInterfaces(ctx, input)
		},
		InputMapperGet:  NetworkInterfaceInputMapperGet,
		InputMapperList: NetworkInterfaceInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeNetworkInterfacesInput) sources.Paginator[*ec2.DescribeNetworkInterfacesOutput, *ec2.Options] {
			return ec2.NewDescribeNetworkInterfacesPaginator(client, params)
		},
		OutputMapper: NetworkInterfaceOutputMapper,
	}
}
