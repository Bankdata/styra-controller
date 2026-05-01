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
