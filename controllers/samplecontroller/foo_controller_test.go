package samplecontroller

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	foov1alpha1 "github.com/pkbhowmick/kubebuilder-sample-controller/apis/samplecontroller/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("Foo controller", func() {

	const (
		FooName            = "test-foo"
		FooDeploymentName  = "test-foo-deployment"
		FooNamespace       = v1.NamespaceDefault
		FooStartingReplica = int32(1)
		FooUpdatedReplica  = int32(2)

		timeout  = time.Minute * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("Foo CRUD Operations", func() {
		It("Will create Foo object and update it's status", func() {
			By("Creating new Foo resource")
			ctx := context.Background()
			foo := &foov1alpha1.Foo{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FooName,
					Namespace: FooNamespace,
				},
				Spec: foov1alpha1.FooSpec{
					DeploymentName: FooDeploymentName,
					Replicas:       pointer.Int32Ptr(FooStartingReplica),
				},
			}
			Eventually(func() bool {
				err := k8sClient.Create(ctx, foo)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Checking created deployment's replica count")
			createdDeployment := &appsv1.Deployment{}
			deploymentLookupKey := types.NamespacedName{
				Namespace: FooNamespace,
				Name:      FooDeploymentName,
			}
			Eventually(func() (int32, error) {
				err := k8sClient.Get(ctx, deploymentLookupKey, createdDeployment)
				if err != nil {
					return -1, err
				}
				return *createdDeployment.Spec.Replicas, nil
			}, timeout, interval).Should(Equal(FooStartingReplica))

			By("Updating the Foo Status")
			fooCopy := foo.DeepCopy()
			fooCopy.Status.Phase = foov1alpha1.FooPhaseRunning
			Eventually(func() bool {
				err := k8sClient.Status().Update(ctx, fooCopy)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval)

			By("Updating the Foo resource")
			fooCopy.Spec.Replicas = pointer.Int32Ptr(FooUpdatedReplica)
			Eventually(func() bool {
				err := k8sClient.Update(ctx, fooCopy)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval)

		})
	})
})
