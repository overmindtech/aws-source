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
type ListGetSource[Input InputType, Output OutputType, ClientStruct ClientStructType, Options OptionsType] struct {
	ItemType    string       // The type of items to return
	Client      ClientStruct // The AWS API client
	AccountID   string       // The AWS account ID
	Region      string       // The AWS region this is related to
	MaxParallel MaxParallel  // How many Get request to run in parallel for a single List request

	// The input to the ListFunc
	ListInput Input

	// A function that returns a paginator for the ListFunc
	ListFuncPaginatorBuilder func(client ClientStruct, input Input) Paginator[Output, Options]

	// A function that accepts the output of a ListFunc and maps this to a slice
	// of uniqueAttributeValues which will be passed to GetFunc
	ListFuncOutputMapper func(output Output) ([]string, error)

	// A function that gets the details of a given item
	GetFunc func(ctx context.Context, scope, query string) (*sdp.Item, error)
}

// Validate Checks that the source has been set up correctly
func (s *ListGetSource[Input, Output, ClientStruct, Options]) Validate() error {
	if s.ListFuncPaginatorBuilder == nil {
		return errors.New("ListFuncPaginatorBuilder is nil")
	}

	if s.ListFuncOutputMapper == nil {
		return errors.New("ListFuncOutputMapper is nil")
	}

	if s.GetFunc == nil {
		return errors.New("GetFunc is nil")
	}

	return nil
}

func (s *ListGetSource[Input, Output, ClientStruct, Options]) Type() string {
	return s.ItemType
}

func (s *ListGetSource[Input, Output, ClientStruct, Options]) Name() string {
	return fmt.Sprintf("%v-source", s.ItemType)
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *ListGetSource[Input, Output, ClientStruct, Options]) Scopes() []string {
	return []string{
		FormatScope(s.AccountID, s.Region),
	}
}

func (s *ListGetSource[Input, Output, ClientStruct, Options]) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
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

	item, err = s.GetFunc(ctx, scope, query)

	if err != nil {
		// TODO: How can we handle NOTFOUND?
		return nil, sdp.NewItemRequestError(err)
	}

	return item, nil
}

// List Lists all available items. This is done by running the ListFunc, then
// passing these results to GetFunc in order to get the details
func (s *ListGetSource[Input, Output, ClientStruct, Options]) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	var output Output
	var err error
	var newItemQueries []string
	items := make([]*sdp.Item, 0)
	itemsChan := make(chan *sdp.Item)
	queries := make(chan string)
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
		for q := range queries {
			<-permissions
			wg.Add(1)
			go func(query string) {
				defer wg.Done()
				item, err := s.GetFunc(ctx, scope, query)

				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
						"query": query,
						"scope": scope,
					}).Error("Error running Get for List item")
				} else {
					itemsChan <- item
				}

				// Give the permission back
				permissions <- struct{}{}
			}(q)
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

	paginator := s.ListFuncPaginatorBuilder(s.Client, s.ListInput)

	for paginator.HasMorePages() {
		output, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		newItemQueries, err = s.ListFuncOutputMapper(output)

		if err != nil {
			return nil, err
		}

		// Push new queries onto the channel for processing
		for _, q := range newItemQueries {
			queries <- q
		}
	}

	// Close queries channel
	close(queries)

	// Wait for all processing to be complete
	<-doneChan

	return items, nil
}

// Search Searches for AWS resources by ARN
func (s *ListGetSource[Input, Output, ClientStruct, Options]) Search(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
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

// Weight Returns the priority weighting of items returned by this sourcs.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *ListGetSource[Input, Output, ClientStruct, Options]) Weight() int {
	return 100
}
