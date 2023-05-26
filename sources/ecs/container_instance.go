package ecs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// ContainerInstanceIncludeFields Fields that we want included by default
var ContainerInstanceIncludeFields = []types.ContainerInstanceField{
	types.ContainerInstanceFieldTags,
	types.ContainerInstanceFieldContainerInstanceHealth,
}

func containerInstanceGetFunc(ctx context.Context, client ECSClient, scope string, input *ecs.DescribeContainerInstancesInput) (*sdp.Item, error) {
	out, err := client.DescribeContainerInstances(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(out.ContainerInstances) != 1 {
		return nil, fmt.Errorf("got %v ContainerInstances, expected 1", len(out.ContainerInstances))
	}

	containerInstance := out.ContainerInstances[0]

	attributes, err := sources.ToAttributesCase(containerInstance)

	if err != nil {
		return nil, err
	}

	// Create an ID param since they don't have anything that uniquely
	// identifies them. This is {clusterName}/{id} e.g.
	// ecs-template-ECSCluster-8nS0WOLbs3nZ/50e9bf71ed57450ca56293cc5a042886
	if a, err := sources.ParseARN(*containerInstance.ContainerInstanceArn); err == nil {
		attributes.Set("id", a.Resource)
	}

	item := sdp.Item{
		Type:            "ecs-container-instance",
		UniqueAttribute: "id",
		Scope:           scope,
		Attributes:      attributes,
	}

	if containerInstance.HealthStatus != nil {
		switch containerInstance.HealthStatus.OverallStatus {
		case types.InstanceHealthCheckStateOk:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.InstanceHealthCheckStateImpaired:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		case types.InstanceHealthCheckStateInsufficientData:
			item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
		case types.InstanceHealthCheckStateInitializing:
			item.Health = sdp.Health_HEALTH_WARNING.Enum()
		}
	}

	if containerInstance.Ec2InstanceId != nil {
		// +overmind:link ec2-instance
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ec2-instance",
				Method: sdp.QueryMethod_GET,
				Query:  *containerInstance.Ec2InstanceId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// These represent the same thing
				In:  true,
				Out: true,
			},
		})
	}

	return &item, nil
}

func containerInstanceListFuncOutputMapper(output *ecs.ListContainerInstancesOutput, input *ecs.ListContainerInstancesInput) ([]*ecs.DescribeContainerInstancesInput, error) {
	inputs := make([]*ecs.DescribeContainerInstancesInput, 0)

	var a *sources.ARN
	var err error

	for _, arn := range output.ContainerInstanceArns {
		a, err = sources.ParseARN(arn)

		if err != nil {
			continue
		}

		sections := strings.Split(a.Resource, "/")

		if len(sections) != 2 {
			return nil, fmt.Errorf("could not split into 2 sections on '/': %v", a.Resource)
		}

		inputs = append(inputs, &ecs.DescribeContainerInstancesInput{
			Cluster: &sections[0],
			ContainerInstances: []string{
				sections[1],
			},
			Include: ContainerInstanceIncludeFields,
		})
	}

	return inputs, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ecs-container-instance
// +overmind:descriptiveType Container Instance
// +overmind:get Get a container instance by ID which consists of {clusterName}/{id}
// +overmind:list List all container instances
// +overmind:search Search for container instances by cluster
// +overmind:group AWS

func NewContainerInstanceSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*ecs.ListContainerInstancesInput, *ecs.ListContainerInstancesOutput, *ecs.DescribeContainerInstancesInput, *ecs.DescribeContainerInstancesOutput, ECSClient, *ecs.Options] {
	return &sources.AlwaysGetSource[*ecs.ListContainerInstancesInput, *ecs.ListContainerInstancesOutput, *ecs.DescribeContainerInstancesInput, *ecs.DescribeContainerInstancesOutput, ECSClient, *ecs.Options]{
		ItemType:  "ecs-container-instance",
		Client:    ecs.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		GetFunc:   containerInstanceGetFunc,
		GetInputMapper: func(scope, query string) *ecs.DescribeContainerInstancesInput {
			// We are using a custom id of {clusterName}/{id} e.g.
			// ecs-template-ECSCluster-8nS0WOLbs3nZ/50e9bf71ed57450ca56293cc5a042886
			sections := strings.Split(query, "/")

			if len(sections) != 2 {
				return nil
			}

			return &ecs.DescribeContainerInstancesInput{
				ContainerInstances: []string{
					sections[1],
				},
				Cluster: &sections[0],
				Include: ContainerInstanceIncludeFields,
			}
		},
		ListInput:   &ecs.ListContainerInstancesInput{},
		DisableList: true, // Tou can't list without a cluster
		ListFuncPaginatorBuilder: func(client ECSClient, input *ecs.ListContainerInstancesInput) sources.Paginator[*ecs.ListContainerInstancesOutput, *ecs.Options] {
			return ecs.NewListContainerInstancesPaginator(client, input)
		},
		SearchInputMapper: func(scope, query string) (*ecs.ListContainerInstancesInput, error) {
			// Custom search by cluster
			return &ecs.ListContainerInstancesInput{
				Cluster: sources.PtrString(query),
			}, nil
		},
		ListFuncOutputMapper: containerInstanceListFuncOutputMapper,
	}
}
