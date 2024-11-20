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

package styra_test

import (
	"net/http"

	"github.com/bankdata/styra-controller/pkg/styra"
	"github.com/patrickmn/go-cache"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(f roundTripFunc) styra.ClientInterface {
	return &styra.Client{
		URL: "http://test.com",
		HTTPClient: http.Client{
			Transport: roundTripFunc(f),
		},
	}
}

func newTestClientWithCache(f roundTripFunc, cache *cache.Cache) styra.ClientInterface {
	return &styra.Client{
		URL: "http://test.com",
		HTTPClient: http.Client{
			Transport: roundTripFunc(f),
		},
		Cache: cache,
	}
}
