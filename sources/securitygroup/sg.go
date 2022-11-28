package securitygroup

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/sdp-go"
)

type SecurityGroupSource struct {
	// Config AWS Config including region and credentials
	Config aws.Config

	// AccountID The id of the account that is being used. This is used by
	// sources as the first element in the context
	AccountID string

	// client The AWS client to use when making requests
	client        *ec2.Client
	clientCreated bool
	clientMutex   sync.Mutex
}

func (s *SecurityGroupSource) Client() *ec2.Client {
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
func (s *SecurityGroupSource) Type() string {
	return "sg-instance"
}

// Descriptive name for the source, used in logging and metadata
func (s *SecurityGroupSource) Name() string {
	return "sg-aws-source"
}

// List of contexts that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *SecurityGroupSource) Contexts() []string {
	return []string{
		fmt.Sprintf("%v.%v", s.AccountID, s.Config.Region),
	}
}

// Get Get a single item with a given context and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *SecurityGroupSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	return nil, &sdp.ItemRequestError{
		ErrorType:   sdp.ItemRequestError_OTHER,
		ErrorString: "not implemented",
		Context:     itemContext,
	}
}

// Find Finds all items in a given context
func (s *SecurityGroupSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	return nil, &sdp.ItemRequestError{
		ErrorType:   sdp.ItemRequestError_OTHER,
		ErrorString: "not implemented",
		Context:     itemContext,
	}
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *SecurityGroupSource) Weight() int {
	return 100
}
