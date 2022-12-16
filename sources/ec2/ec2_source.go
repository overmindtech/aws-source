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

const DefaultMaxResultsPerPage = 100

// Paginator Represents an AWS API Paginator:
// https://aws.github.io/aws-sdk-go-v2/docs/making-requests/#using-paginators
// The Output param should be the type of output that this specific paginator
// returns e.g. *ec2.DescribeInstancesOutput
type Paginator[Output any] interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*ec2.Options)) (Output, error)
}

// EC2Source This Struct allows us to create sources easily despite the
// differences between the many EC2 APIs. Not that paginated APIs should
// populate the `InputMapperPaginated` and `OutputMapperPaginated` fields, where
// non-paginated APIs should use `InputMapper` and `OutputMapper`. The source
// will return an error if you use any other combination
type EC2Source[Input any, Output any] struct {
	MaxResultsPerPage int32  // Max results per page when making API queries
	ItemType          string // The type of items that will be returned

	// The function that should be used to describe the resources that this
	// source is related to
	DescribeFunc func(ctx context.Context, client *ec2.Client, input Input, optFns ...func(*ec2.Options)) (Output, error)

	// A function that returns the input object that will be passed to
	// DescribeFunc for a GET request
	InputMapperGet func(scope, query string) (Input, error)

	// A function that returns the input object that will be passed to
	// DescribeFunc for a LIST request
	InputMapperList func(scope string) (Input, error)

	// A function that returns a paginator for this API. If this is nil, we will
	// assume that the API is not paginated
	PaginatorBuilder func(client *ec2.Client, params Input) Paginator[Output]

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

	if e.MaxResultsPerPage == 0 {
		e.MaxResultsPerPage = DefaultMaxResultsPerPage
	}

	if e.InputMapperGet == nil {
		return errors.New("ec2 source get input mapper is nil")
	}

	if e.InputMapperList == nil {
		return errors.New("ec2 source list input mapper is nil")
	}

	if e.OutputMapper == nil {
		return errors.New("ec2 source output mapper is nil")
	}

	return nil
}

// Paginated returns whether or not this source is using a paginated API
func (e *EC2Source[Input, Output]) Paginated() bool {
	return e.PaginatorBuilder != nil
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
	input, err = e.InputMapperGet(scope, query)

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

	err := e.Validate()

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	var items []*sdp.Item

	if e.Paginated() {
		items, err = e.listPaginated(ctx, scope)
	} else {
		items, err = e.listRegular(ctx, scope)
	}

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	return items, nil
}

// listRegular Lists items from the API when the API is not paginated. Basically
// just calls the API and maps the output once
func (e *EC2Source[Input, Output]) listRegular(ctx context.Context, scope string) ([]*sdp.Item, error) {
	var input Input
	var output Output
	var err error
	var items []*sdp.Item

	input, err = e.InputMapperList(scope)

	if err != nil {
		return nil, err
	}

	output, err = e.DescribeFunc(ctx, e.client, input)

	if err != nil {
		return nil, err
	}

	items, err = e.OutputMapper(scope, output)

	if err != nil {
		return nil, err
	}

	return items, nil
}

// listPaginated Lists all items with a paginated API. This requires that the
// `PaginatorBuilder` be set
func (e *EC2Source[Input, Output]) listPaginated(ctx context.Context, scope string) ([]*sdp.Item, error) {
	var input Input
	var output Output
	var err error
	var newItems []*sdp.Item
	items := make([]*sdp.Item, 0)

	input, err = e.InputMapperList(scope)

	if err != nil {
		return nil, err
	}

	if e.PaginatorBuilder == nil {
		return nil, errors.New("paginator builder is nil, cannot use paginated API")
	}

	paginator := e.PaginatorBuilder(e.client, input)

	for paginator.HasMorePages() {
		output, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		newItems, err = e.OutputMapper(scope, output)

		if err != nil {
			return nil, err
		}

		items = append(items, newItems...)
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
