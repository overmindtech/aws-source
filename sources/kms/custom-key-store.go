package kms

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/kms/types"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func customKeyStoreOutputMapper(_ context.Context, _ *kms.Client, scope string, _ *kms.DescribeCustomKeyStoresInput, output *kms.DescribeCustomKeyStoresOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, customKeyStore := range output.CustomKeyStores {
		attributes, err := sources.ToAttributesCase(customKeyStore, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "kms-custom-key-store",
			UniqueAttribute: "customKeyStoreId",
			Attributes:      attributes,
			Scope:           scope,
		}

		switch customKeyStore.ConnectionState {
		case types.ConnectionStateTypeConnected:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.ConnectionStateTypeConnecting:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.ConnectionStateTypeDisconnected:
			item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
		case types.ConnectionStateTypeFailed:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		case types.ConnectionStateTypeDisconnecting:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		default:
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: "unknown Connection State",
			}
		}

		if customKeyStore.CloudHsmClusterId != nil {
			// +overmind:link cloudhsmv2-cluster
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "cloudhsmv2-cluster",
					Method: sdp.QueryMethod_GET,
					Query:  *customKeyStore.CloudHsmClusterId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the CloudHSM cluster will affect the custom key store
					In: true,
					// Updating the custom key store will not affect the CloudHSM cluster
					Out: false,
				},
			})
		}

		// TODO: Activate this after enabling get vpc by name
		/*
			if customKeyStore.XksProxyConfiguration != nil &&
				customKeyStore.XksProxyConfiguration.VpcEndpointServiceName != nil {
				// +overmind:link ec2-vpc
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-vpc",
						Method: sdp.QueryMethod_GET,
						Query:  *customKeyStore.XksProxyConfiguration.VpcEndpointServiceName,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the VPC will affect the custom key store
						In: true,
						// Updating the custom key store will not affect the VPC
						Out: false,
					},
				})
			}
		*/

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type kms-custom-key-store
// +overmind:descriptiveType Custom Key Store
// +overmind:get Get a custom key store by its ID or name
// +overmind:list List all custom key stores
// +overmind:search Search custom key store by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_kms_custom_key_store.id

func NewCustomKeyStoreSource(client *kms.Client, accountID string, region string) *sources.DescribeOnlySource[*kms.DescribeCustomKeyStoresInput, *kms.DescribeCustomKeyStoresOutput, *kms.Client, *kms.Options] {
	return &sources.DescribeOnlySource[*kms.DescribeCustomKeyStoresInput, *kms.DescribeCustomKeyStoresOutput, *kms.Client, *kms.Options]{
		Region:    region,
		Client:    client,
		AccountID: accountID,
		ItemType:  "kms-custom-key-store",
		DescribeFunc: func(ctx context.Context, client *kms.Client, input *kms.DescribeCustomKeyStoresInput) (*kms.DescribeCustomKeyStoresOutput, error) {
			return client.DescribeCustomKeyStores(ctx, input)
		},
		InputMapperGet: func(_, query string) (*kms.DescribeCustomKeyStoresInput, error) {
			return describeInput(query)
		},
		InputMapperList: func(string) (*kms.DescribeCustomKeyStoresInput, error) {
			return &kms.DescribeCustomKeyStoresInput{}, nil
		},
		OutputMapper: customKeyStoreOutputMapper,
	}
}

func describeInput(query string) (*kms.DescribeCustomKeyStoresInput, error) {
	errMsgFmt := "invalid query format: %s, expected 'id|custom-key-store-id' or 'name|custom-key-store-name'"

	// query can be in the format of:
	// 1) id|custom-key-store-id
	// 2) name|custom-key-store-name

	q := strings.Split(query, "|")
	if len(q) != 2 {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: fmt.Sprintf(errMsgFmt, query),
		}
	}

	switch q[0] {
	case "id":
		return &kms.DescribeCustomKeyStoresInput{
			CustomKeyStoreId: &q[1],
		}, nil
	case "name":
		return &kms.DescribeCustomKeyStoresInput{
			CustomKeyStoreName: &q[1],
		}, nil
	default:
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: fmt.Sprintf(errMsgFmt, query),
		}
	}
}
