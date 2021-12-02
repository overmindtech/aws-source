package elasticloadbalancing

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/sdp-go"
)

type ELBv2Source struct {
	// Config AWS Config including region and credentials
	Config aws.Config

	// AccountID The id of the account that is being used. This is used by
	// sources as the first element in the context
	AccountID string

	// client The AWS client to use when making requests
	client        *elbv2.Client
	clientCreated bool
	clientMutex   sync.Mutex
}

func (s *ELBv2Source) Client() *elbv2.Client {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	// If the client already exists then return it
	if s.clientCreated {
		return s.client
	}

	// Otherwise create a new client from the config
	s.client = elbv2.NewFromConfig(s.Config)
	s.clientCreated = true

	return s.client
}

// Type The type of items that this source is capable of finding
func (s *ELBv2Source) Type() string {
	return "elasticloadbalancerv2"
}

// Descriptive name for the source, used in logging and metadata
func (s *ELBv2Source) Name() string {
	return "elasticloadbalancing-v2-aws-source"
}

// List of contexts that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *ELBv2Source) Contexts() []string {
	return []string{
		fmt.Sprintf("%v.%v", s.AccountID, s.Config.Region),
	}
}

// Get Get a single item with a given context and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *ELBv2Source) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != s.Contexts()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("requested context %v does not match source context %v", itemContext, s.Contexts()[0]),
			Context:     itemContext,
		}
	}

	lbs, err := s.Client().DescribeLoadBalancers(
		ctx,
		&elbv2.DescribeLoadBalancersInput{
			Names: []string{
				query,
			},
		},
	)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	switch len(lbs.LoadBalancers) {
	case 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: "elasticloadbalancer not found",
			Context:     itemContext,
		}
	case 1:
		return mapElasticLoadBalancerV2ToItem(lbs.LoadBalancers[0], itemContext)
	default:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("more than 1 elasticloadbalancer found, found: %v", len(lbs.LoadBalancers)),
			Context:     itemContext,
		}
	}
}

// Find Finds all items in a given context
func (s *ELBv2Source) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != s.Contexts()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("requested context %v does not match source context %v", itemContext, s.Contexts()[0]),
			Context:     itemContext,
		}
	}

	items := make([]*sdp.Item, 0)

	lbs, err := s.Client().DescribeLoadBalancers(
		ctx,
		&elbv2.DescribeLoadBalancersInput{},
	)

	if err != nil {
		return items, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	for _, lb := range lbs.LoadBalancers {
		item, err := mapElasticLoadBalancerV2ToItem(lb, itemContext)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// sen on, so the one with the higher weight value will win.
func (s *ELBv2Source) Weight() int {
	return 100
}

// mapElasticLoadBalancerV2ToItem Maps a load balancer to an item
func mapElasticLoadBalancerV2ToItem(lb types.LoadBalancer, itemContext string) (*sdp.Item, error) {
	attrMap := make(map[string]interface{})

	if lb.LoadBalancerName == nil || *lb.LoadBalancerName == "" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: "elasticloadbalancer was returned with an empty name",
			Context:     itemContext,
		}
	}

	item := sdp.Item{
		Type:            "elasticloadbalancerv2",
		UniqueAttribute: "name",
		Context:         itemContext,
	}

	attrMap["name"] = lb.LoadBalancerName
	attrMap["availabilityZones"] = lb.AvailabilityZones
	attrMap["ipAddressType"] = lb.IpAddressType
	attrMap["scheme"] = lb.Scheme
	attrMap["securityGroups"] = lb.SecurityGroups
	attrMap["type"] = lb.Type

	if lb.CanonicalHostedZoneId != nil {
		attrMap["canonicalHostedZoneId"] = lb.CanonicalHostedZoneId
	}

	if lb.CreatedTime != nil {
		attrMap["createdTime"] = lb.CreatedTime.String()
	}

	if lb.CustomerOwnedIpv4Pool != nil {
		attrMap["customerOwnedIpv4Pool"] = lb.CustomerOwnedIpv4Pool
	}

	if lb.DNSName != nil {
		attrMap["dNSName"] = lb.DNSName

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "dns",
			Method:  sdp.RequestMethod_GET,
			Query:   *lb.DNSName,
			Context: "global",
		})
	}

	if lb.LoadBalancerArn != nil {
		attrMap["loadBalancerArn"] = lb.LoadBalancerArn
	}

	if lb.State != nil {
		attrMap["state"] = lb.State
	}

	if lb.VpcId != nil {
		attrMap["vpcId"] = lb.VpcId

		// TODO: Linked item request to VPC
	}

	attributes, err := sdp.ToAttributes(attrMap)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("error creating attributes: %v", err),
			Context:     itemContext,
		}
	}

	item.Attributes = attributes

	return &item, nil
}
