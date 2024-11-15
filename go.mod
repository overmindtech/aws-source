module github.com/overmindtech/aws-source

go 1.22.5

// Direct dependencies
require (
	github.com/MrAlias/otel-schema-utils v0.2.1-alpha
	github.com/aws/aws-sdk-go-v2 v1.32.4
	github.com/aws/aws-sdk-go-v2/config v1.28.4
	github.com/aws/aws-sdk-go-v2/credentials v1.17.45
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.27.5
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.48.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.41.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.42.4
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.29.5
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.37.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.188.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.49.2
	github.com/aws/aws-sdk-go-v2/service/efs v1.33.5
	github.com/aws/aws-sdk-go-v2/service/eks v1.52.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.28.4
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.41.1
	github.com/aws/aws-sdk-go-v2/service/iam v1.38.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.37.5
	github.com/aws/aws-sdk-go-v2/service/lambda v1.66.0
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.44.2
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.31.5
	github.com/aws/aws-sdk-go-v2/service/rds v1.89.2
	github.com/aws/aws-sdk-go-v2/service/route53 v1.46.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.67.0
	github.com/aws/aws-sdk-go-v2/service/sns v1.33.4
	github.com/aws/aws-sdk-go-v2/service/sqs v1.37.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.0
	github.com/aws/smithy-go v1.22.0
	github.com/getsentry/sentry-go v0.29.1
	github.com/micahhausler/aws-iam-policy v0.4.2
	github.com/nats-io/jwt/v2 v2.7.2
	github.com/nats-io/nkeys v0.4.7
	github.com/overmindtech/discovery v0.31.2
	github.com/overmindtech/sdp-go v0.100.0
	github.com/overmindtech/sdpcache v1.6.4
	github.com/sirupsen/logrus v1.9.3
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.19.0
	go.opentelemetry.io/contrib/detectors/aws/ec2 v1.31.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.56.0
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.31.0
	go.opentelemetry.io/otel/sdk v1.31.0
	go.opentelemetry.io/otel/trace v1.31.0
	go.uber.org/automaxprocs v1.6.0
	golang.org/x/oauth2 v0.24.0 // indirect
	google.golang.org/protobuf v1.35.2
)

// Transitive dependencies
require (
	connectrpc.com/connect v1.17.0 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/auth0/go-jwt-middleware/v2 v2.2.2 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.6 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.19 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.4 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nats.go v1.37.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/otel/metric v1.31.0 // indirect
	go.opentelemetry.io/otel/schema v0.0.7 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20231206192017-f3f8817b8deb // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
