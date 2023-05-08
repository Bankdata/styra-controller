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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv2alpha1 "github.com/bankdata/styra-controller/api/config/v2alpha1"
	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	"github.com/bankdata/styra-controller/pkg/ptr"
	"github.com/bankdata/styra-controller/pkg/styra"
)

type specToUpdateTest struct {
	cfg      *configv2alpha1.ProjectConfig
	ds       *styrav1alpha1.GlobalDatasource
	expected *styra.UpsertDatasourceRequest
}

var _ = DescribeTable("globalDatasourceSpecToUpdate", func(test specToUpdateTest) {
	cfg := test.cfg
	if cfg == nil {
		cfg = &configv2alpha1.ProjectConfig{}
	}
	r := &GlobalDatasourceReconciler{Config: cfg}
	Expect(r.specToUpdate(test.ds)).To(Equal(test.expected))
},
	Entry("ds is nil", specToUpdateTest{
		ds:       nil,
		expected: nil,
	}),

	Entry("zero value", specToUpdateTest{
		ds: &styrav1alpha1.GlobalDatasource{},
		expected: &styra.UpsertDatasourceRequest{
			Enabled: true,
		},
	}),

	Entry("using default git credentials", specToUpdateTest{
		cfg: &configv2alpha1.ProjectConfig{
			GitCredentials: []*configv2alpha1.GitCredential{
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

	Entry("using credentials from secret", specToUpdateTest{
		cfg: &configv2alpha1.ProjectConfig{
			GitCredentials: []*configv2alpha1.GitCredential{
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

	Entry("setting all fields", specToUpdateTest{
		cfg: &configv2alpha1.ProjectConfig{
			GitCredentials: []*configv2alpha1.GitCredential{
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
	cfg      *configv2alpha1.ProjectConfig
	gds      *styrav1alpha1.GlobalDatasource
	dc       *styra.DatasourceConfig
	expected bool
}

var _ = DescribeTable("needsUpdate", func(test needsUpdateTest) {
	cfg := test.cfg
	if cfg == nil {
		cfg = &configv2alpha1.ProjectConfig{}
	}
	r := &GlobalDatasourceReconciler{Config: cfg}
	Expect(r.needsUpdate(test.gds, test.dc)).To(Equal(test.expected))
},
	Entry("nil nil", needsUpdateTest{
		gds:      nil,
		dc:       nil,
		expected: false,
	}),
	Entry("nil gds not nil dc", needsUpdateTest{
		gds:      nil,
		dc:       &styra.DatasourceConfig{},
		expected: false,
	}),
	Entry("nil dc not nil gds", needsUpdateTest{
		gds:      &styrav1alpha1.GlobalDatasource{},
		dc:       nil,
		expected: true,
	}),
	Entry("not nil", needsUpdateTest{
		gds: &styrav1alpha1.GlobalDatasource{},
		dc: &styra.DatasourceConfig{
			Enabled: true,
		},
		expected: false,
	}),
	Entry("name but no credentials", needsUpdateTest{
		gds: &styrav1alpha1.GlobalDatasource{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
		},
		dc: &styra.DatasourceConfig{
			Enabled: true,
		},
		expected: false,
	}),
	Entry("git credentials in secret", needsUpdateTest{
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
	Entry("git credentials in secret in sync", needsUpdateTest{
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
	Entry("git credentials from default", needsUpdateTest{
		cfg: &configv2alpha1.ProjectConfig{
			GitCredentials: []*configv2alpha1.GitCredential{
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
	Entry("git credentials from default in sync", needsUpdateTest{
		cfg: &configv2alpha1.ProjectConfig{
			GitCredentials: []*configv2alpha1.GitCredential{
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
	Entry("everything out of sync", needsUpdateTest{
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
	Entry("everything in sync", needsUpdateTest{
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
