package generators

import (
	"io"

	"istio.io/tools/cmd/kubetype-gen/metadata"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

type typesGenerator struct {
	generator.DefaultGen
	source  metadata.PackageMetadata
	imports namer.ImportTracker
}

// NewTypesGenerator creates a new generator for creating k8s style types.go files
func NewTypesGenerator(source metadata.PackageMetadata) generator.Generator {
	return &typesGenerator{
		DefaultGen: generator.DefaultGen{
			OptionalName: "types",
		},
		source:  source,
		imports: generator.NewImportTracker(),
	}
}

func (g *typesGenerator) Namers(c *generator.Context) namer.NameSystems {
	return NameSystems(g.source.TargetPackage().Path, g.imports)
}

func (g *typesGenerator) Imports(c *generator.Context) []string {
	return g.imports.ImportLines()
}

func (g *typesGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	kubeTypes := g.source.KubeTypes(t)
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	m := map[string]interface{}{
		"KubeType":   nil,
		"RawType":    t,
		"TypeMeta":   c.Universe.Type(types.Name{Name: "TypeMeta", Package: "k8s.io/apimachinery/pkg/apis/meta/v1"}),
		"ObjectMeta": c.Universe.Type(types.Name{Name: "ObjectMeta", Package: "k8s.io/apimachinery/pkg/apis/meta/v1"}),
		"ListMeta":   c.Universe.Type(types.Name{Name: "ListMeta", Package: "k8s.io/apimachinery/pkg/apis/meta/v1"}),
	}
	for _, kubeType := range kubeTypes {
		m["KubeType"] = kubeType
		sw.Do(kubeTypeTemplate, m)
	}
	return sw.Error()
}

const kubeTypeTemplate = `
$- range .RawType.SecondClosestCommentLines $
// $ . $
$- end $
$- range .KubeType.Tags $
// +$ . $
$- end $
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

$ range .RawType.CommentLines $
// $ . $
$- end $
type $.KubeType.Type|public$ struct {
	$.TypeMeta|raw$ ` + "`" + `json:",inline"` + "`" + `
	// +optional
	$.ObjectMeta|raw$ ` + "`" + `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"` + "`" + `

	// Spec defines the implementation of this definition.
	// +optional
	Spec $.RawType|raw$ ` + "`" + `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"` + "`" + `

}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// $.KubeType.Type|public$List is a collection of $.KubeType.Type|publicPlural$.
type $.KubeType.Type|public$List struct {
	$.TypeMeta|raw$ ` + "`" + `json:",inline"` + "`" + `
	// +optional
	$.ListMeta|raw$ ` + "`" + `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"` + "`" + `
	Items           []$.KubeType.Type|raw$ ` + "`" + `json:"items" protobuf:"bytes,2,rep,name=items"` + "`" + `
}
`
