// /*
// Copyright (C) 2023 Bankdata (bankdata@bankdata.dk)

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package styra

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	"github.com/bankdata/styra-controller/pkg/ptr"
	"github.com/bankdata/styra-controller/pkg/styra"
)

type specToUpdateTest struct {
	cfg      *configv2alpha2.ProjectConfig
	ds       *styrav1alpha1.GlobalDatasource
	expected *styra.UpsertDatasourceRequest
}

var _ = ginkgo.DescribeTable("globalDatasourceSpecToUpdate", func(test specToUpdateTest) {
	cfg := test.cfg
	if cfg == nil {
		cfg = &configv2alpha2.ProjectConfig{}
	}
	r := &GlobalDatasourceReconciler{Config: cfg}
	gomega.Expect(r.specToUpdate(test.ds)).To(gomega.Equal(test.expected))
},
	ginkgo.Entry("ds is nil", specToUpdateTest{
		ds:       nil,
		expected: nil,
	}),

	ginkgo.Entry("zero value", specToUpdateTest{
		ds: &styrav1alpha1.GlobalDatasource{},
		expected: &styra.UpsertDatasourceRequest{
			Enabled: true,
		},
	}),

	ginkgo.Entry("using default git credentials", specToUpdateTest{
		cfg: &configv2alpha2.ProjectConfig{
			GitCredentials: []*configv2alpha2.GitCredential{
				{User: "test-user", Password: "test-pw"},
			},
		},
		ds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec:       styrav1alpha1.GlobalDatasourceSpec{Name: "test"},
		},
		expected: &styra.UpsertDatasourceRequest{
			Credentials: "libraries/global/test/git",
			Enabled:     true,
		},
	}),

	ginkgo.Entry("using credentials from secret", specToUpdateTest{
		cfg: &configv2alpha2.ProjectConfig{
			GitCredentials: []*configv2alpha2.GitCredential{
				{User: "test-user", Password: "test-pw"},
			},
		},
		ds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: styrav1alpha1.GlobalDatasourceSpec{
				Name: "test",
				CredentialsSecretRef: &styrav1alpha1.GlobalDatasourceSecretRef{
					Namespace: "test-ns",
					Name:      "test-name",
				},
			},
		},
		expected: &styra.UpsertDatasourceRequest{
			Credentials: "libraries/global/test/git",
			Enabled:     true,
		},
	}),

	ginkgo.Entry("setting all fields", specToUpdateTest{
		cfg: &configv2alpha2.ProjectConfig{
			GitCredentials: []*configv2alpha2.GitCredential{
				{User: "test-user", Password: "test-pw"},
			},
		},
		ds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: styrav1alpha1.GlobalDatasourceSpec{
				Name:        "test",
				Category:    "test-category",
				Description: "test-description",
				URL:         "test-url",
				Reference:   "test reference",
				Enabled:     ptr.Bool(true),
				Commit:      "test",
				Path:        "test path",
				CredentialsSecretRef: &styrav1alpha1.GlobalDatasourceSecretRef{
					Namespace: "test-ns",
					Name:      "test-name",
				},
			},
		},
		expected: &styra.UpsertDatasourceRequest{
			Category:    "test-category",
			Description: "test-description",
			Enabled:     true,
			Commit:      "test",
			Credentials: "libraries/global/test/git",
			Reference:   "test reference",
			URL:         "test-url",
			Path:        "test path",
		},
	}),
)

type needsUpdateTest struct {
	cfg      *configv2alpha2.ProjectConfig
	gds      *styrav1alpha1.GlobalDatasource
	dc       *styra.DatasourceConfig
	expected bool
}

