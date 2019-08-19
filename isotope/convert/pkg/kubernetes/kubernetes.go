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
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/tools/isotope/convert/pkg/consts"
	"istio.io/tools/isotope/convert/pkg/graph"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
)

const (
	// ServiceGraphNamespace is the namespace all service graph related resources
	// (i.e. ConfigMap, Services, and Deployments) will reside in.
	ServiceGraphNamespace = "service-graph"

	numConfigMaps          = 1
	numManifestsPerService = 2

	configVolume           = "config-volume"
	serviceGraphConfigName = "service-graph-config"
)

var (
	serviceGraphAppLabels       = map[string]string{"app": "service-graph"}
	serviceGraphNodeLabels      = map[string]string{"role": "service"}
	prometheusScrapeAnnotations = map[string]string{
		"prometheus.io/scrape": "true"}
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
	environmentName string,
	loadLevel int) ([]byte, error) {
	numServices := len(serviceGraph.Services)
	numManifests := numManifestsPerService*numServices + numConfigMaps
	manifests := make([]string, 0, numManifests)

	appendManifest := func(manifest interface{}) error {
		yamlDoc, err := yaml.Marshal(manifest)
		if err != nil {
			return err
		}
		manifests = append(manifests, string(yamlDoc))
		return nil
	}

	namespace := makeServiceGraphNamespace()
	if err := appendManifest(namespace); err != nil {
		return nil, err
	}

	configMap, err := makeConfigMap(serviceGraph)
	if err != nil {
		return nil, err
	}
	if err := appendManifest(configMap); err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UTC().UnixNano())
	hasRbacPolicy := false
	for _, service := range serviceGraph.Services {
		k8sDeployment := makeDeployment(
			service, serviceNodeSelector, serviceImage,
			serviceMaxIdleConnectionsPerHost,
			loadLevel)
		innerErr := appendManifest(k8sDeployment)
		if innerErr != nil {
			return nil, innerErr
		}

		k8sService := makeService(service)
		innerErr = appendManifest(k8sService)
		if innerErr != nil {
			return nil, innerErr
		}

		// Only generates the RBAC rules when Istio is installed.
		if strings.EqualFold(environmentName, "ISTIO") && service.NumRbacPolicies > 0 {
			hasRbacPolicy = true
			var i int32
			// Generates random RBAC rules for the service.
			for i = 0; i < service.NumRbacPolicies; i++ {
				manifests = append(manifests, generateRbacPolicy(service, false /* allowAll */))
			}
			// Generates "allow-all" RBAC rule for the service.
			manifests = append(manifests, generateRbacPolicy(service, true /* allowAll */))
		}
	}

	fortioDeployment := makeFortioDeployment(
		clientNodeSelector, clientImage)
	if err := appendManifest(fortioDeployment); err != nil {
		return nil, err
	}

	fortioService := makeFortioService()
	if err := appendManifest(fortioService); err != nil {
		return nil, err
	}

	if hasRbacPolicy {
		manifests = append(manifests, generateRbacConfig())
	}

	yamlDocString := strings.Join(manifests, "---\n")
	return []byte(yamlDocString), nil
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

func makeServiceGraphNamespace() (namespace apiv1.Namespace) {
	namespace.APIVersion = "v1"
	namespace.Kind = "Namespace"
	namespace.ObjectMeta.Name = consts.ServiceGraphNamespace
	namespace.ObjectMeta.Labels = map[string]string{"istio-injection": "enabled"}
	timestamp(&namespace.ObjectMeta)
	return
}

func makeConfigMap(
	graph graph.ServiceGraph) (configMap apiv1.ConfigMap, err error) {
	graphYAMLBytes, err := yaml.Marshal(graph)
	if err != nil {
		return
	}
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.ObjectMeta.Name = serviceGraphConfigName
	configMap.ObjectMeta.Namespace = ServiceGraphNamespace
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
	k8sService.ObjectMeta.Namespace = ServiceGraphNamespace
	k8sService.ObjectMeta.Labels = serviceGraphAppLabels
	timestamp(&k8sService.ObjectMeta)
	k8sService.Spec.Ports = []apiv1.ServicePort{{Port: consts.ServicePort, Name: consts.ServicePortName}}
	k8sService.Spec.Selector = map[string]string{"name": service.Name}
	return
}

func makeDeployment(
	service svc.Service, nodeSelector map[string]string,
	serviceImage string, serviceMaxIdleConnectionsPerHost int,
	loadLevel int) (
	k8sDeployment appsv1.Deployment) {

	k8sDeployment.APIVersion = "apps/v1"
	k8sDeployment.Kind = "Deployment"
	k8sDeployment.ObjectMeta.Name = service.Name
	k8sDeployment.ObjectMeta.Namespace = ServiceGraphNamespace
	k8sDeployment.ObjectMeta.Labels = serviceGraphAppLabels
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
							{Name: consts.LoadEnvKey, Value: strconv.Itoa(loadLevel)},
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
