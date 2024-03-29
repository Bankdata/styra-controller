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

package labels_test

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	testv1 "github.com/bankdata/styra-controller/api/test/v1"
	"github.com/bankdata/styra-controller/internal/labels"
)

var _ = ginkgo.Describe("SetManagedBy", func() {
	ginkgo.It("should set the managed-by label", func() {
		o := testv1.Object{}
		labels.SetManagedBy(&o)
		gomega.Ω(o.Labels["app.kubernetes.io/managed-by"]).To(gomega.Equal("styra-controller"))
	})
})

var _ = ginkgo.DescribeTable("HasManagedBy",
	func(o client.Object, expected bool) {
		gomega.Ω(labels.HasManagedBy(o)).To(gomega.Equal(expected))
	},
	ginkgo.Entry(
		"should return false if labels is nil",
		&testv1.Object{},
		false,
	),
	ginkgo.Entry(
		"should return false if label is set to something unexpected",
		&testv1.Object{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
			"app.kubernetes.io/managed-by": "something-unexpected",
		}}},
		false,
	),
	ginkgo.Entry(
		"should return true if label is set as expected",
		&testv1.Object{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
			"app.kubernetes.io/managed-by": "styra-controller",
		}}},
		true,
	),
)

var _ = ginkgo.DescribeTable("ControllerClassMatches",
	func(o client.Object, class string, expected bool) {
		gomega.Ω(labels.ControllerClassMatches(o, class)).To(gomega.Equal(expected))
	},
	ginkgo.Entry(
		"should return false if labels is nil and class is non-empty",
		&testv1.Object{},
		"test",
		false,
	),
	ginkgo.Entry(
		"should return true if labels is nil and class empty",
		&testv1.Object{},
		"",
		true,
	),
	ginkgo.Entry(
		"should return true if label is missing but class is empty",
		&testv1.Object{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}},
		"",
		true,
	),
	ginkgo.Entry(
		"should return true if label value matches",
		&testv1.Object{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
			"styra-controller/class": "test",
		}}},
		"test",
		true,
	),
)
