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
	"sync/atomic"
	"time"

	"istio.io/pkg/log"

	"istio.io/tools/isotope/convert/pkg/graph/svc"
	"istio.io/tools/isotope/convert/pkg/graph/svctype"
	"istio.io/tools/isotope/service/pkg/srv/prometheus"
)

// Handler handles the default endpoint by emulating its Service.
type Handler struct {
	Service      svc.Service
	ServiceTypes map[string]svctype.ServiceType

	responsePayload []byte
	counter         uint64
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	startTime := time.Now()

	prometheus.RecordRequestReceived()

	respond := func(status int, body string) {
		writer.WriteHeader(status)
		if body != "" {
			if _, err := writer.Write([]byte(body)); err != nil {
				log.Errorf("%s", err)
			}
		} else {
			if _, err := writer.Write(h.responsePayload); err != nil {
				log.Errorf("%s", err)
			}
		}

		stopTime := time.Now()
		duration := stopTime.Sub(startTime)
		prometheus.RecordResponseSent(duration, len(h.responsePayload), status)
	}

	// Simulate failure based on the error percentage
	reqCount := atomic.AddUint64(&h.counter, 1)
	if h.Service.ErrorRate > 0 && reqCount >= (10000/uint64(h.Service.ErrorRate*10000)) {
		atomic.StoreUint64(&h.counter, 0) // atomic.CompareAndSwap potentially collides and never resolves to true
		log.Debug("Provoking simulated failure")
		respond(http.StatusInternalServerError, "simulated failure")
		return
	}

	for _, step := range h.Service.Script {
		forwardableHeader := extractForwardableHeader(request.Header)
		err := execute(step, forwardableHeader, h.ServiceTypes)
		if err != nil {
			log.Errorf("%s", err)
			respond(http.StatusInternalServerError, err.Error()+"\n")
			return
		}
	}

	respond(http.StatusOK, "")
}
