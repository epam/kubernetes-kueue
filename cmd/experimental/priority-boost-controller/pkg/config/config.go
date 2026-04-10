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

package config

import (
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"
)

// Configuration defines the configuration for the priority-boost-controller.
type Configuration struct {
	// MinAdmitDuration is the minimum time a workload must be admitted before
	// a post-window preemption signal applies. Set to "0" to disable.
	// Examples: "30m", "1h".
	MinAdmitDuration string `json:"minAdmitDuration,omitempty"`

	// BoostValue is the magnitude of the negative priority-boost applied after
	// MinAdmitDuration while the workload remains admitted (see README).
	BoostValue int32 `json:"boostValue,omitempty"`

	// WorkloadSelector, if set, limits which workloads the controller manages.
	// Workloads whose labels do not match are left with no annotation (any
	// existing annotation is cleared).
	WorkloadSelector *metav1.LabelSelector `json:"workloadSelector,omitempty"`

	// MaxWorkloadPriority, if set, excludes workloads with Spec.Priority greater
	// than this value (nil Spec.Priority is treated as 0). Excluded workloads
	// have any priority-boost annotation cleared.
	MaxWorkloadPriority *int32 `json:"maxWorkloadPriority,omitempty"`
}

// Default returns the default configuration.
func Default() Configuration {
	return Configuration{
		MinAdmitDuration: "0",
		BoostValue:       100000,
	}
}

// ParsedConfiguration holds the parsed, ready-to-use form of Configuration.
type ParsedConfiguration struct {
	MinAdmitDuration    time.Duration
	BoostValue          int32
	WorkloadSelector    labels.Selector
	MaxWorkloadPriority *int32
}

// Parse converts a Configuration into a ParsedConfiguration.
func Parse(cfg Configuration) (*ParsedConfiguration, error) {
	var minAdmitDuration time.Duration
	if cfg.MinAdmitDuration != "" && cfg.MinAdmitDuration != "0" {
		d, err := time.ParseDuration(cfg.MinAdmitDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid minAdmitDuration %q: %w", cfg.MinAdmitDuration, err)
		}
		minAdmitDuration = d
	}

	out := &ParsedConfiguration{
		MinAdmitDuration:    minAdmitDuration,
		BoostValue:          cfg.BoostValue,
		MaxWorkloadPriority: cfg.MaxWorkloadPriority,
	}

	if cfg.WorkloadSelector != nil {
		sel, err := metav1.LabelSelectorAsSelector(cfg.WorkloadSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid workloadSelector: %w", err)
		}
		out.WorkloadSelector = sel
	}

	return out, nil
}

// Load reads the configuration from the given file path.
// If the path is empty, it returns the default configuration.
func Load(path string) (*ParsedConfiguration, error) {
	cfg := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config file %s: %w", path, err)
		}
	}
	return Parse(cfg)
}
