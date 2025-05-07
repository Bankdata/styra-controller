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

package main

import (
	"encoding/json"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/pkg/styra"
	styraclientmock "github.com/bankdata/styra-controller/pkg/styra/mocks"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/kind/pkg/errors"
)

var _ = ginkgo.Describe("ConfigureDecisionExporter", func() {

	type test struct {
		createUpdateSecretError error
		updateWorkspaceError    error
		expectedErrorMsg        string
	}

	// Test cases for ConfigureExporter Decisions
	ginkgo.DescribeTable("ConfigureDecisionExporter", func(test test) {
		// Arrange
		ctrlConfig := &configv2alpha2.ProjectConfig{
			DecisionsExporter: &configv2alpha2.ExporterConfig{
				Enabled: true,
				Kafka: &configv2alpha2.KafkaConfig{
					TLS: &configv2alpha2.TLSConfig{
						ClientCertificateName: "test-cert-name",
						ClientCertificate:     "test-cert",
						ClientKey:             "test-key",
						RootCA:                "test-ca",
						InsecureSkipVerify:    true,
					},
					Brokers:      []string{"broker1", "broker2"},
					RequiredAcks: "WaitForAll",
					Topic:        "test-topic",
				},
				Interval: "1m",
			},
		}

		mockStyraClient := &styraclientmock.ClientInterface{}
		mockStyraClient.On("CreateUpdateSecret", mock.Anything, "test-cert-name", &styra.CreateUpdateSecretsRequest{
			Description: "Client certificate for Kafka",
			Name:        "test-cert",
			Secret:      "test-key",
		}).Return(nil, test.createUpdateSecretError)

		// If createUpdateSecret fails, we should not expect call to UpdateWorkspace
		if test.createUpdateSecretError == nil {
			mockStyraClient.On("UpdateWorkspace", mock.Anything, &styra.UpdateWorkspaceRequest{
				DecisionsExporter: &styra.ExporterConfig{
					Interval: "1m",
					Kafka: &styra.KafkaConfig{
						Authentication: "TLS",
						Brokers:        []string{"broker1", "broker2"},
						RequredAcks:    "WaitForAll",
						Topic:          "test-topic",
						TLS: &styra.KafkaTLS{
							ClientCert:         "test-cert-name",
							RootCA:             "test-ca",
							InsecureSkipVerify: true,
						},
					},
				},
			}).Return(nil, test.updateWorkspaceError)
		}

		// Act
		err := configureExporter(mockStyraClient, ctrlConfig.DecisionsExporter, "DecisionsExporter")

		// Assert
		mockStyraClient.AssertExpectations(ginkgo.GinkgoT())
		if test.createUpdateSecretError != nil || test.updateWorkspaceError != nil {
			gomega.Ω(err).Should(gomega.HaveOccurred())
			gomega.Ω(err.Error()).Should(gomega.Equal(test.expectedErrorMsg))
		} else {
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		}
	},

		ginkgo.Entry("OK", test{
			createUpdateSecretError: nil,
			updateWorkspaceError:    nil,
			expectedErrorMsg:        "",
		}),
		ginkgo.Entry("CreateUpdateSecretFails", test{
			createUpdateSecretError: errors.New("CreateUpdateSecret failed"),
			updateWorkspaceError:    nil,
			expectedErrorMsg:        "CreateUpdateSecret failed",
		}),
		ginkgo.Entry("UpdateWorkspaceFails", test{
			createUpdateSecretError: errors.New("UpdateWorkspace failed"),
			updateWorkspaceError:    nil,
			expectedErrorMsg:        "UpdateWorkspace failed",
		}),
	)
})

