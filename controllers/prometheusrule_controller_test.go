package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	monitoringv1 "github.com/bolindalabs/cortex-alert-operator/api/v1"
)

var _ = Describe("PrometheusRule Controller", func() {
	const (
		PrometheusRuleName      = "test-prometheusrule"
		PrometheusRuleNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating PrometheusRule", func() {
		It("Should call Cortex API with the right arguments", func() {
			By("By applying a new PrometheusRule")
			server.AppendHandlers(
				ghttp.VerifyRequest("POST", "/api/v1/rules/default--test-prometheusrule"),
			)
			server.AppendHandlers(
				ghttp.VerifyRequest("POST", "/api/v1/rules/default--test-prometheusrule"),
			)

			ctx := context.Background()
			prometheusRule := &monitoringv1.PrometheusRule{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PrometheusRule",
					APIVersion: "monitoring.bolinda.digital/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      PrometheusRuleName,
					Namespace: PrometheusRuleNamespace,
				},
				Spec: monitoringv1.PrometheusRuleSpec{
					Groups: []monitoringv1.RuleGroup{
						{
							Name: "./example.rules",
							Rules: []monitoringv1.Rule{
								{
									Alert: "ExampleAlert",
									Expr:  intstr.FromString("vector(1)"),
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, prometheusRule)).Should(Succeed())

			prometheusRuleLookupKey := types.NamespacedName{Name: PrometheusRuleName, Namespace: PrometheusRuleNamespace}
			createdPrometheusRule := &monitoringv1.PrometheusRule{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, prometheusRuleLookupKey, createdPrometheusRule)
				return err == nil
			}, timeout, interval).Should(BeTrue(), "PrometheusRule should be stored and retrievable from K8s")

			Eventually(func() int {
				req := server.ReceivedRequests()
				return len(req)
			}, timeout, interval).Should(Equal(2))
		})
	})
})
