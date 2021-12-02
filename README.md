# AWS Source

This source integrates with AWS, allowing Overmind to pull data about many types of AWS resources.

## Config

All configuration options can be provided via the command line or as environment variables:

| Environment Variable | CLI Flag | Automatic | Description |
|----------------------|----------|-----------|-------------|
| `CONFIG`| `--config` | ✅ | Config file location. Can be used instead of the CLI or environment variables if needed |
| `LOG`| `--log` | ✅ | Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace |
| `NATS_SERVERS`| `--nats-servers` | ✅ | A list of NATS servers to connect to |
| `NATS_NAME_PREFIX`| `--nats-name-prefix` | ✅ | A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client |
| `NATS_CA_FILE`| `--nats-ca-file` | ✅ | Path to the CA file that NATS should use when connecting over TLS |
| `NATS_JWT_FILE`| `--nats-jwt-file` | ✅ | Path to the file containing the user JWT |
| `NATS_NKEY_FILE`| `--nats-nkey-file` | ✅ | Path to the file containing the NKey seed |
| `MAX_PARALLEL`| `--max-parallel` | ✅ | Max number of requests to run in parallel |
| `AUTO_CONFIG` | `--auto-config` | | Use the local AWS config, the same as the AWS CLI could use. This can be set up with `aws configure` |
| `AWS_ACCESS_KEY_ID` | `--aws-access-key-id` | | The ID of the access key to use |
| `AWS_REGION` | `--aws-region` | | The AWS region that this source should operate in |
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
      aws-region: eu-west-2
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
