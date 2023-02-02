/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"

	"github.com/apache/incubator-uniffle/deploy/kubernetes/operator/api/uniffle/v1alpha1"
)

func TestGenerateMakeDataDirCommand(t *testing.T) {
	for _, tt := range []struct {
		name             string
		rssPodSpec       *v1alpha1.RSSPodSpec
		expectedCommands []string
	}{
		{
			name: "empty security context",
			rssPodSpec: &v1alpha1.RSSPodSpec{
				HostPathMounts: map[string]string{
					"/data1": "/mnt/data1",
				},
			},
			expectedCommands: []string{
				"chown -R 0:0 /mnt/data1",
			},
		},
		{
			name: "empty runAsUser field in security context",
			rssPodSpec: &v1alpha1.RSSPodSpec{
				HostPathMounts: map[string]string{
					"/data2": "/mnt/data2",
				},
				SecurityContext: &corev1.PodSecurityContext{
					FSGroup: pointer.Int64(1000),
				},
			},
			expectedCommands: []string{
				"chown -R 0:1000 /mnt/data2",
			},
		},
		{
			name: "non empty field of runAsUser and fsGroup in security context",
			rssPodSpec: &v1alpha1.RSSPodSpec{
				HostPathMounts: map[string]string{
					"/data3": "/mnt/data3",
				},
				SecurityContext: &corev1.PodSecurityContext{
					RunAsUser: pointer.Int64(2000),
					FSGroup:   pointer.Int64(1000),
				},
			},
			expectedCommands: []string{
				"chown -R 2000:1000 /mnt/data3",
			},
		},
	} {
		t.Run(tt.name, func(tc *testing.T) {
			commands := generateMakeDataDirCommand(tt.rssPodSpec)
			if !isEqualStringSlice(commands, tt.expectedCommands) {
				tc.Errorf("unexpected commands: %+v, expected: %+v", commands, tt.expectedCommands)
				return
			}
		})
	}
}

func TestGenerateInitContainers(t *testing.T) {
	// first check resource request
	for _, tt := range []struct {
		name       string
		rssPodSpec *v1alpha1.RSSPodSpec
		resources  *corev1.ResourceRequirements
	}{
		{
			name:       "without security context",
			rssPodSpec: &v1alpha1.RSSPodSpec{},
			resources:  nil,
		},
		{
			name: "security context with host path mapping",
			rssPodSpec: &v1alpha1.RSSPodSpec{
				HostPathMounts: map[string]string{
					"/data3": "/mnt/data3",
				},
				SecurityContext: &corev1.PodSecurityContext{
					RunAsUser: pointer.Int64(2000),
					FSGroup:   pointer.Int64(1000),
				},
				MainContainer: &v1alpha1.MainContainer{
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1G"),
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1G"),
						},
					},
				},
			},
			resources: &corev1.ResourceRequirements{
				Limits: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1G"),
				},
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1G"),
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertion := assert.New(t)
			containers := GenerateInitContainers(tt.rssPodSpec)
			if len(containers) == 0 {
				assertion.Nil(tt.resources)
			} else {
				assertion.Equal(tt.resources, &containers[0].Resources)
			}
		})
	}
}

func isEqualStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
