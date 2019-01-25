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
