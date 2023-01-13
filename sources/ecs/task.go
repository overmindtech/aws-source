package ecs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// TaskIncludeFields Fields that we want included by default
var TaskIncludeFields = []types.TaskField{
	types.TaskFieldTags,
}

func TaskGetFunc(ctx context.Context, client ECSClient, scope string, input *ecs.DescribeTasksInput) (*sdp.Item, error) {
	out, err := client.DescribeTasks(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(out.Tasks) != 1 {
		return nil, fmt.Errorf("expected 1 task, got %v", len(out.Tasks))
	}

	task := out.Tasks[0]

	attributes, err := sources.ToAttributesCase(task)

	if err != nil {
		return nil, err
	}

	if task.TaskArn == nil {
		return nil, errors.New("task has nil ARN")
	}

	a, err := sources.ParseARN(*task.TaskArn)

	if err != nil {
		return nil, err
	}

	// Create unique attribute in the format {clusterName}/{id} e.g.
	// test-ECSCluster-Bt4SqcM3CURk/2ffd7ed376c841bcb0e6795ddb6e72e2
	attributes.Set("id", a.ResourceID)

	item := sdp.Item{
		Type:            "ecs-task",
		UniqueAttribute: "id",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, attachment := range task.Attachments {
		if attachment.Type != nil {
			if *attachment.Type == "ElasticNetworkInterface" {
				if attachment.Id != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "ec2-network-interface",
						Method: sdp.RequestMethod_GET,
						Query:  *attachment.Id,
						Scope:  scope,
					})
				}
			}
		}
	}

	if task.ClusterArn != nil {
		if a, err = sources.ParseARN(*task.ClusterArn); err == nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ecs-cluster",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *task.ClusterArn,
				Scope:  sources.FormatScope(a.AccountID, a.Region),
			})
		}
	}

	if task.ContainerInstanceArn != nil {
		if a, err = sources.ParseARN(*task.ContainerInstanceArn); err == nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ecs-container-instance",
				Method: sdp.RequestMethod_GET,
				Query:  a.ResourceID,
				Scope:  scope,
			})
		}
	}

	for _, container := range task.Containers {
		for _, ni := range container.NetworkInterfaces {
			if ni.Ipv6Address != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *ni.Ipv6Address,
					Scope:  "global",
				})
			}

			if ni.PrivateIpv4Address != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *ni.PrivateIpv4Address,
					Scope:  "global",
				})
			}
		}
	}

	if task.TaskDefinitionArn != nil {
		if a, err = sources.ParseARN(*task.TaskDefinitionArn); err == nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ecs-task-definition",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *task.TaskDefinitionArn,
				Scope:  sources.FormatScope(a.AccountID, a.Region),
			})
		}
	}

	return &item, nil
}

func TaskGetInputMapper(scope, query string) *ecs.DescribeTasksInput {
	// `id` is {clusterName}/{id} so split on '/'
	sections := strings.Split(query, "/")

	if len(sections) != 2 {
		return nil
	}

	return &ecs.DescribeTasksInput{
		Tasks: []string{
			sections[1],
		},
		Cluster: sources.PtrString(sections[0]),
		Include: TaskIncludeFields,
	}
}

func TasksListFuncOutputMapper(output *ecs.ListTasksOutput, input *ecs.ListTasksInput) ([]*ecs.DescribeTasksInput, error) {
	inputs := make([]*ecs.DescribeTasksInput, 0)

	for _, taskArn := range output.TaskArns {
		if a, err := sources.ParseARN(taskArn); err == nil {
			// split the cluster name out
			sections := strings.Split(a.ResourceID, "/")

			if len(sections) != 2 {
				continue
			}

			inputs = append(inputs, &ecs.DescribeTasksInput{
				Tasks: []string{
					sections[1],
				},
				Cluster: &sections[0],
				Include: TaskIncludeFields,
			})
		}
	}

	return inputs, nil
}

func NewTaskSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*ecs.ListTasksInput, *ecs.ListTasksOutput, *ecs.DescribeTasksInput, *ecs.DescribeTasksOutput, ECSClient, *ecs.Options] {
	return &sources.AlwaysGetSource[*ecs.ListTasksInput, *ecs.ListTasksOutput, *ecs.DescribeTasksInput, *ecs.DescribeTasksOutput, ECSClient, *ecs.Options]{
		ItemType:       "ecs-task",
		Client:         ecs.NewFromConfig(config),
		AccountID:      accountID,
		Region:         region,
		GetFunc:        TaskGetFunc,
		ListInput:      &ecs.ListTasksInput{},
		GetInputMapper: TaskGetInputMapper,
		DisableList:    true,
		SearchInputMapper: func(scope, query string) (*ecs.ListTasksInput, error) {
			// Search by cluster
			return &ecs.ListTasksInput{
				Cluster: sources.PtrString(query),
			}, nil
		},
		ListFuncPaginatorBuilder: func(client ECSClient, input *ecs.ListTasksInput) sources.Paginator[*ecs.ListTasksOutput, *ecs.Options] {
			return ecs.NewListTasksPaginator(client, input)
		},
		ListFuncOutputMapper: TasksListFuncOutputMapper,
	}
}