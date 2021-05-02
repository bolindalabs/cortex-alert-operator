/*
Copyright 2021.

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
	"fmt"

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

	Cortex *cortex.Client
}

//+kubebuilder:rbac:groups=monitoring.bolinda.digital,resources=prometheusrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.bolinda.digital,resources=prometheusrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitoring.bolinda.digital,resources=prometheusrules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PrometheusRule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *PrometheusRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("prometheusrule", req.NamespacedName)

	var rule monitoringv1.PrometheusRule
	if err := r.Get(ctx, req.NamespacedName, &rule); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch PrometheusRule")
		return ctrl.Result{}, err
	}

	cortexNamespace := rule.Namespace + "--" + rule.Name

	switch {
	case !r.hasFinalizer(rule) && !r.isDeletionScheduled(rule):
		if err := r.addFinalizer(ctx, rule, log); err != nil {
			log.Error(err, "unable to add finalizer")
			return ctrl.Result{}, err
		}
	case r.isDeletionScheduled(rule):
		if err := r.Cortex.DeleteRuleNamespace(log, cortexNamespace); err != nil {
			log.Error(err, "unable to delete rule namespace")
			return ctrl.Result{}, err
		}
		if err := r.removeFinalizer(ctx, rule, log); err != nil {
			log.Error(err, "unable to remove finalizer")
			return ctrl.Result{}, err
		}
	default:
		for _, g := range rule.Spec.Groups {
			if err := r.Cortex.SetRuleGroup(log, cortexNamespace, g); err != nil {
				log.Error(err, "unable to set rule group")

				if err := r.setStatus(ctx, rule, fmt.Sprintf("unable to set rule group: %v", err)); err != nil {
					log.Error(err, "unable to set rule group")
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}

			if err := r.setStatus(ctx, rule, "synced"); err != nil {
				log.Error(err, "unable to set status")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// setStatus sets PrometheusStatus.
func (r *PrometheusRuleReconciler) setStatus(ctx context.Context, rule monitoringv1.PrometheusRule, status string) error {
	newRule := rule.DeepCopy()
	newRule.Status.SyncStatus = status
	if err := r.Patch(ctx, newRule, client.MergeFrom(&rule)); err != nil {
		return err
	}

	return nil
}

// hasFinalizer checks if PrometheusRule has our finalizer set.
func (r *PrometheusRuleReconciler) hasFinalizer(rule monitoringv1.PrometheusRule) bool {
	return containsString(rule.ObjectMeta.Finalizers, finalizerName)
}

// isDeletionScheduled checks if the current PrometheusRule is scheduled for deletion.
// That means a deletion timestamp is set, but it is not completely deleted as there may be finalizers on that object.
func (r *PrometheusRuleReconciler) isDeletionScheduled(rule monitoringv1.PrometheusRule) bool {
	return !rule.ObjectMeta.DeletionTimestamp.IsZero()
}

// removeFinalizer removes our finalizer from the current PrometheusRule.
func (r *PrometheusRuleReconciler) removeFinalizer(ctx context.Context, rule monitoringv1.PrometheusRule, log logr.Logger) error {
	log.Info("Removing finalizer")

	newRule := rule.DeepCopy()
	newRule.ObjectMeta.Finalizers = removeString(rule.ObjectMeta.Finalizers, finalizerName)
	if err := r.Patch(ctx, newRule, client.MergeFrom(&rule)); err != nil {
		return err
	}

	return nil
}

// addFinalizer patches the current PrometheusRule, so that it contains our finalizer.
func (r *PrometheusRuleReconciler) addFinalizer(ctx context.Context, rule monitoringv1.PrometheusRule, log logr.Logger) error {
	log.Info("Adding finalizer")

	newRule := rule.DeepCopy()
	newRule.ObjectMeta.Finalizers = append(newRule.ObjectMeta.Finalizers, finalizerName)
	if err := r.Patch(ctx, newRule, client.MergeFrom(&rule)); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PrometheusRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1.PrometheusRule{}).
		Complete(r)
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
