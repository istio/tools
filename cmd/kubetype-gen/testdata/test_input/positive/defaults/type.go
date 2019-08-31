// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package defaults

// AllOverridden is for test
// +kubetype-gen
// +kubetype-gen:groupVersion=group2/version2
// +kubetype-gen:package=success/defaults/override
type AllOverridden struct {
	Field string
}

// Defaulted is for test
// +kubetype-gen
type Defaulted struct {
	Field string
}

// NotGenerated is for test
type NotGenerated struct {
	Field string
}
