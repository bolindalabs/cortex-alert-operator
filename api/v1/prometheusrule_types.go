package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PrometheusRuleSpec contains specification parameters for a Rule.
type PrometheusRuleSpec struct {
	// Content of Prometheus rule file
	Groups []RuleGroup `json:"groups,omitempty"`
}

// RuleGroup is a list of sequentially evaluated recording and alerting rules.
type RuleGroup struct {
	Name     string `json:"name" `
	Interval string `json:"interval,omitempty"`
	Rules    []Rule `json:"rules"`
}

// Rule describes an alerting or recording rule.
type Rule struct {
	Record      string             `json:"record,omitempty"`
	Alert       string             `json:"alert,omitempty"`
	Expr        intstr.IntOrString `json:"expr"`
	For         string             `json:"for,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	Annotations map[string]string  `json:"annotations,omitempty"`
}

// PrometheusRuleStatus defines the observed state of PrometheusRule
type PrometheusRuleStatus struct {
	SyncStatus string `json:"sync_status,omitempty"`
}

// +kubebuilder:object:root=true

// PrometheusRule is the Schema for the prometheusrules API
type PrometheusRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrometheusRuleSpec   `json:"spec,omitempty"`
	Status PrometheusRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PrometheusRuleList contains a list of PrometheusRule
type PrometheusRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrometheusRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PrometheusRule{}, &PrometheusRuleList{})
}
