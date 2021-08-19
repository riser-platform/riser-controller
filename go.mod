module riser-controller

go 1.17

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.5 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/google/go-containerregistry v0.0.0-20191108172333-79629ba8e9a1 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/riser-platform/riser-server/api/v1/model v0.0.16
	github.com/riser-platform/riser-server/pkg/sdk v0.0.41
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.17.9
	k8s.io/apimachinery v0.17.9
	k8s.io/client-go v0.17.9
	knative.dev/pkg v0.0.0-20200414233146-0eed424fa4ee
	// knative serving pins to a non tagged version of client-go. Need to address that upstream before upgrading.
	knative.dev/serving v0.14.0
	sigs.k8s.io/controller-runtime v0.5.8
)

require (
	cloud.google.com/go v0.43.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-logr/zapr v0.1.0 // indirect
	github.com/go-ozzo/ozzo-validation/v3 v3.8.1 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.4.1 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 // indirect
	golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	gomodules.xyz/jsonpatch/v2 v2.0.1 // indirect
	google.golang.org/appengine v1.6.1 // indirect
	google.golang.org/protobuf v1.23.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
	k8s.io/apiextensions-apiserver v0.17.8 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29 // indirect
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)
