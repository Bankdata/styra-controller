/*
Copyright (C) 2023 Bankdata (bankdata@bankdata.dk)

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

package ptr_test

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/ptr"
)

var _ = ginkgo.Describe("Bool", func() {
	ginkgo.It("should return a pointer to the boolean", func() {
		gomega.Expect(*ptr.Bool(true)).To(gomega.BeTrue())
		gomega.Expect(*ptr.Bool(false)).To(gomega.BeFalse())
	})
})

var _ = ginkgo.Describe("String", func() {
	ginkgo.It("should return a pointer to the string", func() {
		gomega.Expect(*ptr.String("")).To(gomega.Equal(""))
		gomega.Expect(*ptr.String("test")).To(gomega.Equal("test"))
	})
})

var _ = ginkgo.Describe("Int", func() {
	ginkgo.It("should return a pointer to the int", func() {
		gomega.Expect(*ptr.Int(0)).To(gomega.Equal(0))
		gomega.Expect(*ptr.Int(42)).To(gomega.Equal(42))
	})
})
