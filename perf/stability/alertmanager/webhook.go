// Copyright 2020 Istio Authors. All Rights Reserved.
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
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/alertmanager/template"
)

func webhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data := template.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Fatalf(err.Error())
		return
	}
	log.Printf("Alerts: GroupLabels=%v, CommonLabels=%v", data.GroupLabels, data.CommonLabels)
	//TODO(richardwxn): process alert
	for _, alert := range data.Alerts {
		log.Printf("Alert: status=%s,Labels=%v,Annotations=%v", alert.Status, alert.Labels, alert.Annotations)
	}
	fmt.Fprint(w, "Ok!")
}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Ok!")
}

func main() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/webhook", webhook)
	listenAddress := ":5001"
	log.Printf("listening on: %v", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}