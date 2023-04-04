/*
Copyright 2023 The Kubernetes Authors.

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

package rayjob

import (
	"context"
	"strings"

	rayjobapi "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta1"
	"sigs.k8s.io/kueue/pkg/controller/jobframework"
)

var (
	gvk = schema.GroupVersionKind{
		Group:   rayjobapi.GroupVersion.Group,
		Version: rayjobapi.GroupVersion.Version,
		Kind:    "RayJob",
	}
)

// RayJobReconciler reconciles a Job object
type RayJobReconciler jobframework.JobReconciler

func NewReconciler(
	scheme *runtime.Scheme,
	client client.Client,
	record record.EventRecorder,
	opts ...jobframework.Option) *RayJobReconciler {
	return (*RayJobReconciler)(jobframework.NewReconciler(scheme,
		client,
		record,
		opts...,
	))
}

type RayJob rayjobapi.RayJob

func (j *RayJob) Object() client.Object {
	return (*rayjobapi.RayJob)(j)
}

func (j *RayJob) IsSuspended() bool {
	return j.Spec.Suspend
}

func (j *RayJob) IsActive() bool {
	return j.Status.JobDeploymentStatus != rayjobapi.JobDeploymentStatusSuspended
}

func (j *RayJob) Suspend() {
	j.Spec.Suspend = true
}

func (j *RayJob) ResetStatus() bool {
	if j.Status.StartTime == nil {
		return false
	}
	j.Status.StartTime = nil
	return true
}

func (j *RayJob) GetGVK() schema.GroupVersionKind {
	return gvk
}

func (j *RayJob) PodSets() []kueue.PodSet {
	// len = workerGroups + head
	podSets := make([]kueue.PodSet, len(j.Spec.RayClusterSpec.WorkerGroupSpecs)+1)

	// head
	podSets[0] = kueue.PodSet{
		Name:     "head",
		Template: *j.Spec.RayClusterSpec.HeadGroupSpec.Template.DeepCopy(),
		Count:    1,
	}

	// workers
	for index := range j.Spec.RayClusterSpec.WorkerGroupSpecs {
		wgs := &j.Spec.RayClusterSpec.WorkerGroupSpecs[index]
		replicas := int32(1)
		if wgs.Replicas != nil {
			replicas = *wgs.Replicas
		}
		podSets[index+1] = kueue.PodSet{
			Name:     strings.ToLower(wgs.GroupName),
			Template: *wgs.Template.DeepCopy(),
			Count:    replicas,
		}
	}
	return podSets
}

func applySelectors(dst, src map[string]string) map[string]string {
	if len(dst) == 0 {
		return src
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (j *RayJob) RunWithNodeAffinity(nodeSelectors []map[string]string) {
	j.Spec.Suspend = false
	if len(nodeSelectors) == 0 {
		return
	}

	// head
	headPodSpec := &j.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
	headPodSpec.NodeSelector = applySelectors(headPodSpec.NodeSelector, nodeSelectors[0])

	// workers
	for index := range j.Spec.RayClusterSpec.WorkerGroupSpecs {
		workerPodSpec := &j.Spec.RayClusterSpec.WorkerGroupSpecs[index].Template.Spec
		workerPodSpec.NodeSelector = applySelectors(workerPodSpec.NodeSelector, nodeSelectors[index+1])
	}
}

func cloneSelectors(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (j *RayJob) RestoreNodeAffinity(podSets []kueue.PodSet) {
	// head
	headPodSpec := &j.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
	if !equality.Semantic.DeepEqual(headPodSpec.NodeSelector, podSets[0].Template.Spec.NodeSelector) {
		headPodSpec.NodeSelector = cloneSelectors(podSets[0].Template.Spec.NodeSelector)
	}

	// workers
	for index := range j.Spec.RayClusterSpec.WorkerGroupSpecs {
		workerPodSpec := &j.Spec.RayClusterSpec.WorkerGroupSpecs[index].Template.Spec
		workerPodSpec.NodeSelector = cloneSelectors(podSets[index+1].Template.Spec.NodeSelector)
	}
}

func (j *RayJob) Finished() (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:    kueue.WorkloadFinished,
		Status:  metav1.ConditionTrue,
		Reason:  string(j.Status.JobStatus),
		Message: j.Status.Message,
	}
	finished := false
	switch j.Status.JobDeploymentStatus {
	case rayjobapi.JobDeploymentStatusComplete, rayjobapi.JobDeploymentStatusFailedJobDeploy:
		finished = true

	default:
		finished = false
	}
	return condition, finished
}

func (j *RayJob) EquivalentToWorkload(wl kueue.Workload) bool {
	if len(wl.Spec.PodSets) != (len(j.Spec.RayClusterSpec.WorkerGroupSpecs) + 1) {
		return false
	}

	podSets := wl.Spec.PodSets
	headPodSpec := &j.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
	if podSets[0].Count != 1 {
		return false
	}
	if !equality.Semantic.DeepEqual(headPodSpec.InitContainers, podSets[0].Template.Spec.InitContainers) {
		return false
	}
	if !equality.Semantic.DeepEqual(headPodSpec.Containers, podSets[0].Template.Spec.Containers) {
		return false
	}

	// workers
	for index := range j.Spec.RayClusterSpec.WorkerGroupSpecs {
		if podSets[index+1].Count != *j.Spec.RayClusterSpec.WorkerGroupSpecs[index].Replicas {
			return false
		}

		workerPodSpec := &j.Spec.RayClusterSpec.WorkerGroupSpecs[index].Template.Spec
		if !equality.Semantic.DeepEqual(workerPodSpec.InitContainers, podSets[index+1].Template.Spec.InitContainers) {
			return false
		}
		if !equality.Semantic.DeepEqual(workerPodSpec.Containers, podSets[index+1].Template.Spec.Containers) {
			return false
		}
	}
	return true
}

func (j *RayJob) PriorityClass() string {
	if j.Spec.RayClusterSpec != nil {
		rcs := j.Spec.RayClusterSpec
		if len(rcs.HeadGroupSpec.Template.Spec.PriorityClassName) != 0 {
			return rcs.HeadGroupSpec.Template.Spec.PriorityClassName
		}
		for wi := range rcs.WorkerGroupSpecs {
			w := &rcs.WorkerGroupSpecs[wi]
			if len(w.Template.Spec.PriorityClassName) != 0 {
				return w.Template.Spec.PriorityClassName
			}
		}
	}
	return ""
}

func (j *RayJob) PodsReady() bool {
	return j.Status.RayClusterStatus.State == rayjobapi.Ready
}

// SetupWithManager sets up the controller with the Manager. It indexes workloads
// based on the owning jobs.
func (r *RayJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if !r.EnableRay {
		mgr.GetLogger().Info("RayJob integration is not enabled")
		return nil
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&rayjobapi.RayJob{}).
		Owns(&kueue.Workload{}).
		Complete(r)
}

func SetupIndexes(ctx context.Context, indexer client.FieldIndexer) error {
	return jobframework.SetupWorkloadOwnerIndex(ctx, indexer, gvk)
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;watch;update
//+kubebuilder:rbac:groups=ray.io,resources=rayjobs,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=ray.io,resources=rayjobs/status,verbs=get
//+kubebuilder:rbac:groups=kueue.x-k8s.io,resources=workloads,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kueue.x-k8s.io,resources=workloads/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kueue.x-k8s.io,resources=workloads/finalizers,verbs=update
//+kubebuilder:rbac:groups=kueue.x-k8s.io,resources=resourceflavors,verbs=get;list;watch

func (r *RayJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fjr := (*jobframework.JobReconciler)(r)
	return fjr.ReconcileGenericJob(ctx, req, &RayJob{})
}

func GetWorkloadNameForRayJob(jobName string) string {
	return jobframework.GetWorkloadNameForOwnerWithGVK(jobName, gvk)
}
