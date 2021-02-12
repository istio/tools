// Copyright 2018 Istio Authors
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
	"fmt"
	"net/http"
	"sync"
)

var wg = sync.WaitGroup{}

func main() {
	// Simple server that listens on many ports
	wg.Add(6)
	go listen(":8085")
	go listen(":8086")
	go listen(":8087")

	go listen(":9085")
	go listen(":9086")
	go listen(":9087")
	wg.Wait()
}

func listen(addr string) {
	defer wg.Done()
	fmt.Println("Connecting to", addr)
	server := http.NewServeMux()
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world!\n")
	})
	if err := http.ListenAndServe(addr, server); err != nil {
		panic(err.Error())
	}
}
