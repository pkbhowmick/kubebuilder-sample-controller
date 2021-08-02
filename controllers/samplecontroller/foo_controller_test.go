package samplecontroller

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	foov1alpha1 "github.com/pkbhowmick/kubebuilder-sample-controller/apis/samplecontroller/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("Foo controller", func() {
	Context("When updating foo status", func() {
		It("Will create Foo object and update it's status", func() {
			By("Creating new Foo object")
			foo := &foov1alpha1.Foo{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sample-foo",
					Namespace: v1.NamespaceDefault,
				},
				Spec: foov1alpha1.FooSpec{
					DeploymentName: "foo-deployment",
					Replicas:       pointer.Int32Ptr(1),
				},
			}
			Expect(k8sClient.Create(context.Background(), foo)).Should(Succeed())
		})
	})
})
