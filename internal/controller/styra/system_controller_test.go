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

package styra

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv2alpha1 "github.com/bankdata/styra-controller/api/config/v2alpha1"
	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = DescribeTable("createRolebindingSubjects",
	func(subjects []styrav1beta1.Subject, expectedSubject []*styra.Subject) {
		Ω(createRolebindingSubjects(subjects, &configv2alpha1.SSOConfig{
			IdentityProvider: "BDAD",
			JWTGroupsClaim:   "groups",
		})).To(Equal(expectedSubject))
	},

	Entry("returns same adgroup",
		[]styrav1beta1.Subject{{Kind: "group", Name: "ADTEST"}},
		[]*styra.Subject{{
			Kind: "claim",
			ClaimConfig: &styra.ClaimConfig{
				IdentityProvider: "BDAD",
				Key:              "groups",
				Value:            "ADTEST",
			},
		}},
	),

	Entry("defaults empty kind value to user",
		[]styrav1beta1.Subject{
			{Kind: "user", Name: "test1@test.dk"},
			{Name: "test2@test.dk"},
			{Kind: "group", Name: "ADTEST"},
		},
		[]*styra.Subject{
			{
				Kind: "user",
				ID:   "test1@test.dk",
			}, {
				Kind: "user",
				ID:   "test2@test.dk",
			}, {
				Kind: "claim",
				ClaimConfig: &styra.ClaimConfig{
					IdentityProvider: "BDAD",
					Key:              "groups",
					Value:            "ADTEST",
				},
			},
		},
	),

	Entry("does not return duplicates adgroups",
		[]styrav1beta1.Subject{
			{Kind: "group", Name: "ADTEST"},
			{Kind: "group", Name: "ADTEST1"},
			{Kind: "group", Name: "ADTEST1"},
		},
		[]*styra.Subject{
			{
				Kind: "claim",
				ClaimConfig: &styra.ClaimConfig{
					IdentityProvider: "BDAD",
					Key:              "groups",
					Value:            "ADTEST",
				},
			}, {
				Kind: "claim",
				ClaimConfig: &styra.ClaimConfig{
					IdentityProvider: "BDAD",
					Key:              "groups",
					Value:            "ADTEST1",
				},
			},
		},
	),

	Entry("does not return duplicates users and groups",
		[]styrav1beta1.Subject{
			{Kind: "user", Name: "test@test.dk"},
			{Kind: "user", Name: "test@test.dk"},
			{Kind: "group", Name: "ADTEST"},
			{Kind: "group", Name: "ADTEST"},
		},
		[]*styra.Subject{
			{
				Kind: "user",
				ID:   "test@test.dk",
			}, {
				Kind: "claim",
				ClaimConfig: &styra.ClaimConfig{
					IdentityProvider: "BDAD",
					Key:              "groups",
					Value:            "ADTEST",
				},
			},
		},
	),
)
