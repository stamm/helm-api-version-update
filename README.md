# Update internal struct for helm to update apiVersion

In kubernetes v1.16 some apiVersions (https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.16.md#deprecations-and-removals)[has been removed].

For example, Deployments and Daemonsets now must be created with `apps/v1`, not `extensions/v1beta1`.

If you just update apiVersion in your template, helm see the version you applied before and try to get this apiVersion from kubernetes.

So, this tool try to update apiVersion inside internal structs for helm.

For helm2 - it's grpc inside gzip inside base64 inside config map.
For helm3 - it's gzip inside base64 inside secret.


go run ./cmd -dry-run -helm2 -ns=test-ns


flags:
```
  -dry-run
        (optional) don't update
  -filter string
        (optional) filter for releases
  -helm2
        run only for helm3 releases
  -helm3
        run only for helm2 releases
  -kubeconfig string
        (optional) absolute path to the kubeconfig file (default "$HOME/.kube/config")
  -ns string
        (optional) namespaces separated by comma
  -only-find
        (optional) only display config maps
```
