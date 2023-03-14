package ecs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func (t *TestClient) DescribeContainerInstances(ctx context.Context, params *ecs.DescribeContainerInstancesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeContainerInstancesOutput, error) {
	return &ecs.DescribeContainerInstancesOutput{
		ContainerInstances: []types.ContainerInstance{
			{
				ContainerInstanceArn: sources.PtrString("arn:aws:ecs:eu-west-1:052392120703:container-instance/ecs-template-ECSCluster-8nS0WOLbs3nZ/50e9bf71ed57450ca56293cc5a042886"),
				Ec2InstanceId:        sources.PtrString("i-0e778f25705bc0c84"), // link
				Version:              4,
				VersionInfo: &types.VersionInfo{
					AgentVersion:  sources.PtrString("1.47.0"),
					AgentHash:     sources.PtrString("1489adfa"),
					DockerVersion: sources.PtrString("DockerVersion: 19.03.6-ce"),
				},
				RemainingResources: []types.Resource{
					{
						Name:         sources.PtrString("CPU"),
						Type:         sources.PtrString("INTEGER"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 2028,
					},
					{
						Name:         sources.PtrString("MEMORY"),
						Type:         sources.PtrString("INTEGER"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 7474,
					},
					{
						Name:         sources.PtrString("PORTS"),
						Type:         sources.PtrString("STRINGSET"),
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
						Name:           sources.PtrString("PORTS_UDP"),
						Type:           sources.PtrString("STRINGSET"),
						DoubleValue:    0.0,
						LongValue:      0,
						IntegerValue:   0,
						StringSetValue: []string{},
					},
				},
				RegisteredResources: []types.Resource{
					{
						Name:         sources.PtrString("CPU"),
						Type:         sources.PtrString("INTEGER"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 2048,
					},
					{
						Name:         sources.PtrString("MEMORY"),
						Type:         sources.PtrString("INTEGER"),
						DoubleValue:  0.0,
						LongValue:    0,
						IntegerValue: 7974,
					},
					{
						Name:         sources.PtrString("PORTS"),
						Type:         sources.PtrString("STRINGSET"),
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
						Name:           sources.PtrString("PORTS_UDP"),
						Type:           sources.PtrString("STRINGSET"),
						DoubleValue:    0.0,
						LongValue:      0,
						IntegerValue:   0,
						StringSetValue: []string{},
					},
				},
				Status:            sources.PtrString("ACTIVE"),
				AgentConnected:    true,
				RunningTasksCount: 1,
				PendingTasksCount: 0,
				Attributes: []types.Attribute{
					{
						Name: sources.PtrString("ecs.capability.secrets.asm.environment-variables"),
					},
					{
						Name:  sources.PtrString("ecs.capability.branch-cni-plugin-version"),
						Value: sources.PtrString("a21d3a41-"),
					},
					{
						Name:  sources.PtrString("ecs.ami-id"),
						Value: sources.PtrString("ami-0c9ef930279337028"),
					},
					{
						Name: sources.PtrString("ecs.capability.secrets.asm.bootstrap.log-driver"),
					},
					{
						Name: sources.PtrString("ecs.capability.task-eia.optimized-cpu"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.logging-driver.none"),
					},
					{
						Name: sources.PtrString("ecs.capability.ecr-endpoint"),
					},
					{
						Name: sources.PtrString("ecs.capability.docker-plugin.local"),
					},
					{
						Name: sources.PtrString("ecs.capability.task-cpu-mem-limit"),
					},
					{
						Name: sources.PtrString("ecs.capability.secrets.ssm.bootstrap.log-driver"),
					},
					{
						Name: sources.PtrString("ecs.capability.efsAuth"),
					},
					{
						Name: sources.PtrString("ecs.capability.full-sync"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.30"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.31"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.32"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.logging-driver.fluentd"),
					},
					{
						Name: sources.PtrString("ecs.capability.firelens.options.config.file"),
					},
					{
						Name:  sources.PtrString("ecs.availability-zone"),
						Value: sources.PtrString("eu-west-1a"),
					},
					{
						Name: sources.PtrString("ecs.capability.aws-appmesh"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.logging-driver.awslogs"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.24"),
					},
					{
						Name: sources.PtrString("ecs.capability.task-eni-trunking"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.25"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.26"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.27"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.privileged-container"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.28"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.29"),
					},
					{
						Name:  sources.PtrString("ecs.cpu-architecture"),
						Value: sources.PtrString("x86_64"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.ecr-auth"),
					},
					{
						Name: sources.PtrString("ecs.capability.firelens.fluentbit"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.20"),
					},
					{
						Name:  sources.PtrString("ecs.os-type"),
						Value: sources.PtrString("linux"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.21"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.22"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.23"),
					},
					{
						Name: sources.PtrString("ecs.capability.task-eia"),
					},
					{
						Name: sources.PtrString("ecs.capability.private-registry-authentication.secretsmanager"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.logging-driver.syslog"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.logging-driver.awsfirelens"),
					},
					{
						Name: sources.PtrString("ecs.capability.firelens.options.config.s3"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.logging-driver.json-file"),
					},
					{
						Name: sources.PtrString("ecs.capability.execution-role-awslogs"),
					},
					{
						Name:  sources.PtrString("ecs.vpc-id"),
						Value: sources.PtrString("vpc-0e120717a7263de70"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.17"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.18"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.19"),
					},
					{
						Name: sources.PtrString("ecs.capability.docker-plugin.amazon-ecs-volume-plugin"),
					},
					{
						Name: sources.PtrString("ecs.capability.task-eni"),
					},
					{
						Name: sources.PtrString("ecs.capability.firelens.fluentd"),
					},
					{
						Name: sources.PtrString("ecs.capability.efs"),
					},
					{
						Name: sources.PtrString("ecs.capability.execution-role-ecr-pull"),
					},
					{
						Name: sources.PtrString("ecs.capability.task-eni.ipv6"),
					},
					{
						Name: sources.PtrString("ecs.capability.container-health-check"),
					},
					{
						Name:  sources.PtrString("ecs.subnet-id"),
						Value: sources.PtrString("subnet-0bfdb717a234c01b3"),
					},
					{
						Name:  sources.PtrString("ecs.instance-type"),
						Value: sources.PtrString("t2.large"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.task-iam-role-network-host"),
					},
					{
						Name: sources.PtrString("ecs.capability.container-ordering"),
					},
					{
						Name:  sources.PtrString("ecs.capability.cni-plugin-version"),
						Value: sources.PtrString("55b2ae77-2020.09.0"),
					},
					{
						Name: sources.PtrString("ecs.capability.env-files.s3"),
					},
					{
						Name: sources.PtrString("ecs.capability.pid-ipc-namespace-sharing"),
					},
					{
						Name: sources.PtrString("ecs.capability.secrets.ssm.environment-variables"),
					},
					{
						Name: sources.PtrString("com.amazonaws.ecs.capability.task-iam-role"),
					},
				},
				RegisteredAt:         sources.PtrTime(time.Now()),
				Attachments:          []types.Attachment{}, // There is probably an opportunity for some links here but I don't have example data
				Tags:                 []types.Tag{},
				AgentUpdateStatus:    types.AgentUpdateStatusFailed,
				CapacityProviderName: sources.PtrString("name"),
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

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "i-0e778f25705bc0c84",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewContainerInstanceSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewContainerInstanceSource(config, account, region)

	test := sources.E2ETest{
		Source:            source,
		Timeout:           10 * time.Second,
		SkipNotFoundCheck: true,
	}

	test.Run(t)
}
