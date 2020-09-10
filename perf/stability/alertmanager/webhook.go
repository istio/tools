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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// common variables shared between different monitors for one single test
var (
	client      *spanner.Client
	promclient  api.Client
	v1api       v1.API
	projectID   string
	instance    string
	dbName      string
	clusterName string
	branch      string
	testID      string
)

const (
	prometheusAddr = "http://istio-prometheus.istio-prometheus:9090"
	healthyStatus  = "HEALTHY"
	alertingStatus = "ALERTING"
)

// SingleMonitorStatus represents the status of one single monitor
type SingleMonitorStatus struct {
	Name        string
	Status      string
	Labels      map[string]string
	Annotations string
}

func initPromClient(host string) {
	var err error
	promclient, err = api.NewClient(api.Config{
		Address: host,
	})
	if err != nil {
		log.Fatalf("Error creating client: %v\n", err)
	}
	v1api = v1.NewAPI(promclient)
}

func initSpanner() *spanner.Client {
	projectID = os.Getenv("PROJECT_ID")
	instance = os.Getenv("INSTANCE")
	dbName = os.Getenv("DBNAME")
	clusterName = os.Getenv("CLUSTER_NAME")
	branch = os.Getenv("BRANCH")
	testID = os.Getenv("TESTID")

	ctx := context.Background()
	var err error
	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instance, dbName)
	log.Printf("initializing spanner db: %s\n", db)
	client, err = spanner.NewClient(ctx, db)
	if err != nil {
		log.Fatalf("failed to create spanner client: %v", err)
	}
	return client
}

// initMonitorStatus writes initial MonitorStatus to spanner db.
func initMonitorStatus() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("checking prometheus rules")
	rules, err := v1api.Rules(ctx)
	if err != nil {
		log.Fatalf("error querying Prometheus for alerts: %v\n", err)
	}
	var monitorList []SingleMonitorStatus
	for _, gs := range rules.Groups {
		for _, rs := range gs.Rules {
			switch v := rs.(type) {
			case v1.RecordingRule:
				continue
			case v1.AlertingRule:
				fmt.Printf("adding alerting rule: %s\n", v.Name)
				status := healthyStatus
				if len(v.Alerts) != 0 {
					status = alertingStatus
				}
				monitorList = append(monitorList, SingleMonitorStatus{
					Annotations: v.Annotations.String(),
					Name:        v.Name,
					Status:      status,
				})
			default:
				fmt.Printf("unknown rule type %s", v)
			}
		}
	}
	log.Println("writing initial monitor status to Spanner")
	if err := writeMonitorStatusToDB(monitorList); err != nil {
		log.Fatalf("failed to initialize monitor status in Spanner: %v", err)
	}
}

func webhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	log.Println("handling alert webhook")
	data := template.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("alerts: GroupLabels=%v, CommonLabels=%v", data.GroupLabels, data.CommonLabels)
	var monitorList []SingleMonitorStatus
	var errs error
	for _, alert := range data.Alerts {
		ms, err := convertPromAlertToInternalMonitorStatus(alert)
		if err != nil {
			errs = multierror.Append(errs, err)
		} else {
			monitorList = append(monitorList, ms)
		}
	}
	if errs != nil {
		log.Fatalf("failed to convert prom alert to internal monitor: %v", errs)
	}
	if err := writeMonitorStatusToDB(monitorList); err != nil {
		log.Fatalf("failed to write alert to db: %v", err)
	}
	fmt.Fprint(w, "Ok!")
}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Ok!")
}

// writeMonitorStatusToDB is helper function to convert monitorStatus and write to spanner
func writeMonitorStatusToDB(ms []SingleMonitorStatus) error {
	ctx := context.Background()
	monitorColumns := []string{"MonitorName", "Status", "ProjectID", "ClusterName", "Branch", "UpdatedTime", "TestID"}
	curTime := time.Now()
	var m []*spanner.Mutation

	for _, sms := range ms {
		alertName := sms.Name
		if alertName == "" {
			var ok bool
			if alertName, ok = sms.Labels["alertname"]; !ok {
				return fmt.Errorf("no alertname found from the labels")
			}
		}
		log.Printf("Writing Alert status to Spanner: name=%s, status=%s,Labels=%v,Annotations=%v\n",
			alertName, sms.Status, sms.Labels, sms.Annotations)
		m = append(m, spanner.InsertOrUpdate(dbName, monitorColumns,
			[]interface{}{alertName, sms.Status, projectID, clusterName, branch, curTime, testID}))
	}
	if _, err := client.Apply(ctx, m); err != nil {
		return err
	}
	return nil
}

// convertPromAlertToInternalMonitorStatus is helper function to convert from prometheus Alert to internal SingleMonitorStatus struct.
func convertPromAlertToInternalMonitorStatus(alert template.Alert) (SingleMonitorStatus, error) {
	var sms SingleMonitorStatus
	labels := alert.Labels
	alertName, ok := labels["alertname"]
	if !ok {
		return sms, fmt.Errorf("no alertname found from the labels")
	}
	sms.Name = alertName
	sms.Status = alert.Status
	sms.Labels = alert.Labels
	sms.Annotations = strings.Join(alert.Annotations.Values(), ", ")
	return sms, nil
}

func main() {
	client := initSpanner()
	defer client.Close()
	initPromClient(prometheusAddr)
	initMonitorStatus()
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/webhook", webhook)
	listenAddress := ":5001"
	log.Printf("listening on: %v", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
