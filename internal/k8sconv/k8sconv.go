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

// Package k8sconv contains helpers related to converting data to Kubernetes
// resources.
package k8sconv

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/pkg/styra"
)

// OpaConfToK8sOPAConfigMap creates a corev1.ConfigMap for the OPA based on the
// configuration from Styra. The configmap configures the OPA to communicate to
// an SLP.
func OpaConfToK8sOPAConfigMap(
	opaconf styra.OPAConfig,
	slpURL string,
	opaDefaultConfig configv2alpha2.OPAConfig,
) (corev1.ConfigMap, error) {
	type Service struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
	}

	type Labels struct {
		SystemID   string `yaml:"system-id"`
		SystemType string `yaml:"system-type"`
	}

	type Discovery struct {
		Name    string `yaml:"name"`
		Service string `yaml:"service"`
	}

	type HTTP struct {
		Headers []string `yaml:"headers"`
	}

	type RequestContext struct {
		HTTP HTTP `yaml:"http"`
	}

	type DecisionLogs struct {
		RequestContext RequestContext `yaml:"request_context"`
	}

	type OPAConfigMap struct {
		Services     []Service    `yaml:"services"`
		Labels       Labels       `yaml:"labels"`
		Discovery    Discovery    `yaml:"discovery"`
		DecisionLogs DecisionLogs `yaml:"decision_logs,omitempty"`
	}

	opaConfigMap := OPAConfigMap{
		Services: []Service{{
			Name: "styra",
			URL:  slpURL,
		}},
		Labels: Labels{
			SystemID:   opaconf.SystemID,
			SystemType: opaconf.SystemType,
		},
		Discovery: Discovery{
			Name:    "discovery",
			Service: "styra",
		},
	}

	if opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		opaConfigMap.DecisionLogs = DecisionLogs{
			RequestContext: RequestContext{
				HTTP: HTTP{
					Headers: opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers,
				},
			},
		}
	}

	res, err := yaml.Marshal(&opaConfigMap)
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
) (corev1.ConfigMap, error) {
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
		Name    string `yaml:"name"`
		Prefix  string `yaml:"prefix"`
		Service string `yaml:"service"`
	}

	type HTTP struct {
		Headers []string `yaml:"headers"`
	}

	type RequestContext struct {
		HTTP HTTP `yaml:"http"`
	}

	type DecisionLogs struct {
		RequestContext RequestContext `yaml:"request_context"`
	}

	type OPAConfigMap struct {
		Services     []Service    `yaml:"services"`
		Labels       Labels       `yaml:"labels"`
		Discovery    Discovery    `yaml:"discovery"`
		DecisionLogs DecisionLogs `yaml:"decision_logs,omitempty"`
	}

	opaConfigMap := OPAConfigMap{
		Services: []Service{{
			Name: "styra",
			URL:  opaconf.HostURL,
			Credentials: Credentials{
				Bearer: Bearer{
					TokenPath: "/etc/opa/auth/token",
				},
			},
		},
			{
				Name: "styra-bundles",
				URL:  fmt.Sprintf("%s/bundles", opaconf.HostURL),
				Credentials: Credentials{
					Bearer: Bearer{
						TokenPath: "/etc/opa/auth/token",
					},
				},
			},
		},
		Labels: Labels{
			SystemID:   opaconf.SystemID,
			SystemType: opaconf.SystemType,
		},
		Discovery: Discovery{
			Name:    "discovery",
			Prefix:  fmt.Sprintf("/systems/%s", opaconf.SystemID),
			Service: "styra",
		},
	}

	if opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers != nil {
		opaConfigMap.DecisionLogs = DecisionLogs{
			RequestContext: RequestContext{
				HTTP: HTTP{
					Headers: opaDefaultConfig.DecisionLogs.RequestContext.HTTP.Headers,
				},
			},
		}
	}

	res, err := yaml.Marshal(&opaConfigMap)
	if err != nil {
		return corev1.ConfigMap{}, errors.Wrap(err, "Could not marshal configmap data")
	}

	var cm corev1.ConfigMap
	cm.Data = map[string]string{
		"opa-conf.yaml": string(res),
	}

	return cm, nil
}
