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

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("deepMerge", func() {
	ginkgo.It("should return overlay when base is empty", func() {
		base := map[string]interface{}{}
		overlay := map[string]interface{}{"key": "value"}
		result := deepMerge(base, overlay)
		gomega.Ω(result).Should(gomega.Equal(map[string]interface{}{"key": "value"}))
	})

	ginkgo.It("should return base when overlay is empty", func() {
		base := map[string]interface{}{"key": "value"}
		overlay := map[string]interface{}{}
		result := deepMerge(base, overlay)
		gomega.Ω(result).Should(gomega.Equal(map[string]interface{}{"key": "value"}))
	})

	ginkgo.It("should override scalar values", func() {
		base := map[string]interface{}{"key": "base-value", "other": "keep"}
		overlay := map[string]interface{}{"key": "overlay-value"}
		result := deepMerge(base, overlay)
		gomega.Ω(result).Should(gomega.Equal(map[string]interface{}{
			"key":   "overlay-value",
			"other": "keep",
		}))
	})

	ginkgo.It("should replace slices entirely", func() {
		base := map[string]interface{}{
			"items": []interface{}{"a", "b", "c"},
		}
		overlay := map[string]interface{}{
			"items": []interface{}{"x"},
		}
		result := deepMerge(base, overlay)
		gomega.Ω(result).Should(gomega.Equal(map[string]interface{}{
			"items": []interface{}{"x"},
		}))
	})

	ginkgo.It("should recursively merge nested maps", func() {
		base := map[string]interface{}{
			"nested": map[string]interface{}{
				"keep":     "base-value",
				"override": "base-value",
			},
		}
		overlay := map[string]interface{}{
			"nested": map[string]interface{}{
				"override": "overlay-value",
				"new":      "added",
			},
		}
		result := deepMerge(base, overlay)
		gomega.Ω(result).Should(gomega.Equal(map[string]interface{}{
			"nested": map[string]interface{}{
				"keep":     "base-value",
				"override": "overlay-value",
				"new":      "added",
			},
		}))
	})

	ginkgo.It("should not mutate the base map", func() {
		base := map[string]interface{}{"key": "original"}
		overlay := map[string]interface{}{"key": "changed"}
		_ = deepMerge(base, overlay)
		gomega.Ω(base["key"]).Should(gomega.Equal("original"))
	})

	ginkgo.It("should add overlay keys not present in base", func() {
		base := map[string]interface{}{"a": "1"}
		overlay := map[string]interface{}{"b": "2"}
		result := deepMerge(base, overlay)
		gomega.Ω(result).Should(gomega.Equal(map[string]interface{}{
			"a": "1",
			"b": "2",
		}))
	})

	ginkgo.It("should handle deeply nested structures", func() {
		base := map[string]interface{}{
			"l1": map[string]interface{}{
				"l2": map[string]interface{}{
					"l3": "base",
				},
			},
		}
		overlay := map[string]interface{}{
			"l1": map[string]interface{}{
				"l2": map[string]interface{}{
					"l3": "overlay",
				},
			},
		}
		result := deepMerge(base, overlay)
		l1 := result["l1"].(map[string]interface{})
		l2 := l1["l2"].(map[string]interface{})
		gomega.Ω(l2["l3"]).Should(gomega.Equal("overlay"))
	})

	ginkgo.It("should replace map with scalar when overlay changes type", func() {
		base := map[string]interface{}{
			"key": map[string]interface{}{"nested": "value"},
		}
		overlay := map[string]interface{}{
			"key": "scalar-now",
		}
		result := deepMerge(base, overlay)
		gomega.Ω(result["key"]).Should(gomega.Equal("scalar-now"))
	})
})
