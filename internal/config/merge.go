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

package config

// deepMerge recursively merges overlay into base and returns the result.
// For nested maps, values are merged recursively. For all other types
// (scalars, slices), the overlay value replaces the base value.
// Neither input map is modified; a new map is returned.
func deepMerge(base, overlay map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(base))
	for k, v := range base {
		result[k] = v
	}

	for k, overlayVal := range overlay {
		baseVal, exists := result[k]
		if !exists {
			result[k] = overlayVal
			continue
		}

		baseMap, baseIsMap := toStringMap(baseVal)
		overlayMap, overlayIsMap := toStringMap(overlayVal)

		if baseIsMap && overlayIsMap {
			result[k] = deepMerge(baseMap, overlayMap)
		} else {
			result[k] = overlayVal
		}
	}

	return result
}

// toStringMap attempts to convert a value to map[string]interface{}.
// YAML unmarshaling may produce map[string]interface{} directly.
func toStringMap(v interface{}) (map[string]interface{}, bool) {
	m, ok := v.(map[string]interface{})
	return m, ok
}
