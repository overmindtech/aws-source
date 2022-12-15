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
	// sources as the first element in the scope
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

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *InstanceSource) Scopes() []string {
	return []string{
		fmt.Sprintf("%v.%v", s.AccountID, s.Config.Region),
	}
}

// EC2Client Collects all functions this code uses from the AWS SDK, for test replacement.
type EC2Client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// Get Get a single item with a given scope and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *InstanceSource) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return getImpl(ctx, s.Client(), scope, query)
}

func getImpl(ctx context.Context, client EC2Client, scope string, query string) (*sdp.Item, error) {
	describeInstancesOutput, err := client.DescribeInstances(
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
			Scope:       scope,
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
			Scope:       scope,
		}
	case numReservations == 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("Instance %v not found", query),
			Scope:       scope,
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
			Scope:       scope,
		}
	case numInstances > 1:
		instanceIDs := make([]string, numInstances)

		for i, instance := range reservation.Instances {
			instanceIDs[i] = *instance.InstanceId
		}

		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Request returned > 1 instance. Instance IDs: %v", instanceIDs),
			Scope:       scope,
		}
	}

	// Pull the first instance
	instance := reservation.Instances[0]

	return mapInstanceToItem(instance, scope)
}

// List Lists all items in a given scope
func (s *InstanceSource) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return listImpl(ctx, s.Client(), scope)
}

func listImpl(ctx context.Context, client EC2Client, scope string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)
	instances := make([]types.Instance, 0)
	var maxResults int32 = 100
	var nextToken *string

	for morePages := true; morePages; {
		describeInstancesOutput, err := client.DescribeInstances(
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
				Scope:       scope,
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
		item, _ := mapInstanceToItem(instance, scope)
		items = append(items, item)
	}

	return items, nil
}

func mapInstanceToItem(instance types.Instance, scope string) (*sdp.Item, error) {
	attrs, err := sources.ToAttributesCase(instance)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	item := sdp.Item{
		Type:            "ec2-instance",
		UniqueAttribute: "instanceId",
		Scope:           scope,
		Attributes:      attrs,
	}

	if instance.ImageId != nil {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "ec2-image",
			Method: sdp.RequestMethod_GET,
			Query:  *instance.ImageId,
			Scope:  scope,
		})
	}

	for _, nic := range instance.NetworkInterfaces {
		// IPs
		for _, ip := range nic.Ipv6Addresses {
			if ip.Ipv6Address != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *ip.Ipv6Address,
					Scope:  "global",
				})
			}
		}

		for _, ip := range nic.PrivateIpAddresses {
			if ip.PrivateIpAddress != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *ip.PrivateIpAddress,
					Scope:  "global",
				})
			}
		}

		// Subnet
		if nic.SubnetId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-subnet",
				Method: sdp.RequestMethod_GET,
				Query:  *nic.SubnetId,
				Scope:  scope,
			})
		}

		// VPC
		if nic.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-vpc",
				Method: sdp.RequestMethod_GET,
				Query:  *nic.VpcId,
				Scope:  scope,
			})
		}
	}

	if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "dns",
			Method: sdp.RequestMethod_GET,
			Query:  *instance.PublicDnsName,
			Scope:  "global",
		})
	}

	if instance.PublicIpAddress != nil {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "ip",
			Method: sdp.RequestMethod_GET,
			Query:  *instance.PublicIpAddress,
			Scope:  "global",
		})
	}

	// Security groups
	for _, group := range instance.SecurityGroups {
		if group.GroupId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-securitygroup",
				Method: sdp.RequestMethod_GET,
				Query:  *group.GroupId,
				Scope:  scope,
			})
		}
	}

	return &item, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *InstanceSource) Weight() int {
	return 100
}
