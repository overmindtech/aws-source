# AWS Source

This source integrates with AWS, allowing Overmind to pull data about many types of AWS resources.

## Sources

### elasticloadbalancing-loadbalancer-v2

Gathers information about Elastic Load Balancers (v2) and their target groups, health checks etc. e.g.

```json
{
    "type": "elasticloadbalancing-loadbalancer-v2",
    "uniqueAttribute": "name",
    "attributes": {
        "attrStruct": {
            "availabilityZones": [
                {
                    "loadBalancerAddresses": [],
                    "subnetId": "subnet-021f8be6388f11fcd",
                    "zoneName": "eu-west-2a"
                },
                {
                    "loadBalancerAddresses": [],
                    "subnetId": "subnet-0260e9e4e333abd62",
                    "zoneName": "eu-west-2c"
                },
                {
                    "loadBalancerAddresses": [],
                    "subnetId": "subnet-092453eee2eea7611",
                    "zoneName": "eu-west-2b"
                }
            ],
            "canonicalHostedZoneId": "ZD4D7Y8KGAS4G",
            "createdTime": "2021-12-03 17:07:15.334 +0000 UTC",
            "dNSName": "vpc-0fe83a8d71bd1803ctest-elbv2-d88c129308d731ef.elb.eu-west-2.amazonaws.com",
            "ipAddressType": "ipv4",
            "listeners": [
                {
                    "defaultActions": [
                        {
                            "forwardConfig": {
                                "targetGroups": [
                                    {
                                        "targetGroupArn": "arn:aws:elasticloadbalancing:eu-west-2:177828803798:targetgroup/fake-targets/dcf0f1f60163c003"
                                    }
                                ]
                            },
                            "order": 1,
                            "targetGroupArn": "arn:aws:elasticloadbalancing:eu-west-2:177828803798:targetgroup/fake-targets/dcf0f1f60163c003",
                            "type": "forward"
                        }
                    ],
                    "listenerArn": "arn:aws:elasticloadbalancing:eu-west-2:177828803798:listener/net/vpc-0fe83a8d71bd1803ctest-elbv2/d88c129308d731ef/40157d5d032e19fd",
                    "loadBalancerArn": "arn:aws:elasticloadbalancing:eu-west-2:177828803798:loadbalancer/net/vpc-0fe83a8d71bd1803ctest-elbv2/d88c129308d731ef",
                    "port": 80,
                    "protocol": "TCP"
                }
            ],
            "loadBalancerArn": "arn:aws:elasticloadbalancing:eu-west-2:177828803798:loadbalancer/net/vpc-0fe83a8d71bd1803ctest-elbv2/d88c129308d731ef",
            "name": "vpc-0fe83a8d71bd1803ctest-elbv2",
            "scheme": "internet-facing",
            "state": {
                "code": "active"
            },
            "targetGroups": [
                {
                    "healthCheckEnabled": true,
                    "healthCheckIntervalSeconds": 30,
                    "healthCheckPort": "traffic-port",
                    "healthCheckProtocol": "TCP",
                    "healthCheckTimeoutSeconds": 10,
                    "healthyThresholdCount": 3,
                    "ipAddressType": "ipv4",
                    "loadBalancerArns": [
                        "arn:aws:elasticloadbalancing:eu-west-2:177828803798:loadbalancer/net/vpc-0fe83a8d71bd1803ctest-elbv2/d88c129308d731ef"
                    ],
                    "port": 80,
                    "protocol": "TCP",
                    "targetGroupArn": "arn:aws:elasticloadbalancing:eu-west-2:177828803798:targetgroup/fake-targets/dcf0f1f60163c003",
                    "targetGroupName": "fake-targets",
                    "targetHealthDescriptions": [
                        {
                            "healthCheckPort": "80",
                            "target": {
                                "availabilityZone": "eu-west-2c",
                                "id": "10.174.145.37",
                                "port": 80
                            },
                            "targetHealth": {
                                "description": "Health checks failed",
                                "reason": "Target.FailedHealthChecks",
                                "state": "unhealthy"
                            }
                        },
                        {
                            "healthCheckPort": "80",
                            "target": {
                                "availabilityZone": "eu-west-2a",
                                "id": "10.174.145.5",
                                "port": 80
                            },
                            "targetHealth": {
                                "description": "Health checks failed",
                                "reason": "Target.FailedHealthChecks",
                                "state": "unhealthy"
                            }
                        },
                        {
                            "healthCheckPort": "80",
                            "target": {
                                "availabilityZone": "eu-west-2b",
                                "id": "10.174.145.21",
                                "port": 80
                            },
                            "targetHealth": {
                                "description": "Initial health checks in progress",
                                "reason": "Elb.InitialHealthChecking",
                                "state": "initial"
                            }
                        }
                    ],
                    "targetType": "ip",
                    "unhealthyThresholdCount": 3,
                    "vpcId": "vpc-0fe83a8d71bd1803c"
                }
            ],
            "type": "network",
            "vpcId": "vpc-0fe83a8d71bd1803c"
        }
    },
    "context": "177828803798.eu-west-2",
    "linkedItemRequests": [
        {
            "type": "dns",
            "query": "vpc-0fe83a8d71bd1803ctest-elbv2-d88c129308d731ef.elb.eu-west-2.amazonaws.com",
            "context": "global"
        }
    ]
}
```