var _ = ginkgo.DescribeTable("needsUpdate", func(test needsUpdateTest) {
	cfg := test.cfg
	if cfg == nil {
		cfg = &configv2alpha2.ProjectConfig{}
	}
	r := &GlobalDatasourceReconciler{Config: cfg}
	gomega.Expect(r.needsUpdate(test.gds, test.dc)).To(gomega.Equal(test.expected))
},
	ginkgo.Entry("nil nil", needsUpdateTest{
		gds:      nil,
		dc:       nil,
		expected: false,
	}),
	ginkgo.Entry("nil gds not nil dc", needsUpdateTest{
		gds:      nil,
		dc:       &styra.DatasourceConfig{},
		expected: false,
	}),
	ginkgo.Entry("nil dc not nil gds", needsUpdateTest{
		gds:      &styrav1alpha1.GlobalDatasource{},
		dc:       nil,
		expected: true,
	}),
	ginkgo.Entry("not nil", needsUpdateTest{
		gds: &styrav1alpha1.GlobalDatasource{},
		dc: &styra.DatasourceConfig{
			Enabled: true,
		},
		expected: false,
	}),
	ginkgo.Entry("name but no credentials", needsUpdateTest{
		gds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
		},
		dc: &styra.DatasourceConfig{
			Enabled: true,
		},
		expected: false,
	}),
	ginkgo.Entry("git credentials in secret", needsUpdateTest{
		gds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: styrav1alpha1.GlobalDatasourceSpec{
				Name:                 "test",
				CredentialsSecretRef: &styrav1alpha1.GlobalDatasourceSecretRef{},
			},
		},
		dc: &styra.DatasourceConfig{
			Enabled: true,
		},
		expected: true,
	}),
	ginkgo.Entry("git credentials in secret in sync", needsUpdateTest{
		gds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: styrav1alpha1.GlobalDatasourceSpec{
				Name:                 "test",
				CredentialsSecretRef: &styrav1alpha1.GlobalDatasourceSecretRef{},
			},
		},
		dc: &styra.DatasourceConfig{
			Credentials: "libraries/global/test/git",
			Enabled:     true,
		},
		expected: false,
	}),
	ginkgo.Entry("git credentials from default", needsUpdateTest{
		cfg: &configv2alpha2.ProjectConfig{
			GitCredentials: []*configv2alpha2.GitCredential{
				{User: "test-user", Password: "test-pw"},
			},
		},
		gds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: styrav1alpha1.GlobalDatasourceSpec{
				Name:                 "test",
				CredentialsSecretRef: &styrav1alpha1.GlobalDatasourceSecretRef{},
			},
		},
		dc: &styra.DatasourceConfig{
			Enabled: true,
		},
		expected: true,
	}),
	ginkgo.Entry("git credentials from default in sync", needsUpdateTest{
		cfg: &configv2alpha2.ProjectConfig{
			GitCredentials: []*configv2alpha2.GitCredential{
				{User: "test-user", Password: "test-pw"},
			},
		},
		gds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec:       styrav1alpha1.GlobalDatasourceSpec{Name: "test"},
		},
		dc: &styra.DatasourceConfig{
			Credentials: "libraries/global/test/git",
			Enabled:     true,
		},
		expected: false,
	}),
	ginkgo.Entry("everything out of sync", needsUpdateTest{
		gds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: styrav1alpha1.GlobalDatasourceSpec{
				Name:                 "test",
				Category:             "test-category",
				Description:          "test-description",
				Enabled:              ptr.Bool(false),
				Commit:               "test-commit",
				Reference:            "test-reference",
				URL:                  "test-url",
				Path:                 "test-path",
				CredentialsSecretRef: &styrav1alpha1.GlobalDatasourceSecretRef{},
			},
		},
		dc:       &styra.DatasourceConfig{},
		expected: true,
	}),
	ginkgo.Entry("everything in sync", needsUpdateTest{
		gds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: styrav1alpha1.GlobalDatasourceSpec{
				Name:                 "test",
				Category:             "test-category",
				Description:          "test-description",
				Enabled:              ptr.Bool(true),
				Commit:               "test-commit",
				Reference:            "test-reference",
				URL:                  "test-url",
				Path:                 "test-path",
				CredentialsSecretRef: &styrav1alpha1.GlobalDatasourceSecretRef{},
			},
		},
		dc: &styra.DatasourceConfig{
			Category:    "test-category",
			Description: "test-description",
			Enabled:     true,
			Commit:      "test-commit",
			Reference:   "test-reference",
			URL:         "test-url",
			Path:        "test-path",
			Credentials: "libraries/global/test/git",
		},
		expected: false,
	}),
)
