/*


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
	"github.com/go-logr/logr"
	xxxv1 "github.com/imooc-com/imoocpod-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ImoocPodReconciler reconciles a ImoocPod object
type ImoocPodReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=xxx.bluemoon.com.cn,resources=imoocpods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=xxx.bluemoon.com.cn,resources=imoocpods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

//监听pod 的变化，实现监听的logic
func (r *ImoocPodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("imoocpod", req.NamespacedName)

	// your logic here
	////Fetch the ImoocPod instance 首先获取一个imoocpod的实例
	//instance := &xxxv1.ImoocPod{}
	////r是k8s获取到的实例
	////通过r去对instance进行赋值
	//err := r.Get(context.TODO(), req.NamespacedName, instance)
	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		return ctrl.Result{}, nil
	//	}
	//	return ctrl.Result{}, err
	//}
	//
	////Define a new Pod object
	//pod := newPodForCR(instance)
	//
	////Set ImoocPod instance as the owner and controller
	//if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
	//	return ctrl.Result{}, nil
	//}
	//
	////Check if the Pod already exists
	//found := &v1.Pod{}
	//err = r.Get(context.TODO(), types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}, found)
	//if err != nil && errors.IsNotFound(err) {
	//	r.Log.Info("Create a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	//	err = r.Create(context.TODO(), pod)
	//	if err != nil {
	//		return ctrl.Result{}, err
	//	}
	//	//Pod创建成功，不用重新排队 -dont requeue
	//	return ctrl.Result{}, nil
	//} else if err != nil {
	//	return ctrl.Result{}, err
	//}
	//
	////pod已经存在，不用重新排队
	//r.Log.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return ctrl.Result{}, nil
}

func (r *ImoocPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&xxxv1.ImoocPod{}).
		Owns(&appsv1.Deployment{}).
		//Owns(&corev1.Pod{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}

// newPodForCR retures a busybox pod with  the same name/namespace as the cr
func newPodForCR(cr *xxxv1.ImoocPod) *corev1.Pod {
	lables := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    lables,
		},

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
