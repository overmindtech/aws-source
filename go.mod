module github.com/overmindtech/aws-source

go 1.22.5

// Direct dependencies
require (
	github.com/MrAlias/otel-schema-utils v0.2.1-alpha
	github.com/aws/aws-sdk-go-v2 v1.31.0
	github.com/aws/aws-sdk-go-v2/config v1.27.40
	github.com/aws/aws-sdk-go-v2/credentials v1.17.38
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.26.4
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.44.4
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.39.4
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.41.4
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.28.4
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.35.4
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.180.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.46.4
	github.com/aws/aws-sdk-go-v2/service/efs v1.32.4
	github.com/aws/aws-sdk-go-v2/service/eks v1.49.4
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.27.4
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.38.4
	github.com/aws/aws-sdk-go-v2/service/iam v1.36.4
	github.com/aws/aws-sdk-go-v2/service/kms v1.36.4
	github.com/aws/aws-sdk-go-v2/service/lambda v1.62.2
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.42.4
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.30.4
	github.com/aws/aws-sdk-go-v2/service/rds v1.86.1
	github.com/aws/aws-sdk-go-v2/service/route53 v1.44.4
	github.com/aws/aws-sdk-go-v2/service/s3 v1.64.1
	github.com/aws/aws-sdk-go-v2/service/sns v1.32.4
	github.com/aws/aws-sdk-go-v2/service/sqs v1.35.4
	github.com/aws/aws-sdk-go-v2/service/sts v1.31.4
	github.com/aws/smithy-go v1.22.0
	github.com/getsentry/sentry-go v0.29.0
	github.com/micahhausler/aws-iam-policy v0.4.2
	github.com/nats-io/jwt/v2 v2.7.0
	github.com/nats-io/nkeys v0.4.7
	github.com/overmindtech/discovery v0.28.2
	github.com/overmindtech/sdp-go v0.94.2
	github.com/overmindtech/sdpcache v1.6.4
	github.com/sirupsen/logrus v1.9.3
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.19.0
	go.opentelemetry.io/contrib/detectors/aws/ec2 v1.30.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.55.0
	go.opentelemetry.io/otel v1.30.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.30.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.30.0
	go.opentelemetry.io/otel/sdk v1.30.0
	go.opentelemetry.io/otel/trace v1.30.0
	go.uber.org/automaxprocs v1.6.0
	golang.org/x/oauth2 v0.23.0
	google.golang.org/protobuf v1.34.2
)

// Transitive dependencies
require (
	connectrpc.com/connect v1.17.0 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/auth0/go-jwt-middleware/v2 v2.2.2 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.23.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.27.4 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
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
	go.opentelemetry.io/otel/metric v1.30.0 // indirect
	go.opentelemetry.io/otel/schema v0.0.7 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/exp v0.0.0-20231206192017-f3f8817b8deb // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.66.1 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
