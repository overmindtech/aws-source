package ecs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func (t *TestClient) DescribeContainerInstances(ctx context.Context, params *ecs.DescribeContainerInstancesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeContainerInstancesOutput, error) {
	return &ecs.DescribeContainerInstancesOutput{
		ContainerInstances: []types.ContainerInstance{
			{
				ContainerInstanceArn: adapters.PtrString("arn:aws:ecs:eu-west-1:052392120703:container-instance/ecs-template-ECSCluster-8nS0WOLbs3nZ/50e9bf71ed57450ca56293cc5a042886"),
				Ec2InstanceId:        adapters.PtrString("i-0e778f25705bc0c84"), // link
				Version:              4,
				VersionInfo: &types.VersionInfo{
					AgentVersion:  adapters.PtrString("1.47.0"),
					AgentHash:     adapters.PtrString("1489adfa"),
					DockerVersion: adapters.PtrString("DockerVersion: 19.03.6-ce"),
				},
				RemainingResources: []types.Resource{
					{
						Name:         adapters.PtrString("CPU"),
						Type:         adapters.PtrString("INTEGER"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 2028,
					},
					{
						Name:         adapters.PtrString("MEMORY"),
						Type:         adapters.PtrString("INTEGER"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 7474,
					},
					{
						Name:         adapters.PtrString("PORTS"),
						Type:         adapters.PtrString("STRINGSET"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 0,
						StringSetValue: []string{
							"22",
							"2376",
							"2375",
							"51678",
							"51679",
						},
					},
					{
						Name:           adapters.PtrString("PORTS_UDP"),
						Type:           adapters.PtrString("STRINGSET"),
						DoubleValue:    0.0,
						LongValue:      0,
						IntegerValue:   0,
						StringSetValue: []string{},
					},
				},
				RegisteredResources: []types.Resource{
					{
						Name:         adapters.PtrString("CPU"),
						Type:         adapters.PtrString("INTEGER"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 2048,
					},
					{
						Name:         adapters.PtrString("MEMORY"),
						Type:         adapters.PtrString("INTEGER"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 7974,
					},
					{
						Name:         adapters.PtrString("PORTS"),
						Type:         adapters.PtrString("STRINGSET"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 0,
						StringSetValue: []string{
							"22",
							"2376",
							"2375",
							"51678",
							"51679",
						},
					},
					{
						Name:           adapters.PtrString("PORTS_UDP"),
						Type:           adapters.PtrString("STRINGSET"),
						DoubleValue:    0.0,
						LongValue:      0,
						IntegerValue:   0,
						StringSetValue: []string{},
					},
				},
				Status:            adapters.PtrString("ACTIVE"),
				AgentConnected:    true,
				RunningTasksCount: 1,
				PendingTasksCount: 0,
				Attributes: []types.Attribute{
					{
						Name: adapters.PtrString("ecs.capability.secrets.asm.environment-variables"),
					},
					{
						Name:  adapters.PtrString("ecs.capability.branch-cni-plugin-version"),
						Value: adapters.PtrString("a21d3a41-"),
					},
					{
						Name:  adapters.PtrString("ecs.ami-id"),
						Value: adapters.PtrString("ami-0c9ef930279337028"),
					},
					{
						Name: adapters.PtrString("ecs.capability.secrets.asm.bootstrap.log-driver"),
					},
					{
						Name: adapters.PtrString("ecs.capability.task-eia.optimized-cpu"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.logging-driver.none"),
					},
					{
						Name: adapters.PtrString("ecs.capability.ecr-endpoint"),
					},
					{
						Name: adapters.PtrString("ecs.capability.docker-plugin.local"),
					},
					{
						Name: adapters.PtrString("ecs.capability.task-cpu-mem-limit"),
					},
					{
						Name: adapters.PtrString("ecs.capability.secrets.ssm.bootstrap.log-driver"),
					},
					{
						Name: adapters.PtrString("ecs.capability.efsAuth"),
					},
					{
						Name: adapters.PtrString("ecs.capability.full-sync"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.30"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.31"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.32"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.logging-driver.fluentd"),
					},
					{
						Name: adapters.PtrString("ecs.capability.firelens.options.config.file"),
					},
					{
						Name:  adapters.PtrString("ecs.availability-zone"),
						Value: adapters.PtrString("eu-west-1a"),
					},
					{
						Name: adapters.PtrString("ecs.capability.aws-appmesh"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.logging-driver.awslogs"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.24"),
					},
					{
						Name: adapters.PtrString("ecs.capability.task-eni-trunking"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.25"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.26"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.27"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.privileged-container"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.28"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.29"),
					},
					{
						Name:  adapters.PtrString("ecs.cpu-architecture"),
						Value: adapters.PtrString("x86_64"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.ecr-auth"),
					},
					{
						Name: adapters.PtrString("ecs.capability.firelens.fluentbit"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.20"),
					},
					{
						Name:  adapters.PtrString("ecs.os-type"),
						Value: adapters.PtrString("linux"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.21"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.22"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.23"),
					},
					{
						Name: adapters.PtrString("ecs.capability.task-eia"),
					},
					{
						Name: adapters.PtrString("ecs.capability.private-registry-authentication.secretsmanager"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.logging-driver.syslog"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.logging-driver.awsfirelens"),
					},
					{
						Name: adapters.PtrString("ecs.capability.firelens.options.config.s3"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.logging-driver.json-file"),
					},
					{
						Name: adapters.PtrString("ecs.capability.execution-role-awslogs"),
					},
					{
						Name:  adapters.PtrString("ecs.vpc-id"),
						Value: adapters.PtrString("vpc-0e120717a7263de70"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.17"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.18"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.19"),
					},
					{
						Name: adapters.PtrString("ecs.capability.docker-plugin.amazon-ecs-volume-plugin"),
					},
					{
						Name: adapters.PtrString("ecs.capability.task-eni"),
					},
					{
						Name: adapters.PtrString("ecs.capability.firelens.fluentd"),
					},
					{
						Name: adapters.PtrString("ecs.capability.efs"),
					},
					{
						Name: adapters.PtrString("ecs.capability.execution-role-ecr-pull"),
					},
					{
						Name: adapters.PtrString("ecs.capability.task-eni.ipv6"),
					},
					{
						Name: adapters.PtrString("ecs.capability.container-health-check"),
					},
					{
						Name:  adapters.PtrString("ecs.subnet-id"),
						Value: adapters.PtrString("subnet-0bfdb717a234c01b3"),
					},
					{
						Name:  adapters.PtrString("ecs.instance-type"),
						Value: adapters.PtrString("t2.large"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.task-iam-role-network-host"),
					},
					{
						Name: adapters.PtrString("ecs.capability.container-ordering"),
					},
					{
						Name:  adapters.PtrString("ecs.capability.cni-plugin-version"),
						Value: adapters.PtrString("55b2ae77-2020.09.0"),
					},
					{
						Name: adapters.PtrString("ecs.capability.env-files.s3"),
					},
					{
						Name: adapters.PtrString("ecs.capability.pid-ipc-namespace-sharing"),
					},
					{
						Name: adapters.PtrString("ecs.capability.secrets.ssm.environment-variables"),
					},
					{
						Name: adapters.PtrString("com.amazonaws.ecs.capability.task-iam-role"),
					},
				},
				RegisteredAt:         adapters.PtrTime(time.Now()),
				Attachments:          []types.Attachment{}, // There is probably an opportunity for some links here but I don't have example data
				Tags:                 []types.Tag{},
				AgentUpdateStatus:    types.AgentUpdateStatusFailed,
				CapacityProviderName: adapters.PtrString("name"),
				HealthStatus: &types.ContainerInstanceHealthStatus{
					OverallStatus: types.InstanceHealthCheckStateImpaired,
				},
			},
		},
	}, nil
}

func (t *TestClient) ListContainerInstances(context.Context, *ecs.ListContainerInstancesInput, ...func(*ecs.Options)) (*ecs.ListContainerInstancesOutput, error) {
	return &ecs.ListContainerInstancesOutput{
		ContainerInstanceArns: []string{
			"arn:aws:ecs:eu-west-1:052392120703:container-instance/ecs-template-ECSCluster-8nS0WOLbs3nZ/50e9bf71ed57450ca56293cc5a042886",
		},
	}, nil
}

func TestContainerInstanceGetFunc(t *testing.T) {
	item, err := containerInstanceGetFunc(context.Background(), &TestClient{}, "foo", &ecs.DescribeContainerInstancesInput{})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "i-0e778f25705bc0c84",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewContainerInstanceAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewContainerInstanceAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter:           adapter,
		Timeout:           10 * time.Second,
		SkipNotFoundCheck: true,
	}

	test.Run(t)
}
