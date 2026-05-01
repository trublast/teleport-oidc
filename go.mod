module github.com/gravitational/teleport

go 1.26.0

require (
	cloud.google.com/go/compute v1.54.0
	cloud.google.com/go/compute/metadata v0.9.0
	cloud.google.com/go/container v1.46.0
	cloud.google.com/go/firestore v1.21.0
	cloud.google.com/go/iam v1.5.3
	cloud.google.com/go/kms v1.26.0
	cloud.google.com/go/storage v1.57.1
	connectrpc.com/connect v1.16.2
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.21.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v3 v3.0.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2 v2.4.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql v1.1.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers v1.1.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql v1.1.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis/v2 v2.3.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise v1.1.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.1.0
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c
	github.com/ClickHouse/ch-go v0.69.0
	github.com/ClickHouse/clickhouse-go/v2 v2.42.0
	github.com/HdrHistogram/hdrhistogram-go v1.1.2
	github.com/Microsoft/go-winio v0.6.2
	github.com/ThalesIgnite/crypto11 v1.2.5
	github.com/alecthomas/kingpin/v2 v2.4.0 // replaced
	github.com/alicebob/miniredis/v2 v2.30.5
	github.com/andybalholm/brotli v1.2.0
	github.com/aquasecurity/libbpfgo v0.5.1-libbpf-1.2
	github.com/armon/go-radix v1.0.0
	github.com/aws/aws-sdk-go v1.55.8
	github.com/aws/aws-sdk-go-v2 v1.41.5
	github.com/aws/aws-sdk-go-v2/config v1.32.12
	github.com/aws/aws-sdk-go-v2/credentials v1.19.12
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.13.7
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.20
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.10
	github.com/aws/aws-sdk-go-v2/service/athena v1.40.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.30.2
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.149.3
	github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect v1.22.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.41.1
	github.com/aws/aws-sdk-go-v2/service/glue v1.77.2
	github.com/aws/aws-sdk-go-v2/service/iam v1.31.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.74.1
	github.com/aws/aws-sdk-go-v2/service/redshift v1.44.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.97.3
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.28.6
	github.com/aws/aws-sdk-go-v2/service/sns v1.29.1
	github.com/aws/aws-sdk-go-v2/service/sqs v1.31.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.9
	github.com/aws/aws-sigv4-auth-cassandra-gocql-driver-plugin v0.0.0-20220331165046-e4d000c0d6a6
	github.com/aws/smithy-go v1.24.2
	github.com/beevik/etree v1.6.0
	github.com/buildkite/bintest/v3 v3.1.1
	github.com/charmbracelet/bubbles v0.17.1
	github.com/charmbracelet/bubbletea v0.25.0
	github.com/charmbracelet/lipgloss v0.9.1
	github.com/coreos/go-oidc v2.5.0+incompatible // replaced
	github.com/coreos/go-semver v0.3.1
	github.com/creack/pty v1.1.18
	github.com/crewjam/saml v0.4.14
	github.com/datastax/go-cassandra-native-protocol v0.0.0-20220706104457-5e8aad05cf90
	github.com/digitorus/pkcs7 v0.0.0-20230818184609-3a137a874352
	github.com/dustin/go-humanize v1.0.1
	github.com/elastic/go-elasticsearch/v8 v8.9.0
	github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/fsouza/fake-gcs-server v1.47.4
	github.com/fxamacker/cbor/v2 v2.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/gizak/termui/v3 v3.1.0
	github.com/go-jose/go-jose/v3 v3.0.5
	github.com/go-ldap/ldap/v3 v3.4.5
	github.com/go-logr/logr v1.4.3
	github.com/go-mysql-org/go-mysql v1.5.0 // replaced
	github.com/go-piv/piv-go v1.11.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-webauthn/webauthn v0.8.6
	github.com/gocql/gocql v1.5.2
	github.com/gofrs/flock v0.8.1
	github.com/gogo/protobuf v1.3.2 // replaced
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/google/btree v1.1.3
	github.com/google/go-attestation v0.5.0
	github.com/google/go-cmp v0.7.0
	github.com/google/go-querystring v1.1.0
	github.com/google/go-tpm v0.9.0
	github.com/google/go-tpm-tools v0.4.4
	github.com/google/safetext v0.0.0-20240104143208-7a7d9b3d812f
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid v1.6.0
	github.com/googleapis/gax-go/v2 v2.19.0
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/gravitational/form v0.0.0-20151109031454-c4048f792f70
	github.com/gravitational/license v0.0.0-20231228155916-928ed9ac0335
	github.com/gravitational/oxy v0.0.0-20221029012416-9fbf4c444680
	github.com/gravitational/roundtrip v1.0.2
	github.com/gravitational/teleport/api v0.0.0
	github.com/gravitational/trace v1.3.1
	github.com/gravitational/ttlmap v0.0.0-20171116003245-91fd36b9004c
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.1.0
	github.com/guptarohit/asciigraph v0.5.6
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/icza/mjpeg v0.0.0-20220812133530-f79265a232f2
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa
	github.com/jackc/pgx/v5 v5.9.2
	github.com/jcmturner/gokrb5/v8 v8.4.4
	github.com/johannesboyne/gofakes3 v0.0.0-20210217223559-02ffa763be97
	github.com/jonboulle/clockwork v0.5.0
	github.com/joshlf/go-acl v0.0.0-20200411065538-eae00ae38531
	github.com/json-iterator/go v1.1.12
	github.com/julienschmidt/httprouter v1.3.0 // replaced
	github.com/keys-pub/go-libfido2 v1.5.3-0.20220306005615-8ab03fb1ec27 // replaced
	github.com/mailgun/holster/v3 v3.16.2
	github.com/mailgun/timetools v0.0.0-20170619190023-f3a7b8ffff47
	github.com/mailgun/ttlmap v0.0.0-20170619185759-c1c17f74874f
	github.com/mattn/go-sqlite3 v1.14.28
	github.com/mdlayher/netlink v1.7.2
	github.com/microsoft/go-mssqldb v0.0.0-00010101000000-000000000000 // replaced
	github.com/miekg/pkcs11 v1.1.1
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c
	github.com/moby/term v0.5.2
	github.com/okta/okta-sdk-golang/v2 v2.20.0
	github.com/opensearch-project/opensearch-go/v2 v2.3.0
	github.com/pavlo-v-chernykh/keystore-go/v4 v4.4.1
	github.com/pelletier/go-toml v1.9.5
	github.com/pkg/sftp v1.13.6
	github.com/pquerna/otp v1.4.0
	github.com/prometheus/client_golang v1.23.2
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.67.5
	github.com/redis/go-redis/v9 v9.17.2 // replaced
	github.com/russellhaering/gosaml2 v0.11.0
	github.com/russellhaering/goxmldsig v1.6.0
	github.com/sashabaranov/go-openai v1.14.2
	github.com/schollz/progressbar/v3 v3.13.1
	github.com/sijms/go-ora/v2 v2.7.9
	github.com/sirupsen/logrus v1.9.3
	github.com/snowflakedb/gosnowflake v1.13.3
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	github.com/ucarion/urlpath v0.0.0-20200424170820-7ccc79b76bbb
	github.com/vulcand/predicate v1.2.0 // replaced
	github.com/xanzy/go-gitlab v0.103.0
	go.etcd.io/etcd/api/v3 v3.6.8
	go.etcd.io/etcd/client/v3 v3.6.8
	go.mongodb.org/mongo-driver v1.17.6
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.46.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.68.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0
	go.opentelemetry.io/otel v1.43.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.43.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.43.0
	go.opentelemetry.io/otel/metric v1.43.0
	go.opentelemetry.io/otel/sdk v1.43.0
	go.opentelemetry.io/otel/trace v1.43.0
	go.opentelemetry.io/proto/otlp v1.10.0
	golang.org/x/crypto v0.50.0
	golang.org/x/exp v0.0.0-20251219203646-944ab1f22d93
	golang.org/x/mod v0.34.0
	golang.org/x/net v0.53.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/sync v0.20.0
	golang.org/x/sys v0.43.0
	golang.org/x/term v0.42.0
	golang.org/x/text v0.36.0
	golang.org/x/time v0.15.0
	google.golang.org/api v0.272.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260406210006-6f92a3bedf2d
	google.golang.org/grpc v1.80.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/dnaeon/go-vcr.v3 v3.1.2
	gopkg.in/ini.v1 v1.67.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.36.0
	k8s.io/apiextensions-apiserver v0.36.0
	k8s.io/apimachinery v0.36.0
	k8s.io/apiserver v0.36.0
	k8s.io/cli-runtime v0.36.0
	k8s.io/client-go v0.36.0
	k8s.io/component-base v0.36.0
	k8s.io/klog/v2 v2.140.0
	k8s.io/kubectl v0.36.0
	sigs.k8s.io/controller-runtime v0.24.0
	sigs.k8s.io/controller-tools v0.20.1
	sigs.k8s.io/yaml v1.6.0
	software.sslmate.com/src/go-pkcs12 v0.4.0
)

