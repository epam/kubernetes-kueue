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

package util

import (
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/kueue/apis/kueue/v1beta1"
)

func WorkloadNameCompletionFunc(clientGetter ClientGetter, status bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		clientSet, err := clientGetter.KueueClientSet()
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveError
		}

		namespace, _, err := clientGetter.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveError
		}

		list, err := clientSet.KueueV1beta1().Workloads(namespace).List(cmd.Context(), v1.ListOptions{})
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveError
		}

		filteredItems := make([]v1beta1.Workload, 0, len(list.Items))
		for _, wl := range list.Items {
			if ptr.Deref(wl.Spec.Active, true) == status {
				filteredItems = append(filteredItems, wl)
			}
		}

		validArgs := make([]string, len(filteredItems))
		for i, wl := range filteredItems {
			validArgs[i] = wl.Name
		}

		return validArgs, cobra.ShellCompDirectiveNoFileComp
	}
}
