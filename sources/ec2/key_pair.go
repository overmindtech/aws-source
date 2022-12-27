package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func KeyPairInputMapperGet(scope string, query string) (*ec2.DescribeKeyPairsInput, error) {
	return &ec2.DescribeKeyPairsInput{
		KeyPairIds: []string{
			query,
		},
	}, nil
}

func KeyPairInputMapperList(scope string) (*ec2.DescribeKeyPairsInput, error) {
	return &ec2.DescribeKeyPairsInput{}, nil
}

func KeyPairOutputMapper(scope string, output *ec2.DescribeKeyPairsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, gw := range output.KeyPairs {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(gw)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-key-pair",
			UniqueAttribute: "keyPairId",
			Scope:           scope,
			Attributes:      attrs,
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewKeyPairSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*ec2.DescribeKeyPairsInput, *ec2.DescribeKeyPairsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeKeyPairsInput, *ec2.DescribeKeyPairsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-key-pair",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
			return client.DescribeKeyPairs(ctx, input)
		},
		InputMapperGet:  KeyPairInputMapperGet,
		InputMapperList: KeyPairInputMapperList,
		OutputMapper:    KeyPairOutputMapper,
	}
}
