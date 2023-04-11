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

package v2alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/api/config/v2alpha1"
)

var _ = Describe("ProjectConfig", func() {
	DescribeTable("GetGitCredentialForRepo",
		func(gitCredentials []*v2alpha1.GitCredential, repo string, expected *v2alpha1.GitCredential) {
			c := &v2alpha1.ProjectConfig{
				GitCredentials: gitCredentials,
			}
			Ω(c.GetGitCredentialForRepo(repo)).To(Equal(expected))
		},

		Entry("returns nil if the list of credentials is empty", nil, "test", nil),

		Entry("finds matching credential",
			[]*v2alpha1.GitCredential{
				{
					User:       "",
					Password:   "",
					RepoPrefix: "https://github.com/bankdata",
				},
			},
			"https://github.com/bankdata/styra-controller.git",
			&v2alpha1.GitCredential{
				User:       "",
				Password:   "",
				RepoPrefix: "https://github.com/bankdata",
			},
		),
		Entry("returns longest matching credential",
			[]*v2alpha1.GitCredential{
				{
					User:       "",
					Password:   "",
					RepoPrefix: "https://github.com/bankdata",
				},
				{
					User:       "",
					Password:   "",
					RepoPrefix: "https://github.com/bankdata/styra-controller",
				},
			},
			"https://github.com/bankdata/styra-controller.git",
			&v2alpha1.GitCredential{
				User:       "",
				Password:   "",
				RepoPrefix: "https://github.com/bankdata/styra-controller",
			},
		),
	)

})
