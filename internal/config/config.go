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

// Package config provides utilities for reading configfiles
package config

import (
	"os"

	v1 "github.com/bankdata/styra-controller/api/config/v1"
	"github.com/bankdata/styra-controller/api/config/v2alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// Load loads controller configuration from the given file using the types
// registered in the scheme.
func Load(file string, scheme *runtime.Scheme) (*v2alpha1.ProjectConfig, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "could not read config file")
	}
	return deserialize(bs, scheme)
}

func deserialize(data []byte, scheme *runtime.Scheme) (*v2alpha1.ProjectConfig, error) {
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
	_, gvk, err := decoder.Decode(data, nil, nil)

	if err != nil {
		return nil, errors.Wrap(err, "could not decode config")
	}

	if gvk.Group != v2alpha1.GroupVersion.Group {
		return nil, errors.New("unsupported api group")
	}

	if gvk.Kind != "ProjectConfig" {
		return nil, errors.New("unsupported api kind")
	}

	cfg := &v2alpha1.ProjectConfig{}

	switch gvk.Version {
	case v2alpha1.GroupVersion.Version:
		if _, _, err := decoder.Decode(data, nil, cfg); err != nil {
			return nil, errors.Wrap(err, "could not decode into kind")
		}
	case v1.GroupVersion.Version:
		var v1cfg v1.ProjectConfig
		if _, _, err := decoder.Decode(data, nil, &v1cfg); err != nil {
			return nil, errors.Wrap(err, "could not decode into kind")
		}
		cfg = v1cfg.ToV2Alpha1()
	default:
		return nil, errors.New("unsupported api version")
	}

	return cfg, nil
}
