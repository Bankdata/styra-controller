package styra

import (
	"context"
	"net/http"
	"path"

	"github.com/google/uuid"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("GlobalDatasourceReconciler", func() {
	ginkgo.Describe("Reconcile", ginkgo.Label("integration"), func() {
		ginkgo.It("reconciles GlobalDatasource", func() {
			key := types.NamespacedName{Name: uuid.NewString()}

			toCreate := &styrav1alpha1.GlobalDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
				},
				Spec: styrav1alpha1.GlobalDatasourceSpec{
					Name:     key.Name,
					Category: styrav1alpha1.GlobalDatasourceCategoryGitRego,
					URL:      "http://test.com/test.git",
				},
			}

			ctx := context.Background()

			ginkgo.By("creating the datasource")

			styraClientMock.On(
				"CreateUpdateSecret",
				mock.Anything,
				path.Join("libraries/global", key.Name, "git"),
				&styra.CreateUpdateSecretsRequest{
					Name:   "test-user",
					Secret: "test-secret",
				},
			).Return(&styra.CreateUpdateSecretResponse{
				StatusCode: http.StatusOK,
			}, nil)

			styraClientMock.On(
				"GetDatasource",
				mock.Anything,
				path.Join("global", key.Name),
			).Return(
				nil, &styra.HTTPError{
					StatusCode: http.StatusNotFound,
				},
			)

			styraClientMock.On(
				"UpsertDatasource",
				mock.Anything,
				path.Join("global", key.Name),
				&styra.UpsertDatasourceRequest{
					Category:    "git/rego",
					Enabled:     true,
					Credentials: path.Join("libraries/global", key.Name, "git"),
					URL:         "http://test.com/test.git",
				},
			).Return(&styra.UpsertDatasourceResponse{}, nil)

			gomega.Ω(k8sClient.Create(ctx, toCreate)).To(gomega.Succeed())

			gomega.Eventually(func() bool {
				var gds styrav1alpha1.GlobalDatasource
				if err := k8sClient.Get(ctx, key, &gds); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(gomega.BeTrue())

			gomega.Eventually(func() bool {
				var (
					createUpdateSecret int
					getDatasource      int
					upsertDatasource   int
				)
				for _, call := range styraClientMock.Calls {
					switch call.Method {
					case "CreateUpdateSecret":
						createUpdateSecret++
					case "GetDatasource":
						getDatasource++
					case "UpsertDatasource":
						upsertDatasource++
					}
				}
				return createUpdateSecret == 1 &&
					getDatasource == 1 &&
					upsertDatasource == 1
			}, timeout, interval).Should(gomega.BeTrue())

			styraClientMock.AssertExpectations(ginkgo.GinkgoT())
			resetMock(&styraClientMock.Mock)

			ginkgo.By("using a git credential from a secret")

			gomega.Ω(k8sClient.Create(ctx, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: "default",
				},
				Data: map[string][]byte{
					"name":   []byte("test-user-2"),
					"secret": []byte("test-secret-2"),
				},
			})).To(gomega.Succeed())

			gomega.Eventually(func() bool {
				var s v1.Secret
				return k8sClient.Get(ctx, types.NamespacedName{Name: key.Name, Namespace: "default"}, &s) == nil
			}, timeout, interval).Should(gomega.BeTrue())

			styraClientMock.On(
				"CreateUpdateSecret",
				mock.Anything,
				path.Join("libraries/global", key.Name, "git"),
				&styra.CreateUpdateSecretsRequest{
					Name:   "test-user-2",
					Secret: "test-secret-2",
				},
			).Return(&styra.CreateUpdateSecretResponse{
				StatusCode: http.StatusOK,
			}, nil)

			styraClientMock.On(
				"GetDatasource",
				mock.Anything,
				path.Join("global", key.Name),
			).Return(
				&styra.GetDatasourceResponse{
					StatusCode: http.StatusOK,
					DatasourceConfig: &styra.DatasourceConfig{
						Category:    "git/rego",
						Enabled:     true,
						Credentials: path.Join("libraries/global", key.Name, "git"),
						URL:         "http://test.com/test.git",
					},
				}, nil,
			)

			toCreate.Spec.CredentialsSecretRef = &styrav1alpha1.GlobalDatasourceSecretRef{
				Name:      key.Name,
				Namespace: "default",
			}

			gomega.Ω(k8sClient.Update(ctx, toCreate)).To(gomega.Succeed())

			gomega.Eventually(func() bool {
				var (
					createUpdateSecret int
					getDatasource      int
				)
				for _, call := range styraClientMock.Calls {
					switch call.Method {
					case "CreateUpdateSecret":
						createUpdateSecret++
					case "GetDatasource":
						getDatasource++
					}
				}
				return createUpdateSecret == 1 && getDatasource == 1
			}, timeout, interval).Should(gomega.BeTrue())

			styraClientMock.AssertExpectations(ginkgo.GinkgoT())
		})
	})
})
