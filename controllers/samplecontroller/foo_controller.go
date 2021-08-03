/*
Copyright 2021.

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

package samplecontroller

import (
	"context"
	"errors"
	samplecontrollerv1alpha1 "github.com/pkbhowmick/kubebuilder-sample-controller/apis/samplecontroller/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	ResourceKindSingular = "Foo"
)

// FooReconciler reconciles a Foo object
type FooReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=samplecontroller.example.com,resources=foos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=samplecontroller.example.com,resources=foos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=samplecontroller.example.com,resources=foos/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Foo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *FooReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Reconcile func logic here
	klog.Info("Got Foo event")
	var foo samplecontrollerv1alpha1.Foo

	if err := r.Get(ctx, req.NamespacedName, &foo); err != nil {
		klog.Infof("Foo %q doesn't exist anymore", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deploymentName := foo.Spec.DeploymentName
	if deploymentName == "" {
		return ctrl.Result{}, errors.New("deployment name must be specified")
	}

	// Get the deployment with the name specified in foo.spec
	deplKey := types.NamespacedName{
		Namespace: "default",
		Name:      deploymentName,
	}
	var deployment appsv1.Deployment
	err := r.Get(ctx, deplKey, &deployment)

	if kerr.IsNotFound(err) {
		err = r.Create(ctx, newDeployment(&foo))
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Get(ctx, deplKey, &deployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !metav1.IsControlledBy(&deployment, &foo) {
		return ctrl.Result{}, errors.New("deployment is not owned by Foo")
	}

	// Check Deployment replica count with foo replica count
	// and update if necessary
	if foo.Spec.Replicas != nil && *foo.Spec.Replicas != *deployment.Spec.Replicas {
		err = r.Update(ctx, newDeployment(&foo))
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update Foo status
	fooCopy := foo.DeepCopy()
	fooCopy.Status.Phase = samplecontrollerv1alpha1.FooPhaseRunning
	err = r.Status().Update(ctx, fooCopy)
	if err != nil {
		return ctrl.Result{}, err
	}
	klog.Info("Successfully reconcile Foo object")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FooReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplecontrollerv1alpha1.Foo{}).
		Complete(r)
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(foo *samplecontrollerv1alpha1.Foo) *appsv1.Deployment {
	gv := samplecontrollerv1alpha1.GroupVersion
	labels := map[string]string{
		"app":        "nginx",
		"controller": foo.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      foo.Spec.DeploymentName,
			Namespace: foo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, schema.GroupVersionKind{
					Group:   gv.Group,
					Version: gv.Version,
					Kind:    ResourceKindSingular,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: foo.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
						},
					},
				},
			},
		},
	}
}
