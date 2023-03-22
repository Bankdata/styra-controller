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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("GlobalDatasource", Label("integration"), func() {

	Describe("Default", Label("integration"), func() {
		It("should set deletionProtection", func() {
			gds := newGlobalDatasource()
			ctx := context.Background()
			Ω(k8sClient.Create(ctx, gds)).To(Succeed())
			Ω(gds.Spec.DeletionProtection).NotTo(BeNil())
			Ω(*gds.Spec.DeletionProtection).To(BeTrue())
			Ω(k8sClient.Delete(ctx, gds)).To(Succeed())
		})

		It("should set the enabled true", func() {
			gds := newGlobalDatasource()
			ctx := context.Background()
			Ω(k8sClient.Create(ctx, gds)).To(Succeed())
			Ω(gds.Spec.Enabled).NotTo(BeNil())
			Ω(*gds.Spec.Enabled).To(BeTrue())
			Ω(k8sClient.Delete(ctx, gds)).To(Succeed())
		})

		It("sets default reference", func() {
			gds := newGlobalDatasource()
			ctx := context.Background()
			Ω(k8sClient.Create(ctx, gds)).To(Succeed())
			Ω(gds.Spec.Reference).To(Equal("refs/heads/master"))
			Ω(k8sClient.Delete(ctx, gds)).To(Succeed())
		})

		It("doesn't set default reference if commit is set", func() {
			gds := newGlobalDatasource()
			ctx := context.Background()
			gds.Spec.Commit = "test"
			Ω(k8sClient.Create(ctx, gds)).To(Succeed())
			Ω(gds.Spec.Reference).To(Equal(""))
			Ω(k8sClient.Delete(ctx, gds)).To(Succeed())
		})
	})

	Describe("Validate", func() {
		Context("category is git/rego", func() {
			It("enforces that commit/reference is mutual exclusive", func() {
				gds := newGlobalDatasource()
				ctx := context.Background()
				gds.Spec.Reference = "test-reference"
				gds.Spec.Commit = "test-commit"
				err := k8sClient.Create(ctx, gds)
				Ω(err).To(HaveOccurred())
				var sErr *apierrors.StatusError
				Ω(errors.As(err, &sErr)).To(BeTrue())
				expErrs := field.ErrorList{
					field.Invalid(
						field.NewPath("spec").Child("reference"),
						"test-reference",
						"reference can not be set when commit is specified",
					),
				}
				causes := sErr.ErrStatus.Details.Causes
				Ω(len(causes)).To(Equal(len(expErrs)))
				for i, expErr := range expErrs {
					Ω(string(causes[i].Type)).To(Equal(string(expErr.Type)))
					Ω(causes[i].Message).To(Equal(expErr.ErrorBody()))
					Ω(causes[i].Field).To(Equal(expErr.Field))
				}
			})

			It("requires url to be set", func() {
				gds := newGlobalDatasource()
				ctx := context.Background()
				gds.Spec.URL = ""
				err := k8sClient.Create(ctx, gds)
				Ω(err).To(HaveOccurred())
				var sErr *apierrors.StatusError
				Ω(errors.As(err, &sErr)).To(BeTrue())
				expErrs := field.ErrorList{
					field.Required(
						field.NewPath("spec").Child("url"),
						"category git/rego requires url",
					),
				}
				causes := sErr.ErrStatus.Details.Causes
				Ω(len(causes)).To(Equal(len(expErrs)))
				for i, expErr := range expErrs {
					Ω(string(causes[i].Type)).To(Equal(string(expErr.Type)))
					Ω(causes[i].Message).To(Equal(expErr.ErrorBody()))
					Ω(causes[i].Field).To(Equal(expErr.Field))
				}
			})
		})
	})
})
