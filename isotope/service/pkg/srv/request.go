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

package srv

import (
	"bytes"
	"fmt"
	"net/http"

	"fortio.org/fortio/log"

	"istio.io/tools/isotope/convert/pkg/consts"
	"istio.io/tools/isotope/convert/pkg/graph/size"
)

func sendRequest(
	destName string,
	size size.ByteSize,
	requestHeader http.Header) (*http.Response, error) {
	url := fmt.Sprintf("http://%s:%v", destName, consts.ServicePort)
	request, err := buildRequest(url, size, requestHeader)
	if err != nil {
		return nil, err
	}
	log.Debugf("sending request to %s (%s)", destName, url)
	return http.DefaultClient.Do(request)
}

func buildRequest(
	url string, size size.ByteSize, requestHeader http.Header) (
	*http.Request, error) {
	payload, err := makeRandomByteArray(size)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	copyHeader(request, requestHeader)
	return request, nil
}

func copyHeader(request *http.Request, header http.Header) {
	for key, values := range header {
		request.Header[key] = values
	}
}
