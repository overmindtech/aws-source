module github.com/overmindtech/aws-source

go 1.17

// Direct dependencies
require (
	github.com/aws/aws-sdk-go-v2 v1.12.0
	github.com/aws/aws-sdk-go-v2/config v1.12.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.27.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.11.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.15.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.13.0
	github.com/iancoleman/strcase v0.2.0
	github.com/overmindtech/discovery v0.9.4
	github.com/overmindtech/sdp-go v0.6.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
)

// Transitive dependencies
require (
	github.com/aws/aws-sdk-go-v2/credentials v1.7.0
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.8.0 // indirect
	github.com/aws/smithy-go v1.9.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/nats-io/jwt/v2 v2.2.0 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20211122170419-d7c1d78a50fc // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/overmindtech/sdpcache v0.1.4 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/spf13/afero v1.8.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3 // indirect
	golang.org/x/sys v0.0.0-20220111092808-5a964db01320 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
