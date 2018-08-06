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
	"net/http"
)

var (
	forwardableHeaders = []string{
		"X-Request-Id",
		"X-B3-Traceid",
		"X-B3-Spanid",
		"X-B3-Parentspanid",
		"X-B3-Sampled",
		"X-B3-Flags",
		"X-Ot-Span-Context",
	}
	forwardableHeadersSet = make(map[string]bool, len(forwardableHeaders))
)

func init() {
	for _, key := range forwardableHeaders {
		forwardableHeadersSet[key] = true
	}
}

func extractForwardableHeader(header http.Header) http.Header {
	forwardableHeader := make(http.Header, len(forwardableHeaders))
	for key := range forwardableHeadersSet {
		if values, ok := header[key]; ok {
			forwardableHeader[key] = values
		}
	}
	return forwardableHeader
}
