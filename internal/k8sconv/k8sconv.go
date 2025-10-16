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

// Package k8sconv contains helpers related to converting data to Kubernetes
// resources.
package k8sconv

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/pkg/ocp"
	"github.com/bankdata/styra-controller/pkg/styra"
)

type bearer struct {
	TokenPath string `yaml:"token_path"`
}

type credentials struct {
	Bearer bearer `yaml:"bearer"`
}

type s3credentials struct {
	S3Signing s3signing `yaml:"s3_signing"`
}

type s3signing struct {
	S3EnvironmentCredentials map[string]interface{} `yaml:"environment_credentials"`
}

type authz struct {
	Service  string `yaml:"service"`
	Resource string `yaml:"resource"`
}

type bundle struct {
	Authz authz `yaml:"authz"`
}

type service struct {
	Name        string      `yaml:"name"`
	URL         string      `yaml:"url"`
	Credentials credentials `yaml:"credentials,omitempty"`
}

type s3service struct {
	Name          string        `yaml:"name"`
	URL           string        `yaml:"url"`
	S3Credentials s3credentials `yaml:"credentials"`
}

type labels struct {
	SystemID   string `yaml:"system-id"`
	SystemType string `yaml:"system-type"`
}

type discovery struct {
	Name    string `yaml:"name"`
	Prefix  string `yaml:"prefix,omitempty"`
	Service string `yaml:"service"`
}

type http struct {
	Headers []string `yaml:"headers"`
}

type requestContext struct {
	HTTP http `yaml:"http"`
}

type decisionLogs struct {
	RequestContext requestContext `yaml:"request_context"`
}

type opaConfigMap struct {
	Services     []service    `yaml:"services"`
	Labels       labels       `yaml:"labels"`
	Discovery    discovery    `yaml:"discovery"`
	DecisionLogs decisionLogs `yaml:"decision_logs,omitempty"`
}

type s3opaConfigMap struct {
	Services     []s3service  `yaml:"services"`
	Bundles      bundle       `yaml:"bundles,omitempty"`
	DecisionLogs decisionLogs `yaml:"decision_logs,omitempty"`
}

// OpaConfToK8sOPAConfigMapforOCP creates a ConfigMap for the OPA.
// It configures OPA to fetch bundle from MinIO.
// OpaConfToK8sOPAConfigMapforOCP merges the information given as input into a ConfigMap for OPA
func OpaConfToK8sOPAConfigMapforOCP(
	opaconf ocp.OPAConfig,
	opaDefaultConfig configv2alpha2.OPAConfig,
	customConfig map[string]interface{},
) (corev1.ConfigMap, error) {

	s3opaConfigMap := s3opaConfigMap{
		Bundles: bundle{
			Authz: authz{
				Service:  "s3",
				Resource: opaconf.Resource,
			},
		},
		Services: []s3service{{
			Name: "s3",
			URL:  opaconf.URL,
			S3Credentials: s3credentials{
				S3Signing: s3signing{
					S3EnvironmentCredentials: map[string]interface{}{},
				},
			},
		}},
	}

	if opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		s3opaConfigMap.DecisionLogs = decisionLogs{
			RequestContext: requestContext{
				HTTP: http{
					Headers: opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers,
				},
			},
		}
	}

	opaConfigMapMapStringInterface, err := opaConfigMapToMap(s3opaConfigMap)
	if err != nil {
		return corev1.ConfigMap{}, err
	}

	merged := mergeMaps(opaConfigMapMapStringInterface, customConfig)

	res, err := yaml.Marshal(&merged)
	if err != nil {
		return corev1.ConfigMap{}, errors.Wrap(err, "Could not marshal configmap data")
	}

	var cm corev1.ConfigMap
	cm.Data = map[string]string{
		"opa-conf.yaml": string(res),
	}

	return cm, nil
}

// OpaConfToK8sOPAConfigMap creates a corev1.ConfigMap for the OPA based on the
// configuration from Styra. The configmap configures the OPA to communicate to
// an SLP.
func OpaConfToK8sOPAConfigMap(
	opaconf styra.OPAConfig,
	slpURL string,
	opaDefaultConfig configv2alpha2.OPAConfig,
	customConfig map[string]interface{},
) (corev1.ConfigMap, error) {

	opaConfigMap := opaConfigMap{
		Services: []service{{
			Name: "styra",
			URL:  slpURL,
		}},
		Labels: labels{
			SystemID:   opaconf.SystemID,
			SystemType: opaconf.SystemType,
		},
		Discovery: discovery{
			Name:    "discovery",
			Service: "styra",
		},
	}

	if opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		opaConfigMap.DecisionLogs = decisionLogs{
			RequestContext: requestContext{
				HTTP: http{
					Headers: opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers,
				},
			},
		}
	}

	opaConfigMapMapStringInterface, err := opaConfigMapToMap(opaConfigMap)

	if err != nil {
		return corev1.ConfigMap{}, err
	}

	merged := mergeMaps(opaConfigMapMapStringInterface, customConfig)

	res, err := yaml.Marshal(&merged)
	if err != nil {
		return corev1.ConfigMap{}, errors.Wrap(err, "Could not marshal configmap data")
	}

	var cm corev1.ConfigMap
	cm.Data = map[string]string{
		"opa-conf.yaml": string(res),
	}

	return cm, nil
}

