module github.com/bolindalabs/cortex-alert-operator

go 1.15

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	go.uber.org/zap v1.16.0
	k8s.io/apimachinery v0.17.17
	k8s.io/client-go v0.17.17
	sigs.k8s.io/controller-runtime v0.5.13
)
