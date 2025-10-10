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

package predicate_test

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"

	testv1 "github.com/bankdata/styra-controller/api/test/v1"
	"github.com/bankdata/styra-controller/internal/predicate"
)

var _ = ginkgo.DescribeTable("ControllerClass",
	func(class string, obj client.Object, expected bool) {
		p, err := predicate.ControllerClass(class)
		gomega.Ω(err).NotTo(gomega.HaveOccurred())
		gomega.Ω(p.Create(event.CreateEvent{Object: obj})).To(gomega.Equal(expected))
	},

	ginkgo.Entry("empty class. no label.", "", &testv1.Object{}, true),

	ginkgo.Entry("empty class. label is set.", "", &testv1.Object{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"styra-controller/class": "test"},
		},
	}, false),

	ginkgo.Entry("empty class. label is empty.", "", &testv1.Object{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"styra-controller/class": ""},
		},
	}, false),

	ginkgo.Entry("class set. no label.", "test", &testv1.Object{}, false),

	ginkgo.Entry("class set. label mismatch", "test", &testv1.Object{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"styra-controller/class": "tset"},
		},
	}, false),

	ginkgo.Entry("class set. label match.", "test", &testv1.Object{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"styra-controller/class": "test"},
		},
	}, true),
)
