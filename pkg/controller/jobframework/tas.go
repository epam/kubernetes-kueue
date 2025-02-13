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

type podSetTopologyRequestOptions struct {
	podIndexLabel      *string
	subGroupIndexLabel *string
	subGroupCount      *int32
	ringName           *string
}

type PodSetTopologyRequestOption func(*podSetTopologyRequestOptions)

func WithPodIndexLabel(podIndexLabel string) PodSetTopologyRequestOption {
	return func(opts *podSetTopologyRequestOptions) {
		opts.podIndexLabel = &podIndexLabel
	}
}

func WithSubGroupIndexLabel(subGroupIndexLabel string) PodSetTopologyRequestOption {
	return func(opts *podSetTopologyRequestOptions) {
		opts.subGroupIndexLabel = &subGroupIndexLabel
	}
}

func WithSubGroupCount(subGroupCount int32) PodSetTopologyRequestOption {
	return func(opts *podSetTopologyRequestOptions) {
		opts.subGroupCount = &subGroupCount
	}
}

func WithRingName(ringName string) PodSetTopologyRequestOption {
	return func(psTopologyReq *podSetTopologyRequestOptions) {
		psTopologyReq.ringName = &ringName
	}
}

func PodSetTopologyRequest(meta *metav1.ObjectMeta, opts ...PodSetTopologyRequestOption) *kueue.PodSetTopologyRequest {
	requiredValue, requiredFound := meta.Annotations[kueuealpha.PodSetRequiredTopologyAnnotation]
	preferredValue, preferredFound := meta.Annotations[kueuealpha.PodSetPreferredTopologyAnnotation]

	if requiredFound || preferredFound {
		options := podSetTopologyRequestOptions{}
		for _, opt := range opts {
			opt(&options)
		}

		psTopologyReq := &kueue.PodSetTopologyRequest{
			PodIndexLabel:      options.podIndexLabel,
			SubGroupIndexLabel: options.subGroupIndexLabel,
			SubGroupCount:      options.subGroupCount,
			RingName:           options.ringName,
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
