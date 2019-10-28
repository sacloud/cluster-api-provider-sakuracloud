/*
Copyright 2019 Kazumichi Yamamoto.

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

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
)

var _ = Describe("SakuraCloudClusterReconciler", func() {
	BeforeEach(func() {})
	AfterEach(func() {})

	Context("Reconcile an SakuraCloudCluster", func() {
		It("should not error and requeue the request with insufficient set up", func() {

			ctx := context.Background()

			reconciler := &SakuraCloudClusterReconciler{
				Client: k8sClient,
				Log:    log.Log,
			}

			instance := &infrav1.SakuraCloudCluster{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"}}

			// Create the SakuraCloudCluster object and expect the Reconcile and Deployment to be created
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())

			result, err := reconciler.Reconcile(ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: instance.Namespace,
					Name:      instance.Name,
				},
			})
			Expect(err).To(BeNil())
			Expect(result.RequeueAfter).ToNot(BeZero())
		})
	})
})
