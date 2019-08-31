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

// NoGroupVersion is for test
// test should fail because kubetype-gen:groupVersion is missing
// +kubetype-gen
type NoGroupVersion struct {
	Field string
}

// InvalidGroupVersion is for test
// test should fail because kubetype-gen:groupVersion is invalid (schema.ParseGroupVersion() error)
// +kubetype-gen
// +kubetype-gen:groupVersion=group/version/version
type InvalidGroupVersion struct {
	Field string
}

// EmptyGroup is for test
// test should fail because group/version is missing group
// +kubetype-gen
// +kubetype-gen:groupVersion=groupversion
type EmptyGroup struct {
	Field string
}

// AGoodType is for test
// a good type used for other failures
// +kubetype-gen
// +kubetype-gen:groupVersion=group/version
type AGoodType struct {
	Field string
}

// DuplicateKubeType is for test
// this test should fail because it specifies a kubeType name that is used for another type
// +kubetype-gen
// +kubetype-gen:groupVersion=group/version
// +kubetype-gen:kubeType=AGoodType
type DuplicateKubeType struct {
	Field string
}

// OverlappingPackageGroupType1 is for test
// this test should fail because the target package for the group conflicts with another group (group2.name.io same package as group2)
// +kubetype-gen
// +kubetype-gen:groupVersion=group2.name.io/version
type OverlappingPackageGroupType1 struct {
	Field string
}

// OverlappingPackageGroupType2 is for test
// this test should fail because the target package for the group conflicts with another group (group2.name.io same package as group2)
// +kubetype-gen
// +kubetype-gen:groupVersion=group2/version
type OverlappingPackageGroupType2 struct {
	Field string
}
