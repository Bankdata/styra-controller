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

package config

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/bankdata/styra-controller/api/config/v2alpha2"
)

var _ = ginkgo.DescribeTable("deserialize",
	func(data []byte, expected *v2alpha2.ProjectConfig, shouldErr bool) {
		scheme := runtime.NewScheme()
		err := v2alpha2.AddToScheme(scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		actual, err := deserialize(data, scheme)
		if shouldErr {
			gomega.Ω(err).Should(gomega.HaveOccurred())
		} else {
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		}
		gomega.Ω(actual).Should(gomega.Equal(expected))
	},

	ginkgo.Entry("errors on unexpected api group",
		[]byte(`
apiVersion: myconfig.bankdata.dk/v1
kind: ProjectConfig
styra:
  token: my-token
`),
		nil,
		true,
	),

	ginkgo.Entry("can deserialize v2alpha2",
		[]byte(`
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
styra:
  token: my-token
`),
		&v2alpha2.ProjectConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ProjectConfig",
				APIVersion: v2alpha2.GroupVersion.Identifier(),
			},
			Styra: v2alpha2.StyraConfig{
				Token: "my-token",
			},
		},
		false,
	),
)
