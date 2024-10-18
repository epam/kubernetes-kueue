# Init

For the purpose of this guide [kubebuilder](https://book.kubebuilder.io/) is used to bootstrap the Kubernetes operator that will run the Demo Admission Check Controller.

Check the kubebuilder [quick-start guide](https://book.kubebuilder.io/quick-start) for details.

Choose a domain and repo for the project, for the purpose of this demo the project's domain is `demo-acc.experimental.kueue.x-k8s.io` and the repo `sigs.k8s.io/kueue/cmd/experimental/demo-acc`.

## Init the project

```bash
kubebuilder init --domain demo-acc.experimental.kueue.x-k8s.io --repo sigs.k8s.io/kueue/cmd/experimental/demo-acc
```

## Scaffold the controllers

Note: Kubebuilder should have better support for scaffolding controllers for external API resources in the near future, follow [this pr](https://github.com/kubernetes-sigs/kubebuilder/pull/4171) for details.

```bash
kubebuilder create api --group replace-me --version v1beta1 --kind AdmissionCheck --controller=true  --resource=false
kubebuilder create api --group replace-me --version v1beta1 --kind Workload --controller=true  --resource=false
```

Replace the dummy group with kueue's group.

```bash
sed -i 's/groups=replace-me.\+,resources/groups=kueue.x-k8s.io,resources/g' internal/controller/*.go
yq -i '(.resources[]|select(.group == "replace-me")) |= .domain="x-k8s.io" '  PROJECT
yq -i '(.resources[]|select(.group == "replace-me")) |= .group="kueue" '  PROJECT
```

Regenerate the code

```bash
make genetrate manifests
```

Note: The generated `internal/controller/suite_test.go` is missing the "context" import it should be added.

## Implement the AdmissionCheck reconciler

Go get `kueue`

```bash
go get sigs.k8s.io/kueue@main
```

In `internal/controller/admissioncheck_controller.go`:

Import the kueue api and some helper packages:

```go
import (
	"context"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	kueueapi "sigs.k8s.io/kueue/apis/kueue/v1beta1"
)
```

Define the controller's name

```go
const (
	DemoACCName = "experimental.kueue.x-k8s.io/demo-acc"
)

```

Update the reconciler's logic to set active all ACs managed by this controller.

```go
func (r *AdmissionCheckReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var ac kueueapi.AdmissionCheck
	err := r.Get(ctx, req.NamespacedName, &ac)
	if err != nil {
		// Ignore not found
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Ignore if not managed by demo-acc
	if ac.Spec.ControllerName != DemoACCName {
		return ctrl.Result{}, nil
	}

	// For now, we only want to set it active if not already done
	log.V(2).Info("Reconcile AdmissionCheck")
	if !apimeta.IsStatusConditionTrue(ac.Status.Conditions, kueueapi.AdmissionCheckActive) {
		apimeta.SetStatusCondition(&ac.Status.Conditions, metav1.Condition{
			Type:               kueueapi.AdmissionCheckActive,
			Status:             metav1.ConditionTrue,
			Reason:             "Active",
			Message:            "demo-acc is running",
			ObservedGeneration: ac.Generation,
		})
		log.V(2).Info("Update Active condition")
		return ctrl.Result{}, r.Status().Update(ctx, &ac)
	}
	return ctrl.Result{}, nil
}

```

Finally update `SetupWithManager` to be registered `For` Kueue's AdmissionChecks.

```go
func (r *AdmissionCheckReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kueueapi.AdmissionCheck{}).
		Complete(r)
}
```

## Implement the Workload reconciler

In `internal/controller/workload_controller.go`:

Import the kueue api and some helper packages:
```go
import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	kueueapi "sigs.k8s.io/kueue/apis/kueue/v1beta1"
	"sigs.k8s.io/kueue/pkg/util/admissioncheck"
	"sigs.k8s.io/kueue/pkg/workload"
)

```

As a demo logic the ACC will prevent workloads from being `Admitted` for one minute after their creation.

Define some globals:

```go
const (
	MakeReadyAfter = time.Minute
	ReadyMessage   = "The workload is now ready"
)
```

Update the reconciler's logic:

```go
func (r *WorkloadReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var wl kueueapi.Workload
	err := r.Get(ctx, req.NamespacedName, &wl)
	if err != nil {
		// Ignore not found
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Ignore the workloads that don't have quota reservation or are finished
	if !workload.HasQuotaReservation(&wl) || workload.IsFinished(&wl) {
		return ctrl.Result{}, nil
	}

	// Get the states managed by demo-acc
	managedStatesNames, err := admissioncheck.FilterForController(ctx, r.Client, wl.Status.AdmissionChecks, DemoACCName)

	// Ignore if none are managed by demo-acc
	if len(managedStatesNames) == 0 {
		return ctrl.Result{}, nil
	}

	log.V(2).Info("Reconcile Workload")

	//If we need to wait
	if remaining := time.Until(wl.CreationTimestamp.Add(MakeReadyAfter)); remaining > 0 {
		return ctrl.Result{RequeueAfter: remaining}, nil
	}

	// Mark the states 'Ready' if not done already
	needsUpdate := false
	wlPatch := workload.BaseSSAWorkload(&wl)
	for _, name := range managedStatesNames {
		if acs := workload.FindAdmissionCheck(wl.Status.AdmissionChecks, name); acs.State != kueueapi.CheckStateReady {
			workload.SetAdmissionCheckState(&wlPatch.Status.AdmissionChecks, kueueapi.AdmissionCheckState{
				Name:    name,
				State:   kueueapi.CheckStateReady,
				Message: ReadyMessage,
			})
			needsUpdate = true
		} else {
			workload.SetAdmissionCheckState(&wlPatch.Status.AdmissionChecks, *acs)
		}
	}
	if needsUpdate {
		return ctrl.Result{}, r.Status().Patch(ctx, wlPatch, client.Apply, client.FieldOwner(DemoACCName), client.ForceOwnership)
	}

	return ctrl.Result{}, nil
}

```

Finally update `SetupWithManager` to be registered `For` Kueue's Workloads.

```go
func (r *WorkloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kueueapi.Workload{}).
		Complete(r)
}
```

## Register the Kueue's api in the controller's scheme

In `cmd/main.go`

```diff
@@ -34,6 +34,7 @@ import (
 	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
 	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
 	"sigs.k8s.io/controller-runtime/pkg/webhook"
+	kueueapi "sigs.k8s.io/kueue/apis/kueue/v1beta1"

 	"sigs.k8s.io/kueue/cmd/experimental/demo-acc/internal/controller"
 	// +kubebuilder:scaffold:imports
@@ -46,6 +47,7 @@ var (

 func init() {
 	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
+	utilruntime.Must(kueueapi.AddToScheme(scheme))

 	// +kubebuilder:scaffold:scheme
 }
```

## Test the ACC

### Install Kueue

Install kueue as described in its [installation guide](https://kueue.sigs.k8s.io/docs/installation/).

### Build the image

Note: Since at this point the operator is not defying an API, the `COPY api/ api/` line should be removed to be able to build the image.
Note: Since at this point kueue is using go v1.23 update the builder's base image in the Dockerfile.

```bash
make docker-build
```

Make sure the test cluster has access to the built image (it can be pushed in a image repository or directly in a test cluster).


Skip `make install` since we have no API defined.

```bash
make deploy
```

Make sure the controller's manager is properly running

```bash
kubectl get -n demo-acc-system deployments.apps
```

Set up a single cluster queue environment as described [here](https://kueue.sigs.k8s.io/docs/tasks/manage/administer_cluster_quotas/#single-clusterqueue-and-single-resourceflavor-setup).
This can be done with:
```bash
kubectl apply -f https://kueue.sigs.k8s.io/examples/admin/single-clusterqueue-setup.yaml
```

Create an AdmissionCheck managed by demo-acc

```bash
kubectl apply -f - <<EOF
apiVersion: kueue.x-k8s.io/v1beta1
kind: AdmissionCheck
metadata:
  name: demo-ac
spec:
  controllerName: experimental.kueue.x-k8s.io/demo-acc
EOF
```

It should get marked as `Active`

```bash
kubectl get admissionchecks.kueue.x-k8s.io demo-ac -o=jsonpath='{.status.conditions[?(@.type=="Active")].status}{" -> "}{.status.conditions[?(@.type=="Active")].message}{"\n"}'
```

Should output: `True -> demo-acc is running`

Add the admission check to the `cluster-queue` queue:

```bash
kubectl patch clusterqueues.kueue.x-k8s.io cluster-queue --type='json' -p='[{"op": "add", "path": "/spec/admissionChecks", "value":["demo-ac"]}]'
```

Create a sample-job:

```bash
kubectl create -f https://kueue.sigs.k8s.io/examples/jobs/sample-job.yaml
```

Observe the workload's execution.

Given the job is created at t0
- Since the queue is not used, it will get its quota reserved at t0
- demo-acc will marks its admission check state at t0+1m
- it will be admitted at t0+1m

For example:

```yaml
apiVersion: kueue.x-k8s.io/v1beta1
kind: Workload
metadata:
  creationTimestamp: "2024-10-18T12:37:15Z" #t0
  #<skip>
spec:
  active: true
  #<skip>
status:
  admission:
    clusterQueue: cluster-queue
    #<skip>
  admissionChecks:
  - lastTransitionTime: "2024-10-18T12:38:15Z" #t0+1m
    message: The workload is now ready
    name: demo-ac
    state: Ready
  conditions:
  - lastTransitionTime: "2024-10-18T12:37:15Z" #t0
    message: Quota reserved in ClusterQueue cluster-queue
    observedGeneration: 1
    reason: QuotaReserved
    status: "True"
    type: QuotaReserved
  - lastTransitionTime: "2024-10-18T12:38:15Z" #t0+1m
    message: The workload is admitted
    observedGeneration: 1
    reason: Admitted
    status: "True"
    type: Admitted
  - #<skip>
```

