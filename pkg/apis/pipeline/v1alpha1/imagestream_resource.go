/*
Copyright 2018 The Knative Authors.

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

package v1alpha1

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const os4Registry = "image-registry.openshift-image-registry.svc:5000"

// NewImageStreamResource creates a new NewImageStreamResource from a PipelineResource.
func NewImageStreamResource(r *PipelineResource) (*ImageStreamResource, error) {
	if r.Spec.Type != PipelineResourceTypeIS {
		return nil, fmt.Errorf("ImageStreamResource: Cannot create an ImageStream resource from a %s Pipeline Resource", r.Spec.Type)
	}

	if r.Namespace == "" {
		return nil, fmt.Errorf("ImageStreamResource: Cannot create an ImageStream resource from a %s Pipeline Resource. Namespaces is missing from PipelineResource metadata", r.Name)
	}

	isr := &ImageStreamResource{
		Type: PipelineResourceTypeIS,
		Ns:   r.Namespace,
	}

	for _, param := range r.Spec.Params {
		switch {
		case strings.EqualFold(param.Name, "name"):
			isr.Name = param.Value
			break
		}
	}

	return isr, nil
}

// NewImageStreamResource generates an endpoint where images can be stored in OpenShift.
type ImageStreamResource struct {
	Name string               `json:"name"`
	Type PipelineResourceType `json:"type"`
	Ns   string               `json:"ns"`
}

// GetName returns the name of the resource
func (s ImageStreamResource) GetName() string {
	return s.Name
}

// GetType returns the type of the resource, in this case "image"
func (s ImageStreamResource) GetType() PipelineResourceType {
	return PipelineResourceTypeIS
}

// GetParams returns the resoruce params
func (s ImageStreamResource) GetParams() []Param { return []Param{} }

// Replacements is used for template replacement on an ImageStreamResource inside of a Taskrun.
func (s *ImageStreamResource) Replacements() map[string]string {

	return map[string]string{
		"name": os4Registry + "/" + s.Ns + "/" + s.Name,
		"type": string(s.Type),
	}
}

func (s *ImageStreamResource) GetUploadContainerSpec() ([]corev1.Container, error) {
	return nil, nil
}
func (s *ImageStreamResource) GetDownloadContainerSpec() ([]corev1.Container, error) {
	return nil, nil
}
func (s *ImageStreamResource) SetDestinationDirectory(path string) {
}
