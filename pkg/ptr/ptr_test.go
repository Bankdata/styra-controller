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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/ptr"
)

var _ = Describe("Bool", func() {
	It("should return a pointer to the boolean", func() {
		Expect(*ptr.Bool(true)).To(BeTrue())
		Expect(*ptr.Bool(false)).To(BeFalse())
	})
})

var _ = Describe("String", func() {
	It("should return a pointer to the string", func() {
		Expect(*ptr.String("")).To(Equal(""))
		Expect(*ptr.String("test")).To(Equal("test"))
	})
})

var _ = Describe("Int", func() {
	It("should return a pointer to the int", func() {
		Expect(*ptr.Int(0)).To(Equal(0))
		Expect(*ptr.Int(42)).To(Equal(42))
	})
})
