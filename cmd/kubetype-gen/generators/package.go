package generators

import (
	"istio.io/tools/cmd/kubetype-gen/metadata"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/types"
)

// NewPackageGenerator generates source for a scanned package, specifically k8s styled doc.go, types.go and register.go files
func NewPackageGenerator(source metadata.PackageMetadata, boilerplate []byte) generator.Package {
	return &generator.DefaultPackage{
		PackageName: source.TargetPackage().Name,
		PackagePath: source.TargetPackage().Path,
		HeaderText:  boilerplate,
		PackageDocumentation: []byte(`
// Package has auto-generated kube type wrappers for raw types.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package
`),
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
