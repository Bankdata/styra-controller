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

package config

import "strings"

// StringSlice implements flag.Value for a repeatable string flag.
// Each call to Set appends a value, allowing --flag=a --flag=b syntax.
type StringSlice []string

// String returns the flag value as a comma-separated string.
func (s *StringSlice) String() string {
	return strings.Join(*s, ",")
}

// Set appends a value to the slice. Called once per flag occurrence.
func (s *StringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}
