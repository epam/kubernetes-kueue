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

	"sigs.k8s.io/yaml"
)

// Configuration defines the configuration for the priority-boost-controller.
type Configuration struct {
	// MinAdmitDuration is the minimum time a workload must be admitted before
	// it can be preempted by same-priority workloads. Set to "0" to disable
	// time-sharing protection. Examples: "30m", "1h".
	MinAdmitDuration string `json:"minAdmitDuration,omitempty"`

	// BoostValue is the priority boost applied to an admitted workload during
	// the MinAdmitDuration protection window. Must be large enough that
	// same-base-priority pending workloads cannot preempt the protected workload.
	BoostValue int32 `json:"boostValue,omitempty"`
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
	MinAdmitDuration time.Duration
	BoostValue       int32
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
	return &ParsedConfiguration{
		MinAdmitDuration: minAdmitDuration,
		BoostValue:       cfg.BoostValue,
	}, nil
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
