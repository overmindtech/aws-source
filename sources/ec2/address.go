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
func addressInputMapperGet(scope, query string) (*ec2.DescribeAddressesInput, error) {
	return &ec2.DescribeAddressesInput{
		PublicIps: []string{
			query,
		},
	}, nil
}

// AddressInputMapperList Maps source calls to the correct input for the AZ API
func addressInputMapperList(scope string) (*ec2.DescribeAddressesInput, error) {
	return &ec2.DescribeAddressesInput{}, nil
}

// AddressOutputMapper Maps API output to items
func addressOutputMapper(scope string, _ *ec2.DescribeAddressesInput, output *ec2.DescribeAddressesOutput) ([]*sdp.Item, error) {
	if output == nil {
		return nil, errors.New("empty output")
	}

	items := make([]*sdp.Item, 0)
	var err error
	var attrs *sdp.ItemAttributes

	// An EC2-address, along with an IP is an item that inherently links things
	// and therefore should propagate blast radius in both directions on all
	// links
	bp := &sdp.BlastPropagation{
		In:  true,
		Out: true,
	}

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
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *address.PublicIp,
						Scope:  "global",
					},
					BlastPropagation: bp,
				},
			},
		}

		if address.InstanceId != nil {
			// +overmind:link ec2-instance
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *address.InstanceId,
					Scope:  scope,
				},
				BlastPropagation: bp,
			})
		}

		if address.CarrierIp != nil {
			// +overmind:link ip
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ip",
					Method: sdp.QueryMethod_GET,
					Query:  *address.CarrierIp,
					Scope:  "global",
				},
				BlastPropagation: bp,
			})
		}

		if address.CustomerOwnedIp != nil {
			// +overmind:link ip
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ip",
					Method: sdp.QueryMethod_GET,
					Query:  *address.CustomerOwnedIp,
					Scope:  "global",
				},
				BlastPropagation: bp,
			})
		}

		if address.NetworkInterfaceId != nil {
			// +overmind:link ec2-network-interface
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-network-interface",
					Method: sdp.QueryMethod_GET,
					Query:  *address.NetworkInterfaceId,
					Scope:  scope,
				},
				BlastPropagation: bp,
			})
		}

		if address.PrivateIpAddress != nil {
			// +overmind:link ip
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ip",
					Method: sdp.QueryMethod_GET,
					Query:  *address.PrivateIpAddress,
					Scope:  "global",
				},
				BlastPropagation: bp,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-address
// +overmind:descriptiveType EC2 Address
// +overmind:get Get an EC2 address by Public IP
// +overmind:list List EC2 addresses
// +overmind:search Search for EC2 addresses by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_eip.public_ip
// +overmind:terraform:queryMap aws_eip_association.public_ip

// NewAddressSource Creates a new source for aws-Address resources
func NewAddressSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeAddressesInput, *ec2.DescribeAddressesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeAddressesInput, *ec2.DescribeAddressesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-address",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeAddresses(ctx, input)
		},
		InputMapperGet:  addressInputMapperGet,
		InputMapperList: addressInputMapperList,
		OutputMapper:    addressOutputMapper,
	}
}
