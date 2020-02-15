// Copyright (c) 2018-2020 Splunk Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deploy

import (
	"testing"

	enterprisev1 "github.com/splunk/splunk-operator/pkg/apis/enterprise/v1alpha2"
	"github.com/splunk/splunk-operator/pkg/splunk/enterprise"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReconcileSplunkConfig(t *testing.T) {
	funcCalls := []mockFuncCall{
		{metaName: "*v1.Secret-test-splunk-stack1-search-head-secrets"},
		{metaName: "*v1.ConfigMap-test-splunk-stack1-search-head-defaults"},
	}
	createCalls := map[string][]mockFuncCall{"Get": funcCalls, "Create": funcCalls}
	updateCalls := map[string][]mockFuncCall{"Get": funcCalls}
	current := enterprisev1.SearchHead{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stack1",
			Namespace: "test",
		},
	}
	current.Spec.Defaults = "defaults-yaml"
	revised := current.DeepCopy()
	revised.Spec.Image = "splunk/test"
	reconcile := func(c *mockClient, cr interface{}) error {
		idx := cr.(*enterprisev1.SearchHead)
		spec := idx.Spec.CommonSplunkSpec
		return ReconcileSplunkConfig(c, idx, spec, enterprise.SplunkSearchHead)
	}
	reconcileTester(t, "TestReconcileSplunkConfig", &current, revised, createCalls, updateCalls, reconcile)

	revised.Spec.IndexerRef.Name = "stack2"
	updateCalls["Get"] = []mockFuncCall{
		{metaName: "*v1.Secret-test-splunk-stack2-indexer-secrets"},
		{metaName: "*v1.Secret-test-splunk-stack1-search-head-secrets"},
		{metaName: "*v1.ConfigMap-test-splunk-stack1-search-head-defaults"},
	}
	reconcileTester(t, "TestReconcileSplunkConfig", &current, revised, createCalls, updateCalls, reconcile)
}

func TestApplyConfigMap(t *testing.T) {
	funcCalls := []mockFuncCall{{metaName: "*v1.ConfigMap-test-defaults"}}
	createCalls := map[string][]mockFuncCall{"Get": funcCalls, "Create": funcCalls}
	updateCalls := map[string][]mockFuncCall{"Get": funcCalls, "Update": funcCalls}
	current := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "defaults",
			Namespace: "test",
		},
	}
	revised := current.DeepCopy()
	revised.Data = map[string]string{"a": "b"}
	reconcile := func(c *mockClient, cr interface{}) error {
		return ApplyConfigMap(c, cr.(*corev1.ConfigMap))
	}
	reconcileTester(t, "TestApplyConfigMap", &current, revised, createCalls, updateCalls, reconcile)
}

func TestApplySecret(t *testing.T) {
	funcCalls := []mockFuncCall{{metaName: "*v1.Secret-test-secrets"}}
	createCalls := map[string][]mockFuncCall{"Get": funcCalls, "Create": funcCalls}
	updateCalls := map[string][]mockFuncCall{"Get": funcCalls}
	current := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secrets",
			Namespace: "test",
		},
	}
	revised := current.DeepCopy()
	revised.Data = map[string][]byte{"a": []byte{'1', '2'}}
	reconcile := func(c *mockClient, cr interface{}) error {
		return ApplySecret(c, cr.(*corev1.Secret))
	}
	reconcileTester(t, "TestApplySecret", &current, revised, createCalls, updateCalls, reconcile)
}
