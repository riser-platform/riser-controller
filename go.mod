module riser-controller

go 1.12

// because these packages don't use semver (https://github.com/kubernetes-sigs/kubebuilder/issues/675)
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190722141453-b90922c02518
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.5 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/google/go-containerregistry v0.0.0-20191108172333-79629ba8e9a1 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/pkg/errors v0.8.1
	github.com/riser-platform/riser-server/api/v1/model v0.0.0-20191231160837-2bc69123b600
	github.com/riser-platform/riser/sdk v0.0.0-20191231161402-420e3e7c3087
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472 // indirect
	k8s.io/api v0.0.0-20190830074751-c43c3e1d5a79
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	knative.dev/pkg v0.0.0-20191107185656-884d50f09454 // indirect
	knative.dev/serving v0.10.0
	sigs.k8s.io/controller-runtime v0.2.1

)
