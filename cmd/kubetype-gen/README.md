# kubetype-gen

`kubetype-gen` is a utility for generating Kubernetes type wrappers (`types.go`)
and registration code (`register.go`) for existing types that need to be used
within Kubernetes (e.g. a Custom Resource Definition based on an existing type).

The type definitions take the form:

```go
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SomeType struct {
	v1.TypeMeta `json:",inline"`
	// +optional
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the implementation of this definition.
	// +optional
	Spec pkg.SomeType `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}
```

The `register.go` file takes the standard form, including all the generated
types, as well as their corresponding `List` types.

## Usage

The generator is enabled by applying comment tags (`+tagName`) to types and
packages.

The general command line usage:

`${GOPATH}/bin/kubetype-gen -i <input-packages> -p <output-package>`

For example:

`${GOPATH}/bin/kubetype-gen -i istio.io/api/authentication/... -p istio.io/api/kube/apis`

The tool can also be used with `go run`, e.g. `go run vendor/istio.io/tools/cmd/kubetype-gen/main.go ...`

### Package Tags

The generator is enabled for a source package by adding tags to the package
comments in its `doc.go` file.  The following tags are required:

**+kubetype-gen:groupVersion=group/version**
> Tells the code generator that the types should be registered with the
> *group/version*.  The generated types will be emitted into the package,
> *.../group/version*.

Note, the same *groupVersion* values may be used in multiple packages, resulting
in the types from all those packages being generated into the same output
package and registered with the same *group/version*.

Example `doc.go`:
```go
// Package v1alpha1 tags supporting code generate
//+kubetype-gen:groupVersion=authentication.istio.io/v1alpha1
package v1alpha1

```

### Type Tags

The generator is enabled for a specific type by adding tags to the type's
comments.  The following tags are available:

**+kubetype-gen**
> Generate a Kubernetes type for this type.

**+kubetype-gen:groupVersion=\<group/version>**
> Tells the code generator that the type should be registered with the
> *group/version*.  The generated type will be emitted into the package,
> *.../group/version*.

**+kubetype-gen:kubeType=\<name>**
> The generated Kubenetes type should use the specified *name*.  If this tag is
> not specified, the generated type will use the same name as the type.  This
> tag may be specified multiple times, which will result in a generated type for
> each specified *name*.

**+kubetype-gen:\<name>:tag=\<tag>**
> The generated Kubenetes type named *name* should have the specified *tag*
> added to its generated comments.  The tag should not include the leading `+`
> as this will be added by the generator.  This may be specified multiple times,
> once for each *tag* that should be added to the generated type.

Example type comments:
```go
// +kubetype-gen
// +kubetype-gen:groupVersion=authentication.istio.io/v1alpha1
// +kubetype-gen:kubeType=Policy
// +kubetype-gen:kubeType=MeshPolicy
// +kubetype-gen:MeshPolicy:tag=genclient:nonNamespaced
type Policy struct {
```

### Other Considerations

#### Tags For Other Kubernetes Code Generators

In addition to adding `kubetype-gen` tags to types, you should consider adding
tags used by other Kubernetes code generators.  The comments from the source
types will be copied over directly to the generated Kubernetes types, allowing
you to run other code generators on the generated types (e.g. to generate
clientsets, informers, etc.).  This allows you to control what gets generated
from the other generators.  Note, the generator adds the
`+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object` tag to the
generated types.

#### Generating `DeepCopy()` Functions

If you are using your types in Kubernetes, you will most likely need to be able
to copy them.  This will require you to use `deepcopy-gen` on your source types.
This can be problematic depending on how your types are structured, meaning you
may have to manually implement functions for some types.  You will also want to
run the generator manually so you have direct control over the input packages,
e.g.

```
${GOPATH}/bin/deepcopy-gen -i istio.io/api/authentication/... -O zz_generated.deepcopy -h vendor/istio.io/tools/cmd/kubetype-gen/boilerplate.go.txt
```

#### Protobuf Serializers ####

If using protobuf for the source types, you will need to ensure that the
packages for the source types include a `generated.proto` file.  The Kubernetes
`go-to-protobuf` tool looks for these files in the source packages.  If you are
not using this mechanism, a simple work around is to use
`import public "package/to/some_existing.proto";` statements to include the
upstream `.proto` files for the generated types.  Also be aware that the
 `go-to-protobuf` generator assumes imports use full import path, e.g.
 `istio.io/api/authentication/v1alpha1/policy.proto` vs. something like
 `authentication/v1alpha1/policy.proto`.

When running `go-to-protobuf`, add the packages for the existing types to the
`--apimachinery-packages` using the `-` prefix.  This tells the code generator
to look for `generated.proto` files in those packages, but do not generate
code for the types within them.  This will limit `go-to-protobuf` to generating
code only for the types generated by `kubetype-gen`.