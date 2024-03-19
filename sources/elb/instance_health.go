package elb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"

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

func instanceHealthOutputMapper(_ context.Context, _ *elb.Client, scope string, _ *elb.DescribeInstanceHealthInput, output *elb.DescribeInstanceHealthOutput) ([]*sdp.Item, error) {
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

		if is.State != nil {
			switch *is.State {
			case "InService":
				item.Health = sdp.Health_HEALTH_OK.Enum()
			case "OutOfService":
				item.Health = sdp.Health_HEALTH_ERROR.Enum()
			case "Unknown":
				item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
			}
		}

		if is.InstanceId != nil {
			// +overmind:link ec2-instance
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *is.InstanceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// These are tightly linked
					In:  true,
					Out: true,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type elb-instance-health
// +overmind:descriptiveType ELB Instance Health
// +overmind:get Get instance health by ID ({LoadBalancerName}/{InstanceId})
// +overmind:list List all instance healths
// +overmind:group AWS

func NewInstanceHealthSource(client *elasticloadbalancing.Client, accountID string, region string) *sources.DescribeOnlySource[*elb.DescribeInstanceHealthInput, *elb.DescribeInstanceHealthOutput, *elb.Client, *elb.Options] {
	return &sources.DescribeOnlySource[*elb.DescribeInstanceHealthInput, *elb.DescribeInstanceHealthOutput, *elb.Client, *elb.Options]{
		Region:    region,
		Client:    client,
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
		OutputMapper: instanceHealthOutputMapper,
	}
}
