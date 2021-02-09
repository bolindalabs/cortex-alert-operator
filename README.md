# Cortex Alert Operator

This K8s Operator consumes _PrometheusRules_ as known from using Prometheus with Prometheus Operator.
It then applies these Rules against a Cortex environment.


### Project state

This project is currently in proof of concept state.
It works for our internal test use-case.

We plan on improving configuration of the operator itself.
I.e. currently it is not possible to change the naming behavior of the operator.
While PrometheusRules in Kubernetes are named in a `namespace` + `name` + `groupname`-scheme,
Cortex only supports `namespace` + `groupname`.
With the current behavior, we map a Cortex namespace to a Kubernetes `{namespace}--{name}`.

We plan to allow custom namespace prefixes i.e. for use with different Kubernetes clusters
and to further investigate into supporting other naming schemes.

### Example
```yaml
# original from: https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/user-guides/alerting.md
apiVersion: monitoring.bolinda.digital/v1
kind: PrometheusRule
metadata:
  labels:
    prometheus: example
    role: alert-rules
  name: prometheus-example-rules
spec:
  groups:
  - name: ./example.rules
    rules:
    - alert: ExampleAlert
      expr: vector(1)
```
