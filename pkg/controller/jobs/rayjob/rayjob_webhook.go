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

	rayjobapi "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"sigs.k8s.io/kueue/pkg/controller/jobframework"
	"sigs.k8s.io/kueue/pkg/util/pointer"
)

type RayJobWebhook struct {
	manageJobsWithoutQueueName bool
	enableRay                  bool
}

// SetupWebhook configures the webhook for rayjobapi RayJob.
func SetupRayJobWebhook(mgr ctrl.Manager, opts ...jobframework.Option) error {
	options := jobframework.DefaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	wh := &RayJobWebhook{
		manageJobsWithoutQueueName: options.ManageJobsWithoutQueueName,
		enableRay:                  options.EnableRay,
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(&rayjobapi.RayJob{}).
		WithDefaulter(wh).
		WithValidator(wh).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-ray-io-v1alpha1-rayjob,mutating=true,failurePolicy=fail,sideEffects=None,groups=ray.io,resources=rayjobs,verbs=create,versions=v1alpha1,name=mrayjob.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &RayJobWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *RayJobWebhook) Default(ctx context.Context, obj runtime.Object) error {
	job := obj.(*rayjobapi.RayJob)
	log := ctrl.LoggerFrom(ctx).WithName("job-webhook")
	if !w.enableRay {
		log.V(5).Info("RayJob integration id not enabled", "job", klog.KObj(job))
		return nil
	}
	log.V(5).Info("Applying defaults", "job", klog.KObj(job))

	jobframework.ApplyDefaultForSuspend((*RayJob)(job), w.manageJobsWithoutQueueName)
	return nil
}

// +kubebuilder:webhook:path=/validate-ray-io-v1alpha1-rayjob,mutating=false,failurePolicy=fail,sideEffects=None,groups=ray.io,resources=rayjobs,verbs=update,versions=v1alpha1,name=vrayjob.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &RayJobWebhook{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (w *RayJobWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	job := obj.(*rayjobapi.RayJob)
	log := ctrl.LoggerFrom(ctx).WithName("job-webhook")
	if !w.enableRay {
		log.V(5).Info("RayJob integration id not enabled", "job", klog.KObj(job))
		return nil
	}
	log.Info("Validating create", "job", klog.KObj(job))
	return w.validateCreate(job).ToAggregate()
}

func (w *RayJobWebhook) validateCreate(job *rayjobapi.RayJob) field.ErrorList {
	var allErrors field.ErrorList
	kueueJob := (*RayJob)(job)

	if w.manageJobsWithoutQueueName || jobframework.QueueName(kueueJob) != "" {
		spec := &job.Spec
		specPath := field.NewPath("spec")

		if !spec.ShutdownAfterJobFinishes {
			allErrors = append(allErrors, field.Invalid(specPath.Child("shutdownAfterJobFinishes"), spec.ShutdownAfterJobFinishes, "a kueue managed job should delete the cluster after finishing"))
		}
		// should not want existing cluster
		if len(spec.ClusterSelector) > 0 {
			allErrors = append(allErrors, field.Invalid(specPath.Child("clusterSelector"), spec.ClusterSelector, "a kueue managed job should not use an existing cluster"))
		}

		clusterSpec := spec.RayClusterSpec
		clusterSpecPath := specPath.Child("rayClusterSpec")

		// should not use auto scaler
		if pointer.BoolDeref(clusterSpec.EnableInTreeAutoscaling, false) {
			allErrors = append(allErrors, field.Invalid(clusterSpecPath.Child("enableInTreeAutoscaling"), clusterSpec.EnableInTreeAutoscaling, "a kueue managed job should not use autoscaling"))
		}
		// should limit the worker count to 8 - 1 (max podSets num - cluster head)
		if len(clusterSpec.WorkerGroupSpecs) > 7 {
			allErrors = append(allErrors, field.Invalid(clusterSpecPath.Child("workerGroupSpecs"), clusterSpec.EnableInTreeAutoscaling, "a kueue managed job should define at most 7 worker groups"))
		}
	}

	allErrors = append(allErrors, jobframework.ValidateAnnotationAsCRDName(kueueJob, jobframework.QueueLabel)...)
	return allErrors
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (w *RayJobWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	oldJob := oldObj.(*rayjobapi.RayJob)
	newJob := newObj.(*rayjobapi.RayJob)
	log := ctrl.LoggerFrom(ctx).WithName("job-webhook")
	if !w.enableRay {
		log.V(5).Info("RayJob integration id not enabled", "newJob", klog.KObj(newJob))
		return nil
	}
	log.Info("Validating update", "job", klog.KObj(newJob))
	allErrors := jobframework.ValidateUpdateForQueueName((*RayJob)(oldJob), (*RayJob)(newJob))
	allErrors = append(allErrors, w.validateCreate(newJob)...)
	return allErrors.ToAggregate()
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (w *RayJobWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}
