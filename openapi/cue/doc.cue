// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file describes the available fields for the genopai configuration file.
// Optional fields are indicated with a question mark.

// module corresponds to the Go or CUE module of a repository, which is the
// directory prefix used by packages within the module to import each other.
// For example, for github.com/istio/api, the module is "istio.io/api".
module: string

openapi?: {
	// selfContained specifies whether all references should be included within
	// the OpenAPI output, recursively.
	selfContained: bool | *false

	// fieldFilter defines a regular expression of all fields to omit from the
	// output. It is only allowed to filter fields that add additional
	// constraints. Fields that indicate basic types cannot be removed. It is
	// an error for such fields to be excluded by this filter.
	// Fields are qualified by their Object type. For instance, the
	// minimum field of the schema object is qualified as Schema/minimum.
	fieldFilter?: string

	// expandReferences replaces references with actual objects when generating
	// OpenAPI Schema. It is an error for an CUE value to refer to itself if
	// this option is used.
	expandReferences: bool | *false
}

// The all section specifies settings for generating an aggregate OpenAPI file
// that contains all schema defined all of the directory entries.
all?: {
	// oapiFilename is the filename for the generated OpenAPI output.
	oapiFilename: =~".json$"

	// title overrides the $description associated with a directory.
	title: string

	version: =~#"^v\d."# | *"v1aplha1"
}

// directories is a map of directories, relative to the root, for which to
// process proto files.
directories <Dir>: [...{
	// mode indicates which proto files to include in generated output.
	//   all      include all proto files
	//   perFile  generate a separate OpenAPI file for each proto file
	mode?: "all" | "perFile"

	// protoFiles manually specified the list of proto files to include.
	protoFiles?: [...string]

	// oapiFilename is the filename for the generated OpenAPI output. genoapi
	// will pick a default name if unspecified.
	oapiFilename?: =~".json$"

	// title overrides the $description associated with a directory.
	title?: !=""

	// version specifies the version of an OpenAPI file. The default the version
	// as specified in the directory path.
	version?: =~#"^v\d."#
}]

crd?: {
	// the output directory of the CRD file.
	dir?: string

	// the output filename of the CRDs.
	fileprefix?: string

	istioversion: string

	// the list of APIs that have CRDs generated and their details.
	crdconfigs? <ProtoName>: [...{}]
}
