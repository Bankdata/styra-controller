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

package v1beta1

import (
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bankdata/styra-controller/pkg/ptr"
)

var _ = ginkgo.Describe("Expected", func() {
	ginkgo.Describe("Value", func() {
		ginkgo.It("returns true by default", func() {
			gomega.Expect(Expected{}.Value()).To(gomega.BeTrue())
		})

		ginkgo.It("returns true in invalid configurations", func() {
			gomega.Expect(Expected{
				Boolean: ptr.Bool(false),
				String:  ptr.String("test"),
				Integer: ptr.Int(42),
			}.Value()).To(gomega.BeTrue())
		})

		ginkgo.It("returns the value set", func() {
			gomega.Expect(Expected{Boolean: ptr.Bool(true)}.Value()).To(gomega.BeTrue())
			gomega.Expect(Expected{Boolean: ptr.Bool(false)}.Value()).To(gomega.BeFalse())
			gomega.Expect(Expected{String: ptr.String("")}.Value()).To(gomega.Equal(""))
			gomega.Expect(Expected{String: ptr.String("test")}.Value()).To(gomega.Equal("test"))
			gomega.Expect(Expected{Integer: ptr.Int(0)}.Value()).To(gomega.Equal(0))
			gomega.Expect(Expected{Integer: ptr.Int(42)}.Value()).To(gomega.Equal(42))
		})
	})
})

var _ = ginkgo.Describe("System", func() {

	ginkgo.DescribeTable("SetCondition",
		func(
			conditions []Condition,
			conditionType ConditionType,
			status metav1.ConditionStatus,
			expectedConditions []Condition,
		) {
			ss := System{
				Status: SystemStatus{
					Conditions: conditions,
				},
			}
			ss.setCondition(func() time.Time {
				return time.Time{}
			}, conditionType, status)

			gomega.Ω(ss.Status.Conditions).To(gomega.Equal(expectedConditions))
		},

		ginkgo.Entry("Add first condition", nil,
			ConditionTypeCreatedInStyra, metav1.ConditionTrue,
			[]Condition{
				{
					Type:   ConditionTypeCreatedInStyra,
					Status: metav1.ConditionTrue,
				},
			},
		),

		ginkgo.Entry("Add new condition",
			[]Condition{
				{
					Type:   ConditionTypeCreatedInStyra,
					Status: metav1.ConditionTrue,
				},
			},
			ConditionTypeGitCredentialsUpdated, metav1.ConditionFalse,
			[]Condition{
				{
					Type:   ConditionTypeCreatedInStyra,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   ConditionTypeGitCredentialsUpdated,
					Status: metav1.ConditionFalse,
				},
			},
		),

		ginkgo.Entry("Update status on existing condition",
			[]Condition{
				{
					Type:   ConditionTypeCreatedInStyra,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   ConditionTypeGitCredentialsUpdated,
					Status: metav1.ConditionFalse,
				},
			},
			ConditionTypeGitCredentialsUpdated, metav1.ConditionTrue,
			[]Condition{
				{
					Type:   ConditionTypeCreatedInStyra,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   ConditionTypeGitCredentialsUpdated,
					Status: metav1.ConditionTrue,
				},
			},
		),
	)

	ginkgo.DescribeTable("DisplayName",
		func(system *System, prefix, suffix, expected string) {
			gomega.Expect(system.DisplayName(prefix, suffix)).To(gomega.Equal(expected))
		},

		ginkgo.Entry("only namespace and name", &System{ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace",
			Name:      "name",
		}}, "", "", "namespace/name"),

		ginkgo.Entry("with prefix", &System{ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace",
			Name:      "name",
		}}, "test", "", "test/namespace/name"),

		ginkgo.Entry("also with suffix", &System{ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace",
			Name:      "name",
		}}, "test", "cluster1", "test/namespace/name/cluster1"),
	)

	ginkgo.Describe("GitSecretID", func() {
		ginkgo.It("creates the git secret ID", func() {
			s := &System{Status: SystemStatus{ID: "testid"}}
			gomega.Expect(s.GitSecretID()).To(gomega.Equal("systems/testid/git"))
		})
	})
})
