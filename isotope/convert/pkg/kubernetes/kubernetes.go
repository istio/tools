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

// Package kubernetes converts service graphs into Kubernetes manifests.
package kubernetes

import (
	"errors"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"istio.io/tools/isotope/convert/pkg/consts"
	"istio.io/tools/isotope/convert/pkg/graph"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
)

const (
	numConfigMaps          = 1
	numManifestsPerService = 2

	configVolume           = "config-volume"
	serviceGraphConfigName = "service-graph-config"
)

var (
	serviceGraphAppLabels       = map[string]string{"isotope": "service-graph"}
	serviceGraphNodeLabels      = map[string]string{"role": "service"}
	prometheusScrapeAnnotations = map[string]string{
		"prometheus.io/scrape": "true",
	}
)

// ServiceGraphToKubernetesManifests converts a ServiceGraph to Kubernetes
// manifests.
func ServiceGraphToKubernetesManifests(
	serviceGraph graph.ServiceGraph,
	serviceNodeSelector map[string]string,
	serviceImage string,
	serviceMaxIdleConnectionsPerHost int,
	clientNodeSelector map[string]string,
	clientImage string,
	clientNamespace string,
	environmentName string,
	clusterName string,
	clientDisabled bool) ([]byte, error) {
	numServices := len(serviceGraph.Services)
	numManifests := numManifestsPerService*numServices + numConfigMaps
	manifests := make([]string, 0, numManifests)
	namespaces := []string{}

	appendManifest := func(manifest interface{}) error {
		yamlDoc, err := yaml.Marshal(manifest)
		if err != nil {
			return err
		}
		manifests = append(manifests, string(yamlDoc))
		return nil
	}

	manifestHeader, err := validateServices(clusterName, serviceGraph.Services, manifests)
	if err != nil {
		return nil, err
	}
	manifests = append(manifests, string(manifestHeader))

	// Find all the namespaces with the given cluster
	for _, service := range serviceGraph.Services {
		var ommit = false
		for _, x := range namespaces {
			if x == service.Namespace {
				ommit = true
				break
			}
		}
		if ommit == false && service.Cluster == clusterName {
			namespaces = append(namespaces, service.Namespace)
		}
	}

	if len(namespaces) == 0 {
		configMap, err := makeConfigMap(serviceGraph, "")
		if err != nil {
			return nil, err
		}
		if err := appendManifest(configMap); err != nil {
			return nil, err
		}
	} else {

		for _, namespace := range namespaces {
			configMap, err := makeConfigMap(serviceGraph, namespace)
			if err != nil {
				return nil, err
			}
			if err := appendManifest(configMap); err != nil {
				return nil, err
			}
		}
	}

	for _, service := range serviceGraph.Services {
		if service.Cluster == clusterName || clusterName == "" {
			k8sDeployment := makeDeployment(
				service, serviceNodeSelector, serviceImage,
				serviceMaxIdleConnectionsPerHost)
			innerErr := appendManifest(k8sDeployment)
			if innerErr != nil {
				return nil, innerErr
			}

			k8sService := makeService(service)
			innerErr = appendManifest(k8sService)
			if innerErr != nil {
				return nil, innerErr
			}
		}
	}

	if !clientDisabled {
		fortioDeployment := makeFortioDeployment(
			clientNodeSelector, clientImage, clientNamespace)
		if err := appendManifest(fortioDeployment); err != nil {
			return nil, err
		}

		fortioService := makeFortioService(clientNamespace)
		if err := appendManifest(fortioService); err != nil {
			return nil, err
		}
	}

	yamlDocString := strings.Join(manifests, "---\n")
	return []byte(yamlDocString), nil
}

func validateServices(clusterName string, services []svc.Service, manifest interface{}) ([]byte, error) {
	header := []byte("")
	if clusterName == "" {
		header = append(header, []byte("## WARNING: Cluster name is not supplied. All services will be included in this manifest\n\n")...)
	}
	serviceToAppendCounter := 0
	for _, service := range services {
		if clusterName == service.Cluster || clusterName == "" {
			serviceToAppendCounter++
		}
	}
	if serviceToAppendCounter == 0 {
		return nil, errors.New(fmt.Sprintf("No services found to match clusterName: '%s'", clusterName))
	}
	return append(header, []byte(fmt.Sprintf("## Number of services included in this manifest for cluster '%s' is: %d\n\n", clusterName, serviceToAppendCounter))...), nil
}

