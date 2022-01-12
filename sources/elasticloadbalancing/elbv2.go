package elasticloadbalancing

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
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
	return "elasticloadbalancing_loadbalancer_v2"
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
			ErrorString: "elasticloadbalancing_loadbalancer_v1 not found",
			Context:     itemContext,
		}
	case 1:
		expanded, err := s.ExpandLB(ctx, lbs.LoadBalancers[0])

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: fmt.Sprintf("error during details expansion: %v", err.Error()),
				Context:     itemContext,
			}
		}

		return mapExpandedELBv2ToItem(expanded, itemContext)
	default:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("more than 1 elasticloadbalancing_loadbalancer_v1 found, found: %v", len(lbs.LoadBalancers)),
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
		expanded, err := s.ExpandLB(ctx, lb)

		if err != nil {
			continue
		}

		var item *sdp.Item

		item, err = mapExpandedELBv2ToItem(expanded, itemContext)

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
// sen on, so the one with the higher weight value will win.
func (s *ELBv2Source) Weight() int {
	return 100
}

type ExpandedTargetGroup struct {
	types.TargetGroup

	TargetHealthDescriptions []types.TargetHealthDescription
}

type ExpandedELBv2 struct {
	types.LoadBalancer

	Listeners    []types.Listener
	TargetGroups []ExpandedTargetGroup
}

func (s *ELBv2Source) ExpandLB(ctx context.Context, lb types.LoadBalancer) (*ExpandedELBv2, error) {
	var listenersOutput *elbv2.DescribeListenersOutput
	var targetGroupsOutput *elbv2.DescribeTargetGroupsOutput
	var targetHealthOutput *elbv2.DescribeTargetHealthOutput
	var err error

	// Copy all fields from LB
	expandedELB := ExpandedELBv2{
		LoadBalancer: lb,
	}

	// Get listeners
	listenersOutput, err = s.Client().DescribeListeners(
		ctx,
		&elbv2.DescribeListenersInput{
			LoadBalancerArn: lb.LoadBalancerArn,
		},
	)

	if err != nil {
		return nil, err
	}

	expandedELB.Listeners = listenersOutput.Listeners

	// Get target groups
	targetGroupsOutput, err = s.Client().DescribeTargetGroups(
		ctx,
		&elbv2.DescribeTargetGroupsInput{
			LoadBalancerArn: lb.LoadBalancerArn,
		},
	)

	if err != nil {
		return nil, err
	}

	expandedELB.TargetGroups = make([]ExpandedTargetGroup, 0)

	// For each target group get targets and their health
	for _, tg := range targetGroupsOutput.TargetGroups {
		etg := ExpandedTargetGroup{
			TargetGroup: tg,
		}

		targetHealthOutput, err = s.Client().DescribeTargetHealth(
			ctx,
			&elbv2.DescribeTargetHealthInput{
				TargetGroupArn: tg.TargetGroupArn,
			},
		)

		if err != nil {
			return nil, err
		}

		etg.TargetHealthDescriptions = targetHealthOutput.TargetHealthDescriptions

		expandedELB.TargetGroups = append(expandedELB.TargetGroups, etg)
	}

	return &expandedELB, nil
}

// mapExpandedELBv2ToItem Maps a load balancer to an item
func mapExpandedELBv2ToItem(lb *ExpandedELBv2, itemContext string) (*sdp.Item, error) {
	attrMap := make(map[string]interface{})

	if lb.LoadBalancerName == nil || *lb.LoadBalancerName == "" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: "elasticloadbalancing_loadbalancer_v1 was returned with an empty name",
			Context:     itemContext,
		}
	}

	item := sdp.Item{
		Type:            "elasticloadbalancing_loadbalancer_v2",
		UniqueAttribute: "name",
		Context:         itemContext,
	}

	attrMap["name"] = lb.LoadBalancerName
	attrMap["availabilityZones"] = lb.AvailabilityZones
	attrMap["ipAddressType"] = lb.IpAddressType
	attrMap["scheme"] = lb.Scheme
	attrMap["securityGroups"] = lb.SecurityGroups
	attrMap["type"] = lb.Type
	attrMap["listeners"] = lb.Listeners
	attrMap["targetGroups"] = lb.TargetGroups
	attrMap["canonicalHostedZoneId"] = lb.CanonicalHostedZoneId
	attrMap["loadBalancerArn"] = lb.LoadBalancerArn
	attrMap["customerOwnedIpv4Pool"] = lb.CustomerOwnedIpv4Pool
	attrMap["state"] = lb.State

	if lb.CreatedTime != nil {
		attrMap["createdTime"] = lb.CreatedTime.String()
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

	if lb.VpcId != nil {
		attrMap["vpcId"] = lb.VpcId

		// TODO: Linked item request to VPC
	}

	attributes, err := sources.ToAttributesCase(attrMap)

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
