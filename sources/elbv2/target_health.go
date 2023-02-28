package elbv2

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type TargetHealthUniqueID struct {
	TargetGroupArn   string
	Id               string
	AvailabilityZone *string
	Port             *int32
}

func (id TargetHealthUniqueID) String() string {
	var az string
	var port string

	if id.AvailabilityZone != nil {
		az = *id.AvailabilityZone
	}

	if id.Port != nil {
		port = fmt.Sprint(*id.Port)
	}

	return strings.Join([]string{
		id.TargetGroupArn,
		id.Id,
		az,
		port,
	}, "|")
}

func ToTargetHealthUniqueID(id string) (TargetHealthUniqueID, error) {
	sections := strings.Split(id, "|")

	if len(sections) != 4 {
		return TargetHealthUniqueID{}, fmt.Errorf("cannot parse TargetHealthUniqueID, must have 4 sections, got %v", len(sections))
	}

	healthId := TargetHealthUniqueID{
		TargetGroupArn: sections[0],
		Id:             sections[1],
	}

	if sections[2] != "" {
		healthId.AvailabilityZone = &sections[2]
	}

	if sections[3] != "" {
		port, err := strconv.Atoi(sections[3])

		if err != nil {
			return TargetHealthUniqueID{}, err
		}

		pint32 := int32(port)

		healthId.Port = &pint32
	}

	return healthId, nil
}

func TargetHealthOutputMapper(scope string, input *elbv2.DescribeTargetHealthInput, output *elbv2.DescribeTargetHealthOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, desc := range output.TargetHealthDescriptions {
		attrs, err := sources.ToAttributesCase(desc)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "elbv2-target-health",
			UniqueAttribute: "uniqueId",
			Attributes:      attrs,
			Scope:           scope,
		}

		// Check that we have an input and not a nil pointer
		if input == nil {
			return nil, fmt.Errorf("input cannot be nil")
		}

		if input.TargetGroupArn == nil {
			return nil, fmt.Errorf("target group ARN cannot be nil")
		}

		// Make sure there is actually a target in this result, there always
		// should be but safer to check
		if desc.Target == nil {
			continue
		}

		if desc.Target.Id == nil {
			continue
		}

		id := TargetHealthUniqueID{
			TargetGroupArn:   *input.TargetGroupArn,
			Id:               *desc.Target.Id,
			AvailabilityZone: desc.Target.AvailabilityZone,
			Port:             desc.Target.Port,
		}

		item.Attributes.Set("uniqueId", id.String())

		// See if the ID is an ARN
		a, err := sources.ParseARN(*desc.Target.Id)

		if err == nil {
			switch a.Service {
			case "lambda":
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "lambda-function",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *desc.Target.Id,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			case "elasticloadbalancing":
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "elbv2-load-balancer",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *desc.Target.Id,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		} else {
			// In this case it could be an instance ID or an IP. We will check
			// for IP first
			if net.ParseIP(*desc.Target.Id) != nil {
				// This means it's an IP
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *desc.Target.Id,
					Scope:  "global",
				})
			} else {
				// If all else fails it must be an instance ID
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-instance",
					Method: sdp.RequestMethod_GET,
					Query:  *desc.Target.Id,
					Scope:  scope,
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewTargetHealthSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elbv2.DescribeTargetHealthInput, *elbv2.DescribeTargetHealthOutput, *elbv2.Client, *elbv2.Options] {
	return &sources.DescribeOnlySource[*elbv2.DescribeTargetHealthInput, *elbv2.DescribeTargetHealthOutput, *elbv2.Client, *elbv2.Options]{
		Config:    config,
		Client:    elbv2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elbv2-target-health",
		DescribeFunc: func(ctx context.Context, client *elbv2.Client, input *elbv2.DescribeTargetHealthInput) (*elbv2.DescribeTargetHealthOutput, error) {
			return client.DescribeTargetHealth(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elbv2.DescribeTargetHealthInput, error) {
			id, err := ToTargetHealthUniqueID(query)

			if err != nil {
				return nil, err
			}

			return &elbv2.DescribeTargetHealthInput{
				TargetGroupArn: &id.TargetGroupArn,
				Targets: []types.TargetDescription{
					{
						Id:               &id.Id,
						AvailabilityZone: id.AvailabilityZone,
						Port:             id.Port,
					},
				},
			}, nil
		},
		InputMapperList: func(scope string) (*elbv2.DescribeTargetHealthInput, error) {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: "list not supported for elbv2-target-health, use search",
			}
		},
		InputMapperSearch: func(ctx context.Context, client *elbv2.Client, scope, query string) (*elbv2.DescribeTargetHealthInput, error) {
			// Search by target group ARN
			return &elbv2.DescribeTargetHealthInput{
				TargetGroupArn: &query,
			}, nil
		},
		OutputMapper: TargetHealthOutputMapper,
	}
}
