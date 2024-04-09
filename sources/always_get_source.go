package sources

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdpcache"
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

// AlwaysGetSource This source is designed for AWS APIs that have separate List
// and Get functions. It also assumes that the results of the list function
// cannot be converted directly into items as they do not contain enough
// information, and therefore they always need to be passed to the Get function
// before returning. An example is the `ListClusters` API in EKS which returns a
// list of cluster names.
type AlwaysGetSource[ListInput InputType, ListOutput OutputType, GetInput InputType, GetOutput OutputType, ClientStruct ClientStructType, Options OptionsType] struct {
	ItemType    string       // The type of items to return
	Client      ClientStruct // The AWS API client
	AccountID   string       // The AWS account ID
	Region      string       // The AWS region this is related to
	MaxParallel MaxParallel  // How many Get request to run in parallel for a single List request

	// Disables List(), meaning all calls will return empty results. This does
	// not affect Search()
	DisableList bool

	// A function that gets the details of a given item. This should include the
	// tags if relevant
	GetFunc func(ctx context.Context, client ClientStruct, scope string, input GetInput) (*sdp.Item, error)

	// The input to the ListFunc. This is static
	ListInput ListInput

	// A function that maps from the SDP get inputs to the relevant input for
	// the GetFunc
	GetInputMapper func(scope, query string) GetInput

	// If this is set, Search queries will always use the automatic ARN resolver
	// if the input is an ARN, falling back to the `SearchInputMapper` if it
	// isn't
	AlwaysSearchARNs bool

	// Maps search terms from an SDP Search request into the relevant input for
	// the ListFunc. If this is not set, Search() will handle ARNs like most AWS
	// sources. Note that this and `SearchGetInputMapper` are mutually exclusive
	SearchInputMapper func(scope, query string) (ListInput, error)

	// Maps search terms from an SDP Search request into the relevant input for
	// the GetFunc. If this is not set, Search() will handle ARNs like most AWS
	// sources. Note that this and `SearchInputMapper` are mutually exclusive
	SearchGetInputMapper func(scope, query string) (GetInput, error)

	// A function that returns a paginator for the ListFunc
	ListFuncPaginatorBuilder func(client ClientStruct, input ListInput) Paginator[ListOutput, Options]

	// A function that accepts the output of a ListFunc and maps this to a slice
	// of inputs to pass to the GetFunc. The input used for the ListFunc is also
	// included in case it is required
	ListFuncOutputMapper func(output ListOutput, input ListInput) ([]GetInput, error)

	CacheDuration time.Duration   // How long to cache items for
	cache         *sdpcache.Cache // The sdpcache of this source
	cacheInitMu   sync.Mutex      // Mutex to ensure cache is only initialised once
}

func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) cacheDuration() time.Duration {
	if s.CacheDuration == 0 {
		return DefaultCacheDuration
	}

	return s.CacheDuration
}

func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) ensureCache() {
	s.cacheInitMu.Lock()
	defer s.cacheInitMu.Unlock()

	if s.cache == nil {
		s.cache = sdpcache.NewCache()
	}
}

func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Cache() *sdpcache.Cache {
	s.ensureCache()
	return s.cache
}

// Validate Checks that the source has been set up correctly
func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Validate() error {
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

	if s.SearchGetInputMapper != nil && s.SearchInputMapper != nil {
		return errors.New("SearchGetInputMapper and SearchInputMapper are mutually exclusive")
	}

	return nil
}

func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Type() string {
	return s.ItemType
}

func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Name() string {
	return fmt.Sprintf("%v-source", s.ItemType)
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Scopes() []string {
	return []string{
		FormatScope(s.AccountID, s.Region),
	}
}

func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	var err error
	var item *sdp.Item

	if err = s.Validate(); err != nil {
		return nil, WrapAWSError(err)
	}

	s.ensureCache()
	cacheHit, ck, cachedItems, qErr := s.cache.Lookup(ctx, s.Name(), sdp.QueryMethod_GET, scope, s.ItemType, query, ignoreCache)
	if qErr != nil {
		return nil, qErr
	}
	if cacheHit {
		if len(cachedItems) > 0 {
			return cachedItems[0], nil
		} else {
			return nil, nil
		}
	}

	input := s.GetInputMapper(scope, query)

	item, err = s.GetFunc(ctx, s.Client, scope, input)

	if err != nil {
		err = s.processError(err, ck)
		return nil, err
	}

	s.cache.StoreItem(item, s.cacheDuration(), ck)
	return item, nil
}

// List Lists all available items. This is done by running the ListFunc, then
// passing these results to GetFunc in order to get the details
func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	// Check to see if we have supplied the required functions
	if s.DisableList {
		// In this case we can't run list, so just return empty
		return []*sdp.Item{}, nil
	}

	s.ensureCache()
	cacheHit, ck, cachedItems, qErr := s.cache.Lookup(ctx, s.Name(), sdp.QueryMethod_LIST, scope, s.ItemType, "", ignoreCache)
	if qErr != nil {
		return nil, qErr
	}
	if cacheHit {
		return cachedItems, nil
	}

	items, err := s.listInternal(ctx, scope, s.ListInput)
	if err != nil {
		err = s.processError(err, ck)
		return nil, err
	}

	for _, item := range items {
		s.cache.StoreItem(item, s.cacheDuration(), ck)
	}

	return items, nil
}

