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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	homeworkv1alpha1 "github.com/podprinter/print-operator/api/v1alpha1"
)

// PrintReconciler reconciles a Print object
type PrintReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=homework.podprinter.com,resources=prints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homework.podprinter.com,resources=prints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homework.podprinter.com,resources=prints/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Print object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *PrintReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	printer := &homeworkv1alpha1.Print{}
	err := r.Get(ctx, req.NamespacedName, printer)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			//
			log.Info("Memcached resource nokubet found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - rdequeue the request.
		log.Error(err, "Failed to get Memcached")
		return ctrl.Result{}, err
	}
	log.Info("1차과제 로그찍기: ", "printer.Spec.String", printer.Spec.Printer)

	// deploymet로 x
	// pod가 이미 존재하는지 체크하고, 만들어지지 않았다면 생성한다.
	// loop 목적대로 사용안하고 그냥 pod만 생성해서 로그찍기
	found := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: printer.Name, Namespace: printer.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {

		dep := r.PodForPrint(printer, ctx)

		log.Info("Creating a new Pod - printer Pod 생성") //"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Pod", "Pod.Namespace", dep.Namespace, "Pod.Name", dep.Name)
			return ctrl.Result{}, err
		}
		log.Info("생성 성공", "pod.Namespace", dep.Namespace, "pod.Name", dep.Name)

	} else if err != nil {
		log.Error(err, "Failed to get pod")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
func (r *PrintReconciler) PodForPrint(m *homeworkv1alpha1.Print, ctx context.Context) *corev1.Pod {
	setLog := m.Spec.Printer

	printPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "print",
					Image: "bash",
					Env: []corev1.EnvVar{{
						Name:  "logfrompod",
						Value: setLog,
					}},
					Command: []string{"echo"},
					Args:    []string{"$(logfrompod)"},
				},
			},
		},
	}
	ctrl.SetControllerReference(m, printPod, r.Scheme)
	return printPod
}

// SetupWithManager sets up the controller with the Manager.
func (r *PrintReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homeworkv1alpha1.Print{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
