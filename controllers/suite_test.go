package controllers

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	monitoringv1 "github.com/bolindalabs/cortex-alert-operator/api/v1"
	"github.com/bolindalabs/cortex-alert-operator/controllers/cortex"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var server *ghttp.Server
var prometheusRuleReconciler *PrometheusRuleReconciler

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = monitoringv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	// Initially load some cortexClient
	var cortexClient *cortex.Client
	{
		c := cortex.Config{
			Key:             "",
			Address:         "",
			ID:              "",
			UseLegacyRoutes: false,
		}
		cortexClient, err = cortex.New(c)
		Expect(err).ToNot(HaveOccurred())
	}

	prometheusRuleReconciler = &PrometheusRuleReconciler{
		Client: k8sManager.GetClient(),
		Cortex: cortexClient,
		Log:    ctrl.Log.WithName("controllers").WithName("PrometheusRule"),
	}

	err = prometheusRuleReconciler.SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// Start and Stop test server between each test run.
var _ = BeforeEach(func() {
	server = ghttp.NewServer()
	c := cortex.Config{
		Key:             "",
		Address:         server.URL(),
		ID:              "",
		UseLegacyRoutes: false,
	}
	cortexClient, err := cortex.New(c)
	Expect(err).ToNot(HaveOccurred())

	prometheusRuleReconciler.Cortex = cortexClient
})

var _ = AfterEach(func() {
	server.Close()
})
