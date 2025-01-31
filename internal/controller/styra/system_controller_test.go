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
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.DescribeTable("createRolebindingSubjects",
	func(subjects []styrav1beta1.Subject, expectedSubject []*styra.Subject) {
		gomega.Ω(createRolebindingSubjects(subjects, &configv2alpha2.SSOConfig{
			IdentityProvider: "BDAD",
			JWTGroupsClaim:   "groups",
		})).To(gomega.Equal(expectedSubject))
	},

	ginkgo.Entry("returns same adgroup",
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

	ginkgo.Entry("defaults empty kind value to user",
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

	ginkgo.Entry("does not return duplicates adgroups",
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

	ginkgo.Entry("does not return duplicates users and groups",
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

// test the isURLValid method
var _ = ginkgo.DescribeTable("isURLValid",
	func(url string, expected bool) {
		gomega.Ω(isURLValid(url)).To(gomega.Equal(expected))
	},
	ginkgo.Entry("valid url", "", true),
	ginkgo.Entry("valid url", "https://www.github.com/test/repo.git", true),
	ginkgo.Entry("valid url", "https://www.github.com/test/repo", true),
	ginkgo.Entry("invalid url", "https://www.github.com/[test]/repo", false),
	ginkgo.Entry("invalid url", "https://www.github.com/[test]/repo.git", false),
	ginkgo.Entry("invalid url", "www.google.com", false),
	ginkgo.Entry("invalid url", "google.com", false),
	ginkgo.Entry("invalid url", "google", false),
)
