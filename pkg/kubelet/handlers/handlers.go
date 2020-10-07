/*
Copyright 2020 The Kubernetes Authors.

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
package handlers

/// AllHandlerReasons a collection of all internal eviction handlers
var AllHandlerReasons []string = []string{
	ActiveDeadlineReason,
}

/// IsEvictionReason is the reason an internal eviction reason
func IsEvictionReason(reason string) bool {
	for i, _ := range AllHandlerReasons {
		if reason == AllHandlerReasons[i] {
			return true
		}
	}
	return false
}
