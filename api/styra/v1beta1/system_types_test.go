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

package v1beta1

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bankdata/styra-controller/pkg/ptr"
)

var _ = Describe("Expected", func() {
	Describe("Value", func() {
		It("returns true by default", func() {
			Expect(Expected{}.Value()).To(BeTrue())
		})

		It("returns true in invalid configurations", func() {
			Expect(Expected{
				Boolean: ptr.Bool(false),
				String:  ptr.String("test"),
				Integer: ptr.Int(42),
			}.Value()).To(BeTrue())
		})

		It("returns the value set", func() {
			Expect(Expected{Boolean: ptr.Bool(true)}.Value()).To(BeTrue())
			Expect(Expected{Boolean: ptr.Bool(false)}.Value()).To(BeFalse())
			Expect(Expected{String: ptr.String("")}.Value()).To(Equal(""))
			Expect(Expected{String: ptr.String("test")}.Value()).To(Equal("test"))
			Expect(Expected{Integer: ptr.Int(0)}.Value()).To(Equal(0))
			Expect(Expected{Integer: ptr.Int(42)}.Value()).To(Equal(42))
		})
	})
})

var _ = Describe("System", func() {

	DescribeTable("SetCondition",
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

			Ω(ss.Status.Conditions).To(Equal(expectedConditions))
		},

		Entry("Add first condition", nil,
			ConditionTypeCreatedInStyra, metav1.ConditionTrue,
			[]Condition{
				{
					Type:   ConditionTypeCreatedInStyra,
					Status: metav1.ConditionTrue,
				},
			},
		),

		Entry("Add new condition",
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

		Entry("Update status on existing condition",
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

	DescribeTable("DisplayName",
		func(system *System, prefix, suffix, expected string) {
			Expect(system.DisplayName(prefix, suffix)).To(Equal(expected))
		},

		Entry("only namespace and name", &System{ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace",
			Name:      "name",
		}}, "", "", "namespace/name"),

		Entry("with prefix", &System{ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace",
			Name:      "name",
		}}, "test", "", "test/namespace/name"),

		Entry("also with suffix", &System{ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace",
			Name:      "name",
		}}, "test", "cluster1", "test/namespace/name/cluster1"),
	)

	Describe("GitSecretID", func() {
		It("creates the git secret ID", func() {
			s := &System{Status: SystemStatus{ID: "testid"}}
			Expect(s.GitSecretID()).To(Equal("systems/testid/git"))
		})
	})
})
