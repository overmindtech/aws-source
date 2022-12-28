package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/overmindtech/sdp-go"
)

// DescribeOnlySource Generates a source for AWS APIs that only use a `Describe`
// function for both List and Get operations. EC2 is a good example of this,
// where running Describe with no params returns everything, but params can be
// supplied to reduce the number of results.
type DescribeOnlySource[Input InputType, Output OutputType, ClientStruct ClientStructType, Options OptionsType] struct {
	MaxResultsPerPage int32  // Max results per page when making API queries
	ItemType          string // The type of items that will be returned

	// The function that should be used to describe the resources that this
	// source is related to
	DescribeFunc func(ctx context.Context, client ClientStruct, input Input) (Output, error)

	// A function that returns the input object that will be passed to
	// DescribeFunc for a GET request
	InputMapperGet func(scope, query string) (Input, error)

	// A function that returns the input object that will be passed to
	// DescribeFunc for a LIST request
	InputMapperList func(scope string) (Input, error)

	// A function that returns a paginator for this API. If this is nil, we will
	// assume that the API is not paginated e.g.
	// https://aws.github.io/aws-sdk-go-v2/docs/making-requests/#using-paginators
	PaginatorBuilder func(client ClientStruct, params Input) Paginator[Output, Options]

	// A function that returns a slice of items for a given output. If this is a
	// GET request the EC2 source itself will handle errors if there are too
	// many items returned, so no need to worry about handling that
	OutputMapper func(scope string, output Output) ([]*sdp.Item, error)

	// Config AWS Config including region and credentials
	Config aws.Config

	// AccountID The id of the account that is being used. This is used by
	// sources as the first element in the scope
	AccountID string

	// Client The AWS client to use when making requests
	Client ClientStruct
}

// Validate Checks that the source is correctly set up and returns an error if
// not
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) Validate() error {
	if s.DescribeFunc == nil {
		return errors.New("ec2 source describe func is nil")
	}

	if s.MaxResultsPerPage == 0 {
		s.MaxResultsPerPage = DefaultMaxResultsPerPage
	}

	if s.InputMapperGet == nil {
		return errors.New("ec2 source get input mapper is nil")
	}

	if s.InputMapperList == nil {
		return errors.New("ec2 source list input mapper is nil")
	}

	if s.OutputMapper == nil {
		return errors.New("ec2 source output mapper is nil")
	}

	return nil
}

// Paginated returns whether or not this source is using a paginated API
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) Paginated() bool {
	return s.PaginatorBuilder != nil
}

func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) Type() string {
	return s.ItemType
}

func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) Name() string {
	return fmt.Sprintf("%v-source", s.ItemType)
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) Scopes() []string {
	return []string{
		FormatScope(s.AccountID, s.Config.Region),
	}
}

// Get Get a single item with a given scope and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	var input Input
	var output Output
	var err error
	var items []*sdp.Item

	err = s.Validate()

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	// Get the input object
	input, err = s.InputMapperGet(scope, query)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	// Call the API using the object
	output, err = s.DescribeFunc(ctx, s.Client, input)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	items, err = s.OutputMapper(scope, output)

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
			ErrorString: fmt.Sprintf("%v %v not found", s.Type(), query),
		}
	}

	return items[0], nil
}

// List Lists all items in a given scope
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	err := s.Validate()

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	var items []*sdp.Item

	if s.Paginated() {
		items, err = s.listPaginated(ctx, scope)
	} else {
		items, err = s.listRegular(ctx, scope)
	}

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	return items, nil
}

// Search Searches for AWS resources by ARN
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) Search(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	// Parse the ARN
	a, err := ParseARN(query)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	if arnScope := FormatScope(a.AccountID, a.Region); arnScope != scope {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("ARN scope %v does not match request scope %v", arnScope, scope),
			Scope:       scope,
		}
	}

	item, err := s.Get(ctx, scope, a.ResourceID)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	return []*sdp.Item{item}, nil
}

// listRegular Lists items from the API when the API is not paginated. Basically
// just calls the API and maps the output once
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) listRegular(ctx context.Context, scope string) ([]*sdp.Item, error) {
	var input Input
	var output Output
	var err error
	var items []*sdp.Item

	input, err = s.InputMapperList(scope)

	if err != nil {
		return nil, err
	}

	output, err = s.DescribeFunc(ctx, s.Client, input)

	if err != nil {
		return nil, err
	}

	items, err = s.OutputMapper(scope, output)

	if err != nil {
		return nil, err
	}

	return items, nil
}

// listPaginated Lists all items with a paginated API. This requires that the
// `PaginatorBuilder` be set
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) listPaginated(ctx context.Context, scope string) ([]*sdp.Item, error) {
	var input Input
	var output Output
	var err error
	var newItems []*sdp.Item
	items := make([]*sdp.Item, 0)

	input, err = s.InputMapperList(scope)

	if err != nil {
		return nil, err
	}

	if s.PaginatorBuilder == nil {
		return nil, errors.New("paginator builder is nil, cannot use paginated API")
	}

	paginator := s.PaginatorBuilder(s.Client, input)

	for paginator.HasMorePages() {
		output, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		newItems, err = s.OutputMapper(scope, output)

		if err != nil {
			return nil, err
		}

		items = append(items, newItems...)
	}

	return items, nil
}

// Weight Returns the priority weighting of items returned by this sourcs.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *DescribeOnlySource[Input, Output, ClientStruct, Options]) Weight() int {
	return 100
}
