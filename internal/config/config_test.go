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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/bankdata/styra-controller/api/config/v1"
	"github.com/bankdata/styra-controller/api/config/v2alpha1"
)

var _ = DescribeTable("deserialize",
	func(data []byte, expected *v2alpha1.ProjectConfig, shouldErr bool) {
		scheme := runtime.NewScheme()
		err := v2alpha1.AddToScheme(scheme)
		Ω(err).ShouldNot(HaveOccurred())
		err = v1.AddToScheme(scheme)
		Ω(err).ShouldNot(HaveOccurred())
		actual, err := deserialize(data, scheme)
		if shouldErr {
			Ω(err).Should(HaveOccurred())
		} else {
			Ω(err).ShouldNot(HaveOccurred())
		}
		Ω(actual).Should(Equal(expected))
	},

	Entry("errors on unexpected api group",
		[]byte(`
apiVersion: myconfig.bankdata.dk/v1
kind: ProjectConfig
styra:
  token: my-token
`),
		nil,
		true,
	),

	Entry("can deserialize v1",
		[]byte(`
apiVersion: config.bankdata.dk/v1
kind: ProjectConfig
styraToken: my-token
`),
		&v2alpha1.ProjectConfig{
			Styra: v2alpha1.StyraConfig{
				Token: "my-token",
			},
		},
		false,
	),

	Entry("can deserialize v2alpha1",
		[]byte(`
apiVersion: config.bankdata.dk/v2alpha1
kind: ProjectConfig
styra:
  token: my-token
`),
		&v2alpha1.ProjectConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ProjectConfig",
				APIVersion: v2alpha1.GroupVersion.Identifier(),
			},
			Styra: v2alpha1.StyraConfig{
				Token: "my-token",
			},
		},
		false,
	),
)
