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
