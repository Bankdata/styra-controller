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
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// OPAConfig stores the information retrieved from calling the GET
// /v1/systems/{systemId}/assets/opa-config endpoint in the Styra API.
type OPAConfig struct {
	HostURL    string
	Token      string
	SystemID   string
	SystemType string
}

type getOPAConfigResponse struct {
	Discovery getOPAConfigDiscovery `yaml:"discovery"`
	Labels    getOPAConfigLabels    `yaml:"labels"`
	Services  []getOPAConfigService `yaml:"services"`
}

type getOPAConfigDiscovery struct {
	Name    string `yaml:"name"`
	Prefix  string `yaml:"prefix"`
	Service string `yaml:"service"`
}

type getOPAConfigLabels struct {
	SystemID   string `yaml:"system-id"`
	SystemType string `yaml:"system-type"`
}

type getOPAConfigService struct {
	Credentials getOPAConfigServiceCredentials `yaml:"credentials"`
	URL         string                         `yaml:"url"`
}

type getOPAConfigServiceCredentials struct {
	Bearer getOPAConfigServiceBearerCredentials `yaml:"bearer"`
}

type getOPAConfigServiceBearerCredentials struct {
	Token string `yaml:"token"`
}

// GetOPAConfig calls the GET /v1/systems/{systemId}/assets/opa-config endpoint
// in the Styra API.
func (c *Client) GetOPAConfig(ctx context.Context, systemID string) (OPAConfig, error) {
	res, err := c.request(ctx, http.MethodGet, fmt.Sprintf("/v1/systems/%s/assets/opa-config", systemID), nil)
	if err != nil {
		return OPAConfig{}, errors.Wrap(err, "could not get opaconf file")
	}

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return OPAConfig{}, errors.Wrap(err, "could not read body")
		}

		err = NewHTTPError(res.StatusCode, string(body))
		return OPAConfig{}, err
	}

	var getOPAConfigResponse getOPAConfigResponse
	if err := yaml.NewDecoder(res.Body).Decode(&getOPAConfigResponse); err != nil {
		return OPAConfig{}, errors.Wrap(err, "could not decode opa-config asset response")
	}

	if getOPAConfigResponse.Services == nil {
		return OPAConfig{}, errors.Errorf("No services in opa config")
	}

	opaConfig := OPAConfig{
		HostURL:    getOPAConfigResponse.Services[0].URL,
		Token:      getOPAConfigResponse.Services[0].Credentials.Bearer.Token,
		SystemID:   getOPAConfigResponse.Labels.SystemID,
		SystemType: getOPAConfigResponse.Labels.SystemType,
	}

	return opaConfig, nil
}
