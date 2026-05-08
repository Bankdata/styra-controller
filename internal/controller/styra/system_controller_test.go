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
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/api/styra/v1beta1"
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

// test the isSystemControllerClassOnIgnoreList method
var _ = ginkgo.DescribeTable("isSystemControllerClassOnIgnoreList",
	func(ignoreList []string, labels map[string]string, expected bool) {
		r := &SystemReconciler{
			Config: &configv2alpha2.ProjectConfig{
				SystemControllerClassesIgnoreList: ignoreList,
			},
		}
		system := &v1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
		}
		gomega.Ω(r.isSystemControllerClassOnIgnoreList(system)).To(gomega.Equal(expected))
	},
	ginkgo.Entry("empty ignore list returns false", []string{},
		map[string]string{"styra-controller/class": "myclass"}, false),
	ginkgo.Entry("nil ignore list returns false", nil,
		map[string]string{"styra-controller/class": "myclass"}, false),
	ginkgo.Entry("class in ignore list returns true", []string{"myclass"},
		map[string]string{"styra-controller/class": "myclass"}, true),
	ginkgo.Entry("class not in ignore list returns false", []string{"otherclass"},
		map[string]string{"styra-controller/class": "myclass"}, false),
	ginkgo.Entry("class matches one of multiple entries", []string{"a", "myclass", "b"},
		map[string]string{"styra-controller/class": "myclass"}, true),
	ginkgo.Entry("no labels, empty class in ignore list returns false", []string{""}, nil, false),
	ginkgo.Entry("no labels, non-empty class in ignore list returns false", []string{"myclass"}, nil, false),
)

// test the isSystemNamespaceMatchingSelector method
var _ = ginkgo.DescribeTable("isSystemNamespaceMatchingSelector",
	func(namespaceSelector *configv2alpha2.NamespaceSelector, namespace string, expected bool) {
		r := &SystemReconciler{
			Config: &configv2alpha2.ProjectConfig{
				NamespaceSelector: namespaceSelector,
			},
		}
		system := &v1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
			},
		}
		gomega.Ω(r.isSystemNamespaceMatchingSelector(system)).To(gomega.Equal(expected))
	},
	ginkgo.Entry("nil selector returns true", nil, "mynamespace", true),
	ginkgo.Entry("exact match returns true", &configv2alpha2.NamespaceSelector{
		MatchPatterns: []string{"mynamespace"}}, "mynamespace", true),
	ginkgo.Entry("no match returns false", &configv2alpha2.NamespaceSelector{
		MatchPatterns: []string{"other"}}, "mynamespace", false),
	ginkgo.Entry("wildcard matches namespace", &configv2alpha2.NamespaceSelector{
		MatchPatterns: []string{"*-dev"}}, "mynamespace-dev", true),
	ginkgo.Entry("wildcard does not match unrelated namespace", &configv2alpha2.NamespaceSelector{
		MatchPatterns: []string{"*-dev"}}, "other", false),
	ginkgo.Entry("matches one of multiple patterns", &configv2alpha2.NamespaceSelector{
		MatchPatterns: []string{"*-dev", "*-staging", "*-prod"}}, "mynamespace-staging", true),
	ginkgo.Entry("empty patterns returns true", &configv2alpha2.NamespaceSelector{
		MatchPatterns: []string{}}, "mynamespace", true),
)
