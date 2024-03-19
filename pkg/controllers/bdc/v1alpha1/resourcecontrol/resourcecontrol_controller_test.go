/*
Copyright 2023 KDP(Kubernetes Data Platform).

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

package resourcecontrol

import (
	bdcv1alpha1 "bpaas-core-operator/api/bdc/v1alpha1"
	"bpaas-core-operator/pkg/utils"
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"time"
)

var _ = Describe("Test ResourceControl Controller", func() {
	ctx := context.Background()

	BeforeEach(func() {
		var validDef = utils.ReadContent(filepath.Join("../../../../..", "charts", "bpaas-core-operator", "templates", "deftemplate", "resourcecontrol-def.yaml"))
		// Create namespace
		ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kdp-system"}}

		Eventually(func() error {
			return k8sClient.Create(ctx, &ns)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))

		// Create bdc definition map into configmap
		bdcDefinitionMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bdc-definition-map",
				Namespace: "kdp-system",
			},
			Data: map[string]string{
				"default-ResourceControl": "resourcecontrol-def",
			},
		}
		Eventually(func() error {
			return k8sClient.Create(ctx, &bdcDefinitionMap)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))

		// Create xDefinition
		var def bdcv1alpha1.XDefinition
		err := yaml.Unmarshal([]byte(validDef), &def)
		if err != nil {
			return
		}

		Eventually(func() error {
			return k8sClient.Create(ctx, &def)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "resourcecontrol-def"}, &def)).Should(Succeed())
	})

	Context("When the ResourceControl dependent xDefinition doesn't exist, should occur error", func() {
		It("Applying ResourceControl", func() {
			By("Apply ResourceControl")

			var validInstance = `
apiVersion: bdc.bdos.io/v1alpha1
kind: ResourceControl
metadata:
  labels:
    "bdc.bdos.io/org": "bdctestorg"
  name: resourcecontrol-sample
  annotations:
    "bdc.bdos.io/org": "bdctestorg"
    "bdc.bdos.io/name": "bdc-sample5"      # 必填，需要用于关联BigDataCluster CR中namespace信息
spec:
  # TODO(user): Add fields here
  resourceQuota:
    limits:
      cpu: '1'
      memory: '1024'
    requests:
      cpu: '0.1'
      memory: '128'
      storage: '1Gi'
`
			var resourceControl bdcv1alpha1.ResourceControl
			Expect(yaml.Unmarshal([]byte(validInstance), &resourceControl)).Should(BeNil())
			Expect(k8sClient.Create(ctx, &resourceControl)).Should(Succeed())
		})
	})
})