func combineLabels(a, b map[string]string) map[string]string {
	c := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		c[k] = v
	}
	for k, v := range b {
		c[k] = v
	}
	return c
}

func makeConfigMap(
	graph graph.ServiceGraph, namespace string) (configMap apiv1.ConfigMap, err error) {
	graphYAMLBytes, err := yaml.Marshal(graph)
	if err != nil {
		return
	}
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.ObjectMeta.Name = serviceGraphConfigName
	configMap.ObjectMeta.Namespace = namespace
	configMap.ObjectMeta.Labels = serviceGraphAppLabels
	timestamp(&configMap.ObjectMeta)
	configMap.Data = map[string]string{
		consts.ServiceGraphConfigMapKey: string(graphYAMLBytes),
	}
	return
}

func makeService(service svc.Service) (k8sService apiv1.Service) {
	k8sService.APIVersion = "v1"
	k8sService.Kind = "Service"
	k8sService.ObjectMeta.Name = service.Name
	k8sService.ObjectMeta.Namespace = service.Namespace
	k8sService.ObjectMeta.Labels = combineLabels(
		serviceGraphNodeLabels,
		map[string]string{
			"app": service.Name,
		})
	timestamp(&k8sService.ObjectMeta)
	k8sService.Spec.Ports = []apiv1.ServicePort{{Port: consts.ServicePort, Name: consts.ServicePortName}}
	k8sService.Spec.Selector = map[string]string{"name": service.Name}
	return
}

func makeDeployment(
	service svc.Service, nodeSelector map[string]string,
	serviceImage string, serviceMaxIdleConnectionsPerHost int) (
	k8sDeployment appsv1.Deployment) {
	k8sDeployment.APIVersion = "apps/v1"
	k8sDeployment.Kind = "Deployment"
	k8sDeployment.ObjectMeta.Name = service.Name
	k8sDeployment.ObjectMeta.Namespace = service.Namespace
	k8sDeployment.ObjectMeta.Labels = combineLabels(
		serviceGraphNodeLabels,
		map[string]string{
			"app": service.Name,
		})
	timestamp(&k8sDeployment.ObjectMeta)
	k8sDeployment.Spec = appsv1.DeploymentSpec{
		Replicas: &service.NumReplicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name": service.Name,
			},
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: combineLabels(
					serviceGraphNodeLabels,
					map[string]string{
						"name": service.Name,
					}),
				Annotations: prometheusScrapeAnnotations,
			},
			Spec: apiv1.PodSpec{
				NodeSelector: nodeSelector,
				Containers: []apiv1.Container{
					{
						Name:  consts.ServiceContainerName,
						Image: serviceImage,
						Args: []string{
							fmt.Sprintf(
								"--max-idle-connections-per-host=%v",
								serviceMaxIdleConnectionsPerHost),
						},
						Env: []apiv1.EnvVar{
							{Name: consts.ServiceNameEnvKey, Value: service.Name},
						},
						VolumeMounts: []apiv1.VolumeMount{
							{
								Name:      configVolume,
								MountPath: consts.ConfigPath,
							},
						},
						Ports: []apiv1.ContainerPort{
							{
								ContainerPort: consts.ServicePort,
							},
						},
					},
				},
				Volumes: []apiv1.Volume{
					{
						Name: configVolume,
						VolumeSource: apiv1.VolumeSource{
							ConfigMap: &apiv1.ConfigMapVolumeSource{
								LocalObjectReference: apiv1.LocalObjectReference{
									Name: serviceGraphConfigName,
								},
								Items: []apiv1.KeyToPath{
									{
										Key:  consts.ServiceGraphConfigMapKey,
										Path: consts.ServiceGraphYAMLFileName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	timestamp(&k8sDeployment.Spec.Template.ObjectMeta)
	return
}

func timestamp(objectMeta *metav1.ObjectMeta) {
	objectMeta.CreationTimestamp = metav1.Time{Time: time.Now()}
}
