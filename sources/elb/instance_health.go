package elb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// InstanceHealthName Structured representation of an instance health's unique
// name
type InstanceHealthName struct {
	LoadBalancerName string
	InstanceId       string
}

func (i InstanceHealthName) String() string {
	return fmt.Sprintf("%v/%v", i.LoadBalancerName, i.InstanceId)
}

func ParseInstanceName(name string) (InstanceHealthName, error) {
	sections := strings.Split(name, "/")

	if len(sections) != 2 {
		return InstanceHealthName{}, errors.New("instance health name did not have 2 sections separated by a forward slash")
	}

	return InstanceHealthName{
		LoadBalancerName: sections[0],
		InstanceId:       sections[1],
	}, nil
}

func InstanceHealthOutputMapper(scope string, _ *elb.DescribeInstanceHealthInput, output *elb.DescribeInstanceHealthOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, is := range output.InstanceStates {
		attrs, err := sources.ToAttributesCase(is)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "elb-instance-health",
			UniqueAttribute: "instanceId",
			Attributes:      attrs,
			Scope:           scope,
		}

		if is.InstanceId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-instance",
				Method: sdp.RequestMethod_GET,
				Query:  *is.InstanceId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewInstanceHealthSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elb.DescribeInstanceHealthInput, *elb.DescribeInstanceHealthOutput, *elb.Client, *elb.Options] {
	return &sources.DescribeOnlySource[*elb.DescribeInstanceHealthInput, *elb.DescribeInstanceHealthOutput, *elb.Client, *elb.Options]{
		Config:    config,
		Client:    elb.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elb-instance-health",
		DescribeFunc: func(ctx context.Context, client *elb.Client, input *elb.DescribeInstanceHealthInput) (*elb.DescribeInstanceHealthOutput, error) {
			return client.DescribeInstanceHealth(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elb.DescribeInstanceHealthInput, error) {
			// This has a composite name defined by `InstanceHealthName`
			name, err := ParseInstanceName(query)

			if err != nil {
				return nil, err
			}

			return &elb.DescribeInstanceHealthInput{
				LoadBalancerName: &name.LoadBalancerName,
				Instances: []types.Instance{
					{
						InstanceId: &name.InstanceId,
					},
				},
			}, nil
		},
		InputMapperList: func(scope string) (*elb.DescribeInstanceHealthInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for elb-instance-health, use search",
			}
		},
		OutputMapper: InstanceHealthOutputMapper,
	}
}
