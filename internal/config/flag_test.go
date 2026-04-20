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
	"flag"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("StringSlice", func() {
	ginkgo.It("implements flag.Value interface", func() {
		var s StringSlice
		var _ flag.Value = &s
	})

	ginkgo.It("starts empty and returns empty string", func() {
		var s StringSlice
		gomega.Ω(s.String()).Should(gomega.Equal(""))
		gomega.Ω(s).Should(gomega.BeEmpty())
	})

	ginkgo.It("appends values on each Set call", func() {
		var s StringSlice
		gomega.Ω(s.Set("file1.yaml")).Should(gomega.Succeed())
		gomega.Ω(s.Set("file2.yaml")).Should(gomega.Succeed())
		gomega.Ω(s).Should(gomega.Equal(StringSlice{"file1.yaml", "file2.yaml"}))
	})

	ginkgo.It("returns comma-separated string representation", func() {
		s := StringSlice{"a.yaml", "b.yaml", "c.yaml"}
		gomega.Ω(s.String()).Should(gomega.Equal("a.yaml,b.yaml,c.yaml"))
	})

	ginkgo.It("works with flag.FlagSet", func() {
		var files StringSlice
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.Var(&files, "config", "config file")

		err := fs.Parse([]string{"--config=one.yaml", "--config=two.yaml"})
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(files).Should(gomega.Equal(StringSlice{"one.yaml", "two.yaml"}))
	})

	ginkgo.It("works with flag.FlagSet using space-separated values", func() {
		var files StringSlice
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.Var(&files, "config", "config file")

		err := fs.Parse([]string{"--config", "one.yaml", "--config", "two.yaml"})
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(files).Should(gomega.Equal(StringSlice{"one.yaml", "two.yaml"}))
	})
})
