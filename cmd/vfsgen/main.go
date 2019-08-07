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

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	inputDir := flag.String("i", "", "input directory to process")
	outputFile := flag.String("o", "", "output go file to produce")
	packageName := flag.String("p", "", "name of package in generated go output")
	variableName := flag.String("v", "", "name of variable in generated go output")
	flag.Parse()

	if *inputDir == "" || *outputFile == "" || *packageName == "" || *variableName == "" {
		log.Fatalf("Missing argument")
	}

	templates := http.Dir(*inputDir)
	if err := vfsgen.Generate(templates, vfsgen.Options{
		Filename:     *outputFile,
		PackageName:  *packageName,
		VariableName: *variableName,
	}); err != nil {
		log.Fatalln("vfsgen failed to generate code: ", err)
	}
}
