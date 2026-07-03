/*
Copyright (C) 2025 Bankdata (bankdata@bankdata.dk)

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

// Package opaconfig contains helpers for validating and converting OPA config.
package opaconfig

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

var knownTopLevelKeys = map[string]struct{}{
	"services":                       {},
	"labels":                         {},
	"discovery":                      {},
	"bundle":                         {},
	"bundles":                        {},
	"decision_logs":                  {},
	"status":                         {},
	"plugins":                        {},
	"keys":                           {},
	"default_decision":               {},
	"default_authorization_decision": {},
	"caching":                        {},
	"nd_builtin_cache":               {},
	"persistence_directory":          {},
	"distributed_tracing":            {},
	"metrics_export":                 {},
	"server":                         {},
	"storage":                        {},
}

// ValidateRaw validates raw OPA config bytes and rejects unknown top-level
// keys.
func ValidateRaw(raw []byte) error {
	if len(raw) == 0 {
		return nil
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("could not parse OPA config: %w", err)
	}

	if len(cfg) == 0 {
		return nil
	}

	keys := make([]string, 0)
	for key := range cfg {
		if _, ok := knownTopLevelKeys[key]; !ok {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nil
	}

	sort.Strings(keys)
	return fmt.Errorf("unknown OPA configuration key(s): %s", strings.Join(keys, ", "))
}

// ToMap converts a typed OPA config into a map representation that can be
// merged with other YAML-derived maps.
func ToMap(cfg any) (map[string]interface{}, error) {
	if cfg == nil {
		return nil, nil
	}

	bs, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not marshal OPA config: %w", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(bs, &result); err != nil {
		return nil, fmt.Errorf("could not unmarshal OPA config: %w", err)
	}

	return result, nil
}
