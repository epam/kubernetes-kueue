/*
Copyright The Kubernetes Authors.

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

package singlecluster

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/kueue/pkg/constants"
	"sigs.k8s.io/kueue/pkg/util/testingjobs/pod"
	"sigs.k8s.io/kueue/test/util"
)

var _ = ginkgo.Describe("Apply", func() {
	var (
		ns *corev1.Namespace
	)

	ginkgo.BeforeEach(func() {
		ns = util.CreateNamespaceFromPrefixWithLog(ctx, k8sClient, "pod-e2e-")
	})
	ginkgo.AfterEach(func() {
		gomega.Expect(util.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
	})

	ginkgo.It("should return a NotFound error on .Status().Patch()", func() {
		p := pod.MakePod("pod", ns.Name).Obj()
		//nolint:staticcheck //SA1019: client.Apply is deprecated
		err := k8sClient.Status().Patch(ctx, p, client.Apply, client.FieldOwner(constants.KueueName))
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(errors.IsNotFound(err)).To(gomega.BeTrue())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(p), &corev1.Pod{})
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(errors.IsNotFound(err)).To(gomega.BeTrue())
	})

	ginkgo.It("should return a NotFound error on .Status().Apply()", func() {
		p := corev1ac.Pod("pod", ns.Name)
		err := k8sClient.Status().Apply(ctx, p, client.FieldOwner(constants.KueueName), client.ForceOwnership)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(errors.IsNotFound(err)).To(gomega.BeTrue())

		err = k8sClient.Get(ctx, client.ObjectKey{Name: "pod", Namespace: ns.Name}, &corev1.Pod{})
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(errors.IsNotFound(err)).To(gomega.BeTrue())
	})
})
