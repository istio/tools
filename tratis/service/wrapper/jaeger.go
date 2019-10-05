// Copyright 2019 Istio Authors
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

package wrapper

/*
This package allows one to query the distributed tracers (jaeger, zipkin) API.
*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

/*
The ExtractTraces function uses the Jaeger API to extract
the raw trace data.

API Documentation: https://www.jaegertracing.io/docs/1.13/apis/

NOTE: If Jaeger decides to change the jaeger endpoint, this
function would break.
*/
func ExtractJaegerTraces(toolAddr string, toolPortNum string,
	appEntryPoint string, numTraces int) []byte {
	pageAddress := fmt.Sprintf("http://%s:%s/jaeger/api/traces?service=%s&limit=%d",
		toolAddr, toolPortNum, appEntryPoint, numTraces)
	return RequestAPI(pageAddress)
}

func RequestAPI(pageAddress string) []byte {
	resp, err := http.Get(pageAddress)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalln(http.StatusText(resp.StatusCode))
	} else if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return (body)
}
