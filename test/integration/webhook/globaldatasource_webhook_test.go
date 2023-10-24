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

package webhook

import (
	"context"
	"errors"

	"github.com/google/uuid"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/bankdata/styra-controller/api/styra/v1alpha1"
)

func newGlobalDatasource() *v1alpha1.GlobalDatasource {
	return &v1alpha1.GlobalDatasource{
		ObjectMeta: metav1.ObjectMeta{Name: uuid.NewString()},
		Spec: v1alpha1.GlobalDatasourceSpec{
			Category: v1alpha1.GlobalDatasourceCategoryGitRego,
			URL:      "http://test.com/test.git",
		},
	}
}

var _ = ginkgo.Describe("GlobalDatasource", ginkgo.Label("integration"), func() {

	ginkgo.Describe("Default", ginkgo.Label("integration"), func() {
		ginkgo.It("should set the enabled true", func() {
			gds := newGlobalDatasource()
			ctx := context.Background()
			gomega.Ω(k8sClient.Create(ctx, gds)).To(gomega.Succeed())
			gomega.Ω(gds.Spec.Enabled).NotTo(gomega.BeNil())
			gomega.Ω(*gds.Spec.Enabled).To(gomega.BeTrue())
			gomega.Ω(k8sClient.Delete(ctx, gds)).To(gomega.Succeed())
		})

		ginkgo.It("sets default reference", func() {
			gds := newGlobalDatasource()
			ctx := context.Background()
			gomega.Ω(k8sClient.Create(ctx, gds)).To(gomega.Succeed())
			gomega.Ω(gds.Spec.Reference).To(gomega.Equal("refs/heads/master"))
			gomega.Ω(k8sClient.Delete(ctx, gds)).To(gomega.Succeed())
		})

		ginkgo.It("doesn't set default reference if commit is set", func() {
			gds := newGlobalDatasource()
			ctx := context.Background()
			gds.Spec.Commit = "test"
			gomega.Ω(k8sClient.Create(ctx, gds)).To(gomega.Succeed())
			gomega.Ω(gds.Spec.Reference).To(gomega.Equal(""))
			gomega.Ω(k8sClient.Delete(ctx, gds)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("Validate", func() {
		ginkgo.Context("category is git/rego", func() {
			ginkgo.It("enforces that commit/reference is mutual exclusive", func() {
				gds := newGlobalDatasource()
				ctx := context.Background()
				gds.Spec.Reference = "test-reference"
				gds.Spec.Commit = "test-commit"
				err := k8sClient.Create(ctx, gds)
				gomega.Ω(err).To(gomega.HaveOccurred())
				var sErr *apierrors.StatusError
				gomega.Ω(errors.As(err, &sErr)).To(gomega.BeTrue())
				expErrs := field.ErrorList{
					field.Invalid(
						field.NewPath("spec").Child("reference"),
						"test-reference",
						"reference can not be set when commit is specified",
					),
				}
				causes := sErr.ErrStatus.Details.Causes
				gomega.Ω(len(causes)).To(gomega.Equal(len(expErrs)))
				for i, expErr := range expErrs {
					gomega.Ω(string(causes[i].Type)).To(gomega.Equal(string(expErr.Type)))
					gomega.Ω(causes[i].Message).To(gomega.Equal(expErr.ErrorBody()))
					gomega.Ω(causes[i].Field).To(gomega.Equal(expErr.Field))
				}
			})

			ginkgo.It("requires url to be set", func() {
				gds := newGlobalDatasource()
				ctx := context.Background()
				gds.Spec.URL = ""
				err := k8sClient.Create(ctx, gds)
				gomega.Ω(err).To(gomega.HaveOccurred())
				var sErr *apierrors.StatusError
				gomega.Ω(errors.As(err, &sErr)).To(gomega.BeTrue())
				expErrs := field.ErrorList{
					field.Required(
						field.NewPath("spec").Child("url"),
						"category git/rego requires url",
					),
				}
				causes := sErr.ErrStatus.Details.Causes
				gomega.Ω(len(causes)).To(gomega.Equal(len(expErrs)))
				for i, expErr := range expErrs {
					gomega.Ω(string(causes[i].Type)).To(gomega.Equal(string(expErr.Type)))
					gomega.Ω(causes[i].Message).To(gomega.Equal(expErr.ErrorBody()))
					gomega.Ω(causes[i].Field).To(gomega.Equal(expErr.Field))
				}
			})
		})
	})
})
