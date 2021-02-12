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

package scanner

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/gengo/args"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/types"

	"istio.io/tools/cmd/kubetype-gen/generators"
	"istio.io/tools/cmd/kubetype-gen/metadata"
)

const (
	// enabledTagName is the root tag used to identify types that need a corresponding kube type generated
	enabledTagName = "kubetype-gen"

	// groupVersionTagName is the tag used to identify the k8s group/version associated with the generated types.
	groupVersionTagName = enabledTagName + ":groupVersion"

	// kubeTypeTagName is used to identify the name(s) of the types to be generated from the type with this tag.
	// If this tag is not present, the k8s type will have the same name as the source type.  If this tag is specified
	// multiple times, a k8s type will be generated for each value.
	kubeTypeTagName = enabledTagName + ":kubeType"

	// kubeTagsTagTemplate is used to identify a comment tag that should be added to the generated kubeType.  This
	// allows different sets of tags to be used when a single type is the source for multiple kube types (e.g. where one
	// is namespaced and another is not).  The tag should not be prefixed with '+', as this will be added by the
	// generator.  This may be specified multiple times, once for each tag to be added to the generated type.
	kubeTagsTagTemplate = enabledTagName + ":%s:tag"
)

// Scanner is used to scan input packages for types with kubetype-gen tags
type Scanner struct {
	arguments *args.GeneratorArgs
	context   *generator.Context
}

// Scan the input packages for types with kubetype-gen tags
func (s *Scanner) Scan(c *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	s.arguments = arguments
	s.context = c

	boilerplate, err := arguments.LoadGoBoilerplate()
	if err != nil {
		glog.Fatalf("Failed loading boilerplate: %v", err)
	}

	// scan input packages for kubetype-gen
	metadataStore := metadata.NewMetadataStore(s.getBaseOutputPackage(), &c.Universe)
	fail := false

	glog.V(5).Info("Scanning input packages")
	for _, input := range c.Inputs {
		glog.V(5).Infof("Scanning package %s", input)
		pkg := c.Universe[input]
		if pkg == nil {
			glog.Warningf("Package not found: %s", input)
			continue
		}
		if strings.HasPrefix(arguments.OutputPackagePath, pkg.Path) {
			glog.Warningf("Ignoring package %s as it is located in the output package %s", pkg.Path, arguments.OutputPackagePath)
			continue
		}

		pkgTags := types.ExtractCommentTags("+", pkg.DocComments)

		// group/version for generated types from this package
		defaultGV, err := s.getGroupVersion(pkgTags, nil)
		if err != nil {
			glog.Errorf("Could not calculate Group/Version for package %s: %v", pkg.Path, err)
			fail = true
		} else if defaultGV != nil {
			if len(defaultGV.Group) == 0 {
				glog.Errorf("Invalid Group/Version for package %s, Group not specified for Group/Version: %v", pkg.Path, defaultGV)
				fail = true
			} else {
				glog.V(5).Infof("Default Group/Version for package: %s", defaultGV)
			}
		}

		// scan package for types that need kube types generated
		for _, t := range pkg.Types {
			comments := make([]string, 0, len(t.CommentLines)+len(t.SecondClosestCommentLines))
			comments = append(comments, t.CommentLines...)
			comments = append(comments, t.SecondClosestCommentLines...)
			typeTags := types.ExtractCommentTags("+", comments)
			if _, exists := typeTags[enabledTagName]; exists {
				var gv *schema.GroupVersion
				gv, err = s.getGroupVersion(typeTags, defaultGV)
				if err != nil {
					glog.Errorf("Could not calculate Group/Version for type %s: %v", t, err)
					fail = true
					continue
				} else if gv == nil || len(gv.Group) == 0 {
					glog.Errorf("Invalid Group/Version for type %s: %s", t, gv)
					fail = true
					continue
				}

				packageMetadata := metadataStore.MetadataForGV(gv)
				if packageMetadata == nil {
					glog.Errorf("Could not create metadata for type: %s", t)
					fail = true
					continue
				}

				kubeTypes := s.createKubeTypesForType(t, packageMetadata.TargetPackage())
				glog.V(5).Infof("Kube types %v will be generated with Group/Version %s, for raw type in %s", kubeTypes, gv, t)
				err = packageMetadata.AddMetadataForType(t, kubeTypes...)
				if err != nil {
					glog.Errorf("Error adding metadata source for %s: %v", t, err)
					fail = true
				}
			}
		}
	}

	glog.V(5).Info("Finished scanning input packages")

	validationErrors := metadataStore.Validate()
	if len(validationErrors) > 0 {
		for _, validationErr := range validationErrors {
			glog.Error(validationErr)
		}
		fail = true
	}
	if fail {
		glog.Exit("Errors occurred while scanning input.  See previous output for details.")
	}

	generatorPackages := []generator.Package{}
	for _, source := range metadataStore.AllMetadata() {
		if len(source.RawTypes()) == 0 {
			glog.Warningf("Skipping generation of %s, no types to generate", source.GroupVersion())
			continue
		}
		glog.V(2).Infof("Adding package generator for %s.", source.GroupVersion())
		generatorPackages = append(generatorPackages, generators.NewPackageGenerator(source, boilerplate))
	}
	return generatorPackages
}