// OpaConfToK8sSLPConfigMap creates a ConfigMap for the SLP based on the
// configuration from Styra.
func OpaConfToK8sSLPConfigMap(opaconf styra.OPAConfig) (corev1.ConfigMap, error) {
	type Bearer struct {
		TokenPath string `yaml:"token_path"`
	}

	type Credentials struct {
		Bearer Bearer `yaml:"bearer"`
	}

	type Service struct {
		Name        string      `yaml:"name"`
		URL         string      `yaml:"url"`
		Credentials Credentials `yaml:"credentials"`
	}

	type Labels struct {
		SystemID   string `yaml:"system-id"`
		SystemType string `yaml:"system-type"`
	}

	type Discovery struct {
		Name     string `yaml:"name"`
		Resource string `yaml:"resource"`
		Service  string `yaml:"service"`
	}

	type SLPConfigMap struct {
		Services  []Service `yaml:"services"`
		Labels    Labels    `yaml:"labels"`
		Discovery Discovery `yaml:"discovery"`
	}

	slpConfigMap := SLPConfigMap{
		Services: []Service{{
			Name: "styra",
			URL:  opaconf.HostURL,
			Credentials: Credentials{
				Bearer: Bearer{
					TokenPath: "/etc/slp/auth/token",
				},
			},
		}},
		Labels: Labels{
			SystemID:   opaconf.SystemID,
			SystemType: opaconf.SystemType,
		},
		Discovery: Discovery{
			Name:     "discovery",
			Resource: fmt.Sprintf("/systems/%s/discovery", opaconf.SystemID),
			Service:  "styra",
		},
	}

	res, err := yaml.Marshal(&slpConfigMap)
	if err != nil {
		return corev1.ConfigMap{}, errors.Wrap(err, "could not marshal opa config data")
	}

	var cm corev1.ConfigMap
	cm.Data = map[string]string{
		"slp.yaml": string(res),
	}

	return cm, nil
}

// OpaConfToK8sOPAConfigMapNoSLP creates a ConfigMap for the OPA based on the
// configuration from Styra. The ConfigMap configures the OPA to communicate
// directly to Styra and not via an SLP.
func OpaConfToK8sOPAConfigMapNoSLP(
	opaconf styra.OPAConfig,
	opaDefaultConfig configv2alpha2.OPAConfig,
	customConfig map[string]interface{},
) (corev1.ConfigMap, error) {

	opaConfigMap := opaConfigMap{
		Services: []service{{
			Name: "styra",
			URL:  opaconf.HostURL,
			Credentials: credentials{
				Bearer: bearer{
					TokenPath: "/etc/opa/auth/token",
				},
			},
		},
			{
				Name: "styra-bundles",
				URL:  fmt.Sprintf("%s/bundles", opaconf.HostURL),
				Credentials: credentials{
					Bearer: bearer{
						TokenPath: "/etc/opa/auth/token",
					},
				},
			},
		},
		Labels: labels{
			SystemID:   opaconf.SystemID,
			SystemType: opaconf.SystemType,
		},
		Discovery: discovery{
			Name:    "discovery",
			Prefix:  fmt.Sprintf("/systems/%s", opaconf.SystemID),
			Service: "styra",
		},
	}

	if opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		opaConfigMap.DecisionLogs = decisionLogs{
			RequestContext: requestContext{
				HTTP: http{
					Headers: opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers,
				},
			},
		}
	}

	opaConfigMapMapStringInterface, err := opaConfigMapToMap(opaConfigMap)

	if err != nil {
		return corev1.ConfigMap{}, err
	}

	merged := mergeMaps(opaConfigMapMapStringInterface, customConfig)

	res, err := yaml.Marshal(&merged)
	if err != nil {
		return corev1.ConfigMap{}, errors.Wrap(err, "Could not marshal configmap data")
	}

	var cm corev1.ConfigMap
	cm.Data = map[string]string{
		"opa-conf.yaml": string(res),
	}

	return cm, nil
}

func opaConfigMapToMap(cm interface{}) (map[string]interface{}, error) {
	res, err := yaml.Marshal(&cm)
	if err != nil {
		return nil, errors.Wrap(err, "Could not marshal configmap data")
	}

	var opaConfigMapMapStringInterface map[string]interface{}

	err = yaml.Unmarshal(res, &opaConfigMapMapStringInterface)
	if err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal configmap data to map[string]interface{}")
	}

	return opaConfigMapMapStringInterface, nil
}

// mergeMaps recursively merges two map[string]interface{} variables
func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	mergedMap := make(map[string]interface{})

	// Copy all key-value pairs from map1 to mergedMap
	for key, value := range map1 {
		mergedMap[key] = value
	}

	// Copy all key-value pairs from map2 to mergedMap
	for key, value := range map2 {
		if existingValue, ok := mergedMap[key]; ok {
			// If the key exists in both maps and both values are maps, merge them recursively
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if valueMap, ok := value.(map[string]interface{}); ok {
					mergedMap[key] = mergeMaps(existingMap, valueMap)
					continue
				}
			}
		}
		// Otherwise, overwrite the value from map1 with the value from map2
		mergedMap[key] = value
	}

	return mergedMap
}