// Indirect mailgun dependencies.
// Updating causes breaking changes.
require github.com/mailgun/minheap v0.0.0-20170619185613-3dbe6c6bf55f // indirect

require (
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/distribution/reference v0.6.0
	github.com/google/go-containerregistry v0.20.7
	github.com/opencontainers/go-digest v1.0.0
	github.com/parquet-go/parquet-go v0.26.3
	github.com/sigstore/cosign/v2 v2.6.2
	github.com/sigstore/sigstore v1.10.5
)

require (
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.18.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/longrunning v0.8.0 // indirect
	cloud.google.com/go/monitoring v1.24.3 // indirect
	cloud.google.com/go/pubsub v1.50.1 // indirect
	cloud.google.com/go/pubsub/v2 v2.3.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/go-ntlmssp v0.1.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.31.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.54.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.54.0 // indirect
	github.com/JohnCGriffin/overflow v0.0.0-20211019200055-46fa312c352c // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/apache/arrow-go/v18 v18.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.20.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.17 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cloudflare/cfssl v1.6.4 // indirect
	github.com/cncf/xds/go v0.0.0-20251210132809-ee656c7534f5 // indirect
	github.com/containerd/console v1.0.4-0.20230313162750-1ae8d489ac81 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.18.1 // indirect
	github.com/coreos/go-oidc/v3 v3.17.0 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/coreos/pkg v0.0.0-20220810130054-c7d1c02cb6cf // indirect
	github.com/crewjam/httperr v0.2.0 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20241213102144-19d51d7fe467 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/digitorus/timestamp v0.0.0-20231217203849-220c5c2851b7 // indirect
	github.com/dmarkham/enumer v1.6.1 // indirect
	github.com/docker/cli v29.4.1+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.9.3 // indirect
	github.com/dvsekhvalnov/jose2go v1.7.0 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.3.0 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.36.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.4 // indirect
	github.com/go-chi/chi/v5 v5.2.5 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/analysis v0.24.3 // indirect
	github.com/go-openapi/errors v0.22.7 // indirect
	github.com/go-openapi/jsonpointer v0.22.5 // indirect
	github.com/go-openapi/jsonreference v0.21.5 // indirect
	github.com/go-openapi/loads v0.23.3 // indirect
	github.com/go-openapi/runtime v0.29.3 // indirect
	github.com/go-openapi/spec v0.22.4 // indirect
	github.com/go-openapi/strfmt v0.26.1 // indirect
	github.com/go-openapi/swag v0.25.5 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.5 // indirect
	github.com/go-openapi/swag/conv v0.25.5 // indirect
	github.com/go-openapi/swag/fileutils v0.25.5 // indirect
	github.com/go-openapi/swag/jsonname v0.25.5 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.5 // indirect
	github.com/go-openapi/swag/loading v0.25.5 // indirect
	github.com/go-openapi/swag/mangling v0.25.5 // indirect
	github.com/go-openapi/swag/netutils v0.25.5 // indirect
	github.com/go-openapi/swag/stringutils v0.25.5 // indirect
	github.com/go-openapi/swag/typeutils v0.25.5 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.5 // indirect
	github.com/go-openapi/validate v0.25.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/go-webauthn/x v0.1.4 // indirect
	github.com/gobuffalo/flect v1.0.3 // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/certificate-transparency-go v1.3.2 // indirect
	github.com/google/flatbuffers v24.12.23+incompatible // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-configfs-tsm v0.2.2 // indirect
	github.com/google/go-tspi v0.3.0 // indirect
	github.com/google/renameio/v2 v2.0.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.14 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/in-toto/attestation v1.1.2 // indirect
	github.com/in-toto/in-toto-golang v0.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jedisct1/go-minisign v0.0.0-20230811132847-661be99b8267 // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/joshlf/testutil v0.0.0-20170608050642-b5d8aa79d93d // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/letsencrypt/boulder v0.20260223.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattermost/xml-roundtrip-validator v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/moby/spdystream v0.5.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/muesli/ansi v0.0.0-20211018074035-2e021307bc4b // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481 // indirect
	github.com/nsf/termbox-go v1.1.1 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/parquet-go/bitpack v1.0.0 // indirect
	github.com/parquet-go/jsonlite v1.0.0 // indirect
	github.com/pascaldekloe/name v1.0.1 // indirect
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627 // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pingcap/errors v0.11.5-0.20201126102027-b0a155152ca3 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/xattr v0.4.9 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rs/zerolog v1.28.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/sasha-s/go-deadlock v0.0.0-20180226215254-237a9547c8a5 // indirect
	github.com/sassoftware/relic v7.2.1+incompatible // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.10.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shabbyrobe/gocovmerge v0.0.0-20190829150210-3e036491d500 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726 // indirect
	github.com/siddontang/go-log v0.0.0-20180807004314-8d05993dda07 // indirect
	github.com/sigstore/protobuf-specs v0.5.0 // indirect
	github.com/sigstore/rekor v1.5.0 // indirect
	github.com/sigstore/rekor-tiles/v2 v2.0.1 // indirect
	github.com/sigstore/sigstore-go v1.1.4 // indirect
	github.com/sigstore/timestamp-authority/v2 v2.0.6 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/thales-e-security/pool v0.0.2 // indirect
	github.com/theupdateframework/go-tuf v0.7.0 // indirect
	github.com/theupdateframework/go-tuf/v2 v2.4.1 // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	github.com/transparency-dev/formats v0.0.0-20251017110053-404c0d5b696c // indirect
	github.com/transparency-dev/merkle v0.0.2 // indirect
	github.com/vbatts/tar-split v0.12.2 // indirect
	github.com/weppos/publicsuffix-go v0.50.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yuin/gopher-lua v1.1.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	github.com/zmap/zcrypto v0.0.0-20250129210703-03c45d0bae98 // indirect
	github.com/zmap/zlint/v3 v3.6.6 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.6.8 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/telemetry v0.0.0-20260311193753-579e4da9a98c // indirect
	golang.org/x/tools v0.43.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/genproto v0.0.0-20260316180232-0b37fe3546d5 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260401024825-9d38bb4040a9 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	k8s.io/component-helpers v0.36.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	k8s.io/metrics v0.36.0 // indirect
	k8s.io/streaming v0.36.0 // indirect
	k8s.io/utils v0.0.0-20260210185600-b8788abfbbc2 // indirect
	mvdan.cc/sh/v3 v3.7.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kustomize/api v0.21.1 // indirect
	sigs.k8s.io/kustomize/kustomize/v5 v5.8.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.1 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
)

