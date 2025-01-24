module github.com/overmindtech/aws-source

go 1.22.7

toolchain go1.23.5

// Direct dependencies
require (
	github.com/MrAlias/otel-schema-utils v0.2.1-alpha
	github.com/aws/aws-sdk-go v1.55.6
	github.com/aws/aws-sdk-go-v2 v1.33.0
	github.com/aws/aws-sdk-go-v2/config v1.29.1
	github.com/aws/aws-sdk-go-v2/credentials v1.17.54
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.28.7
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.51.7
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.44.5
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.43.9
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.30.7
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.39.5
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.201.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.53.8
	github.com/aws/aws-sdk-go-v2/service/efs v1.34.6
	github.com/aws/aws-sdk-go-v2/service/eks v1.56.5
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.28.12
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.43.7
	github.com/aws/aws-sdk-go-v2/service/iam v1.38.7
	github.com/aws/aws-sdk-go-v2/service/kms v1.37.13
	github.com/aws/aws-sdk-go-v2/service/lambda v1.69.7
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.44.10
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.32.6
	github.com/aws/aws-sdk-go-v2/service/rds v1.93.7
	github.com/aws/aws-sdk-go-v2/service/route53 v1.48.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.74.0
	github.com/aws/aws-sdk-go-v2/service/sns v1.33.14
	github.com/aws/aws-sdk-go-v2/service/sqs v1.37.9
	github.com/aws/aws-sdk-go-v2/service/ssm v1.56.7
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.9
	github.com/aws/smithy-go v1.22.1
	github.com/getsentry/sentry-go v0.31.1
	github.com/micahhausler/aws-iam-policy v0.4.2
	github.com/overmindtech/discovery v0.33.4
	github.com/overmindtech/sdp-go v0.106.0
	github.com/overmindtech/sdpcache v1.6.4
	github.com/sirupsen/logrus v1.9.3
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.19.0
	go.opentelemetry.io/contrib/detectors/aws/ec2 v1.33.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0
	go.opentelemetry.io/otel v1.33.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.33.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.33.0
	go.opentelemetry.io/otel/sdk v1.33.0
	go.opentelemetry.io/otel/trace v1.33.0
	go.uber.org/automaxprocs v1.6.0
	google.golang.org/protobuf v1.36.3
)

// Transitive dependencies
require (
	connectrpc.com/connect v1.18.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/auth0/go-jwt-middleware/v2 v2.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.28 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.10 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.24.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/jwt/v2 v2.7.3 // indirect
	github.com/nats-io/nats.go v1.38.0 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/schema v0.0.7 // indirect
	go.opentelemetry.io/proto/otlp v1.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/exp v0.0.0-20231206192017-f3f8817b8deb // indirect
	golang.org/x/net v0.32.0 // indirect
	golang.org/x/oauth2 v0.25.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/grpc v1.68.1 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require go.opentelemetry.io/auto/sdk v1.1.0 // indirect
