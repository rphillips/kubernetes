/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package managed

import (
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

var (
	pinnedManagementEnabled      bool
	pinnedManagementFilename     = "/etc/kubernetes/workload-pinning"
	WorkloadsAnnotation          = "workload.openshift.io/management"
	ManagedCapacityLabel         = "management.workload.openshift.io/cores"
	ContainerCpuAnnotationFormat = "io.openshift.workload.management.cpushares/%v"
)

func init() {
	readEnablementFile()
}

func readEnablementFile() {
	if _, err := os.Stat(pinnedManagementFilename); err == nil {
		klog.V(2).Info("Pinned Workload Management Enabled")
		pinnedManagementEnabled = true
	}
}

func IsEnabled() bool {
	return pinnedManagementEnabled
}

func IsPodManaged(pod *v1.Pod) bool {
	if pod.ObjectMeta.Annotations == nil {
		return false
	}
	_, found := pod.ObjectMeta.Annotations[WorkloadsAnnotation]
	return found
}

func ModifyStaticPodForPinnedManagement(pod *v1.Pod) {
	if !IsPodManaged(pod) {
		return
	}
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Annotations[WorkloadsAnnotation] = ""
	updateContainers(pod, pod.Spec.Containers)
	updateContainers(pod, pod.Spec.InitContainers)
}

func updateContainers(pod *v1.Pod, containers []v1.Container) {
	for _, container := range containers {
		if _, ok := container.Resources.Requests[v1.ResourceCPU]; !ok {
			continue
		}
		cpuRequest := container.Resources.Requests[v1.ResourceCPU]
		cpuRequestInMilli := cpuRequest.MilliValue()
		resourceLimit := fmt.Sprintf("%v", MilliCPUToShares(cpuRequestInMilli))

		containerNameKey := fmt.Sprintf(ContainerCpuAnnotationFormat, container.Name)
		pod.Annotations[containerNameKey] = resourceLimit

		newCPURequest := resource.NewMilliQuantity(cpuRequestInMilli*1000, cpuRequest.Format)
		container.Resources.Requests[v1.ResourceName(ManagedCapacityLabel)] = *newCPURequest

		delete(container.Resources.Requests, v1.ResourceCPU)
	}
}