#### `Get`

Gets a specific ELB by name.

**Query format:** The name of the ELB

#### `Find`

Finds all ELBs

### ec2-instance

Get instance info

```json
{
    "type": "ec2-instance",
    "uniqueAttribute": "instanceId",
    "attributes": {
        "attrStruct": {
            "amiLaunchIndex": 0,
            "architecture": "x86_64",
            "blockDeviceMappings": [
                {
                    "deviceName": "/dev/xvda",
                    "ebs": {
                        "attachTime": "2022-01-12T15:21:19Z",
                        "deleteOnTermination": true,
                        "status": "attached",
                        "volumeId": "vol-02d2952cd8fd23a62"
                    }
                }
            ],
            "bootMode": "",
            "capacityReservationSpecification": {
                "capacityReservationPreference": "open"
            },
            "clientToken": "0b501817-8c52-4c6c-b5dd-6f4a336ccd04",
            "cpuOptions": {
                "coreCount": 1,
                "threadsPerCore": 2
            },
            "ebsOptimized": false,
            "enaSupport": true,
            "enclaveOptions": {
                "enabled": false
            },
            "hibernationOptions": {
                "configured": false
            },
            "hypervisor": "xen",
            "imageId": "ami-0fdbd8587b1cf431e",
            "instanceId": "i-09b0c0768577775ef",
            "instanceLifecycle": "",
            "instanceType": "t3.micro",
            "launchTime": "2022-01-12T15:21:18Z",
            "metadataOptions": {
                "httpEndpoint": "enabled",
                "httpProtocolIpv6": "disabled",
                "httpPutResponseHopLimit": 1,
                "httpTokens": "optional",
                "state": "applied"
            },
            "monitoring": {
                "state": "disabled"
            },
            "networkInterfaces": [
                {
                    "attachment": {
                        "attachTime": "2022-01-12T15:21:18Z",
                        "attachmentId": "eni-attach-056b3d8e583a24eb2",
                        "deleteOnTermination": true,
                        "deviceIndex": 0,
                        "networkCardIndex": 0,
                        "status": "attached"
                    },
                    "description": "",
                    "groups": [
                        {
                            "groupId": "sg-012c2822f90f34249",
                            "groupName": "default"
                        }
                    ],
                    "interfaceType": "interface",
                    "ipv6Addresses": [],
                    "macAddress": "06:68:1f:90:08:70",
                    "networkInterfaceId": "eni-0379384dd689e7afc",
                    "ownerId": "177828803798",
                    "privateIpAddress": "10.174.145.13",
                    "privateIpAddresses": [
                        {
                            "primary": true,
                            "privateIpAddress": "10.174.145.13"
                        }
                    ],
                    "sourceDestCheck": true,
                    "status": "in-use",
                    "subnetId": "subnet-0889bae2a717b3ab9",
                    "vpcId": "vpc-0e506f1a2e3074376"
                }
            ],
            "placement": {
                "availabilityZone": "eu-west-2a",
                "groupName": "",
                "tenancy": "default"
            },
            "platform": "",
            "platformDetails": "Linux/UNIX",
            "privateDnsName": "ip-10-174-145-13.eu-west-2.compute.internal",
            "privateDnsNameOptions": {
                "enableResourceNameDnsAAAARecord": false,
                "enableResourceNameDnsARecord": false,
                "hostnameType": "ip-name"
            },
            "privateIpAddress": "10.174.145.13",
            "productCodes": [],
            "publicDnsName": "",
            "rootDeviceName": "/dev/xvda",
            "rootDeviceType": "ebs",
            "securityGroups": [
                {
                    "groupId": "sg-012c2822f90f34249",
                    "groupName": "default"
                }
            ],
            "sourceDestCheck": true,
            "state": {
                "code": 16,
                "name": "running"
            },
            "stateTransitionReason": "",
            "subnetId": "subnet-0889bae2a717b3ab9",
            "tags": [
                {
                    "key": "Purpose",
                    "value": "automated-testing-2022-01-12T15:21:16.012Z"
                },
                {
                    "key": "Name",
                    "value": "automated-testing-2022-01-12T15:21:16.012Z"
                }
            ],
            "usageOperation": "RunInstances",
            "usageOperationUpdateTime": "2022-01-12T15:21:18Z",
            "virtualizationType": "hvm",
            "vpcId": "vpc-0e506f1a2e3074376"
        }
    },
    "context": "177828803798.eu-west-2",
    "linkedItemRequests": [
        {
            "type": "ec2-image",
            "query": "ami-0fdbd8587b1cf431e",
            "context": "177828803798.eu-west-2"
        },
        {
            "type": "ip",
            "query": "10.174.145.13",
            "context": "global"
        },
        {
            "type": "ec2-subnet",
            "query": "subnet-0889bae2a717b3ab9",
            "context": "177828803798.eu-west-2"
        },
        {
            "type": "ec2-vpc",
            "query": "vpc-0e506f1a2e3074376",
            "context": "177828803798.eu-west-2"
        },
        {
            "type": "ec2-securitygroup",
            "query": "sg-012c2822f90f34249",
            "context": "177828803798.eu-west-2"
        }
    ]
}
```

