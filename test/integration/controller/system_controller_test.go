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
	"fmt"
	"net/http"
	"reflect"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
	"github.com/bankdata/styra-controller/internal/finalizer"
	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/bankdata/styra-controller/pkg/ocp"
	"github.com/bankdata/styra-controller/pkg/ptr"
)

var _ = ginkgo.Describe("SystemReconciler.Reconcile", ginkgo.Label("integration"), func() {
	ginkgo.It("should reconcile OCP system", func() {
		sourceControl := styrav1beta1.SourceControl{
			Origin: styrav1beta1.GitRepo{
				URL: "https://github.com/test/repo.git",
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

		//Called in createSourceIfNotExists for datasources
		ocpClientMock.On("GetSource", mock.Anything, "path-to-datasource").Return(&ocp.GetSourceResponse{
			StatusCode: http.StatusNotFound,
		},
			httperror.NewHTTPError(http.StatusNotFound, "404")).Once()
		ocpClientMock.On("GetSource", mock.Anything, "path-to-datasource").Return(&ocp.GetSourceResponse{
			StatusCode: http.StatusOK,
			Source: &ocp.SourceConfig{
				Name: "path-to-datasource",
			},
		}, nil)

		// Called in createSourceIfNotExists
		ocpClientMock.On("PutSource", mock.Anything, "path-to-datasource", &ocp.PutSourceRequest{
			Name: "path-to-datasource",
		}).Return(&ocp.PutSourceResponse{
			StatusCode: http.StatusOK,
		}, nil).Once()

		// Called in reconcileSystemSource
		ocpClientMock.On("PutSource", mock.Anything, "default-ocp-system", &ocp.PutSourceRequest{
			Name: "default-ocp-system",
			Git: &ocp.GitConfig{
				Repo:          "https://github.com/test/repo.git",
				CredentialID:  "github-credentials",
				IncludedFiles: []string{"*.rego"},
				ExcludedFiles: []string{"*_test.rego"},
			},
		}).Return(&ocp.PutSourceResponse{
			StatusCode: http.StatusOK,
		}, nil)

		// Called in reconcileSystemBundle
		ocpClientMock.On("PutBundle", mock.Anything, &ocp.PutBundleRequest{
			Name: "default-ocp-system",
			ObjectStorage: ocp.ObjectStorage{
				AmazonS3: &ocp.AmazonS3{
					Bucket:      "test-bucket",
					Key:         "bundles/default-ocp-system/bundle.tar.gz",
					Region:      "eu-west-1",
					URL:         "s3-url",
					Credentials: "s3-credentials",
				},
			},
			Requirements: []ocp.Requirement{
				{
					Source: "path-to-datasource",
				},
				{
					Source: "default-ocp-system",
				},
				{
					Source: "library1",
				},
			},
			Revision: `$"data:{crypto.sha256(concat("", {x | x := input.sources[_].sql.hash}))},` +
				`git-sha:{input.sources["default-ocp-system"].git.commit},` +
				`libraries:{crypto.sha256(concat("", {x | some y in ["library1"]; x := input.sources[y].git.commit}))}"`,
		}).Return(nil)

		gomega.Expect(k8sClient.Create(ctx, toCreate)).To(gomega.Succeed())

		// Assert that the System has all the correct statuses.
		gomega.Eventually(func() bool {
			fetched := &styrav1beta1.System{}
			if err := k8sClient.Get(ctx, key, fetched); err != nil {
				return false
			}

			finalizerIsSet := finalizer.IsSet(fetched)
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
				systemStatusIsReady &&
				systemStatusPhaseIsCreated &&
				systemStatusFailureMessageIsEmpty &&
				conditionOPASecretUpdated &&
				conditionOPAConfigMapUpdated &&
				conditionOPAUpToDate
		}, timeout, interval).Should(gomega.BeTrue())

		// Assert that the secret has the correct name and is created.
		gomega.Eventually(func() bool {
			fetched := &corev1.Secret{}
			secretKey := types.NamespacedName{Name: fmt.Sprintf("%s-opa-secret", key.Name), Namespace: key.Namespace}
			return k8sClient.Get(ctx, secretKey, fetched) == nil
		}, timeout, interval).Should(gomega.BeTrue())

		// Assert that the configmap has the correct name and content.
		gomega.Eventually(func() bool {
			fetched := &corev1.ConfigMap{}
			var actualMap, expectedMap map[string]interface{}

			cmKey := types.NamespacedName{Name: fmt.Sprintf("%s-opa-config", key.Name), Namespace: key.Namespace}
			if fetchSucceeded := k8sClient.Get(ctx, cmKey, fetched) == nil; !fetchSucceeded {
				return false
			}

			actualYAML := fetched.Data["opa-conf.yaml"]
			expectedYAML := `bundles:
  authz:
    resource: bundles/default-ocp-system/bundle.tar.gz
    service: bundle-server
decision_logs:
  reporting:
    max_delay_seconds: 60
    min_delay_seconds: 5
    upload_size_limit_bytes: 1024
  resource_path: /logs
  service: decision-api
labels:
  namespace: default
  unique-name: default-ocp-system
services:
- credentials:
    bearer:
      token_path: token-path-bundle-server
  name: bundle-server
  url: https://bundle-server-url/test-bucket
- credentials:
    bearer:
      token_path: token-path-decision-api
  name: decision-api
  url: decision-api-url
`

			if err := yaml.Unmarshal([]byte(actualYAML), &actualMap); err != nil {
				return false
			}
			if err := yaml.Unmarshal([]byte(expectedYAML), &expectedMap); err != nil {
				return false
			}

			return reflect.DeepEqual(expectedMap, actualMap)
		}, timeout, interval).Should(gomega.BeTrue())

		// Assert exact OCP call counts during reconciliation.
		gomega.Eventually(func() bool {
			var (
				getSourceDatasource int
				putSourceDatasource int
				putSourceSystem     int
				putBundleSystem     int
			)
			for _, call := range ocpClientMock.Calls {
				switch call.Method {
				case "GetSource":
					if len(call.Arguments) >= 2 && call.Arguments.Get(1) == "path-to-datasource" {
						getSourceDatasource++
					}
				case "PutSource":
					if len(call.Arguments) >= 2 {
						switch call.Arguments.Get(1) {
						case "path-to-datasource":
							putSourceDatasource++
						case "default-ocp-system":
							putSourceSystem++
						}
					}
				case "PutBundle":
					if len(call.Arguments) >= 2 {
						if req, ok := call.Arguments.Get(1).(*ocp.PutBundleRequest); ok && req.Name == "default-ocp-system" {
							putBundleSystem++
						}
					}
				}
			}

			return getSourceDatasource == 4 && putSourceDatasource == 1 && putSourceSystem == 3 && putBundleSystem == 3
		}, timeout, interval).Should(gomega.BeTrue())

		resetMock(&ocpClientMock.Mock)

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

		resetMock(&ocpClientMock.Mock)
	})
})
