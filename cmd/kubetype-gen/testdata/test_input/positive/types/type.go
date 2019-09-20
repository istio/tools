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

package types

// Type1 is for test
// +kubetype-gen
// +kubetype-gen:groupVersion=group/version
type Type1 struct {
	Field string
}

// NameOverride is for test
// +kubetype-gen
// +kubetype-gen:groupVersion=group/version
// +kubetype-gen:kubeType=Type2
type NameOverride struct {
	Field string
}

// MultipleNames is for test
// +kubetype-gen
// +kubetype-gen:groupVersion=group/version
// +kubetype-gen:kubeType=Type3
// +kubetype-gen:kubeType=Type4
// +kubetype-gen:Type4:tag=sometag=somevalue
type MultipleNames struct {
	Field string
}

// EmptyKubeType is for test
// +kubetype-gen
// +kubetype-gen:groupVersion=group/version
// +kubetype-gen:kubeType
type EmptyKubeType struct {
	Field string
}

// ComplexGroupVersionKubeType is for test
// +kubetype-gen
// +kubetype-gen:groupVersion=group2.test.io/version
// +kubetype-gen:kubeType
type ComplexGroupVersionKubeType struct {
	Field string
}

// +kubetype-gen
// +kubetype-gen:groupVersion=group/version

// SecondCommentsKubeType is for test
type SecondCommentsKubeType struct {
	Field string
}
