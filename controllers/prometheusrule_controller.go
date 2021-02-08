/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1 "github.com/bolindalabs/cortex-alert-operator/api/v1"
	"github.com/bolindalabs/cortex-alert-operator/controllers/cortex"
)

const finalizerName = "prometheus.monitoring.bolinda.digital"

// PrometheusRuleReconciler reconciles a PrometheusRule object
type PrometheusRuleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	cortex cortex.Client
}

// +kubebuilder:rbac:groups=monitoring.bolinda.digital,resources=prometheusrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.bolinda.digital,resources=prometheusrules/status,verbs=get;update;patch

func (r *PrometheusRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("prometheusrule", req.NamespacedName)

	var rule monitoringv1.PrometheusRule
	if err := r.Get(ctx, req.NamespacedName, &rule); err != nil {
		log.Error(err, "unable to fetch PrometheusRule")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cortexNamespace := rule.Namespace + "--" + rule.Name

	if !r.hasFinalizer(rule) && !r.isDeletionScheduled(rule) {
		if err := r.addFinalizer(rule); err != nil {
			log.Error(err, "unable to add finalizer")
			return ctrl.Result{}, err
		}
	}

	if r.isDeletionScheduled(rule) {
		if err := r.cortex.DeleteRuleNamespace(cortexNamespace); err != nil {
			log.Error(err, "unable to delete rule namespace")
			return ctrl.Result{}, err
		}

		if err := r.removeFinalizer(rule); err != nil {
			log.Error(err, "unable to remove finalizer")
			return ctrl.Result{}, err
		}
	} else {
		for _, g := range rule.Spec.Groups {
			if err := r.cortex.SetRuleGroup(cortexNamespace, g); err != nil {
				log.Error(err, "unable to set rule group")
				return ctrl.Result{}, err
			}
		}
	}

	rule.Status.SyncStatus = "Synced"
	if err := r.Status().Update(context.Background(), &rule); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PrometheusRuleReconciler) hasFinalizer(rule monitoringv1.PrometheusRule) bool {
	return containsString(rule.ObjectMeta.Finalizers, finalizerName)
}

func (r *PrometheusRuleReconciler) isDeletionScheduled(rule monitoringv1.PrometheusRule) bool {
	return !rule.ObjectMeta.DeletionTimestamp.IsZero()
}

func (r *PrometheusRuleReconciler) removeFinalizer(rule monitoringv1.PrometheusRule) error {
	rule.ObjectMeta.Finalizers = removeString(rule.ObjectMeta.Finalizers, finalizerName)
	if err := r.Update(context.Background(), &rule); err != nil {
		return err
	}

	return nil
}

func (r *PrometheusRuleReconciler) addFinalizer(rule monitoringv1.PrometheusRule) error {
	rule.ObjectMeta.Finalizers = append(rule.ObjectMeta.Finalizers, finalizerName)
	if err := r.Update(context.Background(), &rule); err != nil {
		return err
	}

	return nil
}

func (r *PrometheusRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1.PrometheusRule{}).
		Complete(r)
}

func (r *PrometheusRuleReconciler) deleteExternalResources(rule *monitoringv1.PrometheusRule) error {
	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
