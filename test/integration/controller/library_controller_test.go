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

package styra

import (
	"context"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	"github.com/bankdata/styra-controller/pkg/ocp"
)

var _ = ginkgo.Describe("LibraryReconciler.Reconcile", ginkgo.Label("integration"), func() {
	ginkgo.It("reconciles Library with OCP", func() {
		key := types.NamespacedName{Name: "uniquename", Namespace: "default"}
		toCreate := &styrav1alpha1.Library{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: styrav1alpha1.LibrarySpec{
				Name: key.Name,
				SourceControl: &styrav1alpha1.SourceControl{
					LibraryOrigin: &styrav1alpha1.GitRepo{
						Path:      "libraries/library",
						Reference: "refs/heads/master",
						Commit:    "commit-sha",
						URL:       "https://github.com/test/repo.git",
					},
				},
				Description: "description",
			},
		}
		ctx := context.Background()
		ginkgo.By("creating the Library")

		// Mock the OCP PutSource call for the library
		ocpClientMock.On("PutSource", mock.Anything, key.Name, &ocp.PutSourceRequest{
			Name: key.Name,
			Git: &ocp.GitConfig{
				Repo:          "https://github.com/test/repo.git",
				Commit:        "commit-sha",
				Reference:     "refs/heads/master",
				Path:          "libraries/library",
				CredentialID:  "github-credentials",
				IncludedFiles: []string{"*.rego"},
				ExcludedFiles: []string{"*_test.rego"},
			},
		}).Return(&ocp.PutSourceResponse{}, nil)

		gomega.Ω(k8sClient.Create(ctx, toCreate)).
			To(gomega.Succeed())

		gomega.Eventually(func() bool {
			var k8sLib styrav1alpha1.Library
			if err := k8sClient.Get(ctx, key, &k8sLib); err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(gomega.BeTrue())

		// Verify the OCP client was called as expected
		gomega.Eventually(func() bool {
			var putSourceCalls int
			for _, call := range ocpClientMock.Calls {
				if call.Method == "PutSource" && len(call.Arguments) >= 2 && call.Arguments.Get(1) == key.Name {
					putSourceCalls++
				}
			}
			return putSourceCalls == 1
		}, timeout, interval).Should(gomega.BeTrue())
		ocpClientMock.AssertExpectations(ginkgo.GinkgoT())

		resetMock(&ocpClientMock.Mock)

	})
})
