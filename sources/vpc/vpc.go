package vpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type VpcSource struct {
	// Config AWS Config including region and credentials
	Config aws.Config

	// AccountID The id of the account that is being used. This is used by
	// sources as the first element in the scope
	AccountID string

	// client The AWS client to use when making requests
	client        *ec2.Client
	clientCreated bool
	clientMutex   sync.Mutex
}

func (s *VpcSource) Client() *ec2.Client {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	// If the client already exists then return it
	if s.clientCreated {
		return s.client
	}

	// Otherwise create a new client from the config
	s.client = ec2.NewFromConfig(s.Config)
	s.clientCreated = true

	return s.client
}

// Type The type of items that this source is capable of finding
func (s *VpcSource) Type() string {
	return "ec2-vpc"
}

// Descriptive name for the source, used in logging and metadata
func (s *VpcSource) Name() string {
	return "vpc-aws-source"
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *VpcSource) Scopes() []string {
	return []string{
		fmt.Sprintf("%v.%v", s.AccountID, s.Config.Region),
	}
}

// VpcClient Collects all functions this code uses from the AWS SDK, for test replacement.
type VpcClient interface {
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
}

// Get Get a single item with a given scope and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *VpcSource) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return getImpl(ctx, s.Client(), query, scope)
}

func getImpl(ctx context.Context, client VpcClient, query string, scope string) (*sdp.Item, error) {
	describeVpcsOutput, err := client.DescribeVpcs(
		ctx,
		&ec2.DescribeVpcsInput{
			VpcIds: []string{
				query,
			},
		},
	)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	numVpcs := len(describeVpcsOutput.Vpcs)

	switch {
	case numVpcs > 1:
		VpcIDs := make([]string, numVpcs)

		for i, Vpc := range describeVpcsOutput.Vpcs {
			VpcIDs[i] = *Vpc.VpcId
		}

		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Request returned > 1 Vpc, cannot determine instance. Vpcs: %v", VpcIDs),
			Scope:       scope,
		}
	case numVpcs == 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("Vpc %v not found", query),
			Scope:       scope,
		}
	}

	return mapVpcToItem(&describeVpcsOutput.Vpcs[0], scope)
}

// List Lists all items in a given scope
func (s *VpcSource) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return listImpl(ctx, s.Client(), scope)
}

func listImpl(ctx context.Context, client VpcClient, scope string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)
	Vpcs := make([]types.Vpc, 0)
	var maxResults int32 = 100
	var nextToken *string

	for morePages := true; morePages; {
		describeVpcsOutput, err := client.DescribeVpcs(
			ctx,
			&ec2.DescribeVpcsInput{
				MaxResults: &maxResults,
				NextToken:  nextToken,
			},
		)

		if err != nil {
			return items, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		Vpcs = append(Vpcs, describeVpcsOutput.Vpcs...)

		// If there is more data we should store the token so that we can use
		// that. We also need to set morePages to true so that the loop runs
		// again
		nextToken = describeVpcsOutput.NextToken
		morePages = (nextToken != nil)
	}

	// Convert to items
	for _, Vpc := range Vpcs {
		item, _ := mapVpcToItem(&Vpc, scope)
		items = append(items, item)
	}

	return items, nil
}

func mapVpcToItem(vpc *types.Vpc, scope string) (*sdp.Item, error) {
	var err error
	var attrs *sdp.ItemAttributes
	attrs, err = sources.ToAttributesCase(vpc)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	item := sdp.Item{
		Type:            "ec2-vpc",
		UniqueAttribute: "vpcId",
		Scope:           scope,
		Attributes:      attrs,
	}

	return &item, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *VpcSource) Weight() int {
	return 100
}
