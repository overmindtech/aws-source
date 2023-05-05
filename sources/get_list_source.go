package sources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/overmindtech/sdp-go"
)

// GetListSource A source for AWS APIs where the Get and List functions both
// return the full item, such as many of the IAM APIs
type GetListSource[AWSItem AWSItemType, ClientStruct ClientStructType, Options OptionsType] struct {
	ItemType               string        // The type of items that will be returned
	Client                 ClientStruct  // The AWS API client
	AccountID              string        // The AWS account ID
	Region                 string        // The AWS region this is related to
	SupportGlobalResources bool          // If true, this will also support resources in the "aws" scope which are global
	CacheDuration          time.Duration // How long to cache items for

	// Disables List(), meaning all calls will return empty results. This does
	// not affect Search()
	DisableList bool

	// GetFunc Gets the details of a specific item, returns the AWS
	// representation of that item, and an error
	GetFunc func(ctx context.Context, client ClientStruct, scope string, query string) (AWSItem, error)

	// ListFunc Lists all items that it can find. Returning a slice of AWS items
	ListFunc func(ctx context.Context, client ClientStruct, scope string) ([]AWSItem, error)

	// Optional search func that will be used for Search Requests. If this is
	// unset, Search will simply use ARNs
	SearchFunc func(ctx context.Context, client ClientStruct, scope string, query string) ([]AWSItem, error)

	// ItemMapper Maps an AWS representation of an item to the SDP version
	ItemMapper func(scope string, awsItem AWSItem) (*sdp.Item, error)
}

// Validate Checks that the source has been set up correctly
func (s *GetListSource[AWSItem, ClientStruct, Options]) Validate() error {
	if s.GetFunc == nil {
		return errors.New("GetFunc is nil")
	}

	if !s.DisableList {
		if s.ListFunc == nil {
			return errors.New("ListFunc is nil")
		}
	}

	if s.ItemMapper == nil {
		return errors.New("ItemMapper is nil")
	}

	return nil
}

func (s *GetListSource[AWSItem, ClientStruct, Options]) Type() string {
	return s.ItemType
}

func (s *GetListSource[AWSItem, ClientStruct, Options]) Name() string {
	return fmt.Sprintf("%v-source", s.ItemType)
}

// DefaultCacheDuration Returns the default cache duration for this source
func (s *GetListSource[AWSItem, ClientStruct, Options]) DefaultCacheDuration() time.Duration {
	if s.CacheDuration == 0 {
		return 10 * time.Minute
	}

	return s.CacheDuration
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *GetListSource[AWSItem, ClientStruct, Options]) Scopes() []string {
	scopes := make([]string, 0)

	scopes = append(scopes, FormatScope(s.AccountID, s.Region))

	if s.SupportGlobalResources {
		scopes = append(scopes, "aws")
	}

	return scopes
}

// hasScope Returns whether or not this source has the given scope
func (s *GetListSource[AWSItem, ClientStruct, Options]) hasScope(scope string) bool {
	if scope == "aws" && s.SupportGlobalResources {
		// There is a special global "account" that is used for global resources
		// called "aws"
		return true
	}

	for _, s := range s.Scopes() {
		if s == scope {
			return true
		}
	}

	return false
}

func (s *GetListSource[AWSItem, ClientStruct, Options]) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if !s.hasScope(scope) {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	awsItem, err := s.GetFunc(ctx, s.Client, scope, query)

	if err != nil {
		return nil, WrapAWSError(err)
	}

	item, err := s.ItemMapper(scope, awsItem)

	if err != nil {
		return nil, WrapAWSError(err)
	}

	return item, nil
}

// List Lists all available items. This is done by running the ListFunc, then
// passing these results to GetFunc in order to get the details
func (s *GetListSource[AWSItem, ClientStruct, Options]) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if !s.hasScope(scope) {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	if s.DisableList {
		return []*sdp.Item{}, nil
	}

	awsItems, err := s.ListFunc(ctx, s.Client, scope)

	if err != nil {
		return nil, WrapAWSError(err)
	}

	items := make([]*sdp.Item, 0)

	var item *sdp.Item

	for _, awsItem := range awsItems {
		item, err = s.ItemMapper(scope, awsItem)

		if err != nil {
			continue
		}

		items = append(items, item)
	}

	return items, nil
}

// Search Searches for AWS resources by ARN
func (s *GetListSource[AWSItem, ClientStruct, Options]) Search(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	if !s.hasScope(scope) {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
		}
	}

	if s.SearchFunc != nil {
		return s.SearchCustom(ctx, scope, query)
	} else {
		return s.SearchARN(ctx, scope, query)
	}
}

func (s *GetListSource[AWSItem, ClientStruct, Options]) SearchARN(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
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

	item, err := s.Get(ctx, scope, a.ResourceID())

	if err != nil {
		return nil, WrapAWSError(err)
	}

	return []*sdp.Item{item}, nil
}

func (s *GetListSource[AWSItem, ClientStruct, Options]) SearchCustom(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	awsItems, err := s.SearchFunc(ctx, s.Client, scope, query)

	if err != nil {
		return nil, WrapAWSError(err)
	}

	items := make([]*sdp.Item, 0)
	var item *sdp.Item

	for _, awsItem := range awsItems {
		item, err = s.ItemMapper(scope, awsItem)

		if err != nil {
			continue
		}

		items = append(items, item)
	}

	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *GetListSource[AWSItem, ClientStruct, Options]) Weight() int {
	return 100
}
