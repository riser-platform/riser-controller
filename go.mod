module riser-controller

go 1.14

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.5 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/google/go-containerregistry v0.0.0-20191108172333-79629ba8e9a1 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/riser-platform/riser-server/api/v1/model v0.0.10
	github.com/riser-platform/riser-server/pkg/sdk v0.0.32
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472 // indirect
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	knative.dev/pkg v0.0.0-20200414233146-0eed424fa4ee
	knative.dev/serving v0.14.0
	sigs.k8s.io/controller-runtime v0.4.0
)