var _ = ginkgo.Describe("ConfigureActivityExporter", func() {

	type test struct {
		createUpdateSecretError error
		updateWorkspaceError    error
		expectedErrorMsg        string
		enabled                 bool
		deleteSecretError       error
	}

	// Test cases for ConfigureExporter Decisions
	ginkgo.DescribeTable("ConfigureActivityExporter", func(test test) {
		// Arrange
		ctrlConfig := &configv2alpha2.ProjectConfig{
			ActivityExporter: &configv2alpha2.ExporterConfig{
				Enabled: test.enabled,
				Kafka: &configv2alpha2.KafkaConfig{
					TLS: &configv2alpha2.TLSConfig{
						ClientCertificateName: "test-cert-name",
						ClientCertificate:     "test-cert",
						ClientKey:             "test-key",
						RootCA:                "test-ca",
						InsecureSkipVerify:    true,
					},
					Brokers:      []string{"broker1", "broker2"},
					RequiredAcks: "WaitForAll",
					Topic:        "test-topic",
				},
				Interval: "1m",
			},
		}

		mockStyraClient := &styraclientmock.ClientInterface{}

		if test.enabled {
			mockStyraClient.On("CreateUpdateSecret", mock.Anything, "test-cert-name", &styra.CreateUpdateSecretsRequest{
				Description: "Client certificate for Kafka",
				Name:        "test-cert",
				Secret:      "test-key",
			}).Return(nil, test.createUpdateSecretError)

			// If createUpdateSecret fails, we should not expect call to UpdateWorkspace
			if test.createUpdateSecretError == nil {
				mockStyraClient.On("UpdateWorkspace", mock.Anything, &styra.UpdateWorkspaceRequest{
					ActivityExporter: &styra.ExporterConfig{
						Interval: "1m",
						Kafka: &styra.KafkaConfig{
							Authentication: "TLS",
							Brokers:        []string{"broker1", "broker2"},
							RequredAcks:    "WaitForAll",
							Topic:          "test-topic",
							TLS: &styra.KafkaTLS{
								ClientCert:         "test-cert-name",
								RootCA:             "test-ca",
								InsecureSkipVerify: true,
							},
						},
					},
				}).Return(nil, test.updateWorkspaceError)
			}
		} else {
			mockStyraClient.On("DeleteSecret", mock.Anything, "test-cert-name").Return(nil, test.deleteSecretError)

			if test.deleteSecretError == nil {
				rawJSON := json.RawMessage("{\"activity_exporter\":null}")
				mockStyraClient.On("UpdateWorkspaceRaw", mock.Anything, rawJSON).Return(nil, test.updateWorkspaceError)
			}
		}

		// Act
		err := configureExporter(mockStyraClient, ctrlConfig.ActivityExporter, "ActivityExporter")

		// Assert
		mockStyraClient.AssertExpectations(ginkgo.GinkgoT())
		if test.createUpdateSecretError != nil || test.updateWorkspaceError != nil || test.deleteSecretError != nil {
			gomega.Ω(err).Should(gomega.HaveOccurred())
			gomega.Ω(err.Error()).Should(gomega.Equal(test.expectedErrorMsg))
		} else {
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		}
	},

		ginkgo.Entry("OK", test{
			createUpdateSecretError: nil,
			updateWorkspaceError:    nil,
			expectedErrorMsg:        "",
			enabled:                 true,
			deleteSecretError:       nil,
		}),
		ginkgo.Entry("CreateUpdateSecretFails", test{
			createUpdateSecretError: errors.New("CreateUpdateSecret failed"),
			updateWorkspaceError:    nil,
			expectedErrorMsg:        "CreateUpdateSecret failed",
			enabled:                 true,
			deleteSecretError:       nil,
		}),
		ginkgo.Entry("UpdateWorkspaceFails", test{
			createUpdateSecretError: nil,
			updateWorkspaceError:    errors.New("UpdateWorkspace failed"),
			expectedErrorMsg:        "UpdateWorkspace failed",
			enabled:                 true,
			deleteSecretError:       nil,
		}),
		ginkgo.Entry("OKRemovingExporter", test{
			createUpdateSecretError: nil,
			updateWorkspaceError:    nil,
			expectedErrorMsg:        "",
			enabled:                 false,
			deleteSecretError:       nil,
		}),
		ginkgo.Entry("DeleteSecretFails", test{
			createUpdateSecretError: nil,
			updateWorkspaceError:    nil,
			expectedErrorMsg:        "DeleteSecret failed",
			enabled:                 false,
			deleteSecretError:       errors.New("DeleteSecret failed"),
		}),
		ginkgo.Entry("UpdateWorkspaceRemoveExporterFails", test{
			createUpdateSecretError: nil,
			updateWorkspaceError:    errors.New("UpdateWorkspaceRemoveExporter failed"),
			expectedErrorMsg:        "UpdateWorkspaceRemoveExporter failed",
			enabled:                 false,
			deleteSecretError:       nil,
		}),
	)
})
