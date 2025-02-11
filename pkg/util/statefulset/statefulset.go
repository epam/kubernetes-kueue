/*
Copyright 2025 The Kubernetes Authors.

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

package statefulset

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	podcontroller "sigs.k8s.io/kueue/pkg/controller/jobs/pod"
	utilpod "sigs.k8s.io/kueue/pkg/util/pod"
)

func GetStatefulSet(ctx context.Context, c client.Client, key types.NamespacedName) (*appsv1.StatefulSet, error) {
	if key.Name == "" || key.Namespace == "" {
		return nil, nil
	}
	sts := &appsv1.StatefulSet{}
	if err := c.Get(ctx, key, sts); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return sts, nil
}

func UngateAndFinalizePod(sts *appsv1.StatefulSet, pod *corev1.Pod) bool {
	var updated bool

	if shouldUngatePod(sts, pod) && utilpod.Ungate(pod, podcontroller.SchedulingGateName) {
		updated = true
	}

	if shouldFinalizePod(sts, pod) && controllerutil.RemoveFinalizer(pod, podcontroller.PodFinalizer) {
		updated = true
	}

	return updated
}

func shouldUngatePod(sts *appsv1.StatefulSet, pod *corev1.Pod) bool {
	return sts == nil || sts.Status.CurrentRevision != sts.Status.UpdateRevision &&
		sts.Status.CurrentRevision == pod.Labels[appsv1.ControllerRevisionHashLabelKey]
}

func shouldFinalizePod(sts *appsv1.StatefulSet, pod *corev1.Pod) bool {
	return shouldUngatePod(sts, pod) || utilpod.IsTerminated(pod)
}
