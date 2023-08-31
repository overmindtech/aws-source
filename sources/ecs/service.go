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

func serviceGetFunc(ctx context.Context, client ECSClient, scope string, input *ecs.DescribeServicesInput) (*sdp.Item, error) {
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

	if service.Status != nil {
		switch *service.Status {
		case "ACTIVE":
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case "DRAINING":
			item.Health = sdp.Health_HEALTH_WARNING.Enum()
		case "INACTIVE":
			item.Health = nil
		}
	}

	var a *sources.ARN

	if service.ClusterArn != nil {
		if a, err = sources.ParseARN(*service.ClusterArn); err == nil {
			// +overmind:link ecs-cluster
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ecs-cluster",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *service.ClusterArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the cluster will affect the service
					In: true,
					// The service should be able to affect the cluster
					Out: false,
				},
			})
		}
	}

	for _, lb := range service.LoadBalancers {
		if lb.TargetGroupArn != nil {
			if a, err = sources.ParseARN(*lb.TargetGroupArn); err == nil {
				// +overmind:link elbv2-target-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "elbv2-target-group",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *lb.TargetGroupArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// These are tightly linked
						In:  true,
						Out: true,
					},
				})
			}
		}
	}

	for _, sr := range service.ServiceRegistries {
		if sr.RegistryArn != nil {
			if a, err = sources.ParseARN(*sr.RegistryArn); err == nil {
				// +overmind:link servicediscovery-service
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "servicediscovery-service",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *sr.RegistryArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// These are tightly linked
						In:  true,
						Out: true,
					},
				})
			}
		}
	}

	if service.TaskDefinition != nil {
		if a, err = sources.ParseARN(*service.TaskDefinition); err == nil {
			// +overmind:link ecs-task-definition
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ecs-task-definition",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *service.TaskDefinition,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the task definition will affect the service
					In: true,
					// The service shouldn't affect the task definition itself
					Out: false,
				},
			})
		}
	}

	for _, deployment := range service.Deployments {
		if deployment.TaskDefinition != nil {
			if a, err = sources.ParseARN(*deployment.TaskDefinition); err == nil {
				// +overmind:link ecs-task-definition
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ecs-task-definition",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *deployment.TaskDefinition,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the task definition will affect the service
						In: true,
						// The service shouldn't affect the task definition itself
						Out: false,
					},
				})
			}
		}

		for _, strategy := range deployment.CapacityProviderStrategy {
			if strategy.CapacityProvider != nil {
				// +overmind:link ecs-capacity-provider
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ecs-capacity-provider",
						Method: sdp.QueryMethod_GET,
						Query:  *strategy.CapacityProvider,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the capacity provider will affect the service
						In: true,
						// The service shouldn't affect the capacity provider itself
						Out: false,
					},
				})
			}
		}

		if deployment.NetworkConfiguration != nil {
			if deployment.NetworkConfiguration.AwsvpcConfiguration != nil {
				for _, subnet := range deployment.NetworkConfiguration.AwsvpcConfiguration.Subnets {
					// +overmind:link ec2-subnet
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-subnet",
							Method: sdp.QueryMethod_GET,
							Query:  subnet,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the subnet will affect the service
							In: true,
							// The service shouldn't affect the subnet
							Out: false,
						},
					})
				}

				for _, sg := range deployment.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups {
					// +overmind:link ec2-security-group
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ecs-security-group",
							Method: sdp.QueryMethod_GET,
							Query:  sg,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the security group will affect the service
							In: true,
							// The service shouldn't affect the security group
							Out: false,
						},
					})
				}
			}
		}

		if deployment.ServiceConnectConfiguration != nil {
			for _, svc := range deployment.ServiceConnectConfiguration.Services {
				for _, alias := range svc.ClientAliases {
					if alias.DnsName != nil {
						// +overmind:link dns
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "dns",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *alias.DnsName,
								Scope:  "global",
							},
							BlastPropagation: &sdp.BlastPropagation{
								// DNS always links
								In:  true,
								Out: true,
							},
						})
					}
				}
			}
		}

		for _, cr := range deployment.ServiceConnectResources {
			if cr.DiscoveryArn != nil {
				if a, err = sources.ParseARN(*cr.DiscoveryArn); err == nil {
					// +overmind:link servicediscovery-service
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "servicediscovery-service",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *cr.DiscoveryArn,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// These are tightly linked
							In:  true,
							Out: true,
						},
					})
				}
			}
		}
	}

	if service.NetworkConfiguration != nil {
		if service.NetworkConfiguration.AwsvpcConfiguration != nil {
			for _, subnet := range service.NetworkConfiguration.AwsvpcConfiguration.Subnets {
				// +overmind:link ec2-subnet
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-subnet",
						Method: sdp.QueryMethod_GET,
						Query:  subnet,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the subnet will affect the service
						In: true,
						// The service shouldn't affect the subnet
						Out: false,
					},
				})
			}

			for _, sg := range service.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups {
				// +overmind:link ec2-security-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-security-group",
						Method: sdp.QueryMethod_GET,
						Query:  sg,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the security group will affect the service
						In: true,
						// The service shouldn't affect the security group
						Out: false,
					},
				})
			}
		}
	}

	for _, id := range taskSetIds {
		// +overmind:link ecs-task-set
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ecs-task-set",
				Method: sdp.QueryMethod_GET,
				Query:  id,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// These are tightly linked
				In:  true,
				Out: true,
			},
		})
	}

	return &item, nil
}

func serviceListFuncOutputMapper(output *ecs.ListServicesOutput, input *ecs.ListServicesInput) ([]*ecs.DescribeServicesInput, error) {
	inputs := make([]*ecs.DescribeServicesInput, 0)

	var a *sources.ARN
	var err error

	for _, arn := range output.ServiceArns {
		a, err = sources.ParseARN(arn)

		if err != nil {
			continue
		}

		sections := strings.Split(a.Resource, "/")

		if len(sections) != 3 {
			return nil, fmt.Errorf("could not split into 3 sections on '/': %v", a.Resource)
		}

		inputs = append(inputs, &ecs.DescribeServicesInput{
			Cluster: &sections[1],
			Services: []string{
				sections[2],
			},
			Include: ServiceIncludeFields,
		})
	}

	return inputs, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ecs-service
// +overmind:descriptiveType ECS Service
// +overmind:get Get an ECS service by full name ({clusterName}/{id})
// +overmind:list List all ECS services
// +overmind:search Search for ECS services by cluster
// +overmind:group AWS
// +overmind:terraform:queryMap ${aws_ecs_service.cluster}/${aws_ecs_service.name}

func NewServiceSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*ecs.ListServicesInput, *ecs.ListServicesOutput, *ecs.DescribeServicesInput, *ecs.DescribeServicesOutput, ECSClient, *ecs.Options] {
	return &sources.AlwaysGetSource[*ecs.ListServicesInput, *ecs.ListServicesOutput, *ecs.DescribeServicesInput, *ecs.DescribeServicesOutput, ECSClient, *ecs.Options]{
		ItemType:    "ecs-service",
		Client:      ecs.NewFromConfig(config),
		AccountID:   accountID,
		Region:      region,
		GetFunc:     serviceGetFunc,
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
		ListFuncOutputMapper: serviceListFuncOutputMapper,
	}
}
