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
	"context"
	"net/http"
	"path"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("LibraryReconciler.Reconcile", ginkgo.Label("integration"), func() {
	ginkgo.It("reconciles Library", func() {
		key := types.NamespacedName{Name: "uniquename", Namespace: "default"}
		toCreate := &styrav1alpha1.Library{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: styrav1alpha1.LibrarySpec{
				Name: key.Name,
				Subjects: []styrav1alpha1.LibrarySubject{
					{
						Kind: styrav1alpha1.LibrarySubjectKindUser,
						Name: "user1@mail.com",
					},
					{
						// kind
						Name: "user2@mail.com",
					},
					{
						Kind: styrav1alpha1.LibrarySubjectKindGroup,
						Name: "testGroup",
					},
				},
				SourceControl: &styrav1alpha1.SourceControl{
					LibraryOrigin: &styrav1alpha1.GitRepo{
						Path:      "libraries/library",
						Reference: "refs/heads/master",
						Commit:    "commit-sha",
						URL:       "github.com",
					},
				},
				Description: "description",
				Datasources: []styrav1alpha1.LibraryDatasource{
					{
						Path:        "oidc/sandbox",
						Description: "desc",
					},
					{
						Path:        "oidc/sandbox/2",
						Description: "desc2",
					},
				},
			},
		}
		ctx := context.Background()
		ginkgo.By("creating the Library")

		styraClientMock.On("CreateUpdateSecret",
			mock.Anything,
			path.Join("libraries", key.Name, "git"),
			&styra.CreateUpdateSecretsRequest{
				Name:   "test-user",
				Secret: "test-secret",
			},
		).Return(&styra.CreateUpdateSecretResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("GetLibrary", mock.Anything, key.Name).
			Return(nil, &styra.HTTPError{StatusCode: http.StatusNotFound}).Once()

		defaultSourceControl := &styra.LibrarySourceControlConfig{
			LibraryOrigin: &styra.LibraryGitRepoConfig{
				Commit:      "commit-sha",
				Credentials: path.Join("libraries", key.Name, "git"),
				Path:        "libraries/library",
				Reference:   "refs/heads/master",
				URL:         "github.com",
			},
		}

		styraClientMock.On("UpsertLibrary", mock.Anything, key.Name, &styra.UpsertLibraryRequest{
			Description:   "description",
			ReadOnly:      true,
			SourceControl: defaultSourceControl,
		}).Return(&styra.UpsertLibraryResponse{}, nil).Once()

		// New reconciliation started

		styraClientMock.On("CreateUpdateSecret",
			mock.Anything,
			path.Join("libraries", key.Name, "git"),
			&styra.CreateUpdateSecretsRequest{
				Name:   "test-user",
				Secret: "test-secret",
			},
		).Return(&styra.CreateUpdateSecretResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("GetLibrary", mock.Anything, key.Name).Return(&styra.GetLibraryResponse{
			Statuscode: http.StatusOK,
			LibraryEntityExpanded: &styra.LibraryEntityExpanded{
				DataSources:   []styra.LibraryDatasourceConfig{},
				Description:   "description",
				ID:            key.Name,
				ReadOnly:      true,
				SourceControl: defaultSourceControl,
			},
		}, nil).Once()

		styraClientMock.On("UpsertDatasource", mock.Anything, path.Join("libraries", key.Name, "oidc/sandbox"),
			&styra.UpsertDatasourceRequest{
				Category: "rest",
				Enabled:  true,
			}).Return(&styra.UpsertDatasourceResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		webhookMock.On(
			"LibraryDatasourceChanged",
			mock.Anything,
			mock.Anything,
			path.Join("libraries", key.Name, "oidc/sandbox"),
		).Return(nil).Once()

		styraClientMock.On("UpsertDatasource", mock.Anything, path.Join("libraries", key.Name, "oidc/sandbox/2"),
			&styra.UpsertDatasourceRequest{
				Category: "rest",
				Enabled:  true,
			}).Return(&styra.UpsertDatasourceResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		webhookMock.On(
			"LibraryDatasourceChanged",
			mock.Anything,
			mock.Anything,
			path.Join("libraries", key.Name, "oidc/sandbox/2"),
		).Return(nil).Once()

		// createUsersIfMissing:
		styraClientMock.On("GetUser", mock.Anything, "user1@mail.com").
			Return(&styra.GetUserResponse{
				StatusCode: http.StatusNotFound,
			}, nil).Once()

		styraClientMock.On("CreateInvitation", mock.Anything, false, "user1@mail.com").
			Return(&styra.CreateInvitationResponse{}, nil).Once()

		styraClientMock.On("GetUser", mock.Anything, "user2@mail.com").
			Return(&styra.GetUserResponse{
				StatusCode: http.StatusOK,
			}, nil).Once()

		// deleteIncorrectRoleBindings:
		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindLibrary,
			ResourceID:   toCreate.Spec.Name,
		}).Return(&styra.ListRoleBindingsV2Response{
			Rolebindings: []*styra.RoleBindingConfig{},
		}, nil).Once() //Test deletions also?

		// createRoleBindingIfMissing:
		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindLibrary,
			ResourceID:   toCreate.Spec.Name,
		}).Return(&styra.ListRoleBindingsV2Response{
			Rolebindings: []*styra.RoleBindingConfig{},
		}, nil).Once()

		styraClientMock.On("CreateRoleBinding", mock.Anything, &styra.CreateRoleBindingRequest{
			ResourceFilter: &styra.ResourceFilter{
				ID:   toCreate.Spec.Name,
				Kind: styra.RoleBindingKindLibrary,
			},
			RoleID: styra.RoleLibraryViewer,
			Subjects: []*styra.Subject{
				{
					ID:   "user1@mail.com",
					Kind: styra.SubjectKindUser,
				}, {
					ID:   "user2@mail.com",
					Kind: styra.SubjectKindUser,
				}, {
					Kind: styra.SubjectKindClaim,
					ClaimConfig: &styra.ClaimConfig{
						IdentityProvider: "AzureAD Bankdata",
						Key:              "groups",
						Value:            "testGroup",
					},
				},
			},
		}).Return(&styra.CreateRoleBindingResponse{}, nil).Once()

		// updateRoleBindingIfNeeded:
		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindLibrary,
			ResourceID:   toCreate.Spec.Name,
		}).Return(&styra.ListRoleBindingsV2Response{
			Rolebindings: []*styra.RoleBindingConfig{
				{
					Subjects: []*styra.Subject{
						{
							ID:   "user1@mail.com",
							Kind: styra.SubjectKindUser,
						}, {
							ID:   "user2@mail.com",
							Kind: styra.SubjectKindUser,
						}, {
							Kind: styra.SubjectKindClaim,
							ClaimConfig: &styra.ClaimConfig{
								IdentityProvider: "AzureAD Bankdata",
								Key:              "groups",
								Value:            "testGroup",
							},
						},
					},
					RoleID: "LibraryViewer",
				},
			},
		}, nil).Once()

		gomega.Ω(k8sClient.Create(ctx, toCreate)).
			To(gomega.Succeed())

		gomega.Eventually(func() bool {
			var k8sLib styrav1alpha1.Library
			if err := k8sClient.Get(ctx, key, &k8sLib); err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			var (
				createUpdateSecret int
				getLibrary         int
				upsertLibrary      int
				upsertDatasource   int
				datasourceChanged  int
				getUser            int
				createInvitation   int
				createRoleBinding  int
				listRoleBindings   int
			)

			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "CreateUpdateSecret":
					createUpdateSecret++
				case "GetLibrary":
					getLibrary++
				case "UpsertLibrary":
					upsertLibrary++
				case "UpsertDatasource":
					upsertDatasource++
				case "GetUser":
					getUser++
				case "CreateInvitation":
					createInvitation++
				case "CreateRoleBinding":
					createRoleBinding++
				case "ListRoleBindingsV2":
					listRoleBindings++
				}
			}

			for _, call := range webhookMock.Calls {
				switch call.Method {
				case "LibraryDatasourceChanged":
					datasourceChanged++
				}
			}

			return createUpdateSecret == 2 &&
				getLibrary == 2 &&
				upsertLibrary == 1 &&
				upsertDatasource == 2 &&
				datasourceChanged == 2 &&
				getUser == 2 &&
				createInvitation == 1 &&
				listRoleBindings == 3 &&
				createRoleBinding == 1
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&webhookMock.Mock)
		resetMock(&styraClientMock.Mock)

		styraClientMock.AssertExpectations(ginkgo.GinkgoT())

	})
})
