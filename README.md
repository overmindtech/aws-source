# AWS Source

This source integrates with AWS, allowing Overmind to pull data about many types of AWS resources.

## Required Permissions

This source requires the following IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:Describe*",
        "cloudfront:Get*",
        "cloudfront:List*",
        "cloudwatch:Describe*",
        "cloudwatch:ListTagsForResource",
        "directconnect:Describe*",
        "dynamodb:Describe*",
        "dynamodb:List*",
        "ec2:Describe*",
        "ecs:Describe*",
        "ecs:List*",
        "eks:Describe*",
        "eks:List*",
        "elasticfilesystem:Describe*",
        "elasticloadbalancing:Describe*",
        "iam:Get*",
        "iam:List*",
        "lambda:Get*",
        "lambda:List*",
        "network-firewall:Describe*",
        "network-firewall:List*",
        "networkmanager:Describe*",
        "networkmanager:Get*",
        "networkmanager:List*",
        "rds:Describe*",
        "rds:ListTagsForResource",
        "route53:Get*",
        "route53:List*",
        "s3:GetBucket*",
        "s3:ListAllMyBuckets",
        "sns:Get*",
        "sns:List*",
        "sqs:Get*",
        "sqs:List*"
      ],
      "Resource": "*"
    }
  ]
}
```

## Naming Conventions

Types are named to match the `describe-*`, `get-*` or `list-*` command within the AWS CLI, with the service that they are part of as a prefix. For example to get the details if a security group you would run:

```
aws ec2 describe-security-groups
```

Meaning that the type in Overmind should be:

```
ec2-security-group
```

Note that plurals should be converted to their singular form hence `security-groups` becomes `security-group`

## Rate limiting

For EC2 APIs this sources uses the [same throttling methods as EC2 does](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/throttling.html), with the bucket size and refill rate set to 50% of the total. This means that the source will never use more than 50% of the available requests, including refil;ls when the bucket is empty.

## Config

All configuration options can be provided via the command line or as environment variables:

| Environment Variable    | CLI Flag                  | Automatic | Description                                                                                                                                                                                           |
|-------------------------|---------------------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `CONFIG`                | `--config`                | ✅         | Config file location. Can be used instead of the CLI or environment variables if needed                                                                                                               |
| `LOG`                   | `--log`                   | ✅         | Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace                                                                                                                        |
| `NATS_SERVERS`          | `--nats-servers`          | ✅         | A list of NATS servers to connect to                                                                                                                                                                  |
| `NATS_NAME_PREFIX`      | `--nats-name-prefix`      | ✅         | A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client |
| `NATS_JWT`              | `--nats-jwt`              | ✅         | The JWT token that should be used to authenticate to NATS, provided in raw format e.g. `eyJ0eXAiOiJKV1Q{...}`                                                                                         |
| `NATS_NKEY_SEED`        | `--nats-nkey-seed`        | ✅         | The NKey seed which corresponds to the NATS JWT e.g. `SUAFK6QUC{...}`                                                                                                                                 |
| `MAX_PARALLEL`          | `--max-parallel`          | ✅         | Max number of requests to run in parallel                                                                                                                                                             |
| `AUTO_CONFIG`           | `--auto-config`           |           | Use the local AWS config, the same as the AWS CLI could use. This can be set up with `aws configure`                                                                                                  |
| `AWS_REGIONS`           | `--aws-region`            |           | Comma-separated list of AWS regions that this source should operate in                                                                                                                                |
| `AWS_ACCESS_STRATEGY`   | `--aws-access-strategy`   |           | The strategy to use to access this customer's AWS account. Valid values: 'access-key', 'external-id', 'sso-profile', 'defaults'. Default: 'defaults'.                                                 |
| `AWS_ACCESS_KEY_ID`     | `--aws-access-key-id`     |           | The ID of the access key to use                                                                                                                                                                       |
| `AWS_SECRET_ACCESS_KEY` | `--aws-secret-access-key` |           | The secret access key to use for auth                                                                                                                                                                 |
| `AWS_EXTERNAL_ID`       | `--aws-external-id`       |           | The external ID to use when assuming the customer's role                                                                                                                                              |
| `AWS_TARGET_ROLE_ARN`   | `--aws-target-role-arn`   |           | The role to assume in the customer's account                                                                                                                                                          |
| `AWS_PROFILE`           | `--aws-profile`           |           | The AWS SSO Profile to use. Defaults to $AWS_PROFILE, then whatever the AWS SDK's SSO config defaults to                                                                                              |

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

`{api}-{described_thing}`

**API:** The name of the api as it appears in the AWS CLI. Get a full list using `aws help`

**Described Thing:** What is being described as derived from the name of the CLI action. For example if the action was `describe-instances` then the described thing would be `instance`. Note that plurals should be converted so singular.

Some full examples of this naming convention therefore are:

* `aws ec2 describe-instances`: ec2-instance
* `aws elbv2 describe-rules`: elbv2-rule

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

### Generating Docs

Source data for docs is stored in `docs-data` and can be generated using:

Ensure that [`docgen`](https://github.com/overmindtech/docgen) is installed.

From the root of the project run:
```shell
go generate ./...
```
