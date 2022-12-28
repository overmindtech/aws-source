module github.com/overmindtech/aws-source

go 1.19

// Direct dependencies
require (
	github.com/aws/aws-sdk-go-v2 v1.17.3
	github.com/aws/aws-sdk-go-v2/config v1.18.7
	github.com/aws/aws-sdk-go-v2/credentials v1.13.7
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.77.0
	github.com/aws/aws-sdk-go-v2/service/eks v1.26.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.14.25
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.18.28
	github.com/aws/aws-sdk-go-v2/service/s3 v1.29.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.17.7
	github.com/iancoleman/strcase v0.2.0
	github.com/nats-io/jwt/v2 v2.3.0
	github.com/nats-io/nkeys v0.3.0
	github.com/overmindtech/connect v0.5.0
	github.com/overmindtech/discovery v0.16.0
	github.com/overmindtech/sdp-go v0.14.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.14.0
)

// Transitive dependencies
require (
	github.com/RoaringBitmap/roaring v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.28 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.11 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/bits-and-blooms/bitset v1.4.0 // indirect
	github.com/blevesearch/bleve/v2 v2.3.6 // indirect
	github.com/blevesearch/bleve_index_api v1.0.5 // indirect
	github.com/blevesearch/geo v0.1.17 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.1.4 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.0.9 // indirect
	github.com/blevesearch/zapx/v11 v11.3.7 // indirect
	github.com/blevesearch/zapx/v12 v12.3.7 // indirect
	github.com/blevesearch/zapx/v13 v13.3.7 // indirect
	github.com/blevesearch/zapx/v14 v14.3.7 // indirect
	github.com/blevesearch/zapx/v15 v15.3.8 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/nats-io/nats.go v1.22.1 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/overmindtech/api-client v0.5.0 // indirect
	github.com/overmindtech/sdpcache v0.4.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/crypto v0.4.0 // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/oauth2 v0.3.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/aws/aws-sdk-go-v2/service/route53 v1.25.2
