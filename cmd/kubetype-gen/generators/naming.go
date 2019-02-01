package generators

import (
	"k8s.io/gengo/namer"
)

// NameSystems used by the kubetype generator
func NameSystems(generatedPackage string, tracker namer.ImportTracker) namer.NameSystems {
	return namer.NameSystems{
		"public":       namer.NewPublicNamer(0),
		"raw":          namer.NewRawNamer(generatedPackage, tracker),
		"publicPlural": namer.NewPublicPluralNamer(map[string]string{}),
	}
}

// DefaultNameSystem to use if none is specified
func DefaultNameSystem() string {
	return "public"
}
