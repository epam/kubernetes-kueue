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
	"encoding/json"
	"fmt"
	"slices"
	"strconv"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"

	utiltas "sigs.k8s.io/kueue/pkg/util/tas"
	"sigs.k8s.io/kueue/pkg/workload"
)

const (
	// RuleTypeRequired is the enforcement mode that blocks admission into over-crowded domains.
	RuleTypeRequired = "Required"
	// RuleTypePreferred is the enforcement mode that penalizes over-crowded domains but still
	// allows admission.
	RuleTypePreferred = "Preferred"

	// maxRulesAlpha is the maximum number of spreading rules allowed in the alpha milestone.
	maxRulesAlpha = 2
)

// Config is the parsed JSON object from the
// [alpha].kueue.x-k8s.io/topology-spreading annotation.
type Config struct {
	// WorkloadLabelSelector is a Kubernetes label selector string that identifies
	// the set of workloads (same namespace) to count when evaluating spreading.
	WorkloadLabelSelector string `json:"workload-label-selector"`
	// Rules is the list of spreading rules.
	Rules []Rule `json:"rules"`
}

// Rule is a single spreading rule inside a Config.
type Rule struct {
	// Key is the topology domain label key (must match a level in the ResourceFlavor's Topology).
	Key string `json:"key"`
	// MaxDomainPercentage is an integer in [1,99] expressed as a decimal string.
	// It caps the fraction of matching workloads that may reside in any single domain.
	MaxDomainPercentage string `json:"max-domain-percentage"`
	// Type is the enforcement mode: "Required" (default) or "Preferred".
	Type string `json:"type,omitempty"`
}

// maxDomainPct returns the integer percentage value of r.MaxDomainPercentage.
// The annotation validator guarantees it is in [1,99] before this is called.
func (r *Rule) maxDomainPct() int {
	v, _ := strconv.Atoi(r.MaxDomainPercentage)
	return v
}

// IsRequired returns true if the rule is Required (default when Type is empty).
func (r *Rule) IsRequired() bool {
	return r.Type == "" || r.Type == RuleTypeRequired
}

// ParseAndValidate parses the annotation value and validates the resulting Config.
// It returns a descriptive error when the annotation is malformed or contains
// out-of-range values.
func ParseAndValidate(annotationValue string) (*Config, error) {
	cfg := &Config{}
	if err := json.Unmarshal([]byte(annotationValue), cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.WorkloadLabelSelector == "" {
		return nil, fmt.Errorf("workload-label-selector is required")
	}

	if _, err := labels.Parse(cfg.WorkloadLabelSelector); err != nil {
		return nil, fmt.Errorf("invalid workload-label-selector %q: %w", cfg.WorkloadLabelSelector, err)
	}

	if len(cfg.Rules) == 0 {
		return nil, fmt.Errorf("rules must not be empty")
	}

	if len(cfg.Rules) > maxRulesAlpha {
		return nil, fmt.Errorf("rules must have at most %d entries (alpha limit), got %d", maxRulesAlpha, len(cfg.Rules))
	}

	for i, r := range cfg.Rules {
		if r.Key == "" {
			return nil, fmt.Errorf("rules[%d].key is required", i)
		}
		if r.MaxDomainPercentage == "" {
			return nil, fmt.Errorf("rules[%d].max-domain-percentage is required", i)
		}
		pct, err := strconv.Atoi(r.MaxDomainPercentage)
		if err != nil || pct < 1 || pct > 99 {
			return nil, fmt.Errorf("rules[%d].max-domain-percentage must be an integer in [1,99], got %q", i, r.MaxDomainPercentage)
		}
		if r.Type != "" && r.Type != RuleTypeRequired && r.Type != RuleTypePreferred {
			return nil, fmt.Errorf("rules[%d].type must be %q or %q, got %q", i, RuleTypeRequired, RuleTypePreferred, r.Type)
		}
	}
	return cfg, nil
}

// IsBanned returns true when domain D should be banned given its current count,
// the total count of matching workloads, and the maximum domain percentage cap.
//
// Admission constraint (from KEP-13746 table / "Equivalently" clause):
//
//	count(D) / total < max_domain_percentage / 100
//
// A domain is admissible when the above strict inequality holds; we ban it when:
//
//	count(D) / total >= max_domain_percentage / 100
//
// Using integer arithmetic (no floating-point division): banned when
//
//	100 * count(D) >= max_domain_percentage * total
//
// Cold start (total == 0): no domain is banned so the first workload is
// always admitted.
func IsBanned(count, total, maxPct int) bool {
	if total == 0 {
		return false
	}
	return 100*count >= maxPct*total
}

// DomainCounts returns per-domain workload counts for a given topology key
// and the total number of matching workloads (those that have an assignment at
// that topology key level).
//
// "domain value" for a workload is the topology label value at the given level
// taken from its TopologyAssignment. Workloads that span multiple domains (e.g.
// training jobs placed across zones) are counted once per unique domain they
// occupy at that level.
//
// Only workloads in the given namespace whose labels match selector are counted.
// workloads is the slice of admitted workload.Info objects to search.
func DomainCounts(wls []*workload.Info, namespace string, selector labels.Selector, topologyKey string) (map[string]int, int) {
	domainCounts := make(map[string]int)
	total := 0

	for _, wl := range wls {
		if wl.Obj.Namespace != namespace {
			continue
		}
		if !selector.Matches(labels.Set(wl.Obj.Labels)) {
			continue
		}
		if wl.Obj.Status.Admission == nil {
			continue
		}

		// Collect unique zone values across all PodSets of this workload.
		zonesForWorkload := sets.New[string]()
		for _, psa := range wl.Obj.Status.Admission.PodSetAssignments {
			if psa.TopologyAssignment == nil {
				continue
			}
			ta := utiltas.InternalFrom(psa.TopologyAssignment)
			if ta == nil {
				continue
			}
			levelIdx := slices.Index(ta.Levels, topologyKey)
			if levelIdx < 0 {
				continue
			}
			for _, domain := range ta.Domains {
				if levelIdx < len(domain.Values) {
					zonesForWorkload.Insert(domain.Values[levelIdx])
				}
			}
		}
		if zonesForWorkload.Len() > 0 {
			total++
			for zone := range zonesForWorkload {
				domainCounts[zone]++
			}
		}
	}
	return domainCounts, total
}

// BannedDomains computes, for each Rule in cfg that is of type Required, the set
// of banned domain values for its topology key.
//
// wls is the full slice of admitted workloads that the caller has already
// filtered to the target namespace.
//
// Returns a map: topology key → set of banned domain values.
// Only Required rules contribute to the result; Preferred rules are handled
// separately via scoring.
func BannedDomains(cfg *Config, wls []*workload.Info, namespace string) (map[string]sets.Set[string], error) {
	selector, err := labels.Parse(cfg.WorkloadLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector: %w", err)
	}

	result := make(map[string]sets.Set[string])

	for _, rule := range cfg.Rules {
		if !rule.IsRequired() {
			continue
		}
		counts, total := DomainCounts(wls, namespace, selector, rule.Key)
		maxPct := rule.maxDomainPct()
		banned := sets.New[string]()
		for domain, count := range counts {
			if IsBanned(count, total, maxPct) {
				banned.Insert(domain)
			}
		}
		if banned.Len() > 0 {
			if existing, ok := result[rule.Key]; ok {
				result[rule.Key] = existing.Union(banned)
			} else {
				result[rule.Key] = banned
			}
		}
	}
	return result, nil
}
