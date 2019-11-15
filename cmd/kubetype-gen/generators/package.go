// Copyright 2019 Istio Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package generators

import (
	"fmt"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/types"

	"istio.io/tools/cmd/kubetype-gen/metadata"
)

// NewPackageGenerator generates source for a scanned package, specifically k8s styled doc.go, types.go and register.go files
func NewPackageGenerator(source metadata.PackageMetadata, boilerplate []byte) generator.Package {
	return &generator.DefaultPackage{
		PackageName: source.TargetPackage().Name,
		PackagePath: source.TargetPackage().Path,
		HeaderText:  boilerplate,
		PackageDocumentation: []byte(fmt.Sprintf(`
// Package has auto-generated kube type wrappers for raw types.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package
// +groupName=%s
`, source.GroupVersion().Group)),
		FilterFunc: func(c *generator.Context, t *types.Type) bool {
			for _, it := range source.RawTypes() {
				if t == it {
					return true
				}
			}
			return false
		},
		GeneratorList: []generator.Generator{
			// generate types.go
			NewTypesGenerator(source),
			// generate register.go
			NewRegisterGenerator(source),
			generator.DefaultGen{
				OptionalName: "doc",
			},
		},
	}
}
