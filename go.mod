module riser-controller

go 1.12

// because these packages don't use semver (https://github.com/kubernetes-sigs/kubebuilder/issues/675)
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190722141453-b90922c02518
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

require (
	cloud.google.com/go v0.40.0 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/riser-platform/riser-server/api/v1/model v0.0.0-20191104153455-93fb7d7070d2
	github.com/riser-platform/riser/sdk v0.0.0-20191104155518-e7f41d859d58
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472 // indirect
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/sys v0.0.0-20190830080133-08d80c9d36de // indirect
	google.golang.org/appengine v1.6.1 // indirect
	k8s.io/api v0.0.0-20190830074751-c43c3e1d5a79
	k8s.io/apimachinery v0.0.0-20191102025618-50aa20a7b23f
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/controller-runtime v0.2.1

)
