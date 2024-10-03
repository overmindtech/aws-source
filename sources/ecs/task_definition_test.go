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

func (t *TestClient) DescribeTaskDefinition(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
	return &ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &types.TaskDefinition{
			TaskDefinitionArn: sources.PtrString("arn:aws:ecs:eu-west-1:052392120703:task-definition/ecs-template-ecs-demo-app:1"),
			ContainerDefinitions: []types.ContainerDefinition{
				{
					Name:   sources.PtrString("simple-app"),
					Image:  sources.PtrString("httpd:2.4"),
					Cpu:    10,
					Memory: sources.PtrInt32(300),
					Links:  []string{},
					PortMappings: []types.PortMapping{
						{
							ContainerPort: sources.PtrInt32(80),
							HostPort:      sources.PtrInt32(0),
							Protocol:      types.TransportProtocolTcp,
							AppProtocol:   types.ApplicationProtocolHttp,
						},
					},
					Essential:  sources.PtrBool(true),
					EntryPoint: []string{},
					Command:    []string{},
					Environment: []types.KeyValuePair{
						{
							Name:  sources.PtrString("DATABASE_SERVER"),
							Value: sources.PtrString("database01.my-company.com"),
						},
					},
					EnvironmentFiles: []types.EnvironmentFile{},
					MountPoints: []types.MountPoint{
						{
							SourceVolume:  sources.PtrString("my-vol"),
							ContainerPath: sources.PtrString("/usr/local/apache2/htdocs"),
							ReadOnly:      sources.PtrBool(false),
						},
					},
					VolumesFrom: []types.VolumeFrom{
						{
							SourceContainer: sources.PtrString("container"),
						},
					},
					Secrets: []types.Secret{
						{
							Name:      sources.PtrString("secrets-manager"),
							ValueFrom: sources.PtrString("arn:aws:secretsmanager:us-west-2:123456789012:secret:my-path/my-secret-name-1a2b3c"), // link
						},
						{
							Name:      sources.PtrString("ssm"),
							ValueFrom: sources.PtrString("arn:aws:ssm:us-east-2:123456789012:parameter/prod-123"), // link
						},
					},
					DnsServers:       []string{},
					DnsSearchDomains: []string{},
					ExtraHosts: []types.HostEntry{
						{
							Hostname:  sources.PtrString("host"),
							IpAddress: sources.PtrString("127.0.0.1"),
						},
					},
					DockerSecurityOptions: []string{},
					DockerLabels:          map[string]string{},
					Ulimits:               []types.Ulimit{},
					LogConfiguration: &types.LogConfiguration{
						LogDriver: types.LogDriverAwslogs,
						Options: map[string]string{
							"awslogs-group":         "ECSLogGroup-ecs-template",
							"awslogs-region":        "eu-west-1",
							"awslogs-stream-prefix": "ecs-demo-app",
						},
						SecretOptions: []types.Secret{
							{
								Name:      sources.PtrString("secrets-manager"),
								ValueFrom: sources.PtrString("arn:aws:secretsmanager:us-west-2:123456789012:secret:my-path/my-secret-name-1a2b3c"), // link
							},
							{
								Name:      sources.PtrString("ssm"),
								ValueFrom: sources.PtrString("arn:aws:ssm:us-east-2:123456789012:parameter/prod-123"), // link
							},
						},
					},
					SystemControls:    []types.SystemControl{},
					DependsOn:         []types.ContainerDependency{},
					DisableNetworking: sources.PtrBool(false),
					FirelensConfiguration: &types.FirelensConfiguration{
						Type:    types.FirelensConfigurationTypeFluentd,
						Options: map[string]string{},
					},
					HealthCheck:            &types.HealthCheck{},
					Hostname:               sources.PtrString("hostname"),
					Interactive:            sources.PtrBool(false),
					LinuxParameters:        &types.LinuxParameters{},
					MemoryReservation:      sources.PtrInt32(100),
					Privileged:             sources.PtrBool(false),
					PseudoTerminal:         sources.PtrBool(false),
					ReadonlyRootFilesystem: sources.PtrBool(false),
					RepositoryCredentials:  &types.RepositoryCredentials{}, // Skipping the link here for now, if you need it, add it in a PR
					ResourceRequirements:   []types.ResourceRequirement{},
					StartTimeout:           sources.PtrInt32(1),
					StopTimeout:            sources.PtrInt32(1),
					User:                   sources.PtrString("foo"),
					WorkingDirectory:       sources.PtrString("/"),
				},
				{
					Name:      sources.PtrString("busybox"),
					Image:     sources.PtrString("busybox"),
					Cpu:       10,
					Memory:    sources.PtrInt32(200),
					Essential: sources.PtrBool(false),
					EntryPoint: []string{
						"sh",
						"-c",
					},
					Command: []string{
						"/bin/sh -c \"while true; do echo '<html> <head> <title>Amazon ECS Sample App</title> <style>body {margin-top: 40px; background-color: #333;} </style> </head><body> <div style=color:white;text-align:center> <h1>Amazon ECS Sample App</h1> <h2>Congratulations!</h2> <p>Your application is now running on a container in Amazon ECS.</p>' > top; /bin/date > date ; echo '</div></body></html>' > bottom; cat top date bottom > /usr/local/apache2/htdocs/index.html ; sleep 1; done\"",
					},
					VolumesFrom: []types.VolumeFrom{
						{
							SourceContainer: sources.PtrString("simple-app"),
						},
					},
					DockerLabels: map[string]string{},
					LogConfiguration: &types.LogConfiguration{
						LogDriver: types.LogDriverAwslogs,
						Options: map[string]string{
							"awslogs-group":         "ECSLogGroup-ecs-template",
							"awslogs-region":        "eu-west-1",
							"awslogs-stream-prefix": "ecs-demo-app",
						},
					},
				},
			},
			Family:   sources.PtrString("ecs-template-ecs-demo-app"),
			Revision: 1,
			Volumes: []types.Volume{
				{
					Name: sources.PtrString("my-vol"),
					Host: &types.HostVolumeProperties{
						SourcePath: sources.PtrString("/"),
					},
				},
			},
			Status: types.TaskDefinitionStatusActive,
			RequiresAttributes: []types.Attribute{
				{
					Name: sources.PtrString("com.amazonaws.ecs.capability.logging-driver.awslogs"),
				},
				{
					Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.19"),
				},
				{
					Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.17"),
				},
				{
					Name: sources.PtrString("com.amazonaws.ecs.capability.docker-remote-api.1.18"),
				},
			},
			PlacementConstraints: []types.TaskDefinitionPlacementConstraint{},
			Compatibilities: []types.Compatibility{
				"EXTERNAL",
				"EC2",
			},
			RegisteredAt:   sources.PtrTime(time.Now()),
			RegisteredBy:   sources.PtrString("arn:aws:sts::052392120703:assumed-role/AWSReservedSSO_AWSAdministratorAccess_c1c3c9c54821c68a/dylan@overmind.tech"),
			Cpu:            sources.PtrString("cpu"),
			DeregisteredAt: sources.PtrTime(time.Now()),
			EphemeralStorage: &types.EphemeralStorage{
				SizeInGiB: 1,
			},
			ExecutionRoleArn:        sources.PtrString("arn:aws:iam:us-east-2:123456789012:role/foo"), // link
			InferenceAccelerators:   []types.InferenceAccelerator{},
			IpcMode:                 types.IpcModeHost,
			Memory:                  sources.PtrString("memory"),
			NetworkMode:             types.NetworkModeAwsvpc,
			PidMode:                 types.PidModeHost,
			ProxyConfiguration:      nil,
			RequiresCompatibilities: []types.Compatibility{},
			RuntimePlatform: &types.RuntimePlatform{
				CpuArchitecture:       types.CPUArchitectureX8664,
				OperatingSystemFamily: types.OSFamilyLinux,
			},
			TaskRoleArn: sources.PtrString("arn:aws:iam:us-east-2:123456789012:role/bar"), // link
		},
	}, nil
}

