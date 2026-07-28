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

package spreading

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	"sigs.k8s.io/kueue/pkg/workload"
)

func TestParseAndValidate(t *testing.T) {
	tests := map[string]struct {
		annotation  string
		wantErr     bool
		wantErrText string
		wantCfg     *Config
	}{
		"valid required rule": {
			annotation: `{"workload-label-selector":"app=svc","rules":[{"key":"topology.kubernetes.io/zone","max-domain-percentage":"45"}]}`,
			wantCfg: &Config{
				WorkloadLabelSelector: "app=svc",
				Rules: []Rule{
					{Key: "topology.kubernetes.io/zone", MaxDomainPercentage: "45"},
				},
			},
		},
		"valid preferred rule": {
			annotation: `{"workload-label-selector":"app=svc","rules":[{"key":"topology.kubernetes.io/zone","max-domain-percentage":"45","type":"Preferred"}]}`,
			wantCfg: &Config{
				WorkloadLabelSelector: "app=svc",
				Rules: []Rule{
					{Key: "topology.kubernetes.io/zone", MaxDomainPercentage: "45", Type: "Preferred"},
				},
			},
		},
		"two rules": {
			annotation: `{"workload-label-selector":"app=svc","rules":[{"key":"zone","max-domain-percentage":"45"},{"key":"rack","max-domain-percentage":"22","type":"Preferred"}]}`,
			wantCfg: &Config{
				WorkloadLabelSelector: "app=svc",
				Rules: []Rule{
					{Key: "zone", MaxDomainPercentage: "45"},
					{Key: "rack", MaxDomainPercentage: "22", Type: "Preferred"},
				},
			},
		},
		"invalid JSON": {
			annotation:  `not-json`,
			wantErr:     true,
			wantErrText: "invalid JSON",
		},
		"missing workload-label-selector": {
			annotation:  `{"rules":[{"key":"zone","max-domain-percentage":"45"}]}`,
			wantErr:     true,
			wantErrText: "workload-label-selector is required",
		},
		"invalid label selector": {
			annotation:  `{"workload-label-selector":"invalid[","rules":[{"key":"zone","max-domain-percentage":"45"}]}`,
			wantErr:     true,
			wantErrText: "invalid workload-label-selector",
		},
		"empty rules": {
			annotation:  `{"workload-label-selector":"app=svc","rules":[]}`,
			wantErr:     true,
			wantErrText: "rules must not be empty",
		},
		"too many rules": {
			annotation:  `{"workload-label-selector":"app=svc","rules":[{"key":"z1","max-domain-percentage":"45"},{"key":"z2","max-domain-percentage":"45"},{"key":"z3","max-domain-percentage":"45"}]}`,
			wantErr:     true,
			wantErrText: "at most 2 entries",
		},
		"missing key": {
			annotation:  `{"workload-label-selector":"app=svc","rules":[{"max-domain-percentage":"45"}]}`,
			wantErr:     true,
			wantErrText: "key is required",
		},
		"missing max-domain-percentage": {
			annotation:  `{"workload-label-selector":"app=svc","rules":[{"key":"zone"}]}`,
			wantErr:     true,
			wantErrText: "max-domain-percentage is required",
		},
		"max-domain-percentage out of range (0)": {
			annotation:  `{"workload-label-selector":"app=svc","rules":[{"key":"zone","max-domain-percentage":"0"}]}`,
			wantErr:     true,
			wantErrText: "must be an integer in [1,99]",
		},
		"max-domain-percentage out of range (100)": {
			annotation:  `{"workload-label-selector":"app=svc","rules":[{"key":"zone","max-domain-percentage":"100"}]}`,
			wantErr:     true,
			wantErrText: "must be an integer in [1,99]",
		},
		"invalid type": {
			annotation:  `{"workload-label-selector":"app=svc","rules":[{"key":"zone","max-domain-percentage":"45","type":"Invalid"}]}`,
			wantErr:     true,
			wantErrText: `must be "Required" or "Preferred"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg, err := ParseAndValidate(tc.annotation)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrText)
				}
				if tc.wantErrText != "" {
					if !contains(err.Error(), tc.wantErrText) {
						t.Errorf("error %q does not contain %q", err.Error(), tc.wantErrText)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantCfg, cfg); diff != "" {
				t.Errorf("Config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestIsBanned(t *testing.T) {
	tests := map[string]struct {
		count, total, maxPct int
		wantBanned           bool
	}{
		// Cold start: no workloads yet → nothing is banned.
		"cold start: first workload, any domain": {count: 0, total: 0, maxPct: 45, wantBanned: false},
		// 45% cap: 1/6 = 16.7% < 45% → not banned.
		"well below cap": {count: 1, total: 6, maxPct: 45, wantBanned: false},
		// 45% cap: 2/6 = 33.3% < 45% → not banned.
		"below cap by one": {count: 2, total: 6, maxPct: 45, wantBanned: false},
		// 45% cap: 3/6 = 50% >= 45% → banned.
		"at or over cap": {count: 3, total: 6, maxPct: 45, wantBanned: true},
		// 50% cap: 3/6 = 50% >= 50% (boundary) → banned.
		"exactly at 50% cap": {count: 3, total: 6, maxPct: 50, wantBanned: true},
		// 50% cap: 2/6 = 33.3% < 50% → not banned.
		"below 50% cap": {count: 2, total: 6, maxPct: 50, wantBanned: false},
		// KEP walk-through table (3 domains, max=34%):
		"kep table row 1 zone A (cold start)":      {count: 0, total: 0, maxPct: 34, wantBanned: false},
		"kep table row 2 zone A (1/1=100% ≥ 34%)":  {count: 1, total: 1, maxPct: 34, wantBanned: true},
		"kep table row 2 zone B (0/1=0% < 34%)":    {count: 0, total: 1, maxPct: 34, wantBanned: false},
		"kep table row 4 zone A (1/3=33.3% < 34%)": {count: 1, total: 3, maxPct: 34, wantBanned: false},
		"kep table row 5 zone A (2/4=50% >= 34%)":  {count: 2, total: 4, maxPct: 34, wantBanned: true},
		"kep table row 5 zone B (1/4=25% < 34%)":   {count: 1, total: 4, maxPct: 34, wantBanned: false},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsBanned(tc.count, tc.total, tc.maxPct)
			if got != tc.wantBanned {
				t.Errorf("IsBanned(%d,%d,%d) = %v, want %v", tc.count, tc.total, tc.maxPct, got, tc.wantBanned)
			}
		})
	}
}

// buildWL creates a minimal workload.Info with topology assignment for testing.
func buildWL(namespace string, lbls map[string]string, levels []string, domainValues [][]string) *workload.Info {
	wl := &kueue.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wl",
			Namespace: namespace,
			Labels:    lbls,
		},
	}

	if len(domainValues) > 0 {
		ta := buildTopologyAssignment(levels, domainValues)
		wl.Status.Admission = &kueue.Admission{
			PodSetAssignments: []kueue.PodSetAssignment{
				{
					Name:               "main",
					TopologyAssignment: ta,
				},
			},
		}
	}

	return &workload.Info{Obj: wl}
}

// buildTopologyAssignment creates a simple kueue.TopologyAssignment for a workload
// with one pod per domain.
func buildTopologyAssignment(levels []string, domainValues [][]string) *kueue.TopologyAssignment {
	ta := &kueue.TopologyAssignment{
		Levels: levels,
		Slices: make([]kueue.TopologyAssignmentSlice, len(domainValues)),
	}

	for i, vals := range domainValues {
		valuesPerLevel := make([]kueue.TopologyAssignmentSliceLevelValues, len(vals))
		for j, v := range vals {
			vCopy := v
			valuesPerLevel[j] = kueue.TopologyAssignmentSliceLevelValues{
				Universal: &vCopy,
			}
		}
		ta.Slices[i] = kueue.TopologyAssignmentSlice{
			DomainCount:    1,
			ValuesPerLevel: valuesPerLevel,
			PodCounts: kueue.TopologyAssignmentSlicePodCounts{
				Universal: func() *int32 { v := int32(1); return &v }(),
			},
		}
	}
	return ta
}

func TestDomainCounts(t *testing.T) {
	zoneKey := "topology.kubernetes.io/zone"
	levels := []string{zoneKey}

	tests := map[string]struct {
		wls         []*workload.Info
		namespace   string
		selector    labels.Selector
		topologyKey string
		wantCounts  map[string]int
		wantTotal   int
	}{
		"empty": {
			namespace:   "ns1",
			selector:    labels.Everything(),
			topologyKey: zoneKey,
			wantCounts:  map[string]int{},
			wantTotal:   0,
		},
		"single workload single zone": {
			wls: []*workload.Info{
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"us-east-1"}}),
			},
			namespace:   "ns1",
			selector:    labels.Everything(),
			topologyKey: zoneKey,
			wantCounts:  map[string]int{"us-east-1": 1},
			wantTotal:   1,
		},
		"two workloads in different zones": {
			wls: []*workload.Info{
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-b"}}),
			},
			namespace:   "ns1",
			selector:    labels.Everything(),
			topologyKey: zoneKey,
			wantCounts:  map[string]int{"zone-a": 1, "zone-b": 1},
			wantTotal:   2,
		},
		"three workloads, two in same zone": {
			wls: []*workload.Info{
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-b"}}),
			},
			namespace:   "ns1",
			selector:    labels.Everything(),
			topologyKey: zoneKey,
			wantCounts:  map[string]int{"zone-a": 2, "zone-b": 1},
			wantTotal:   3,
		},
		"different namespace excluded": {
			wls: []*workload.Info{
				buildWL("ns2", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
			},
			namespace:   "ns1",
			selector:    labels.Everything(),
			topologyKey: zoneKey,
			wantCounts:  map[string]int{},
			wantTotal:   0,
		},
		"non-matching selector excluded": {
			wls: []*workload.Info{
				buildWL("ns1", map[string]string{"app": "other"}, levels, [][]string{{"zone-a"}}),
			},
			namespace:   "ns1",
			selector:    labels.SelectorFromSet(labels.Set{"app": "svc"}),
			topologyKey: zoneKey,
			wantCounts:  map[string]int{},
			wantTotal:   0,
		},
		"workload without topology assignment excluded": {
			wls: []*workload.Info{
				buildWL("ns1", map[string]string{"app": "svc"}, nil, nil),
			},
			namespace:   "ns1",
			selector:    labels.Everything(),
			topologyKey: zoneKey,
			wantCounts:  map[string]int{},
			wantTotal:   0,
		},
		"topology key not in levels excluded": {
			wls: []*workload.Info{
				buildWL("ns1", map[string]string{"app": "svc"}, []string{"rack"}, [][]string{{"rack-1"}}),
			},
			namespace:   "ns1",
			selector:    labels.Everything(),
			topologyKey: zoneKey, // zone not in levels
			wantCounts:  map[string]int{},
			wantTotal:   0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			counts, total := DomainCounts(tc.wls, tc.namespace, tc.selector, tc.topologyKey)
			if diff := cmp.Diff(tc.wantCounts, counts); diff != "" {
				t.Errorf("DomainCounts mismatch (-want +got):\n%s", diff)
			}
			if total != tc.wantTotal {
				t.Errorf("total = %d, want %d", total, tc.wantTotal)
			}
		})
	}
}

func TestBannedDomains(t *testing.T) {
	zoneKey := "topology.kubernetes.io/zone"
	levels := []string{zoneKey}

	tests := map[string]struct {
		cfg        *Config
		wls        []*workload.Info
		wantBanned map[string]sets.Set[string]
	}{
		"no workloads yet – nothing banned": {
			cfg: &Config{
				WorkloadLabelSelector: "app=svc",
				Rules:                 []Rule{{Key: zoneKey, MaxDomainPercentage: "45"}},
			},
			wantBanned: map[string]sets.Set[string]{},
		},
		"one zone over cap gets banned": {
			cfg: &Config{
				WorkloadLabelSelector: "app=svc",
				Rules:                 []Rule{{Key: zoneKey, MaxDomainPercentage: "45"}},
			},
			// total=3, count_A=3: 3/3=100% >= 45% → zone-a is banned
			wls: []*workload.Info{
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
			},
			wantBanned: map[string]sets.Set[string]{
				zoneKey: sets.New("zone-a"),
			},
		},
		"preferred rule does not produce banned domains": {
			cfg: &Config{
				WorkloadLabelSelector: "app=svc",
				Rules:                 []Rule{{Key: zoneKey, MaxDomainPercentage: "45", Type: "Preferred"}},
			},
			wls: []*workload.Info{
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
				buildWL("ns1", map[string]string{"app": "svc"}, levels, [][]string{{"zone-a"}}),
			},
			wantBanned: map[string]sets.Set[string]{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := BannedDomains(tc.cfg, tc.wls, "ns1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantBanned, got); diff != "" {
				t.Errorf("BannedDomains mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
