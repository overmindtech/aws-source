package availabilityzone

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

type AvailabilityZoneSource struct {
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

func (s *AvailabilityZoneSource) Client() *ec2.Client {
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
func (s *AvailabilityZoneSource) Type() string {
	return "ec2-availabilityzone"
}

// Descriptive name for the source, used in logging and metadata
func (s *AvailabilityZoneSource) Name() string {
	return "az-aws-source"
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *AvailabilityZoneSource) Scopes() []string {
	return []string{
		fmt.Sprintf("%v.%v", s.AccountID, s.Config.Region),
	}
}

// AvailabilityZoneClient Collects all functions this code uses from the AWS SDK, for test replacement.
type AvailabilityZoneClient interface {
	DescribeAvailabilityZones(ctx context.Context, params *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error)
}

// Get Get a single item with a given scope and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *AvailabilityZoneSource) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return getImpl(ctx, s.Client(), query, scope)
}

func getImpl(ctx context.Context, client AvailabilityZoneClient, query string, scope string) (*sdp.Item, error) {
	describeAvailabilityZonesOutput, err := client.DescribeAvailabilityZones(
		ctx,
		&ec2.DescribeAvailabilityZonesInput{
			ZoneNames: []string{
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

	numAvailabilityZones := len(describeAvailabilityZonesOutput.AvailabilityZones)

	switch {
	case numAvailabilityZones > 1:
		AvailabilityZoneNames := make([]string, numAvailabilityZones)

		for i, AvailabilityZone := range describeAvailabilityZonesOutput.AvailabilityZones {
			AvailabilityZoneNames[i] = *AvailabilityZone.ZoneName
		}

		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Request returned > 1 AvailabilityZone, cannot determine instance. AvailabilityZones: %v", AvailabilityZoneNames),
			Scope:       scope,
		}
	case numAvailabilityZones == 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("AvailabilityZone %v not found", query),
			Scope:       scope,
		}
	}

	return mapAvailabilityZoneToItem(&describeAvailabilityZonesOutput.AvailabilityZones[0], scope)
}

// List Lists all items in a given scope
func (s *AvailabilityZoneSource) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return listImpl(ctx, s.Client(), scope)
}

func listImpl(ctx context.Context, client AvailabilityZoneClient, scope string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	describeAvailabilityZonesOutput, err := client.DescribeAvailabilityZones(
		ctx,
		&ec2.DescribeAvailabilityZonesInput{},
	)

	if err != nil {
		return items, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	// Convert to items
	for _, AvailabilityZone := range describeAvailabilityZonesOutput.AvailabilityZones {
		item, _ := mapAvailabilityZoneToItem(&AvailabilityZone, scope)
		items = append(items, item)
	}

	return items, nil
}

func mapAvailabilityZoneToItem(az *types.AvailabilityZone, scope string) (*sdp.Item, error) {
	var err error
	var attrs *sdp.ItemAttributes
	attrs, err = sources.ToAttributesCase(az)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	item := sdp.Item{
		Type:            "ec2-availabilityzone",
		UniqueAttribute: "zoneName",
		Scope:           scope,
		Attributes:      attrs,
	}

	// Link to region
	if az.RegionName != nil {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "ec2-region",
			Method: sdp.RequestMethod_GET,
			Query:  *az.RegionName,
			Scope:  scope,
		})
	}

	return &item, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *AvailabilityZoneSource) Weight() int {
	return 100
}
