package golang

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"istio.io/tools/protoc-gen-crds/pkg/crd"
	"istio.io/tools/protoc-gen-crds/pkg/crd/openapi"
	"istio.io/tools/protoc-gen-crds/pkg/naming"
)

const definitionTmpl = `//
// GENERATED CODE -- DO NOT EDIT
//
package {{.PackageName}}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type {{.StructName}} struct {
	metav1.TypeMeta ` + "`json:\",inline\"`" + `

	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta ` + "`json:\"metadata,omitempty\" protobuf:\"bytes,1,opt,name=metadata\"`" + `

	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec {{.SpecName}} ` + "`json:\"spec,omitempty\" protobuf:\"bytes,2,opt,name=spec\"`" + `
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type {{.ListName}} struct {
	metav1.TypeMeta ` + "`json:\",inline\"`" + `

	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta ` + "`json:\"metadata,omitempty\" protobuf:\"bytes,1,opt,name=metadata\"`" + `

	// Items is the list of Ingress.
	Items []{{.StructName}} ` + "`json:\"items\" protobuf:\"bytes,2,rep,name=items\"`" + `
}
{{ range $oid, $type := .ObjectTypes }}

type {{$type.Name}} struct { {{ range $fid, $field := $type.Fields }}
	{{asGoFieldName $field.Name}} {{asGoType $field.Type}} {{jsonSpec $field}}{{end}}
}
{{end}}
`

var funcMap = map[string]interface{}{
	"asGoType":      asGoType,
	"asGoFieldName": asGoFieldName,
	"jsonSpec":      jsonSpec,
}

var definitionTemplate = template.
	Must(template.New("template").Funcs(funcMap).Parse(definitionTmpl))

type definitionContext struct {
	PackageName string
	StructName  string
	ListName    string
	SpecName    string

	ObjectTypes []*openapi.Object
}

// Emit go code for the given Custom Resource Definition.
func Emit(m *crd.ResourceDefinition) (string, error) {
	contents := ""

	var c definitionContext
	c.PackageName = m.Spec.Version
	c.StructName = naming.PascalCase(m.Spec.Names.Kind)
	c.SpecName = naming.PascalCase(m.Spec.Names.Kind) + "Spec"
	c.ListName = naming.PascalCase(m.Spec.Names.Kind + "List")

	// Get object types in order
	types := make(map[*openapi.Object]struct{})
	addTypes(types, m.Spec.Validation)

	var ordered []*openapi.Object
	for t := range types {
		ordered = append(ordered, t)
	}
	sort.Slice(ordered, func(i, j int) bool {
		ti := ordered[i]
		tj := ordered[j]
		if ti == m.Spec.Validation {
			return true
		}
		if tj == m.Spec.Validation {
			return false
		}
		return strings.Compare(ti.Name, tj.Name) < 0
	})
	c.ObjectTypes = ordered

	var b bytes.Buffer
	if err := definitionTemplate.Execute(&b, c); err != nil {
		return "", err
	}
	contents += string(b.Bytes())

	return contents, nil
}

func addTypes(types map[*openapi.Object]struct{}, t openapi.Type) {
	switch ty := t.(type) {
	case *openapi.Object:
		types[ty] = struct{}{}

		for _, f := range ty.Fields {
			addTypes(types, f.Type)
		}

	case *openapi.Array:
		addTypes(types, ty.ElementType)
	}
}

func asGoFieldName(s string) (string, error) {
	return naming.PascalCase(s), nil
}

func asGoType(t openapi.Type) (string, error) {
	switch st := t.(type) {
	case *openapi.Primitive:
		return asGoPrimitive(st)
	case *openapi.Array:
		el, err := asGoType(st.ElementType)
		if err != nil {
			return "", err
		}
		return "[]" + el, nil
	case *openapi.Object:
		return "*" + st.Name, nil
	default:
		return "", fmt.Errorf("unknown type: %v", t)
	}
}

func asGoPrimitive(p *openapi.Primitive) (string, error) {
	switch p.Type {
	case openapi.TypeInt32:
		return "int", nil
	case openapi.TypeString:
		return "string", nil
	case openapi.TypeBool:
		return "bool", nil
	default:
		return "", fmt.Errorf("unknown primitive type: %v", p)
	}
}

func jsonSpec(f openapi.Field) string {
	s := fmt.Sprintf("`json:\"%s,omitempty\"`", naming.CamelCase(f.Name))
	return s
}
