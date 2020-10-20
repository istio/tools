// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protomodel

import (
	"fmt"
	"os"
	"sort"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// PackageDescriptor describes a package, which is a composition of proto files.
type PackageDescriptor struct {
	baseDesc
	Files []*FileDescriptor
	Name  string
}

func newPackageDescriptor(name string, desc []*descriptor.FileDescriptorProto, perFile bool) *PackageDescriptor {
	p := &PackageDescriptor{
		Name: name,
	}

	for _, fd := range desc {
		f := newFileDescriptor(fd, p)
		p.Files = append(p.Files, f)

		// The package's file is one that documents the package statement.
		// The first file to do this "wins".
		loc := f.find(newPathVector(packagePath))
		if loc != nil {
			if p.loc == nil {
				if loc.GetLeadingComments() != "" || loc.GetTrailingComments() != "" {
					p.loc = loc
					p.file = f
					// Inherit only f's frontMatter, don't get title from one file
				}
			} else if !perFile {
				leading := loc.GetLeadingComments()
				trailing := loc.GetTrailingComments()
				if leading != "" || trailing != "" {
					_, _ = fmt.Fprintf(os.Stderr, "WARNING: package %v has a conflicting package comment in file %v.\n",
						name, f.GetName())
					_, _ = fmt.Fprintf(os.Stderr, "Previous:\n%v\n%v\nCurrent:\n%v\n%v\n", p.loc.GetLeadingComments(), p.loc.GetTrailingComments(), leading, trailing)
				}
			}
		}
	}

	// Make the documenting file first. This ensures its messages show up first on the output
	// If we need more granularity with explicitly multiple file ordering, we can add some sort of annotation,
	// but or now this is sufficient.
	sort.SliceStable(p.Files, func(i, j int) bool {
		return p.Files[i] == p.file
	})

	return p
}
