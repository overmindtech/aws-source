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

// ServiceIncludeFields Fields that we want included by default
var ServiceIncludeFields = []types.ServiceField{
	types.ServiceFieldTags,
}

func ServiceGetFunc(ctx context.Context, client ECSClient, scope string, input *ecs.DescribeServicesInput) (*sdp.Item, error) {
	out, err := client.DescribeServices(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(out.Services) != 1 {
		return nil, fmt.Errorf("got %v Services, expected 1", len(out.Services))
	}

	service := out.Services[0]

	// Before we convert to attributes we want to extract the task sets to link
	// to and then delete the info. This because the response embeds the entire
	// task set which is unnecessary since it'll be returned by ecs-task-set
	taskSetIds := make([]string, 0)

	for _, ts := range service.TaskSets {
		if ts.Id != nil {
			taskSetIds = append(taskSetIds, *ts.Id)
		}
	}

	service.TaskSets = []types.TaskSet{}

	attributes, err := sources.ToAttributesCase(service)

	if err != nil {
		return nil, err
	}

	if service.ServiceArn != nil {
		if a, err := sources.ParseARN(*service.ServiceArn); err == nil {
			attributes.Set("serviceFullName", a.Resource)
		}
	}

	item := sdp.Item{
		Type:            "ecs-service",
		UniqueAttribute: "serviceFullName",
		Scope:           scope,
		Attributes:      attributes,
	}

	var a *sources.ARN

	if service.ClusterArn != nil {
		if a, err = sources.ParseARN(*service.ClusterArn); err == nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ecs-cluster",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *service.ClusterArn,
				Scope:  sources.FormatScope(a.AccountID, a.Region),
			})
		}
	}

	for _, lb := range service.LoadBalancers {
		if lb.TargetGroupArn != nil {
			if a, err = sources.ParseARN(*lb.TargetGroupArn); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "elbv2-target-group",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *lb.TargetGroupArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}
	}

	for _, sr := range service.ServiceRegistries {
		if sr.RegistryArn != nil {
			if a, err = sources.ParseARN(*sr.RegistryArn); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "servicediscovery-service",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *sr.RegistryArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}
	}

	if service.TaskDefinition != nil {
		if a, err = sources.ParseARN(*service.TaskDefinition); err == nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ecs-task-definition",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *service.TaskDefinition,
				Scope:  sources.FormatScope(a.AccountID, a.Region),
			})
		}
	}

	for _, deployment := range service.Deployments {
		if deployment.TaskDefinition != nil {
			if a, err = sources.ParseARN(*deployment.TaskDefinition); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ecs-task-definition",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *deployment.TaskDefinition,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, strategy := range deployment.CapacityProviderStrategy {
			if strategy.CapacityProvider != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ecs-capacity-provider",
					Method: sdp.RequestMethod_GET,
					Query:  *strategy.CapacityProvider,
					Scope:  scope,
				})
			}
		}

		if deployment.NetworkConfiguration != nil {
			if deployment.NetworkConfiguration.AwsvpcConfiguration != nil {
				for _, subnet := range deployment.NetworkConfiguration.AwsvpcConfiguration.Subnets {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ec2-subnet",
						Method: sdp.RequestMethod_GET,
						Query:  subnet,
						Scope:  scope,
					})
				}

				for _, sg := range deployment.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ecs-security-group",
						Method: sdp.RequestMethod_GET,
						Query:  sg,
						Scope:  scope,
					})
				}
			}
		}

		if deployment.ServiceConnectConfiguration != nil {
			for _, svc := range deployment.ServiceConnectConfiguration.Services {
				for _, alias := range svc.ClientAliases {
					if alias.DnsName != nil {
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
							Type:   "dns",
							Method: sdp.RequestMethod_GET,
							Query:  *alias.DnsName,
							Scope:  "global",
						})
					}
				}
			}
		}

		for _, cr := range deployment.ServiceConnectResources {
			if cr.DiscoveryArn != nil {
				if a, err = sources.ParseARN(*cr.DiscoveryArn); err == nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "servicediscovery-service",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *cr.DiscoveryArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}
	}

	if service.NetworkConfiguration != nil {
		if service.NetworkConfiguration.AwsvpcConfiguration != nil {
			for _, subnet := range service.NetworkConfiguration.AwsvpcConfiguration.Subnets {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.RequestMethod_GET,
					Query:  subnet,
					Scope:  scope,
				})
			}

			for _, sg := range service.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ecs-security-group",
					Method: sdp.RequestMethod_GET,
					Query:  sg,
					Scope:  scope,
				})
			}
		}
	}

	for _, id := range taskSetIds {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "ecs-task-set",
			Method: sdp.RequestMethod_GET,
			Query:  id,
			Scope:  scope,
		})
	}

	return &item, nil
}

func ServiceListFuncOutputMapper(output *ecs.ListServicesOutput, input *ecs.ListServicesInput) ([]*ecs.DescribeServicesInput, error) {
	inputs := make([]*ecs.DescribeServicesInput, 0)

	var a *sources.ARN
	var err error

	for _, arn := range output.ServiceArns {
		a, err = sources.ParseARN(arn)

		if err != nil {
			continue
		}

		sections := strings.Split(a.Resource, "/")

		if len(sections) != 2 {
			return nil, fmt.Errorf("could not split into 2 sections on '/': %v", a.Resource)
		}

		inputs = append(inputs, &ecs.DescribeServicesInput{
			Cluster: &sections[0],
			Services: []string{
				sections[1],
			},
			Include: ServiceIncludeFields,
		})
	}

	return inputs, nil
}

func NewServiceSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*ecs.ListServicesInput, *ecs.ListServicesOutput, *ecs.DescribeServicesInput, *ecs.DescribeServicesOutput, ECSClient, *ecs.Options] {
	return &sources.AlwaysGetSource[*ecs.ListServicesInput, *ecs.ListServicesOutput, *ecs.DescribeServicesInput, *ecs.DescribeServicesOutput, ECSClient, *ecs.Options]{
		ItemType:    "ecs-service",
		Client:      ecs.NewFromConfig(config),
		AccountID:   accountID,
		Region:      region,
		GetFunc:     ServiceGetFunc,
		DisableList: true,
		GetInputMapper: func(scope, query string) *ecs.DescribeServicesInput {
			// We are using a custom id of {clusterName}/{id} e.g.
			// ecs-template-ECSCluster-8nS0WOLbs3nZ/ecs-template-service-i0mQKzkhDI2C
			sections := strings.Split(query, "/")

			if len(sections) != 2 {
				return nil
			}

			return &ecs.DescribeServicesInput{
				Services: []string{
					sections[1],
				},
				Cluster: &sections[0],
				Include: ServiceIncludeFields,
			}
		},
		ListInput: &ecs.ListServicesInput{},
		ListFuncPaginatorBuilder: func(client ECSClient, input *ecs.ListServicesInput) sources.Paginator[*ecs.ListServicesOutput, *ecs.Options] {
			return ecs.NewListServicesPaginator(client, input)
		},
		SearchInputMapper: func(scope, query string) (*ecs.ListServicesInput, error) {
			// Custom search by cluster
			return &ecs.ListServicesInput{
				Cluster: sources.PtrString(query),
			}, nil
		},
		ListFuncOutputMapper: ServiceListFuncOutputMapper,
	}
}