// Update also `ignore` in .github/dependabot.yml.
replace (
	github.com/alecthomas/kingpin/v2 => github.com/gravitational/kingpin/v2 v2.1.11-0.20230515143221-4ec6b70ecd33
	github.com/coreos/go-oidc => github.com/gravitational/go-oidc v0.1.1
	github.com/crewjam/saml => github.com/gravitational/saml v0.4.15-teleport.1
	github.com/datastax/go-cassandra-native-protocol => github.com/gravitational/go-cassandra-native-protocol v0.0.0-20221005103706-b9e66c056e90
	github.com/go-mysql-org/go-mysql => github.com/gravitational/go-mysql v1.5.0-teleport.1
	github.com/gogo/protobuf => github.com/gravitational/protobuf v1.3.2-teleport.1
	github.com/gravitational/teleport/api => ./api
	github.com/julienschmidt/httprouter => github.com/gravitational/httprouter v1.3.1-0.20220408074523-c876c5e705a5
	github.com/keys-pub/go-libfido2 => github.com/gravitational/go-libfido2 v1.5.3-teleport.1
	github.com/microsoft/go-mssqldb => github.com/gravitational/go-mssqldb v0.11.1-0.20230331180905-0f76f1751cd3
	// replace module github.com/moby/spdystream until https://github.com/moby/spdystream/pull/91 merges and deps are updated
	// otherwise tests fail with a data race detection.
	github.com/moby/spdystream => github.com/gravitational/spdystream v0.0.0-20230512133543-4e46862ca9bf
	github.com/redis/go-redis/v9 => github.com/gravitational/redis/v9 v9.0.2-teleport.2
	github.com/sijms/go-ora/v2 => github.com/gravitational/go-ora/v2 v2.0.0-20230821114616-e2a9f1131a46
	github.com/vulcand/predicate => github.com/gravitational/predicate v1.3.1
)