// listInternal Accepts a ListInput and runs the List logic against it
func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) listInternal(ctx context.Context, scope string, input ListInput) ([]*sdp.Item, error) {
	var output ListOutput
	var err error
	items := make([]*sdp.Item, 0)
	itemsChan := make(chan *sdp.Item)
	getInputs := make(chan GetInput)
	doneChan := make(chan struct{})

	if err = s.Validate(); err != nil {
		return nil, WrapAWSError(err)
	}

	// Create a channel of permissions to allow only a certain number of Get requests to tun in parallel
	permissions := make(chan struct{}, s.MaxParallel.Value())
	for i := 0; i < s.MaxParallel.Value(); i++ {
		permissions <- struct{}{}
	}

	// Create a process to take queries and run them using Get
	go func() {
		defer sentry.Recover()
		var wg sync.WaitGroup
		for i := range getInputs {
			<-permissions
			wg.Add(1)
			go func(input GetInput) {
				defer sentry.Recover()
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
		defer sentry.Recover()
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
func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	ck := sdpcache.CacheKeyFromParts(s.Name(), sdp.QueryMethod_SEARCH, scope, s.ItemType, query)

	var items []*sdp.Item
	var err error

	if s.SearchInputMapper == nil && s.SearchGetInputMapper == nil {
		items, err = s.SearchARN(ctx, scope, query, ignoreCache)
	} else {
		// If we should always look for ARNs first, do that
		if s.AlwaysSearchARNs {
			if _, err = ParseARN(query); err == nil {
				items, err = s.SearchARN(ctx, scope, query, ignoreCache)
			} else {
				items, err = s.SearchCustom(ctx, scope, query)
			}
		} else {
			items, err = s.SearchCustom(ctx, scope, query)
		}
	}

	if err != nil {
		err = s.processError(err, ck)
		return nil, err
	}

	for _, item := range items {
		s.cache.StoreItem(item, s.cacheDuration(), ck)
	}

	return items, nil
}

// SearchCustom Searches using custom mapping logic. The SearchInputMapper is
// used to create an input for ListFunc, at which point the usual logic is used
func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) SearchCustom(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	var items []*sdp.Item

	ck := sdpcache.CacheKeyFromParts(s.Name(), sdp.QueryMethod_SEARCH, scope, s.ItemType, query)

	if s.SearchInputMapper != nil {
		input, err := s.SearchInputMapper(scope, query)

		if err != nil {
			err = s.processError(err, ck)
			return nil, err
		}

		items, err = s.listInternal(ctx, scope, input)

		if err != nil {
			err = s.processError(err, ck)
			return nil, err
		}
	} else if s.SearchGetInputMapper != nil {
		input, err := s.SearchGetInputMapper(scope, query)

		if err != nil {
			err = s.processError(err, ck)
			return nil, err
		}

		item, err := s.GetFunc(ctx, s.Client, scope, input)

		if err != nil {
			err = s.processError(err, ck)
			return nil, err
		}

		items = []*sdp.Item{item}
	} else {
		return nil, errors.New("SearchCustom called without SearchInputMapper or SearchGetInputMapper")
	}

	for _, item := range items {
		s.cache.StoreItem(item, s.cacheDuration(), ck)
	}
	return items, nil
}

func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) SearchARN(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	// Parse the ARN
	a, err := ParseARN(query)

	if err != nil {
		return nil, WrapAWSError(err)
	}

	if arnScope := FormatScope(a.AccountID, a.Region); arnScope != scope {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("ARN scope %v does not match request scope %v", arnScope, scope),
			Scope:       scope,
		}
	}

	item, err := s.Get(ctx, scope, a.ResourceID(), ignoreCache)
	if err != nil {
		return nil, WrapAWSError(err)
	}

	return []*sdp.Item{item}, nil
}

// Weight Returns the priority weighting of items returned by this sourcs.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) Weight() int {
	return 100
}

// Processes an error returned by the AWS API so that it can be handled by
// Overmind. This includes extracting the correct error type, wrapping in an SDP
// error, and caching that error if it is non-transient (like a 404)
func (s *AlwaysGetSource[ListInput, ListOutput, GetInput, GetOutput, ClientStruct, Options]) processError(err error, cacheKey sdpcache.CacheKey) error {
	var sdpErr *sdp.QueryError

	if err != nil {
		sdpErr = WrapAWSError(err)

		// Only cache the error if is something that won't be fixed by retrying
		if sdpErr.GetErrorType() == sdp.QueryError_NOTFOUND || sdpErr.GetErrorType() == sdp.QueryError_NOSCOPE {
			s.cache.StoreError(sdpErr, s.cacheDuration(), cacheKey)
		}
	}

	return sdpErr
}
