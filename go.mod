module github.com/overmindtech/aws-source

go 1.22.4

// Direct dependencies
require (
	github.com/MrAlias/otel-schema-utils v0.2.1-alpha
	github.com/aws/aws-sdk-go-v2 v1.30.0
	github.com/aws/aws-sdk-go-v2/config v1.27.21
	github.com/aws/aws-sdk-go-v2/credentials v1.17.21
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.41.1
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.37.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.39.1
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.26.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.33.2
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.165.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.43.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.30.1
	github.com/aws/aws-sdk-go-v2/service/eks v1.44.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.25.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.32.1
	github.com/aws/aws-sdk-go-v2/service/iam v1.33.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.55.1
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.39.1
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.27.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.80.1
	github.com/aws/aws-sdk-go-v2/service/route53 v1.41.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.56.1
	github.com/aws/aws-sdk-go-v2/service/sns v1.30.1
	github.com/aws/aws-sdk-go-v2/service/sqs v1.33.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.29.1
	github.com/aws/smithy-go v1.20.2
	github.com/getsentry/sentry-go v0.28.1
	github.com/iancoleman/strcase v0.2.0
	github.com/nats-io/jwt/v2 v2.5.7
	github.com/nats-io/nkeys v0.4.7
	github.com/overmindtech/discovery v0.27.6
	github.com/overmindtech/sdp-go v0.76.0
	github.com/overmindtech/sdpcache v1.6.4
	github.com/sirupsen/logrus v1.9.3
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.19.0
	go.opentelemetry.io/contrib/detectors/aws/ec2 v1.27.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.52.0
	go.opentelemetry.io/otel v1.27.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.27.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.27.0
	go.opentelemetry.io/otel/sdk v1.27.0
	go.opentelemetry.io/otel/trace v1.27.0
	go.uber.org/automaxprocs v1.5.3
	google.golang.org/protobuf v1.34.2
)

// Transitive dependencies
require (
	connectrpc.com/connect v1.16.2 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/auth0/go-jwt-middleware/v2 v2.2.1 // indirect
	github.com/aws/aws-sdk-go v1.53.6 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.21.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.25.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.2 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nats.go v1.36.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/schema v0.0.7 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/exp v0.0.0-20231206192017-f3f8817b8deb // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240520151616-dc85e6b867a5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240515191416-fc5f0ca64291 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
