package ec2

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type InstanceSource struct {
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

func (s *InstanceSource) Client() *ec2.Client {
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
func (s *InstanceSource) Type() string {
	return "ec2-instance"
}

// Descriptive name for the source, used in logging and metadata
func (s *InstanceSource) Name() string {
	return "ec2-aws-source"
}

// List of contexts that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *InstanceSource) Contexts() []string {
	return []string{
		fmt.Sprintf("%v.%v", s.AccountID, s.Config.Region),
	}
}

// Get Get a single item with a given context and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *InstanceSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != s.Contexts()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("requested context %v does not match source context %v", itemContext, s.Contexts()[0]),
			Context:     itemContext,
		}
	}

	describeInstancesOutput, err := s.Client().DescribeInstances(
		ctx,
		&ec2.DescribeInstancesInput{
			InstanceIds: []string{
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

	numReservations := len(describeInstancesOutput.Reservations)

	switch {
	case numReservations > 1:
		reservationIDs := make([]string, numReservations)

		for i, reservation := range describeInstancesOutput.Reservations {
			reservationIDs[i] = *reservation.ReservationId
		}

		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Request returned > 1 reservation, cannot determine instance. Reservations: %v", reservationIDs),
			Context:     itemContext,
		}
	case numReservations == 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("Instance %v not found", query),
			Context:     itemContext,
		}
	}

	// Pull out the first and only reservation
	reservation := describeInstancesOutput.Reservations[0]

	numInstances := len(reservation.Instances)

	switch {
	case numInstances == 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("Instance %v not found", query),
			Context:     itemContext,
		}
	case numInstances > 1:
		instanceIDs := make([]string, numInstances)

		for i, instance := range reservation.Instances {
			instanceIDs[i] = *instance.InstanceId
		}

		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Request returned > 1 instance. Instance IDs: %v", instanceIDs),
			Context:     itemContext,
		}
	}

	// Pull the first instance
	instance := reservation.Instances[0]

	return mapInstanceToItem(instance, itemContext)
}

// Find Finds all items in a given context
func (s *InstanceSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != s.Contexts()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("requested context %v does not match source context %v", itemContext, s.Contexts()[0]),
			Context:     itemContext,
		}
	}

	items := make([]*sdp.Item, 0)
	instances := make([]types.Instance, 0)
	var maxResults int32 = 100
	var nextToken *string

	for morePages := true; morePages; {
		describeInstancesOutput, err := s.Client().DescribeInstances(
			ctx,
			&ec2.DescribeInstancesInput{
				MaxResults: &maxResults,
				NextToken:  nextToken,
			},
		)

		if err != nil {
			return items, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Context:     itemContext,
			}
		}

		for _, reservation := range describeInstancesOutput.Reservations {
			instances = append(instances, reservation.Instances...)
		}

		// If there is more data we should store the token so that we can use
		// that. We also need to set morePages to true so that the loop runs
		// again
		nextToken = describeInstancesOutput.NextToken
		morePages = (nextToken != nil)
	}

	// Convert to items
	for _, instance := range instances {
		item, _ := mapInstanceToItem(instance, itemContext)
		items = append(items, item)
	}

	return items, nil
}

func mapInstanceToItem(instance types.Instance, itemContext string) (*sdp.Item, error) {
	var err error
	var attrs *sdp.ItemAttributes
	attrs, err = sources.ToAttributesCase(instance)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	item := sdp.Item{
		Type:            "ec2-instance",
		UniqueAttribute: "instanceId",
		Context:         itemContext,
		Attributes:      attrs,
	}

	if instance.ImageId != nil {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "ec2-image",
			Method:  sdp.RequestMethod_GET,
			Query:   *instance.ImageId,
			Context: itemContext,
		})
	}

	for _, nic := range instance.NetworkInterfaces {
		// IPs
		for _, ip := range nic.Ipv6Addresses {
			if ip.Ipv6Address != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:    "ip",
					Method:  sdp.RequestMethod_GET,
					Query:   *ip.Ipv6Address,
					Context: "global",
				})
			}
		}

		for _, ip := range nic.PrivateIpAddresses {
			if ip.PrivateIpAddress != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:    "ip",
					Method:  sdp.RequestMethod_GET,
					Query:   *ip.PrivateIpAddress,
					Context: "global",
				})
			}
		}

		// Subnet
		if nic.SubnetId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:    "ec2-subnet",
				Method:  sdp.RequestMethod_GET,
				Query:   *nic.SubnetId,
				Context: itemContext,
			})
		}

		// VPC
		if nic.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:    "ec2-vpc",
				Method:  sdp.RequestMethod_GET,
				Query:   *nic.VpcId,
				Context: itemContext,
			})
		}
	}

	if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "dns",
			Method:  sdp.RequestMethod_GET,
			Query:   *instance.PublicDnsName,
			Context: "global",
		})
	}

	if instance.PublicIpAddress != nil {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "ip",
			Method:  sdp.RequestMethod_GET,
			Query:   *instance.PublicIpAddress,
			Context: "global",
		})
	}

	// Security groups
	for _, group := range instance.SecurityGroups {
		if group.GroupId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:    "ec2-securitygroup",
				Method:  sdp.RequestMethod_GET,
				Query:   *group.GroupId,
				Context: itemContext,
			})
		}
	}

	return &item, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// sen on, so the one with the higher weight value will win.
func (s *InstanceSource) Weight() int {
	return 100
}
