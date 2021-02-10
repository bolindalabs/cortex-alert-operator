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

// +kubebuilder:rbac:groups=monitoring.bolinda.digital,resources=prometheusrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.bolinda.digital,resources=prometheusrules/status,verbs=get;update;patch

func (r *PrometheusRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
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

	if !r.hasFinalizer(rule) && !r.isDeletionScheduled(rule) {
		if err := r.addFinalizer(ctx, rule, log); err != nil {
			log.Error(err, "unable to add finalizer")
			return ctrl.Result{}, err
		}
	}

	if r.isDeletionScheduled(rule) {
		if err := r.Cortex.DeleteRuleNamespace(log, cortexNamespace); err != nil {
			log.Error(err, "unable to delete rule namespace")
			return ctrl.Result{}, err
		}

		if err := r.removeFinalizer(ctx, rule, log); err != nil {
			log.Error(err, "unable to remove finalizer")
			return ctrl.Result{}, err
		}
	} else {
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
