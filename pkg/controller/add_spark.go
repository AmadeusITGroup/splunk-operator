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

package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	enterprisev1 "github.com/splunk/splunk-operator/pkg/apis/enterprise/v1beta1"
	splcommon "github.com/splunk/splunk-operator/pkg/splunk/common"
	splctrl "github.com/splunk/splunk-operator/pkg/splunk/controller"
	spark "github.com/splunk/splunk-operator/pkg/splunk/spark"
)

func init() {
	SplunkControllersToAdd = append(SplunkControllersToAdd, SparkController{})
}

// blank assignment to verify that SparkController implements SplunkController
var _ splctrl.SplunkController = &SparkController{}

// SparkController is used to manage Spark custom resources
type SparkController struct{}

// GetInstance returns an instance of the custom resource managed by the controller
func (ctrl SparkController) GetInstance() splcommon.MetaObject {
	return &enterprisev1.Spark{
		TypeMeta: metav1.TypeMeta{
			APIVersion: enterprisev1.APIVersion,
			Kind:       "Spark",
		},
	}
}

// GetWatchTypes returns a list of types owned by the controller that it would like to receive watch events for
func (ctrl SparkController) GetWatchTypes() []runtime.Object {
	return []runtime.Object{&appsv1.Deployment{}, &corev1.Secret{}}
}

// Reconcile is used to perform an idempotent reconciliation of the custom resource managed by this controller
func (ctrl SparkController) Reconcile(client client.Client, cr splcommon.MetaObject) (reconcile.Result, error) {
	instance := cr.(*enterprisev1.Spark)
	return spark.ApplySpark(client, instance)
}
