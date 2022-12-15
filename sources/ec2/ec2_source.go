package ec2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/sdp-go"
)

type EC2Source[Input any, Output any] struct {
	ItemType string // The type of items that will be returned

	// The funciton that should be used to describe the resources that this
	// source is related to
	DescribeFunc func(ctx context.Context, client *ec2.Client, input Input, optFns ...func(*ec2.Options)) (Output, error)

	// A function that returns the input object that will be passed to
	// DescribeFunc for a given set of scope, query and method
	InputMapper func(scope string, query string, method sdp.RequestMethod) (Input, error)

	// A function that returns a slice of items for a given output. If this is a
	// GET request the EC2 source itself will handle errors if there are too
	// many items returned, so no need to worry about handling that
	OutputMapper func(scope string, output Output) ([]*sdp.Item, error)

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

func (e *EC2Source[Input, Output]) Client() *ec2.Client {
	e.clientMutex.Lock()
	defer e.clientMutex.Unlock()

	// If the client already exists then return it
	if e.clientCreated {
		return e.client
	}

	// Otherwise create a new client from the config
	e.client = ec2.NewFromConfig(e.Config)
	e.clientCreated = true

	return e.client
}

// Validate Checks that the source is correctly set up and returns an error if
// not
func (e *EC2Source[Input, Output]) Validate() error {
	if e.DescribeFunc == nil {
		return errors.New("ec2 source describe func is nil")
	}

	if e.InputMapper == nil {
		return errors.New("ec2 source input mapper is nil")
	}

	if e.OutputMapper == nil {
		return errors.New("ec2 source output mapper is nil")
	}

	return nil
}

func (e *EC2Source[Input, Output]) Type() string {
	return e.ItemType
}

func (e *EC2Source[Input, Output]) Name() string {
	return fmt.Sprintf("%v-source", e.ItemType)
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (e *EC2Source[Input, Output]) Scopes() []string {
	return []string{
		fmt.Sprintf("%v.%v", e.AccountID, e.Config.Region),
	}
}

// Get Get a single item with a given scope and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (e *EC2Source[Input, Output]) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != e.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, e.Scopes()[0]),
		}
	}

	var input Input
	var output Output
	var err error
	var items []*sdp.Item

	err = e.Validate()

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	// Get the input object
	input, err = e.InputMapper(scope, query, sdp.RequestMethod_GET)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	// Call the API using the object
	output, err = e.DescribeFunc(ctx, e.Client(), input)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	items, err = e.OutputMapper(scope, output)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	numItems := len(items)

	switch {
	case numItems > 1:
		itemNames := make([]string, len(items))

		// Get the names for logging
		for i := range items {
			itemNames[i] = items[i].GloballyUniqueName()
		}

		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Request returned > 1 item for a GET request. Items: %v", strings.Join(itemNames, ", ")),
		}
	case numItems == 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("%v %v not found", e.Type(), query),
		}
	}

	return items[0], nil
}

// List Lists all items in a given scope
func (e *EC2Source[Input, Output]) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != e.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, e.Scopes()[0]),
		}
	}

	var input Input
	var output Output
	var err error
	var items []*sdp.Item

	err = e.Validate()

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	// Get the input object
	input, err = e.InputMapper(scope, "", sdp.RequestMethod_LIST)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	// Call the API using the object
	output, err = e.DescribeFunc(ctx, e.Client(), input)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	items, err = e.OutputMapper(scope, output)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (e *EC2Source[Input, Output]) Weight() int {
	return 100
}
