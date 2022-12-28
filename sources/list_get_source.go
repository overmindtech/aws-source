package sources

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/overmindtech/sdp-go"
	log "github.com/sirupsen/logrus"
)

// MaxParallel An integer that defaults to 10
type MaxParallel int

// Value Get the value of MaxParallel, defaulting to 10
func (m MaxParallel) Value() int {
	if m == 0 {
		return 10
	}

	return int(m)
}

// ListGetSource This source is designed for AWS APIs that have separate List
// and Get functions. It also assumes that the results of the list function
// cannot be converted directly into items as they do not contain enough
// information, and therefore they need to be passed to the Get function before
// returning. An example is the `ListClusters` API in EKS which returns a list
// of cluster names.
type ListGetSource[ListInput InputType, ListOutput OutputType, GetInput InputType, GetOutput OutputType, ClientStruct ClientStructType, Options OptionsType] struct {
	ItemType    string       // The type of items to return
	Client      ClientStruct // The AWS API client
	AccountID   string       // The AWS account ID
	Region      string       // The AWS region this is related to
	MaxParallel MaxParallel  // How many Get request to run in parallel for a single List request

	// Disables List(), meaning all calls will return empty results. This does
	// not affect Search()
	DisableList bool

	// A function that gets the details of a given item
	GetFunc func(ctx context.Context, client ClientStruct, scope string, input GetInput) (*sdp.Item, error)

	// The input to the ListFunc. This is static
	ListInput ListInput

	// A function that maps from the GDP get inputs to the relevant input for
	// the GetFunc
	GetInputMapper func(scope, query string) GetInput

	// Maps search terms from an SDP Search request into the relevant input for
	// the ListFunc. If this is not set, Search() will handle ARNs like most AWS
	// sources
	SearchInputMapper func(scope, query string) (ListInput, error)

	// A function that returns a paginator for the ListFunc
	ListFuncPaginatorBuilder func(client ClientStruct, input ListInput) Paginator[ListOutput, Options]

	// A function that accepts the output of a ListFunc and maps this to a slice
	// of inputs to pass to the GetFunc. The input used for the ListFunc is also
	// included in case it is required
	ListFuncOutputMapper func(output ListOutput, input ListInput) ([]GetInput, error)
}

// Validate Checks that the source has been set up correctly
func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Validate() error {
	if s.ListFuncPaginatorBuilder == nil {
		return errors.New("ListFuncPaginatorBuilder is nil")
	}

	if s.ListFuncOutputMapper == nil {
		return errors.New("ListFuncOutputMapper is nil")
	}

	if s.GetFunc == nil {
		return errors.New("GetFunc is nil")
	}

	if s.GetInputMapper == nil {
		return errors.New("GetInputMapper is nil")
	}

	return nil
}

func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Type() string {
	return s.ItemType
}

func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Name() string {
	return fmt.Sprintf("%v-source", s.ItemType)
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Scopes() []string {
	return []string{
		FormatScope(s.AccountID, s.Region),
	}
}

func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	var err error
	var item *sdp.Item

	if err = s.Validate(); err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	input := s.GetInputMapper(scope, query)

	item, err = s.GetFunc(ctx, s.Client, scope, input)

	if err != nil {
		// TODO: How can we handle NOTFOUND?
		return nil, sdp.NewItemRequestError(err)
	}

	return item, nil
}

// List Lists all available items. This is done by running the ListFunc, then
// passing these results to GetFunc in order to get the details
func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	// Check to see if we have supplied the required functions
	if s.DisableList {
		// In this case we can't run list, so just return empty
		return []*sdp.Item{}, nil
	}

	return s.listInternal(ctx, scope, s.ListInput)
}

// listInternal Accepts a ListInput and runs the List logic against it
func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) listInternal(ctx context.Context, scope string, input ListInput) ([]*sdp.Item, error) {
	var output ListOutput
	var err error
	items := make([]*sdp.Item, 0)
	itemsChan := make(chan *sdp.Item)
	getInputs := make(chan GetInput)
	doneChan := make(chan struct{})

	if err = s.Validate(); err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	// Create a channel of permissions to allow only a certain number of Get requests to tun in parallel
	permissions := make(chan struct{}, s.MaxParallel.Value())
	for i := 0; i < s.MaxParallel.Value(); i++ {
		permissions <- struct{}{}
	}

	// Create a process to take queries and run them using Get
	go func() {
		var wg sync.WaitGroup
		for i := range getInputs {
			<-permissions
			wg.Add(1)
			go func(input GetInput) {
				defer wg.Done()
				item, err := s.GetFunc(ctx, s.Client, scope, input)

				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
						"input": input,
						"scope": scope,
					}).Error("Error running Get for List item")
				} else {
					itemsChan <- item
				}

				// Give the permission back
				permissions <- struct{}{}
			}(i)
		}

		// Wait for all Gets to finish
		wg.Wait()

		// Close channel as there will be no more items
		close(itemsChan)
	}()

	// Create a process to collect items
	go func() {
		for item := range itemsChan {
			items = append(items, item)
		}

		// Close the doneChan to signal that everything is done
		close(doneChan)
	}()

	paginator := s.ListFuncPaginatorBuilder(s.Client, input)
	var newGetInputs []GetInput

	for paginator.HasMorePages() {
		output, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		newGetInputs, err = s.ListFuncOutputMapper(output, input)

		if err != nil {
			return nil, err
		}

		// Push new queries onto the channel for processing
		for _, q := range newGetInputs {
			getInputs <- q
		}
	}

	// Close queries channel
	close(getInputs)

	// Wait for all processing to be complete
	<-doneChan

	return items, nil
}

// Search Searches for AWS resources by ARN
func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Search(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	if s.SearchInputMapper == nil {
		return s.SearchARN(ctx, scope, query)
	} else {
		return s.SearchCustom(ctx, scope, query)
	}
}

// SearchCustom Searches using custom mapping logic. The SearchInputMapper is
// used to create an input for ListFunc, at which point the usual logic is used
func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) SearchCustom(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	input, err := s.SearchInputMapper(scope, query)

	if err != nil {
		return nil, sdp.NewItemRequestError(err)
	}

	return s.listInternal(ctx, scope, input)
}

func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) SearchARN(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
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

// Weight Returns the priority weighting of items returned by this sourcs.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *ListGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Weight() int {
	return 100
}
