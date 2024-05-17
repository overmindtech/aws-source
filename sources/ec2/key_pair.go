package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func keyPairInputMapperGet(scope string, query string) (*ec2.DescribeKeyPairsInput, error) {
	return &ec2.DescribeKeyPairsInput{
		KeyNames: []string{
			query,
		},
	}, nil
}

func keyPairInputMapperList(scope string) (*ec2.DescribeKeyPairsInput, error) {
	return &ec2.DescribeKeyPairsInput{}, nil
}

func keyPairOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeKeyPairsInput, output *ec2.DescribeKeyPairsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, gw := range output.KeyPairs {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(gw, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-key-pair",
			UniqueAttribute: "keyName",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(gw.Tags),
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-key-pair
// +overmind:descriptiveType Key Pair
// +overmind:get Get a key pair by name
// +overmind:list List all key pairs
// +overmind:search Search for key pairs by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_key_pair.id

func NewKeyPairSource(client *ec2.Client, accountID string, region string) *sources.DescribeOnlySource[*ec2.DescribeKeyPairsInput, *ec2.DescribeKeyPairsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeKeyPairsInput, *ec2.DescribeKeyPairsOutput, *ec2.Client, *ec2.Options]{
		Region:    region,
		Client:    client,
		AccountID: accountID,
		ItemType:  "ec2-key-pair",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
			return client.DescribeKeyPairs(ctx, input)
		},
		InputMapperGet:  keyPairInputMapperGet,
		InputMapperList: keyPairInputMapperList,
		OutputMapper:    keyPairOutputMapper,
	}
}
