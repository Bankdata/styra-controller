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
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/bankdata/styra-controller/api/styra/v1beta1"
	"github.com/bankdata/styra-controller/pkg/ptr"
)

var _ = Describe("System", Label("integration"), func() {
	var ss *v1beta1.System
	var key = types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: "default",
	}

	BeforeEach(func() {
		if !Label("integration").MatchesLabelFilter(GinkgoLabelFilter()) {
			return
		}

		ss = &v1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
		}

		Ω(k8sClient.Create(ctx, ss)).To(Succeed())
	})

	AfterEach(func() {
		if !Label("integration").MatchesLabelFilter(GinkgoLabelFilter()) {
			return
		}

		Ω(k8sClient.Delete(ctx, ss)).To(Succeed())
	})

	Describe("Default", func() {
		Describe("GitRepo.default", func() {
			It("should set defaults for missing values", func() {
				ss.Spec.SourceControl = &v1beta1.SourceControl{}

				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				Ω(ss.Spec.SourceControl).NotTo(BeNil())
				Ω(ss.Spec.SourceControl.Origin.Reference).To(Equal("refs/heads/master"))
			})

			It("should not set a default part 1", func() {
				ss.Spec.SourceControl = &v1beta1.SourceControl{
					Origin: v1beta1.GitRepo{
						Commit: "commitsha",
					},
				}

				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				Ω(ss.Spec.SourceControl).NotTo(BeNil())
				Ω(ss.Spec.SourceControl.Origin.Reference).To(Equal(""))
				Ω(ss.Spec.SourceControl.Origin.Commit).To(Equal("commitsha"))
			})

			It("should not set a default part 2", func() {
				ss.Spec.SourceControl = &v1beta1.SourceControl{
					Origin: v1beta1.GitRepo{
						Reference: "reference",
					},
				}

				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				Ω(ss.Spec.SourceControl).NotTo(BeNil())
				Ω(ss.Spec.SourceControl.Origin.Reference).To(Equal("reference"))
				Ω(ss.Spec.SourceControl.Origin.Commit).To(Equal(""))
			})

			It("should not set a default part 3", func() {
				ss.Spec.SourceControl = &v1beta1.SourceControl{
					Origin: v1beta1.GitRepo{
						Commit:    "commitsha",
						Reference: "reference",
					},
				}

				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				Ω(ss.Spec.SourceControl).NotTo(BeNil())
				Ω(ss.Spec.SourceControl.Origin.Reference).To(Equal("reference"))
				Ω(ss.Spec.SourceControl.Origin.Commit).To(Equal("commitsha"))
			})
		})
	})

	Describe("Validate", func() {
		Describe("SystemSpec.validateDecisionMappingNames", func() {
			It("should validate that names are mutually exclusive", func() {
				By("providing unique names we dont get an error")
				ss.Spec.DecisionMappings = []v1beta1.DecisionMapping{
					{},
					{Name: "test1"},
					{Name: "test2"},
				}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())

				By("having the same names we get errors")
				ss.Spec.DecisionMappings = []v1beta1.DecisionMapping{
					{},
					{},
					{Name: "test"},
					{Name: "test1"},
					{Name: "test"},
					{Name: "test2"},
				}
				err := k8sClient.Update(ctx, ss)
				Ω(err).To(HaveOccurred())
				var sErr *apierrors.StatusError
				Ω(errors.As(err, &sErr)).To(BeTrue())
				path := field.NewPath("spec").Child("decisionMappings")
				expErrs := field.ErrorList{
					field.Duplicate(path.Index(0).Child("name"), ""),
					field.Duplicate(path.Index(1).Child("name"), ""),
					field.Duplicate(path.Index(2).Child("name"), "test"),
					field.Duplicate(path.Index(4).Child("name"), "test"),
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

		Describe("DecisionMapping.validateColumnKeys", func() {
			It("should validate that keys are mutually exclusive", func() {
				By("providing unique keys we dont get an error")
				ss.Spec.DecisionMappings = []v1beta1.DecisionMapping{
					{
						Columns: []v1beta1.ColumnMapping{
							{},
							{Key: "test1"},
							{Key: "test2"},
						},
					},
				}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())

				By("reusing keys we get errors")
				ss.Spec.DecisionMappings = []v1beta1.DecisionMapping{
					{
						Columns: []v1beta1.ColumnMapping{
							{},
							{},
							{Key: "test"},
							{Key: "test1"},
							{Key: "test"},
							{Key: "test2"},
						},
					},
				}
				err := k8sClient.Update(ctx, ss)
				Ω(err).To(HaveOccurred())
				var sErr *apierrors.StatusError
				Ω(errors.As(err, &sErr)).To(BeTrue())
				path := field.NewPath("spec").Child("decisionMappings").
					Index(0).Child("columns")
				expErrs := field.ErrorList{
					field.Duplicate(path.Index(0).Child("key"), ""),
					field.Duplicate(path.Index(1).Child("key"), ""),
					field.Duplicate(path.Index(2).Child("key"), "test"),
					field.Duplicate(path.Index(4).Child("key"), "test"),
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

		Describe("Expected.validate", func() {
			It("should ensure mutual exclusivity of fields", func() {
				By("not causing validation violations")
				ss.Spec.DecisionMappings = []v1beta1.DecisionMapping{}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				ss.Spec.DecisionMappings = []v1beta1.DecisionMapping{{}}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{Expected: &v1beta1.Expected{}}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{Expected: &v1beta1.Expected{
					Boolean: ptr.Bool(true),
				}}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{Expected: &v1beta1.Expected{
					String: ptr.String("test"),
				}}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{Expected: &v1beta1.Expected{
					Integer: ptr.Int(0),
				}}
				Ω(k8sClient.Update(ctx, ss)).To(Succeed())

				By("setting all fields")
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{Expected: &v1beta1.Expected{
					Boolean: ptr.Bool(true),
					String:  ptr.String("test"),
					Integer: ptr.Int(0),
				}}
				err := k8sClient.Update(ctx, ss)
				Ω(err).To(HaveOccurred())
				var sErr *apierrors.StatusError
				Ω(errors.As(err, &sErr)).To(BeTrue())
				path := field.NewPath("spec").Child("decisionMappings").
					Index(0).Child("allowed").Child("expected")
				msg := "only one of boolean, string or int should be set"
				expErrs := field.ErrorList{
					field.Forbidden(path.Child("boolean"), msg),
					field.Forbidden(path.Child("integer"), msg),
					field.Forbidden(path.Child("string"), msg),
				}
				causes := sErr.ErrStatus.Details.Causes
				Ω(len(causes)).To(Equal(len(expErrs)))
				for i, expErr := range expErrs {
					Ω(string(causes[i].Type)).To(Equal(string(expErr.Type)))
					Ω(causes[i].Message).To(Equal(expErr.ErrorBody()))
					Ω(causes[i].Field).To(Equal(expErr.Field))
				}

				By("setting boolean and string")
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{Expected: &v1beta1.Expected{
					Boolean: ptr.Bool(true),
					String:  ptr.String("test"),
				}}
				err = k8sClient.Update(ctx, ss)
				Ω(err).To(HaveOccurred())
				Ω(errors.As(err, &sErr)).To(BeTrue())
				expErrs = field.ErrorList{
					field.Forbidden(path.Child("boolean"), msg),
					field.Forbidden(path.Child("string"), msg),
				}
				causes = sErr.ErrStatus.Details.Causes
				Ω(len(causes)).To(Equal(len(expErrs)))
				for i, expErr := range expErrs {
					Ω(string(causes[i].Type)).To(Equal(string(expErr.Type)))
					Ω(causes[i].Message).To(Equal(expErr.ErrorBody()))
					Ω(causes[i].Field).To(Equal(expErr.Field))
				}

				By("setting boolean and integer")
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{Expected: &v1beta1.Expected{
					Boolean: ptr.Bool(true),
					Integer: ptr.Int(0),
				}}
				err = k8sClient.Update(ctx, ss)
				Ω(err).To(HaveOccurred())
				Ω(errors.As(err, &sErr)).To(BeTrue())
				expErrs = field.ErrorList{
					field.Forbidden(path.Child("boolean"), msg),
					field.Forbidden(path.Child("integer"), msg),
				}
				causes = sErr.ErrStatus.Details.Causes
				Ω(len(causes)).To(Equal(len(expErrs)))
				for i, expErr := range expErrs {
					Ω(string(causes[i].Type)).To(Equal(string(expErr.Type)))
					Ω(causes[i].Message).To(Equal(expErr.ErrorBody()))
					Ω(causes[i].Field).To(Equal(expErr.Field))
				}

				By("setting string and integer")
				ss.Spec.DecisionMappings[0].Allowed = &v1beta1.AllowedMapping{Expected: &v1beta1.Expected{
					String:  ptr.String("test"),
					Integer: ptr.Int(0),
				}}
				err = k8sClient.Update(ctx, ss)
				Ω(err).To(HaveOccurred())
				Ω(errors.As(err, &sErr)).To(BeTrue())
				expErrs = field.ErrorList{
					field.Forbidden(path.Child("integer"), msg),
					field.Forbidden(path.Child("string"), msg),
				}
				causes = sErr.ErrStatus.Details.Causes
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
