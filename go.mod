module github.com/overmindtech/aws-source

go 1.19

// Direct dependencies
require (
	github.com/aws/aws-sdk-go-v2 v1.19.0
	github.com/aws/aws-sdk-go-v2/config v1.18.28
	github.com/aws/aws-sdk-go-v2/credentials v1.13.27
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.28.10
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.26.3
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.20.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.105.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.28.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.20.4
	github.com/aws/aws-sdk-go-v2/service/eks v1.27.15
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.15.13
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.19.14
	github.com/aws/aws-sdk-go-v2/service/iam v1.21.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.37.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.46.2
	github.com/aws/aws-sdk-go-v2/service/route53 v1.28.4
	github.com/aws/aws-sdk-go-v2/service/s3 v1.37.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.3
	github.com/aws/smithy-go v1.13.5
	github.com/getsentry/sentry-go v0.22.0
	github.com/iancoleman/strcase v0.2.0
	github.com/nats-io/jwt/v2 v2.4.1
	github.com/nats-io/nkeys v0.4.4
	github.com/overmindtech/discovery v0.21.7
	github.com/overmindtech/sdp-go v0.37.0
	github.com/sirupsen/logrus v1.9.3
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.16.0
	go.opentelemetry.io/contrib/detectors/aws/ec2 v1.17.0
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.16.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.16.0
	go.opentelemetry.io/otel/sdk v1.16.0
	google.golang.org/protobuf v1.31.0
)

// Transitive dependencies
require (
	github.com/auth0/go-jwt-middleware/v2 v2.1.0 // indirect
	github.com/aws/aws-sdk-go v1.44.285 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.13 // indirect
	github.com/bufbuild/connect-go v1.9.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/glog v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nats.go v1.28.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/overmindtech/api-client v0.14.0 // indirect
	github.com/overmindtech/sdpcache v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.16.0 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	go.opentelemetry.io/otel/trace v1.16.0
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/oauth2 v0.10.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230526203410-71b5a4ffd15e // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230525234035-dd9d682886f9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/grpc v1.55.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
