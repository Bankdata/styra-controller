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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
	"github.com/bankdata/styra-controller/internal/finalizer"
	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/bankdata/styra-controller/pkg/ocp"
	"github.com/bankdata/styra-controller/pkg/ptr"
	"github.com/bankdata/styra-controller/pkg/s3"
	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("SystemReconciler.Reconcile", ginkgo.Label("integration"), func() {
	ginkgo.It("should reconcile", func() {
		sourceControl := styrav1beta1.SourceControl{
			Origin: styrav1beta1.GitRepo{
				URL: "https://github.com/test/repo.git",
			},
		}

		spec := styrav1beta1.SystemSpec{
			DeletionProtection: ptr.Bool(false),
			SourceControl:      &sourceControl,
		}

		key := types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		}

		toCreate := &styrav1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}

		cfg := &styra.SystemConfig{
			ID:       "system_id",
			Name:     key.String(),
			ReadOnly: true,
		}

		ctx := context.Background()

		ginkgo.By("Creating the system")

		styraClientMock.On("GetSystemByName", mock.Anything, key.String()).Return(&styra.GetSystemResponse{
			StatusCode:   http.StatusOK,
			SystemConfig: nil,
		}, nil).Once()

		styraClientMock.On("CreateSystem", mock.Anything, mock.Anything).Return(&styra.CreateSystemResponse{
			StatusCode:   http.StatusOK,
			SystemConfig: cfg,
		}, nil).Once()

		styraClientMock.On("DeletePolicy", mock.Anything, "systems/system_id/rules").Return(&styra.DeletePolicyResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()
		styraClientMock.On("DeletePolicy", mock.Anything, "systems/system_id/test").Return(&styra.DeletePolicyResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("UpdateSystem", mock.Anything, "system_id", &styra.UpdateSystemRequest{
			SystemConfig: &styra.SystemConfig{
				Name:           key.String(),
				Type:           "custom",
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						URL:         sourceControl.Origin.URL,
						Credentials: "systems/system_id/git",
						Path:        "",
						Commit:      "",
						Reference:   "",
					},
				},
			},
		}).Return(&styra.UpdateSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				Name:           key.String(),
				Type:           "custom",
				ReadOnly:       true,
				ID:             cfg.ID,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						URL:         sourceControl.Origin.URL,
						Credentials: "systems/system_id/git",
						Path:        "",
						Commit:      "",
						Reference:   "",
					},
				},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("CreateRoleBinding", mock.Anything, &styra.CreateRoleBindingRequest{
			ResourceFilter: &styra.ResourceFilter{
				ID:   cfg.ID,
				Kind: styra.RoleBindingKindSystem,
			},
			RoleID:   styra.RoleSystemViewer,
			Subjects: []*styra.Subject{},
		}).Return(&styra.CreateRoleBindingResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// new reconcile as we create opatoken secret that we are watching
		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					ID:             "system_id",
					Name:           key.String(),
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
					SourceControl: &styra.SourceControlConfig{
						Origin: styra.GitRepoConfig{
							URL:         sourceControl.Origin.URL,
							Credentials: "systems/system_id/git",
							Path:        "",
							Commit:      "",
							Reference:   "",
						},
					},
				},
			}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// new reconcile as we create opa configmap that we are watching
		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					ID:             "system_id",
					Name:           key.String(),
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
					SourceControl: &styra.SourceControlConfig{
						Origin: styra.GitRepoConfig{
							URL:         sourceControl.Origin.URL,
							Credentials: "systems/system_id/git",
							Path:        "",
							Commit:      "",
							Reference:   "",
						},
					},
				},
			}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Expect(k8sClient.Create(ctx, toCreate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			if err := k8sClient.Get(ctx, key, fetched); err != nil {
				return false
			}

			return finalizer.IsSet(fetched) &&
				fetched.Status.ID != "" &&
				fetched.Status.Phase == styrav1beta1.SystemPhaseCreated &&
				fetched.Status.Ready
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			fetched := &corev1.Secret{}
			key := types.NamespacedName{Name: fmt.Sprintf("%s-opa-token", key.Name), Namespace: key.Namespace}
			return k8sClient.Get(ctx, key, fetched) == nil && string(fetched.Data["token"]) == "opa-token-123"
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			fetched := &corev1.ConfigMap{}
			var actualMap, expectedMap map[string]interface{}

			key := types.NamespacedName{Name: fmt.Sprintf("%s-opa", key.Name), Namespace: key.Namespace}
			if fetchSuceeded := k8sClient.Get(ctx, key, fetched) == nil; !fetchSuceeded {
				return false
			}

			actualYAML := fetched.Data["opa-conf.yaml"]
			expectedYAML := `services:
- name: styra
  url: styra-url-123
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
- name: styra-bundles
  url: styra-url-123/bundles
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
labels:
  system-id: system_id
  system-type: custom
discovery:
  name: discovery
  prefix: /systems/system_id
  service: styra
`

			if err := yaml.Unmarshal([]byte(actualYAML), &actualMap); err != nil {
				return false
			}
			if err := yaml.Unmarshal([]byte(expectedYAML), &expectedMap); err != nil {
				return false
			}

			return reflect.DeepEqual(expectedMap, actualMap)
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			var (
				getSystem          int
				getSystemByName    int
				createSystem       int
				updateSystem       int
				deletePolicy       int
				getUsers           int
				rolebindingsListed int
				createRoleBinding  int
				getOPAConfig       int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "CreateSystem":
					createSystem++
				case "UpdateSystem":
					updateSystem++
				case "GetSystemByName":
					getSystemByName++
				case "DeletePolicy":
					deletePolicy++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					rolebindingsListed++
				case "CreateRoleBinding":
					createRoleBinding++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 2 &&
				getSystemByName == 1 &&
				createSystem == 1 &&
				updateSystem == 1 &&
				deletePolicy == 2 &&
				getUsers == 3 &&
				rolebindingsListed == 3 &&
				createRoleBinding == 1 &&
				getOPAConfig == 3
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Setting a Local Plane in System")

		toUpdate := &styrav1beta1.System{}
		gomega.Expect(k8sClient.Get(ctx, key, toUpdate)).To(gomega.Succeed())

		toUpdate.Spec.LocalPlane = &styrav1beta1.LocalPlane{
			Name: "default_local_plane",
		}

		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					Name:           key.String(),
					Type:           "custom",
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
					SourceControl: &styra.SourceControlConfig{
						Origin: styra.GitRepoConfig{
							URL:         sourceControl.Origin.URL,
							Credentials: "systems/system_id/git",
							Path:        "",
							Commit:      "",
							Reference:   "",
						},
					},
				},
			}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// new reconcile as we create opa configmap that we watch
		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					Name:           key.String(),
					Type:           "custom",
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
					SourceControl: &styra.SourceControlConfig{
						Origin: styra.GitRepoConfig{
							URL:         sourceControl.Origin.URL,
							Credentials: "systems/system_id/git",
							Path:        "",
							Commit:      "",
							Reference:   "",
						},
					},
				},
			}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// new reconcile as we create slp configmap that we watch
		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					Name:           key.String(),
					Type:           "custom",
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
					SourceControl: &styra.SourceControlConfig{
						Origin: styra.GitRepoConfig{
							URL:         sourceControl.Origin.URL,
							Credentials: "systems/system_id/git",
							Path:        "",
							Commit:      "",
							Reference:   "",
						},
					},
				},
			}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Expect(k8sClient.Update(ctx, toUpdate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			//opa configmap
			fetched := &corev1.ConfigMap{}
			var actualMap, expectedMap map[string]interface{}

			key := types.NamespacedName{Name: fmt.Sprintf("%s-opa", key.Name), Namespace: key.Namespace}
			if fetchSuceeded := k8sClient.Get(ctx, key, fetched) == nil; !fetchSuceeded {
				return false
			}

			actualYAML := fetched.Data["opa-conf.yaml"]
			expectedYAML := `services:
- name: styra
  url: http://default_local_plane/v1
labels:
  system-id: system_id
  system-type: custom
discovery:
  name: discovery
  service: styra
`

			if err := yaml.Unmarshal([]byte(actualYAML), &actualMap); err != nil {
				return false
			}
			if err := yaml.Unmarshal([]byte(expectedYAML), &expectedMap); err != nil {
				return false
			}

			return reflect.DeepEqual(expectedMap, actualMap)
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			//slp configmap
			fetched := &corev1.ConfigMap{}
			key := types.NamespacedName{Name: fmt.Sprintf("%s-slp", key.Name), Namespace: key.Namespace}
			fetchSuceeded := k8sClient.Get(ctx, key, fetched) == nil
			expectedConfigMapContent := `services:
- name: styra
  url: styra-url-123
  credentials:
    bearer:
      token_path: /etc/slp/auth/token
labels:
  system-id: system_id
  system-type: custom
discovery:
  name: discovery
  resource: /systems/system_id/discovery
  service: styra
`
			return fetchSuceeded && fetched.Data["slp.yaml"] == expectedConfigMapContent
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			var (
				getSystem          int
				getUsers           int
				rolebindingsListed int
				getOPAConfig       int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					rolebindingsListed++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 3 &&
				rolebindingsListed == 3 &&
				getOPAConfig == 3
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Updating decision mappings")

		//We set status.conditions as toUpdate is not updated with the
		//new conditions unless we do this get. So to ensure that
		//toUpdate is fully updated we do this get.
		gomega.Expect(k8sClient.Get(ctx, key, toUpdate)).To(gomega.Succeed())

		toUpdate.Spec.DecisionMappings = []styrav1beta1.DecisionMapping{
			{},
			{
				Name: "test",
				Allowed: &styrav1beta1.AllowedMapping{
					Expected: &styrav1beta1.Expected{
						Boolean: ptr.Bool(true),
					},
					Path: "test",
				},
				Columns: []styrav1beta1.ColumnMapping{
					{Key: "test", Path: "test"},
				},
				Reason: styrav1beta1.ReasonMapping{
					Path: "reason",
				},
			},
		}

		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					Name:     key.String(),
					Type:     "custom",
					ReadOnly: true,
				},
			}, nil).Once()

		styraClientMock.On("UpdateSystem", mock.Anything, "system_id", &styra.UpdateSystemRequest{
			SystemConfig: &styra.SystemConfig{
				Name:           key.String(),
				Type:           "custom",
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						URL:         sourceControl.Origin.URL,
						Credentials: "systems/system_id/git",
						Path:        "",
						Commit:      "",
						Reference:   "",
					},
				},
				DecisionMappings: map[string]styra.DecisionMapping{
					"": {},
					"test": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Path:     "test",
						},
						Columns: []styra.DecisionMappingColumn{
							{Key: "test", Path: "test"},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "reason",
						},
					},
				},
			},
		}).Return(&styra.UpdateSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             "system_id",
				Name:           key.String(),
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						URL:         sourceControl.Origin.URL,
						Credentials: "systems/system_id/git",
						Path:        "",
						Commit:      "",
						Reference:   "",
					},
				},
				DecisionMappings: map[string]styra.DecisionMapping{
					"": {},
					"test": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Path:     "test",
						},
						Columns: []styra.DecisionMappingColumn{
							{Key: "test", Path: "test"},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "reason",
						},
					},
				},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Expect(k8sClient.Update(ctx, toUpdate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			var (
				getSystem          int
				updateSystem       int
				getUsers           int
				listRolebindingsV2 int
				getOPAConfig       int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "UpdateSystem":
					updateSystem++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					listRolebindingsV2++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}

			return getSystem == 1 &&
				updateSystem == 1 &&
				getUsers == 1 &&
				listRolebindingsV2 == 1 &&
				getOPAConfig == 1
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Adding users")

		gomega.Expect(k8sClient.Get(ctx, key, toUpdate)).To(gomega.Succeed())

		toUpdate.Spec.Subjects = []styrav1beta1.Subject{
			{Kind: styrav1beta1.SubjectKindUser, Name: "test1@test.com"},
			{Kind: styrav1beta1.SubjectKindUser, Name: "test2@test.com"},
		}

		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(&styra.GetSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             "system_id",
				Name:           key.String(),
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						URL:         sourceControl.Origin.URL,
						Credentials: "systems/system_id/git",
						Path:        "",
						Commit:      "",
						Reference:   "",
					},
				},
				DecisionMappings: map[string]styra.DecisionMapping{
					"": {},
					"test": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Path:     "test",
						},
						Columns: []styra.DecisionMappingColumn{
							{Key: "test", Path: "test"},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "reason",
						},
					},
				},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test3@test.com", Enabled: true},
				{ID: "test4@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("CreateInvitation", mock.Anything, false, "test1@test.com").
			Return(&styra.CreateInvitationResponse{
				StatusCode: http.StatusOK,
			}, nil).Once()
		styraClientMock.On("InvalidateCache", mock.Anything).Return(nil).Once()

		styraClientMock.On("CreateInvitation", mock.Anything, false, "test2@test.com").
			Return(&styra.CreateInvitationResponse{
				StatusCode: http.StatusOK,
			}, nil).Once()
		styraClientMock.On("InvalidateCache", mock.Anything).Return(nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("UpdateRoleBindingSubjects", mock.Anything, "1", &styra.UpdateRoleBindingSubjectsRequest{
			Subjects: []*styra.Subject{
				{ID: "test1@test.com", Kind: styra.SubjectKindUser},
				{ID: "test2@test.com", Kind: styra.SubjectKindUser},
			},
		}).Return(&styra.UpdateRoleBindingSubjectsResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Expect(k8sClient.Update(ctx, toUpdate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			var (
				getSystem                 int
				getUsers                  int
				createInvitation          int
				invalidateCache           int
				listRolebindingsV2        int
				updateRoleBindingSubjects int
				getOPAConfig              int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "GetUsers":
					getUsers++
				case "CreateInvitation":
					createInvitation++
				case "InvalidateCache":
					invalidateCache++
				case "ListRoleBindingsV2":
					listRolebindingsV2++
				case "UpdateRoleBindingSubjects":
					updateRoleBindingSubjects++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 1 &&
				getUsers == 1 &&
				createInvitation == 2 &&
				invalidateCache == 2 &&
				listRolebindingsV2 == 1 &&
				updateRoleBindingSubjects == 1 &&
				getOPAConfig == 1
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Adding groups")
		gomega.Expect(k8sClient.Get(ctx, key, toUpdate)).To(gomega.Succeed())

		toUpdate.Spec.Subjects = []styrav1beta1.Subject{
			{Kind: styrav1beta1.SubjectKindUser, Name: "test1@test.com"},
			{Kind: styrav1beta1.SubjectKindUser, Name: "test2@test.com"},
			{Kind: styrav1beta1.SubjectKindGroup, Name: "Group1"},
			{Kind: styrav1beta1.SubjectKindGroup, Name: "Group2"},
		}

		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(&styra.GetSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             "system_id",
				Name:           key.String(),
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						URL:         sourceControl.Origin.URL,
						Credentials: "systems/system_id/git",
						Path:        "",
						Commit:      "",
						Reference:   "",
					},
				},
				DecisionMappings: map[string]styra.DecisionMapping{
					"": {},
					"test": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Path:     "test",
						},
						Columns: []styra.DecisionMappingColumn{
							{Key: "test", Path: "test"},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "reason",
						},
					},
				},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode: http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer,
				Subjects: []*styra.Subject{
					{ID: "test1@test.com", Kind: styra.SubjectKindUser},
					{ID: "test2@test.com", Kind: styra.SubjectKindUser},
				}}},
		}, nil).Once()

		styraClientMock.On("UpdateRoleBindingSubjects", mock.Anything, "1", &styra.UpdateRoleBindingSubjectsRequest{
			Subjects: []*styra.Subject{
				{ID: "test1@test.com", Kind: styra.SubjectKindUser},
				{ID: "test2@test.com", Kind: styra.SubjectKindUser},
				{
					Kind:        styra.SubjectKindClaim,
					ClaimConfig: &styra.ClaimConfig{IdentityProvider: "AzureAD Bankdata", Key: "groups", Value: "Group1"},
				},
				{
					Kind:        styra.SubjectKindClaim,
					ClaimConfig: &styra.ClaimConfig{IdentityProvider: "AzureAD Bankdata", Key: "groups", Value: "Group2"},
				},
			},
		}).Return(&styra.UpdateRoleBindingSubjectsResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Expect(k8sClient.Update(ctx, toUpdate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			var (
				getSystem                 int
				getUsers                  int
				listRolebindingsV2        int
				updateRoleBindingSubjects int
				getOPAConfig              int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					listRolebindingsV2++
				case "UpdateRoleBindingSubjects":
					updateRoleBindingSubjects++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 1 &&
				getUsers == 1 &&
				listRolebindingsV2 == 1 &&
				updateRoleBindingSubjects == 1 &&
				getOPAConfig == 1
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Removing subjects and reconciling a user with excess priviledges")

		gomega.Expect(k8sClient.Get(ctx, key, toUpdate)).To(gomega.Succeed())

		toUpdate.Spec.Subjects = []styrav1beta1.Subject{}

		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(&styra.GetSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				ID:             "system_id",
				Name:           key.String(),
				ReadOnly:       true,
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						URL:         sourceControl.Origin.URL,
						Credentials: "systems/system_id/git",
						Path:        "",
						Commit:      "",
						Reference:   "",
					},
				},
				DecisionMappings: map[string]styra.DecisionMapping{
					"": {},
					"test": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Path:     "test",
						},
						Columns: []styra.DecisionMappingColumn{
							{Key: "test", Path: "test"},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "reason",
						},
					},
				},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			Rolebindings: []*styra.RoleBindingConfig{
				{
					ID:     "1",
					RoleID: styra.RoleSystemViewer,
					Subjects: []*styra.Subject{
						{ID: "test", Kind: "test"},
						{ID: "test1@test.com", Kind: styra.SubjectKindUser},
						{ID: "test2@test.com", Kind: styra.SubjectKindUser},
						{
							Kind:        styra.SubjectKindClaim,
							ClaimConfig: &styra.ClaimConfig{IdentityProvider: "AzureAD Bankdata", Key: "groups", Value: "Group1"},
						},
						{
							Kind:        styra.SubjectKindClaim,
							ClaimConfig: &styra.ClaimConfig{IdentityProvider: "AzureAD Bankdata", Key: "groups", Value: "Group2"},
						},
					},
				},
				{
					ID:     "2",
					RoleID: "test",
					Subjects: []*styra.Subject{
						{ID: "test", Kind: "test"},
						{ID: "test1@test.com", Kind: styra.SubjectKindUser},
					},
				},
			},
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("UpdateRoleBindingSubjects", mock.Anything, "1", &styra.UpdateRoleBindingSubjectsRequest{
			Subjects: []*styra.Subject{
				{ID: "test", Kind: "test"},
			},
		}).Return(&styra.UpdateRoleBindingSubjectsResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("UpdateRoleBindingSubjects", mock.Anything, "2", &styra.UpdateRoleBindingSubjectsRequest{
			Subjects: []*styra.Subject{
				{ID: "test", Kind: "test"},
			},
		}).Return(&styra.UpdateRoleBindingSubjectsResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Expect(k8sClient.Update(ctx, toUpdate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			var (
				getSystem                 int
				getUsers                  int
				listRolebindingsV2        int
				updateRoleBindingSubjects int
				getOPAConfig              int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					listRolebindingsV2++
				case "UpdateRoleBindingSubjects":
					updateRoleBindingSubjects++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 1 &&
				getUsers == 1 &&
				listRolebindingsV2 == 1 &&
				updateRoleBindingSubjects == 2 &&
				getOPAConfig == 1
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)
		resetMock(&webhookMock.Mock)

		ginkgo.By("Setting a datasource")

		gomega.Expect(k8sClient.Get(ctx, key, toUpdate)).To(gomega.Succeed())

		toUpdate.Spec.Datasources = []styrav1beta1.Datasource{{Path: "test"}}

		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					ID:             "system_id",
					Name:           key.String(),
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
					SourceControl: &styra.SourceControlConfig{
						Origin: styra.GitRepoConfig{
							URL:         sourceControl.Origin.URL,
							Credentials: "systems/system_id/git",
							Path:        "",
							Commit:      "",
							Reference:   "",
						},
					},
					Datasources: []*styra.DatasourceConfig{
						{
							Category: "rest",
							ID:       "systems/system_id/ignore",
						},
						{
							Category: "rest",
							ID:       "systems/system_id/delete",
						},
					},
					DecisionMappings: map[string]styra.DecisionMapping{
						"": {},
						"test": {
							Allowed: &styra.DecisionMappingAllowed{
								Expected: true,
								Path:     "test",
							},
							Columns: []styra.DecisionMappingColumn{
								{Key: "test", Path: "test"},
							},
							Reason: &styra.DecisionMappingReason{
								Path: "reason",
							},
						},
					},
				},
			}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		styraClientMock.On(
			"UpsertDatasource",
			mock.Anything,
			fmt.Sprintf("systems/%s/%s", "system_id", "test"),
			&styra.UpsertDatasourceRequest{Category: "rest"},
		).Return(&styra.UpsertDatasourceResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On(
			"DeleteDatasource",
			mock.Anything,
			"systems/system_id/delete",
		).Return(&styra.DeleteDatasourceResponse{}, nil).Once()

		webhookMock.On(
			"SystemDatasourceChanged",
			mock.Anything,
			mock.Anything,
			"system_id",
			"systems/system_id/test",
		).Return(nil).Once()

		gomega.Expect(k8sClient.Update(ctx, toUpdate)).To(gomega.Succeed())
		gomega.Eventually(func() bool {
			var (
				getSystem          int
				getUsers           int
				listRolebindingsV2 int
				upsertDatasource   int
				getOPAConfig       int
				datasourceChanged  int
				deleteDatasource   int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					listRolebindingsV2++
				case "GetOPAConfig":
					getOPAConfig++
				case "UpsertDatasource":
					upsertDatasource++
				case "DeleteDatasource":
					deleteDatasource++
				}
			}

			for _, call := range webhookMock.Calls {
				switch call.Method {
				case "SystemDatasourceChanged":
					datasourceChanged++
				}
			}

			return getSystem == 1 &&
				getUsers == 1 &&
				listRolebindingsV2 == 1 &&
				upsertDatasource == 1 &&
				getOPAConfig == 1 &&
				datasourceChanged == 1 &&
				deleteDatasource == 1
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)
		resetMock(&webhookMock.Mock)

		ginkgo.By("Setting credentialsSecretName")

		gomega.Expect(k8sClient.Get(ctx, key, toUpdate)).To(gomega.Succeed())

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Data: map[string][]byte{
				"name":   []byte("git-user"),
				"secret": []byte("git-password"),
			},
		}
		gomega.Expect(k8sClient.Create(ctx, secret)).To(gomega.Succeed())

		toUpdate.Spec.SourceControl = &styrav1beta1.SourceControl{
			Origin: styrav1beta1.GitRepo{
				CredentialsSecretName: key.Name,
			},
		}

		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(&styra.GetSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             "system_id",
				Name:           key.String(),
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},

				DecisionMappings: map[string]styra.DecisionMapping{
					"": {},
					"test": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Path:     "test",
						},
						Columns: []styra.DecisionMappingColumn{
							{
								Key:  "test",
								Path: "test",
							},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "reason",
						},
					},
				},
				Datasources: []*styra.DatasourceConfig{
					{
						ID:       "systems/system_id/test",
						Category: "rest",
					},
				},
			},
		}, nil).Once()

		styraClientMock.On("CreateUpdateSecret", mock.Anything, "systems/system_id/git", &styra.CreateUpdateSecretsRequest{
			Name:   "git-user",
			Secret: "git-password",
		}).Return(&styra.CreateUpdateSecretResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("UpdateSystem", mock.Anything, "system_id", &styra.UpdateSystemRequest{
			SystemConfig: &styra.SystemConfig{
				Name:           key.String(),
				ReadOnly:       true,
				Type:           "custom",
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				DecisionMappings: map[string]styra.DecisionMapping{
					"": {},
					"test": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Path:     "test",
						},
						Columns: []styra.DecisionMappingColumn{
							{
								Key:  "test",
								Path: "test",
							},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "reason",
						},
					},
				},
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						Credentials: "systems/system_id/git",
					},
				},
			},
		}).Return(&styra.UpdateSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             "system_id",
				Name:           key.String(),
				ReadOnly:       true,
				Type:           "custom",
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				DecisionMappings: map[string]styra.DecisionMapping{
					"": {},
					"test": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Path:     "test",
						},
						Columns: []styra.DecisionMappingColumn{
							{
								Key:  "test",
								Path: "test",
							},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "reason",
						},
					},
				},
				SourceControl: &styra.SourceControlConfig{
					Origin: styra.GitRepoConfig{
						Credentials: "systems/system_id/git",
					},
				},
				Datasources: []*styra.DatasourceConfig{
					{
						ID:       "systems/system_id/test",
						Category: "rest",
					},
				},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Expect(k8sClient.Update(ctx, toUpdate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			//check if conditions are correct
			fetched := &styrav1beta1.System{}
			key := types.NamespacedName{Name: toUpdate.Name, Namespace: toUpdate.Namespace}
			if err := k8sClient.Get(ctx, key, fetched); err != nil {
				return false
			}

			numberOfConditions := len(fetched.Status.Conditions)
			if numberOfConditions != 10 {
				return false
			}

			var (
				createdInStyra,
				gitCredentialsUpdated,
				subjectsUpdated,
				datasourcesUpdated,
				opaTokenUpdated,
				opaConfigMapUpdated,
				slpConfigMapUpdated,
				systemConfigUpdated,
				opaUpToDate,
				slpUpToDate metav1.ConditionStatus
			)

			for _, c := range fetched.Status.Conditions {
				switch c.Type {
				case styrav1beta1.ConditionTypeCreatedInStyra:
					createdInStyra = c.Status

				case styrav1beta1.ConditionTypeGitCredentialsUpdated:
					gitCredentialsUpdated = c.Status

				case styrav1beta1.ConditionTypeSubjectsUpdated:
					subjectsUpdated = c.Status

				case styrav1beta1.ConditionTypeDatasourcesUpdated:
					datasourcesUpdated = c.Status

				case styrav1beta1.ConditionTypeOPATokenUpdated:
					opaTokenUpdated = c.Status

				case styrav1beta1.ConditionTypeOPAConfigMapUpdated:
					opaConfigMapUpdated = c.Status

				case styrav1beta1.ConditionTypeSLPConfigMapUpdated:
					slpConfigMapUpdated = c.Status

				case styrav1beta1.ConditionTypeSystemConfigUpdated:
					systemConfigUpdated = c.Status

				case styrav1beta1.ConditionTypeOPAUpToDate:
					opaUpToDate = c.Status

				case styrav1beta1.ConditionTypeSLPUpToDate:
					slpUpToDate = c.Status
				}
			}

			return createdInStyra == metav1.ConditionTrue &&
				gitCredentialsUpdated == metav1.ConditionTrue &&
				subjectsUpdated == metav1.ConditionTrue &&
				datasourcesUpdated == metav1.ConditionTrue &&
				opaTokenUpdated == metav1.ConditionTrue &&
				opaConfigMapUpdated == metav1.ConditionTrue &&
				slpConfigMapUpdated == metav1.ConditionTrue &&
				systemConfigUpdated == metav1.ConditionTrue &&
				opaUpToDate == metav1.ConditionTrue &&
				slpUpToDate == metav1.ConditionTrue
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			var (
				getSystem          int
				getUsers           int
				createUpdateSecret int
				updateSystem       int
				listRolebindingsV2 int
				getOPAConfig       int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "GetUsers":
					getUsers++
				case "CreateUpdateSecret":
					createUpdateSecret++
				case "UpdateSystem":
					updateSystem++
				case "ListRoleBindingsV2":
					listRolebindingsV2++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 1 &&
				getUsers == 1 &&
				createUpdateSecret == 1 &&
				updateSystem == 1 &&
				listRolebindingsV2 == 1 &&
				getOPAConfig == 1
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Deleting the system")

		styraClientMock.On("DeleteSystem", mock.Anything, "system_id").Return(&styra.DeleteSystemResponse{
			StatusCode: http.StatusOK,
		}, nil)

		gomega.Expect(k8sClient.Delete(ctx, toCreate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			err := k8sClient.Get(ctx, key, fetched)
			return k8serrors.IsNotFound(err)
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Connecting to an existing Styra System by name with wrong deltabundle setting")

		spec2 := styrav1beta1.SystemSpec{
			DeletionProtection: ptr.Bool(false),
		}

		key2 := types.NamespacedName{
			Name:      "test2",
			Namespace: "default",
		}

		toCreate2 := &styrav1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key2.Name,
				Namespace: key2.Namespace,
			},
			Spec: spec2,
		}

		cfg2 := &styra.SystemConfig{
			ID:             "system_id2",
			Name:           key2.String(),
			ReadOnly:       true,
			BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: true},
		}

		styraClientMock.On("GetSystemByName", mock.Anything, key2.String()).Return(&styra.GetSystemResponse{
			StatusCode:   http.StatusOK,
			SystemConfig: cfg2,
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   cfg2.ID,
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg2.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg2.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil)

		styraClientMock.On("UpdateSystem", mock.Anything, cfg2.ID, &styra.UpdateSystemRequest{
			SystemConfig: &styra.SystemConfig{
				Name:           key2.String(),
				Type:           "custom",
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
			},
		}).Return(&styra.UpdateSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				Name:           key2.String(),
				Type:           "custom",
				ReadOnly:       true,
				ID:             cfg2.ID,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("GetSystem", mock.Anything, cfg2.ID).Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					ID:             cfg2.ID,
					Name:           key2.String(),
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				},
			}, nil)

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id2",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil)

		gomega.Expect(k8sClient.Create(ctx, toCreate2)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			var (
				getSystem          int
				getSystemByName    int
				createSystem       int
				deletePolicy       int
				getUsers           int
				listRolebindingsV2 int
				createRoleBinding  int
				getOPAConfig       int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "CreateSystem":
					createSystem++
				case "GetSystemByName":
					getSystemByName++
				case "DeletePolicy":
					deletePolicy++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					listRolebindingsV2++
				case "CreateRoleBinding":
					createRoleBinding++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 2 &&
				getSystemByName == 1 &&
				createSystem == 0 &&
				deletePolicy == 0 &&
				getUsers == 3 &&
				listRolebindingsV2 == 3 &&
				createRoleBinding == 0 &&
				getOPAConfig == 3
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			if err := k8sClient.Get(ctx, key2, fetched); err != nil {
				return false
			}
			return finalizer.IsSet(fetched) &&
				fetched.Status.ID == "system_id2" &&
				fetched.Status.Phase == styrav1beta1.SystemPhaseCreated &&
				fetched.Status.Ready
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		// Create new system, which was not ReadOnly, so it gets updated
		ginkgo.By("Changing ReadOnly flag in system")

		key3 := types.NamespacedName{
			Name:      "test3",
			Namespace: "default",
		}

		// system v1beta1
		toCreate3 := &styrav1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key3.Name,
				Namespace: key3.Namespace,
			},
			Spec: spec2,
		}

		cfg3 := &styra.SystemConfig{
			ID:             "system_id3",
			Name:           key3.String(),
			ReadOnly:       false,
			BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
		}

		styraClientMock.On("GetSystemByName", mock.Anything, key3.String()).Return(&styra.GetSystemResponse{
			StatusCode:   http.StatusOK,
			SystemConfig: cfg3,
		}, nil).Once()

		styraClientMock.On("UpdateSystem", mock.Anything, "system_id3", &styra.UpdateSystemRequest{
			SystemConfig: &styra.SystemConfig{
				Name:           key3.String(),
				ReadOnly:       true,
				Type:           "custom",
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
			},
		}).Return(&styra.UpdateSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             cfg3.ID,
				Name:           key3.String(),
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id3",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id3").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id3",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil)

		// new reconcile as we create opatoken secret that we are watching
		styraClientMock.On("GetSystem", mock.Anything, "system_id3").Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					ID:             "system_id3",
					Name:           key3.String(),
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
				},
			}, nil)

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id3",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil)

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		gomega.Expect(k8sClient.Create(ctx, toCreate3)).To(gomega.Succeed())
		gomega.Eventually(func() bool {
			var (
				getSystem          int
				getSystemByName    int
				createSystem       int
				deletePolicy       int
				getUsers           int
				listRolebindingsV2 int
				createRoleBinding  int
				getOPAConfig       int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "CreateSystem":
					createSystem++
				case "GetSystemByName":
					getSystemByName++
				case "DeletePolicy":
					deletePolicy++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					listRolebindingsV2++
				case "CreateRoleBinding":
					createRoleBinding++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 2 &&
				getSystemByName == 1 &&
				createSystem == 0 &&
				deletePolicy == 0 &&
				getUsers == 3 &&
				listRolebindingsV2 == 3 &&
				createRoleBinding == 0 &&
				getOPAConfig == 3
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}

			if err := k8sClient.Get(ctx, key3, fetched); err != nil {
				return false
			}
			return finalizer.IsSet(fetched) &&
				fetched.Status.ID == "system_id3" &&
				fetched.Status.Phase == styrav1beta1.SystemPhaseCreated &&
				fetched.Status.Ready
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Creating a system with non-default delta bundle setting and custom settings")

		key4 := types.NamespacedName{
			Name:      "test4",
			Namespace: "default",
		}

		cfg4 := styra.SystemConfig{
			ID:       "system4_id",
			Name:     key4.Name,
			ReadOnly: true,
		}

		customConfig := map[string]interface{}{
			"distributed_tracing": map[string]interface{}{
				"type":    "grpc",
				"address": "localhost:1234",
			},
		}

		customSettingsJSON, err := json.Marshal(customConfig)
		if err != nil {
			return
		}

		toCreate4 := &styrav1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key4.Name,
				Namespace: key4.Namespace,
			},
			Spec: styrav1beta1.SystemSpec{
				DeletionProtection: ptr.Bool(false),
				EnableDeltaBundles: ptr.Bool(true),
				CustomOPAConfig: &runtime.RawExtension{
					Raw: customSettingsJSON,
				},
			},
		}

		styraClientMock.On("GetSystemByName", mock.Anything, key4.String()).Return(&styra.GetSystemResponse{
			StatusCode:   http.StatusOK,
			SystemConfig: nil,
		}, nil).Once()

		styraClientMock.On("CreateSystem", mock.Anything, mock.Anything).Return(&styra.CreateSystemResponse{
			StatusCode:   http.StatusOK,
			SystemConfig: &cfg4,
		}, nil).Once()

		styraClientMock.On("DeletePolicy", mock.Anything, "systems/system4_id/rules").Return(&styra.DeletePolicyResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()
		styraClientMock.On("DeletePolicy", mock.Anything, "systems/system4_id/test").Return(&styra.DeletePolicyResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("UpdateSystem", mock.Anything, cfg4.ID, &styra.UpdateSystemRequest{
			SystemConfig: &styra.SystemConfig{
				Name:           key4.String(),
				Type:           "custom",
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: true},
			},
		}).Return(&styra.UpdateSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				Name:           key4.String(),
				Type:           "custom",
				ReadOnly:       true,
				ID:             cfg4.ID,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: true},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   cfg4.ID,
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("CreateRoleBinding", mock.Anything, &styra.CreateRoleBindingRequest{
			ResourceFilter: &styra.ResourceFilter{
				ID:   cfg4.ID,
				Kind: styra.RoleBindingKindSystem,
			},
			RoleID:   styra.RoleSystemViewer,
			Subjects: []*styra.Subject{},
		}).Return(&styra.CreateRoleBindingResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg4.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg4.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// new reconcile as we create opatoken secret that we are watching
		styraClientMock.On("GetSystem", mock.Anything, cfg4.ID).Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					ID:             cfg4.ID,
					Name:           key4.String(),
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: true},
				},
			}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   cfg4.ID,
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg4.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg4.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// new reconcile as we create opa configmap that we are watching
		styraClientMock.On("GetSystem", mock.Anything, cfg4.ID).Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					ID:             cfg4.ID,
					Name:           key4.String(),
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: true},
				},
			}, nil).Twice()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   "system_id",
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, "system_id").Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   "system_id",
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// new reconcile as we create opa configmap that we are watching
		styraClientMock.On("GetSystem", mock.Anything, cfg4.ID).Return(
			&styra.GetSystemResponse{
				StatusCode: http.StatusOK,
				SystemConfig: &styra.SystemConfig{
					ID:             cfg4.ID,
					Name:           key4.String(),
					ReadOnly:       true,
					BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: true},
				},
			}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: []styra.User{
				{ID: "test1@test.com", Enabled: true},
				{ID: "test2@test.com", Enabled: true},
			},
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   cfg4.ID,
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg4.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg4.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Expect(k8sClient.Create(ctx, toCreate4)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			if err := k8sClient.Get(ctx, key4, fetched); err != nil {
				return false
			}
			return finalizer.IsSet(fetched) &&
				fetched.Status.ID != "" &&
				fetched.Status.Phase == styrav1beta1.SystemPhaseCreated &&
				fetched.Status.Ready
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			fetched := &corev1.Secret{}
			key := types.NamespacedName{Name: fmt.Sprintf("%s-opa-token", key4.Name), Namespace: key4.Namespace}
			return k8sClient.Get(ctx, key, fetched) == nil && string(fetched.Data["token"]) == "opa-token-123"
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			fetched := &corev1.ConfigMap{}
			var actualMap, expectedMap map[string]interface{}

			key := types.NamespacedName{Name: fmt.Sprintf("%s-opa", key4.Name), Namespace: key4.Namespace}
			if fetchSuceeded := k8sClient.Get(ctx, key, fetched) == nil; !fetchSuceeded {
				return false
			}
			actualYAML := fetched.Data["opa-conf.yaml"]
			expectedYAML := `services:
- name: styra
  url: styra-url-123
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
- name: styra-bundles
  url: styra-url-123/bundles
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
labels:
  system-id: system4_id
  system-type: custom
discovery:
  name: discovery
  prefix: /systems/system4_id
  service: styra
distributed_tracing:
  type: grpc
  address: localhost:1234
`

			if err := yaml.Unmarshal([]byte(actualYAML), &actualMap); err != nil {
				return false
			}
			if err := yaml.Unmarshal([]byte(expectedYAML), &expectedMap); err != nil {
				return false
			}

			return reflect.DeepEqual(expectedMap, actualMap)

		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			var (
				getSystem          int
				getSystemByName    int
				createSystem       int
				deletePolicy       int
				getUsers           int
				rolebindingsListed int
				createRoleBinding  int
				getOPAConfig       int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "CreateSystem":
					createSystem++
				case "GetSystemByName":
					getSystemByName++
				case "DeletePolicy":
					deletePolicy++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					rolebindingsListed++
				case "CreateRoleBinding":
					createRoleBinding++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}
			return getSystem == 2 &&
				getSystemByName == 1 &&
				createSystem == 1 &&
				deletePolicy == 2 &&
				getUsers == 3 &&
				rolebindingsListed == 3 &&
				createRoleBinding == 1 &&
				getOPAConfig == 3
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		ginkgo.By("Creating the system will fail due to invalid url")

		key5 := types.NamespacedName{
			Name:      "test5",
			Namespace: "default",
		}

		sourceControl5 := styrav1beta1.SourceControl{
			Origin: styrav1beta1.GitRepo{
				URL: "https://github.com/[test]/repo.git",
			},
		}
		spec5 := styrav1beta1.SystemSpec{
			DeletionProtection: ptr.Bool(false),
			SourceControl:      &sourceControl5,
		}

		toCreate5 := &styrav1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key5.Name,
				Namespace: key5.Namespace,
			},
			Spec: spec5,
		}

		// Mock the GetSystemByName call to return nil, indicating the system does not exist
		styraClientMock.On("GetSystemByName", mock.Anything, key5.String()).Return(&styra.GetSystemResponse{
			StatusCode:   http.StatusOK,
			SystemConfig: nil,
		}, nil).Times(13)

		// Create the system in the Kubernetes cluster
		gomega.Expect(k8sClient.Create(ctx, toCreate5)).To(gomega.Succeed())

		// Verify that the system creation failed due to invalid URL
		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			err := k8sClient.Get(ctx, key5, fetched)
			if err != nil {
				return false
			}

			return fetched.Status.Phase == styrav1beta1.SystemPhaseFailed && !fetched.Status.Ready
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			var (
				getSystemByName int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystemByName":
					getSystemByName++
				}
			}

			return getSystemByName == 13
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)

		styraClientMock.AssertExpectations(ginkgo.GinkgoT())
	})
})