func (t *TestClient) ListTaskDefinitions(context.Context, *ecs.ListTaskDefinitionsInput, ...func(*ecs.Options)) (*ecs.ListTaskDefinitionsOutput, error) {
	return &ecs.ListTaskDefinitionsOutput{
		TaskDefinitionArns: []string{
			"arn:aws:ecs:eu-west-1:052392120703:task-definition/ecs-template-ecs-demo-app:1",
		},
	}, nil
}

func TestTaskDefinitionGetFunc(t *testing.T) {
	item, err := taskDefinitionGetFunc(context.Background(), &TestClient{}, "foo", &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: sources.PtrString("ecs-template-ecs-demo-app:1"),
	})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "secretsmanager-secret",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:secretsmanager:us-west-2:123456789012:secret:my-path/my-secret-name-1a2b3c",
			ExpectedScope:  "123456789012.us-west-2",
		},
		{
			ExpectedType:   "ssm-parameter",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:ssm:us-east-2:123456789012:parameter/prod-123",
			ExpectedScope:  "123456789012.us-east-2",
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam:us-east-2:123456789012:role/foo",
			ExpectedScope:  "123456789012.us-east-2",
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam:us-east-2:123456789012:role/bar",
			ExpectedScope:  "123456789012.us-east-2",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "database01.my-company.com",
			ExpectedScope:  "global",
		},
	}

	tests.Execute(t, item)
}

func TestNewTaskDefinitionSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewTaskDefinitionSource(client, account, region)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
