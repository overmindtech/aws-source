package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// AddressInputMapperGet Maps source calls to the correct input for the AZ API
func AddressInputMapperGet(scope, query string) (*ec2.DescribeAddressesInput, error) {
	return &ec2.DescribeAddressesInput{
		PublicIps: []string{
			query,
		},
	}, nil
}

// AddressInputMapperList Maps source calls to the correct input for the AZ API
func AddressInputMapperList(scope string) (*ec2.DescribeAddressesInput, error) {
	return &ec2.DescribeAddressesInput{}, nil
}

// AddressOutputMapper Maps API output to items
func AddressOutputMapper(scope string, output *ec2.DescribeAddressesOutput) ([]*sdp.Item, error) {
	if output == nil {
		return nil, errors.New("empty output")
	}

	items := make([]*sdp.Item, 0)
	var err error
	var attrs *sdp.ItemAttributes

	for _, address := range output.Addresses {
		attrs, err = sources.ToAttributesCase(address)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "ec2-address",
			UniqueAttribute: "publicIp",
			Scope:           scope,
			Attributes:      attrs,
			LinkedItemRequests: []*sdp.ItemRequest{
				{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *address.PublicIp,
					Scope:  "global",
				},
			},
		}

		if address.InstanceId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-instance",
				Method: sdp.RequestMethod_GET,
				Query:  *address.InstanceId,
				Scope:  scope,
			})
		}

		if address.CarrierIp != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ip",
				Method: sdp.RequestMethod_GET,
				Query:  *address.CarrierIp,
				Scope:  "global",
			})
		}

		if address.CustomerOwnedIp != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ip",
				Method: sdp.RequestMethod_GET,
				Query:  *address.CustomerOwnedIp,
				Scope:  "global",
			})
		}

		if address.NetworkInterfaceId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-network-interface",
				Method: sdp.RequestMethod_GET,
				Query:  *address.NetworkInterfaceId,
				Scope:  scope,
			})
		}

		if address.PrivateIpAddress != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ip",
				Method: sdp.RequestMethod_GET,
				Query:  *address.PrivateIpAddress,
				Scope:  "global",
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

// NewAddressSource Creates a new source for aws-Address resources
func NewAddressSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*ec2.DescribeAddressesInput, *ec2.DescribeAddressesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeAddressesInput, *ec2.DescribeAddressesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-address",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {
			return client.DescribeAddresses(ctx, input)
		},
		InputMapperGet:  AddressInputMapperGet,
		InputMapperList: AddressInputMapperList,
		OutputMapper:    AddressOutputMapper,
	}
}
