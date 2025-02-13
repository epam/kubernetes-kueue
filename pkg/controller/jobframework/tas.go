/*
Copyright 2024 The Kubernetes Authors.

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

package jobframework

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kueuealpha "sigs.k8s.io/kueue/apis/kueue/v1alpha1"
	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta1"
)

type PodSetTopologyRequestOption func(*kueue.PodSetTopologyRequest)

func WithPodIndexLabel(podIndexLabel string) PodSetTopologyRequestOption {
	return func(psTopologyReq *kueue.PodSetTopologyRequest) {
		psTopologyReq.PodIndexLabel = &podIndexLabel
	}
}

func WithSubGroupIndexLabel(subGroupIndexLabel string) PodSetTopologyRequestOption {
	return func(psTopologyReq *kueue.PodSetTopologyRequest) {
		psTopologyReq.SubGroupIndexLabel = &subGroupIndexLabel
	}
}

func WithSubGroupCount(subGroupCount int32) PodSetTopologyRequestOption {
	return func(psTopologyReq *kueue.PodSetTopologyRequest) {
		psTopologyReq.SubGroupCount = &subGroupCount
	}
}

func WithRingName(ringName string) PodSetTopologyRequestOption {
	return func(psTopologyReq *kueue.PodSetTopologyRequest) {
		psTopologyReq.RingName = &ringName
	}
}

func PodSetTopologyRequest(meta *metav1.ObjectMeta, options ...PodSetTopologyRequestOption) *kueue.PodSetTopologyRequest {
	requiredValue, requiredFound := meta.Annotations[kueuealpha.PodSetRequiredTopologyAnnotation]
	preferredValue, preferredFound := meta.Annotations[kueuealpha.PodSetPreferredTopologyAnnotation]

	if requiredFound || preferredFound {
		psTopologyReq := &kueue.PodSetTopologyRequest{}
		for _, option := range options {
			option(psTopologyReq)
		}
		if requiredFound {
			psTopologyReq.Required = &requiredValue
		} else {
			psTopologyReq.Preferred = &preferredValue
		}
		return psTopologyReq
	}
	return nil
}
