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

package metadata

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/gengo/types"
)

// KubeType is the interface representing a type to be generated.
type KubeType interface {
	RawType() *types.Type
	Type() *types.Type
	Tags() []string
}

// PackageMetadata is the interface used to provide source data used by the package generators.
type PackageMetadata interface {
	// GroupVersion is the k8s Group/Version to use for the generated types.
	GroupVersion() *schema.GroupVersion

	// TargetPackage is the package into which the k8s types will be generated.
	TargetPackage() *types.Package

	// RawTypes is the list of types for which k8s types should be generated.
	RawTypes() []*types.Type

	// KubeTypes is the list of k8s types to be generated for the given rawType.
	KubeTypes(rawType *types.Type) []KubeType

	// AllKubeTypes is the list of all k8s types to be generated
	AllKubeTypes() []KubeType

	// AddMetadata is used to add metadata collected by the scanner.
	AddMetadataForType(rawType *types.Type, kubeTypes ...KubeType) error

	// Validate is used to validate the metadata prior to generation
	Validate() []error
}

// Store is used to store/access the source metadata collected by the scanner
type Store interface {

	// MetadataForGV returns the package metadata associated with the Group/Version
	MetadataForGV(gv *schema.GroupVersion) PackageMetadata

	// AllMetadata returns the source metadata.
	AllMetadata() []PackageMetadata

	// Validate is used to validate the metadata prior to generation
	Validate() []error
}

type kubeTypeMetadata struct {
	rawType  *types.Type
	kubeType *types.Type
	tags     []string
}

type packageMetadata struct {
	groupVersion        *schema.GroupVersion
	targetPackage       *types.Package
	rawTypes            []*types.Type
	allKubeTypes        []KubeType
	kubeTypesForRawType map[*types.Type][]KubeType
}

type metadataStore struct {
	baseOutputPackage *types.Package
	universe          *types.Universe
	metadataForGV     map[string]PackageMetadata
	metadata          []PackageMetadata
}

// NewMetadataStore returns a new store used for collecting source metadata used by the generator.
func NewMetadataStore(baseOutputPackage *types.Package, universe *types.Universe) Store {
	return &metadataStore{
		baseOutputPackage: baseOutputPackage,
		universe:          universe,
		metadataForGV:     map[string]PackageMetadata{},
		metadata:          []PackageMetadata{},
	}
}

func (s *metadataStore) MetadataForGV(gv *schema.GroupVersion) PackageMetadata {
	simpleGV := schema.GroupVersion{Group: strings.SplitN(gv.Group, ".", 2)[0], Version: gv.Version}
	existing := s.metadataForGV[simpleGV.String()]
	if existing == nil {
		glog.V(5).Infof("Creating new PackageMetadata for  Group/Version %s", gv)
		existing = &packageMetadata{
			groupVersion:        gv,
			targetPackage:       s.createTargetPackage(gv),
			rawTypes:            []*types.Type{},
			allKubeTypes:        []KubeType{},
			kubeTypesForRawType: map[*types.Type][]KubeType{},
		}
		s.metadataForGV[simpleGV.String()] = existing
		s.metadata = append(s.metadata, existing)
	} else if gv.Group != existing.GroupVersion().Group {
		glog.Errorf("Overlapping packages for Group/Versions %s and %s", gv, existing.GroupVersion())
		return nil
	}
	return existing
}

func (s *metadataStore) AllMetadata() []PackageMetadata {
	return s.metadata
}

func (s *metadataStore) Validate() []error {
	errorSlice := []error{}
	for _, pm := range s.metadata {
		errorSlice = append(errorSlice, pm.Validate()...)
	}
	return errorSlice
}

func (s *metadataStore) createTargetPackage(gv *schema.GroupVersion) *types.Package {
	groupPath := strings.SplitN(gv.Group, ".", 2)
	targetPackage := s.universe.Package(filepath.Join(s.baseOutputPackage.Path, groupPath[0], gv.Version))
	targetPackage.Name = gv.Version
	return targetPackage
}

func (m *packageMetadata) GroupVersion() *schema.GroupVersion {
	return m.groupVersion
}

func (m *packageMetadata) TargetPackage() *types.Package {
	return m.targetPackage
}

func (m *packageMetadata) RawTypes() []*types.Type {
	return m.rawTypes
}

func (m *packageMetadata) KubeTypes(rawType *types.Type) []KubeType {
	return m.kubeTypesForRawType[rawType]
}

func (m *packageMetadata) AllKubeTypes() []KubeType {
	return m.allKubeTypes
}

func (m *packageMetadata) AddMetadataForType(rawType *types.Type, kubeTypes ...KubeType) error {
	if _, exists := m.kubeTypesForRawType[rawType]; exists {
		// we should never get here.  this means we've scanned a type twice
		return fmt.Errorf("type %s already added to scanned metadata", rawType)
	}

	m.rawTypes = append(m.rawTypes, rawType)
	m.kubeTypesForRawType[rawType] = kubeTypes
	m.allKubeTypes = append(m.allKubeTypes, kubeTypes...)

	return nil
}

func (m *packageMetadata) Validate() []error {
	duplicates := map[*types.Type][]*types.Type{}

	// check for duplicates
	sort.Slice(m.allKubeTypes, func(i, j int) bool {
		comp := strings.Compare(m.allKubeTypes[i].Type().Name.Name, m.allKubeTypes[j].Type().Name.Name)
		if comp == 0 {
			if sources, exists := duplicates[m.allKubeTypes[i].Type()]; exists {
				duplicates[m.allKubeTypes[i].Type()] = append(sources, m.allKubeTypes[i].RawType())
			} else {
				duplicates[m.allKubeTypes[i].Type()] = []*types.Type{m.allKubeTypes[j].RawType(), m.allKubeTypes[i].RawType()}
			}
			return false
		}
		return comp < 0
	})
	if len(duplicates) > 0 {
		errorSlice := make([]error, 0, len(duplicates))
		for kubeType, rawTypes := range duplicates {
			errorSlice = append(errorSlice, fmt.Errorf("duplicate kube types specified for %s.  duplicated by: %v", kubeType, rawTypes))
		}
		return errorSlice
	}
	return []error{}
}

// NewKubeType returns a new KubeType object representing the source type, the target type and its comment tags
func NewKubeType(rawType *types.Type, kubeType *types.Type, tags []string) KubeType {
	return &kubeTypeMetadata{
		rawType:  rawType,
		kubeType: kubeType,
		tags:     tags,
	}
}

func (k *kubeTypeMetadata) RawType() *types.Type {
	return k.rawType
}

func (k *kubeTypeMetadata) Type() *types.Type {
	return k.kubeType
}

func (k *kubeTypeMetadata) Tags() []string {
	return k.tags
}

func (k *kubeTypeMetadata) String() string {
	return fmt.Sprintf("%s => %s%s", k.rawType, k.kubeType, k.tags)
}
