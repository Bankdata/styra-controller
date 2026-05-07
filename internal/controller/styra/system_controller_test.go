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

package styra

import (
	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/api/styra/v1beta1"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// test the isURLValid method
var _ = ginkgo.DescribeTable("isURLValid",
	func(url string, expected bool) {
		gomega.Ω(isURLValid(url)).To(gomega.Equal(expected))
	},
	ginkgo.Entry("valid url", "", true),
	ginkgo.Entry("valid url", "https://www.github.com/test/repo.git", true),
	ginkgo.Entry("valid url", "https://www.github.com/test/repo", true),
	ginkgo.Entry("invalid url", "https://www.github.com/[test]/repo", false),
	ginkgo.Entry("invalid url", "https://www.github.com/[test]/repo.git", false),
	ginkgo.Entry("invalid url", "www.google.com", false),
	ginkgo.Entry("invalid url", "google.com", false),
	ginkgo.Entry("invalid url", "google", false),
)

// test the isNamespaceExcluded method
var _ = ginkgo.DescribeTable("isNamespaceExcluded",
	func(namespaceExclusionSelector *configv2alpha2.NamespaceExclusionSelector, systemNamespace string, expected bool) {
		reconciler := &SystemReconciler{
			Config: &configv2alpha2.ProjectConfig{
				NamespaceExclusionSelector: namespaceExclusionSelector,
			},
		}
		system := v1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: systemNamespace,
			},
		}
		gomega.Ω(reconciler.isNamespaceExcluded(&system)).To(gomega.Equal(expected))
	},
	ginkgo.Entry("no exclusion selector",
		nil, "namespace-dev", false),
	ginkgo.Entry("empty exclusion selector",
		&configv2alpha2.NamespaceExclusionSelector{}, "namespace-dev", false),
	ginkgo.Entry("no match multiple patterns",
		&configv2alpha2.NamespaceExclusionSelector{MatchPatterns: []string{"dev-*"}}, "namespace-dev", false),
	ginkgo.Entry("match single pattern",
		&configv2alpha2.NamespaceExclusionSelector{MatchPatterns: []string{"*-dev"}}, "namespace-dev", true),
	ginkgo.Entry("no match single pattern",
		&configv2alpha2.NamespaceExclusionSelector{MatchPatterns: []string{"*-staging", "*-prod"}},
		"namespace-dev", false),
	ginkgo.Entry("match multiple patterns",
		&configv2alpha2.NamespaceExclusionSelector{MatchPatterns: []string{"*-dev", "*-staging"}},
		"namespace-dev", true),
)