var _ = ginkgo.Describe("SystemReconciler.Reconcile1", ginkgo.Label("integration"), func() {
	ginkgo.It("should reconcile", func() {
		key := types.NamespacedName{
			Name:      "test-pod-restart",
			Namespace: "default",
		}

		ctx := context.Background()

		cfg := &styra.SystemConfig{
			ID:       "system_id",
			Name:     key.String(),
			ReadOnly: true,
		}

		ginkgo.By("Empty System already has ID but does not exist in Styra")

		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(&styra.GetSystemResponse{
			StatusCode:   http.StatusNotFound,
			SystemConfig: nil,
		}, &httperror.HTTPError{
			StatusCode: http.StatusNotFound,
			Body:       "nil",
		}).Once()

		styraClientMock.On("PutSystem",
			mock.Anything,
			mock.MatchedBy(func(req *styra.PutSystemRequest) bool {
				matchesDecisionmapping := len(req.SystemConfig.DecisionMappings) == 0
				matchesDescription := req.SystemConfig.Description == cfg.Description
				matchesSourceControl := req.SystemConfig.SourceControl == nil

				return matchesDecisionmapping &&
					matchesDescription &&
					matchesSourceControl
			}),
			"system_id",
			map[string]string{"If-None-Match": "*"},
		).Return(&styra.PutSystemResponse{
			StatusCode:   http.StatusOK,
			SystemConfig: cfg,
		}, nil).Once()

		styraClientMock.On("DeletePolicy", mock.Anything, "systems/system_id/rules").Return(&styra.DeletePolicyResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()
		styraClientMock.On("DeletePolicy", mock.Anything, "systems/system_id/test").Return(&styra.DeletePolicyResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("UpdateSystem",
			mock.Anything,
			cfg.ID,
			mock.MatchedBy(func(req *styra.UpdateSystemRequest) bool {
				return req.SystemConfig.Name == key.String() &&
					req.SystemConfig.ReadOnly == true &&
					req.SystemConfig.BundleDownload.DeltaBundles == false &&
					req.SystemConfig.Type == "custom"
			}),
		).Return(&styra.UpdateSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				Name:     key.String(),
				Type:     "custom",
				ReadOnly: true,
				ID:       cfg.ID,
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: nil,
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   cfg.ID,
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("CreateRoleBinding",
			mock.Anything,
			mock.MatchedBy(func(req *styra.CreateRoleBindingRequest) bool {
				return req.ResourceFilter.ID == cfg.ID
			}),
		).Return(&styra.CreateRoleBindingResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// New reconcile as we create opa token (secret) that we are watching
		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(&styra.GetSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             cfg.ID,
				Name:           key.String(),
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: nil,
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   cfg.ID,
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// New reconcile as we create opa configmap that we are watching
		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(&styra.GetSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             cfg.ID,
				Name:           key.String(),
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: nil,
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   cfg.ID,
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		// New reconcile as we create SLP ConfigMap that we are watching
		styraClientMock.On("GetSystem", mock.Anything, "system_id").Return(&styra.GetSystemResponse{
			StatusCode: http.StatusOK,
			SystemConfig: &styra.SystemConfig{
				ID:             cfg.ID,
				Name:           key.String(),
				ReadOnly:       true,
				BundleDownload: &styra.BundleDownloadConfig{DeltaBundles: false},
			},
		}, nil).Once()

		styraClientMock.On("GetUsers", mock.Anything).Return(&styra.GetUsersResponse{
			Users: nil,
		}, false, nil).Once()

		styraClientMock.On("ListRoleBindingsV2", mock.Anything, &styra.ListRoleBindingsV2Params{
			ResourceKind: styra.RoleBindingKindSystem,
			ResourceID:   cfg.ID,
		}).Return(&styra.ListRoleBindingsV2Response{
			StatusCode:   http.StatusOK,
			Rolebindings: []*styra.RoleBindingConfig{{ID: "1", RoleID: styra.RoleSystemViewer}},
		}, nil).Once()

		styraClientMock.On("GetOPAConfig", mock.Anything, cfg.ID).Return(styra.OPAConfig{
			HostURL:    "styra-url-123",
			SystemID:   cfg.ID,
			Token:      "opa-token-123",
			SystemType: "custom",
		}, nil).Once()

		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			if err := k8sClient.Get(ctx, key, fetched); err != nil {
				return false
			}
			return finalizer.IsSet(fetched) &&
				fetched.Status.ID == "system_id" &&
				fetched.Status.Phase == styrav1beta1.SystemPhaseCreated &&
				fetched.Status.Ready
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			fetched := &appsv1.StatefulSet{}
			key := types.NamespacedName{Name: fmt.Sprintf("%s-slp", key.Name), Namespace: key.Namespace}
			err := k8sClient.Get(ctx, key, fetched)
			if err != nil {
				return false
			}

			timestamp, exists := fetched.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"]
			if !exists {
				return false
			}

			_, err = time.Parse(time.RFC3339, timestamp)
			return err == nil
		}, timeout, interval).Should(gomega.BeTrue())

		gomega.Eventually(func() bool {
			var (
				getSystem          int
				putSystem          int
				deletePolicy       int
				getUsers           int
				rolebindingsListed int
				getOPAConfig       int
			)
			for _, call := range styraClientMock.Calls {
				switch call.Method {
				case "GetSystem":
					getSystem++
				case "PutSystem":
					putSystem++
				case "DeletePolicy":
					deletePolicy++
				case "GetUsers":
					getUsers++
				case "ListRoleBindingsV2":
					rolebindingsListed++
				case "GetOPAConfig":
					getOPAConfig++
				}
			}

			return getSystem == 4 &&
				putSystem == 1 &&
				deletePolicy == 2 &&
				getUsers == 4 &&
				rolebindingsListed == 4 &&
				getOPAConfig == 4
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)
	})
})

var _ = ginkgo.Describe("SystemReconciler.ReconcileOCPSystem", ginkgo.Label("integration"), func() {
	ginkgo.It("should reconcile", func() {

		expectedGitCredentialsID := "github-credentials"
		expectedS3CredentialsID := "s3-credentials"

		sourceControl := styrav1beta1.SourceControl{
			Origin: styrav1beta1.GitRepo{
				URL:       "https://github.com/test/repo.git",
				Reference: "refs/heads/master",
				Path:      "policy",
			},
		}

		spec := styrav1beta1.SystemSpec{
			DeletionProtection: ptr.Bool(false),
			SourceControl:      &sourceControl,
			Datasources: []styrav1beta1.Datasource{{
				Path: "path/to/datasource",
			}},
		}

		key := types.NamespacedName{
			Name:      "ocp-system",
			Namespace: "default",
		}

		toCreate := &styrav1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Labels: map[string]string{
					"styra-controller/control-plane": "opa-control-plane",
				},
			},
			Spec: spec,
		}

		ctx := context.Background()

		ginkgo.By("Creating the OCP system")

		//Called in createSourceIfNotExists
		ocpClientMock.On("GetSource", mock.Anything, "path-to-datasource").Return(&ocp.GetSourceResponse{
			StatusCode: http.StatusNotFound,
		},
			httperror.NewHTTPError(http.StatusNotFound, "404")).Once()

		// Called in createSourceIfNotExists
		ocpClientMock.On("PutSource", mock.Anything, "path-to-datasource", &ocp.PutSourceRequest{
			Name: "path-to-datasource",
		}).Return(&ocp.PutSourceResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		webhookMock.On(
			"SystemDatasourceChangedOCP",
			mock.Anything,
			mock.Anything,
			"path-to-datasource",
		).Return(nil).Once()

		// Called in reconcileSystemSource
		ocpClientMock.On("PutSource", mock.Anything, "default-ocp-system", &ocp.PutSourceRequest{
			Name: "default-ocp-system",
			Git: &ocp.GitConfig{
				Repo:          "https://github.com/test/repo.git",
				Reference:     "refs/heads/master",
				Path:          "policy",
				CredentialID:  expectedGitCredentialsID,
				IncludedFiles: []string{"*.rego"},
			},
		}).Return(&ocp.PutSourceResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		// Called in reconcileSystemBundle
		ocpClientMock.On("PutBundle", mock.Anything, &ocp.PutBundleRequest{
			Name: "default-ocp-system",
			ObjectStorage: ocp.ObjectStorage{
				AmazonS3: &ocp.AmazonS3{
					Bucket:      "test-bucket",
					Key:         "bundles/default-ocp-system/bundle.tar.gz",
					Region:      "eu-west-1",
					URL:         "s3-url",
					Credentials: expectedS3CredentialsID,
				},
			},
			Requirements: []ocp.Requirement{
				{
					Source: "library1",
				},
				{
					Source: "path-to-datasource",
				},
				{
					Source: "default-ocp-system",
				},
			},
		}).Return(nil).Once()

		// Called in reconcileS3Credentials
		// First time, the user does not exist in S3
		s3ClientMock.On("UserExists", mock.Anything, "Access-Key-test-bucket-default-ocp-system").Return(false, nil).Once()

		// Called in reconcileS3Credentials
		s3ClientMock.On("CreateSystemBundleUser", mock.Anything,
			"Access-Key-test-bucket-default-ocp-system", "test-bucket", "default-ocp-system").
			Return("system-user-secret-key", nil).Once()

		// #### New reconcile when updating secret ###

		//Called in createSourceIfNotExists
		ocpClientMock.On("GetSource", mock.Anything, "path-to-datasource").Return(&ocp.GetSourceResponse{
			StatusCode: http.StatusOK,
			Source: &ocp.SourceConfig{
				Name: "path-to-datasource",
			},
		}, nil).Once()

		// Called in reconcileSystemSource
		ocpClientMock.On("PutSource", mock.Anything, "default-ocp-system", &ocp.PutSourceRequest{
			Name: "default-ocp-system",
			Git: &ocp.GitConfig{
				Repo:          "https://github.com/test/repo.git",
				Reference:     "refs/heads/master",
				Path:          "policy",
				CredentialID:  expectedGitCredentialsID,
				IncludedFiles: []string{"*.rego"},
			},
		}).Return(&ocp.PutSourceResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		// Called in reconcileSystemBundle
		ocpClientMock.On("PutBundle", mock.Anything, &ocp.PutBundleRequest{
			Name: "default-ocp-system",
			ObjectStorage: ocp.ObjectStorage{
				AmazonS3: &ocp.AmazonS3{
					Bucket:      "test-bucket",
					Key:         "bundles/default-ocp-system/bundle.tar.gz",
					Region:      "eu-west-1",
					URL:         "s3-url",
					Credentials: expectedS3CredentialsID,
				},
			},
			Requirements: []ocp.Requirement{
				{
					Source: "library1",
				},
				{
					Source: "path-to-datasource",
				},
				{
					Source: "default-ocp-system",
				},
			},
		}).Return(nil).Once()

		// Called in reconcileS3Credentials
		s3ClientMock.On("UserExists", mock.Anything, "Access-Key-test-bucket-default-ocp-system").Return(true, nil).Once()

		// #### New reconcile when updating configmap ###

		//Called in createSourceIfNotExists
		ocpClientMock.On("GetSource", mock.Anything, "path-to-datasource").Return(&ocp.GetSourceResponse{
			StatusCode: http.StatusOK,
			Source: &ocp.SourceConfig{
				Name: "path-to-datasource",
			},
		}, nil).Once()

		// Called in reconcileSystemSource
		ocpClientMock.On("PutSource", mock.Anything, "default-ocp-system", &ocp.PutSourceRequest{
			Name: "default-ocp-system",
			Git: &ocp.GitConfig{
				Repo:          "https://github.com/test/repo.git",
				Reference:     "refs/heads/master",
				Path:          "policy",
				CredentialID:  expectedGitCredentialsID,
				IncludedFiles: []string{"*.rego"},
			},
		}).Return(&ocp.PutSourceResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		// Called in reconcileSystemBundle
		ocpClientMock.On("PutBundle", mock.Anything, &ocp.PutBundleRequest{
			Name: "default-ocp-system",
			ObjectStorage: ocp.ObjectStorage{
				AmazonS3: &ocp.AmazonS3{
					Bucket:      "test-bucket",
					Key:         "bundles/default-ocp-system/bundle.tar.gz",
					Region:      "eu-west-1",
					URL:         "s3-url",
					Credentials: expectedS3CredentialsID,
				},
			},
			Requirements: []ocp.Requirement{
				{
					Source: "library1",
				},
				{
					Source: "path-to-datasource",
				},
				{
					Source: "default-ocp-system",
				},
			},
		}).Return(nil).Once()

		// Called in reconcileS3Credentials
		s3ClientMock.On("UserExists", mock.Anything, "Access-Key-test-bucket-default-ocp-system").Return(true, nil).Once()

		gomega.Expect(k8sClient.Create(ctx, toCreate)).To(gomega.Succeed())

		// Assert that the System has all the correct statuses.
		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			if err := k8sClient.Get(ctx, key, fetched); err != nil {
				return false
			}

			finalizerIsSet := finalizer.IsSet(fetched)
			emptyID := fetched.Status.ID == ""
			systemStatusIsReady := fetched.Status.Ready
			systemStatusPhaseIsCreated := fetched.Status.Phase == styrav1beta1.SystemPhaseCreated
			systemStatusFailureMessageIsEmpty := fetched.Status.FailureMessage == ""

			conditionOPASecretUpdated := false
			conditionOPAConfigMapUpdated := false
			conditionOPAUpToDate := false

			for _, condition := range fetched.Status.Conditions {
				if condition.Type == styrav1beta1.ConditionTypeOPASecretUpdated && condition.Status == metav1.ConditionTrue {
					conditionOPASecretUpdated = true
				}
				if condition.Type == styrav1beta1.ConditionTypeOPAConfigMapUpdated && condition.Status == metav1.ConditionTrue {
					conditionOPAConfigMapUpdated = true
				}
				if condition.Type == styrav1beta1.ConditionTypeOPAUpToDate && condition.Status == metav1.ConditionTrue {
					conditionOPAUpToDate = true
				}
			}

			return finalizerIsSet &&
				emptyID &&
				systemStatusIsReady &&
				systemStatusPhaseIsCreated &&
				systemStatusFailureMessageIsEmpty &&
				conditionOPASecretUpdated &&
				conditionOPAConfigMapUpdated &&
				conditionOPAUpToDate
		}, timeout, interval).Should(gomega.BeTrue())

		// Assert that the secret has correct name and content
		gomega.Eventually(func() bool {
			fetched := &corev1.Secret{}
			key := types.NamespacedName{Name: fmt.Sprintf("%s-opa-secret", key.Name), Namespace: key.Namespace}
			return k8sClient.Get(ctx, key, fetched) == nil &&
				string(fetched.Data[s3.AWSSecretNameKeyID]) == "Access-Key-test-bucket-default-ocp-system" &&
				string(fetched.Data[s3.AWSSecretNameSecretKey]) == "system-user-secret-key"
		}, timeout, interval).Should(gomega.BeTrue())

		// Assert that the configmap has the correct name and content.
		gomega.Eventually(func() bool {
			fetched := &corev1.ConfigMap{}
			var actualMap, expectedMap map[string]interface{}

			key := types.NamespacedName{Name: fmt.Sprintf("%s-opa-config", key.Name), Namespace: key.Namespace}
			if fetchSuceeded := k8sClient.Get(ctx, key, fetched) == nil; !fetchSuceeded {
				return false
			}

			actualYAML := fetched.Data["opa-conf.yaml"]
			expectedYAML := `bundles:
  authz:
    resource: bundles/default-ocp-system/bundle.tar.gz
    service: s3
services:
- credentials:
    s3_signing:
      environment_credentials: {}
  name: s3
  url: s3-url/test-bucket
`

			if err := yaml.Unmarshal([]byte(actualYAML), &actualMap); err != nil {
				return false
			}
			if err := yaml.Unmarshal([]byte(expectedYAML), &expectedMap); err != nil {
				return false
			}

			equal := reflect.DeepEqual(expectedMap, actualMap)
			if !equal {
				fmt.Println("Actual", string(actualYAML))
				fmt.Println("Expected", string(expectedYAML))
			}
			return equal
		}, timeout, interval).Should(gomega.BeTrue())

		// Assert the correct amount of calls have been made to mocks.
		gomega.Eventually(func() bool {
			var (
				getSource int
				putSource int
				putBundle int
			)
			for _, call := range ocpClientMock.Calls {
				switch call.Method {
				case "GetSource":
					getSource++
				case "PutSource":
					putSource++
				case "PutBundle":
					putBundle++
				}
			}

			expected := getSource == 3 && putSource == 4 && putBundle == 3
			if !expected {
				fmt.Println("GetSource expected 3, was", getSource)
				fmt.Println("PutSource expected 4, was", putSource)
				fmt.Println("PutBundle expected 3, was", putBundle)
			}

			return expected
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&ocpClientMock.Mock)

		gomega.Eventually(func() bool {
			var (
				userExists             int
				createSystemBundleUser int
			)
			for _, call := range s3ClientMock.Calls {
				switch call.Method {
				case "UserExists":
					userExists++
				case "CreateSystemBundleUser":
					createSystemBundleUser++
				}
			}

			expected := userExists == 3 && createSystemBundleUser == 1
			if !expected {
				fmt.Println("UserExists expected 3, was", userExists)
				fmt.Println("CreateSystemBundleUser expected 1, was", createSystemBundleUser)
			}

			return expected
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&s3ClientMock.Mock)

		gomega.Eventually(func() bool {
			var (
				systemDatasourceChangedOCP int
			)
			for _, call := range webhookMock.Calls {
				switch call.Method {
				case "SystemDatasourceChangedOCP":
					systemDatasourceChangedOCP++
				}
			}

			expected := systemDatasourceChangedOCP == 1
			if !expected {
				fmt.Println("SystemDatasourceChangedOCP expected 1, was", systemDatasourceChangedOCP)
			}

			return expected
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&webhookMock.Mock)

		ginkgo.By("Deleting the system")

		ocpClientMock.On("DeleteBundle", mock.Anything, "default-ocp-system").Return(nil)
		ocpClientMock.On("DeleteSource", mock.Anything, "default-ocp-system").Return(nil)
		ocpClientMock.On("DeleteSource", mock.Anything, "path-to-datasource").Return(nil)

		gomega.Expect(k8sClient.Delete(ctx, toCreate)).To(gomega.Succeed())

		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			err := k8sClient.Get(ctx, key, fetched)
			return k8serrors.IsNotFound(err)
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&styraClientMock.Mock)
	})
})