func (s *Scanner) getGroupVersion(tags map[string][]string, defaultGV *schema.GroupVersion) (*schema.GroupVersion, error) {
	if value, exists := tags[groupVersionTagName]; exists && len(value) > 0 {
		gv, err := schema.ParseGroupVersion(value[0])
		if err == nil {
			return &gv, nil
		}
		return nil, fmt.Errorf("invalid group version '%s' specified: %v", value[0], err)
	}
	return defaultGV, nil
}

func (s *Scanner) getBaseOutputPackage() *types.Package {
	return s.context.Universe.Package(s.arguments.OutputPackagePath)
}

func (s *Scanner) createKubeTypesForType(t *types.Type, outputPackage *types.Package) []metadata.KubeType {
	namesForType := s.kubeTypeNamesForType(t)
	newKubeTypes := make([]metadata.KubeType, 0, len(namesForType))
	for _, name := range namesForType {
		tags := s.getTagsForKubeType(t, name)
		newKubeTypes = append(newKubeTypes, metadata.NewKubeType(t, s.context.Universe.Type(types.Name{Name: name, Package: outputPackage.Path}), tags))
	}
	return newKubeTypes
}

func (s *Scanner) kubeTypeNamesForType(t *types.Type) []string {
	names := []string{}
	comments := make([]string, 0, len(t.CommentLines)+len(t.SecondClosestCommentLines))
	comments = append(comments, t.CommentLines...)
	comments = append(comments, t.SecondClosestCommentLines...)
	tags := types.ExtractCommentTags("+", comments)
	if value, exists := tags[kubeTypeTagName]; exists {
		if len(value) == 0 || len(value[0]) == 0 {
			glog.Errorf("Invalid value specified for +%s in type %s.  Using default name %s.", kubeTypeTagName, t, t.Name.Name)
			names = append(names, t.Name.Name)
		} else {
			for _, name := range value {
				if len(name) > 0 {
					names = append(names, name)
				}
			}
		}
	} else {
		names = append(names, t.Name.Name)
	}
	return names
}

func (s *Scanner) getTagsForKubeType(t *types.Type, name string) []string {
	tagName := fmt.Sprintf(kubeTagsTagTemplate, name)
	comments := make([]string, 0, len(t.CommentLines)+len(t.SecondClosestCommentLines))
	comments = append(comments, t.CommentLines...)
	comments = append(comments, t.SecondClosestCommentLines...)
	tags := types.ExtractCommentTags("+", comments)
	if value, exists := tags[tagName]; exists {
		return value
	}
	return []string{}
}