#### `Get`

Gets a specific instance by ID.

**Query format:** The ID of the instance e.g. `i-09b0c0768577775ef`

#### `Find`

Finds all instances

## Config

All configuration options can be provided via the command line or as environment variables:

| Environment Variable | CLI Flag | Automatic | Description |
|----------------------|----------|-----------|-------------|
| `CONFIG`| `--config` | ✅ | Config file location. Can be used instead of the CLI or environment variables if needed |
| `LOG`| `--log` | ✅ | Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace |
| `NATS_SERVERS`| `--nats-servers` | ✅ | A list of NATS servers to connect to |
| `NATS_NAME_PREFIX`| `--nats-name-prefix` | ✅ | A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client |
| `NATS_JWT` | `--nats-jwt` | ✅ | The JWT token that should be used to authenticate to NATS, provided in raw format e.g. `eyJ0eXAiOiJKV1Q{...}` |
| `NATS_NKEY_SEED` | `--nats-nkey-seed` | ✅ | The NKey seed which corresponds to the NATS JWT e.g. `SUAFK6QUC{...}` |
| `MAX_PARALLEL`| `--max-parallel` | ✅ | Max number of requests to run in parallel |
| `AUTO_CONFIG` | `--auto-config` | | Use the local AWS config, the same as the AWS CLI could use. This can be set up with `aws configure` |
| `AWS_ACCESS_KEY_ID` | `--aws-access-key-id` | | The ID of the access key to use |
| `AWS_REGIONS` | `--aws-region` | | Comma-separated list of AWS regions that this source should operate in |
| `AWS_SECRET_ACCESS_KEY` | `--aws-secret-access-key` | | The secret access key to use for auth |

### `srcman` config

When running in srcman, all of the above parameters marked with a checkbox are provided automatically, any additional parameters must be provided under the `config` key. These key-value pairs will become files in the `/etc/srcman/config` directory within the container.

```yaml
apiVersion: srcman.example.com/v0
kind: Source
metadata:
  name: source-sample
spec:
  image: ghcr.io/overmindtech/aws-source:latest
  replicas: 2
  manager: manager-sample
  config:
    # Example config values (not real credentials, don't panic)
    source.yaml: |
      aws-access-key-id: HEGMPEX0232FKZ45FSGV
      aws-regions: eu-west-2,us-west-2
      aws-secret-access-key: ATOXsYtO1xBG3GjbPIWi7iIN0hZYY3gdmUhaEEC5

```

**NOTE:** Remove the above boilerplate once you know what configuration will be required.

### Health Check

The source hosts a health check on `:8080/healthz` which will return an error if NATS is not connected. An example Kubernetes readiness probe is:

```yaml
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
```

## Development

### Source Type Naming Convention

The naming convention for types is as follows:

`{api}-{described_thing}-{version (optional)}`

**API:** The name of the api as it appears when making requests e.g. if you make requests to `https://ec2.amazonaws.com/?Action=CreateRouteTable` then the API name would be `ec2`. 

**Described Thing:** What is being described as derived from the name of the API action. For example if the action was `DescribeInstances` then the described thing would be `instance`. Note that plurals should be converted so singular.

**Version:** This is an optional parameter used when required. For example Elastic Load Balancing has the concept of `v1` and `v2` load balancers. However in the API these versions are `2012-06-01` and `2015-12-01` respectively. In cases like this, the more human readable name should be used such as `v2`

Some full examples of this naming convention therefore are:

* ec2-instance
* elasticloadbalancing-loadbalancer-v2

Check the [AWS API Documentation](https://docs.aws.amazon.com/) in order to predict other names.

### Running Locally

The source CLI can be interacted with locally by running:

```shell
go run main.go --help
```

### Testing

Tests in this package can be run using:

```shell
go test ./...
```

Note that these tests require building real AWS resources in order to test against them. This mean you'll need local credentials and running the test will actually build real resources (and clean them up). At time of writing the tests cost approx $0.03 per run.

### Packaging

Docker images can be created manually using `docker build`, but GitHub actions also exist that are able to create, tag and push images. Images will be build for the `main` branch, and also for any commits tagged with a version such as `v1.2.0`
