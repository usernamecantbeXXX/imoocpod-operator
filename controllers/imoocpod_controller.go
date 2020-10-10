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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;

//监听pod 的变化，实现监听的logic
func (r *ImoocPodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	//_ = r.Log.WithValues("imoocpod", req.NamespacedName)
	r.Log.Info("Recociling ImoocPod", "Request.Namespace", req.NamespacedName, "Request.Name", req.Name)

	// your logic here
	//Fetch the ImoocPod instance 首先获取一个imoocpod的实例
	instance := &xxxv1.ImoocPod{}
	//r是k8s获取到的实例
	//通过r去对instance进行赋值
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	//1：获取 name 对应的所有的pod列表
	lbls := labels.Set{
		"app": instance.Name, //资源的label来查找对应的pod
	}
	existingPods := &corev1.PodList{}
	err = r.List(context.TODO(), existingPods, &client.ListOptions{
		Namespace:     req.Namespace,
		LabelSelector: labels.SelectorFromSet(lbls),
	})
	if err != nil {
		r.Log.Error(err, "取已经存在的pod失败")
		return ctrl.Result{}, err
	}
	//2：获取到pod列表中的pod name
	var existingPodNames []string
	for _, pod := range existingPods.Items {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	//3：update pod.status != 运行中的status
	//比较DeepEqual
	status := xxxv1.ImoocPodStatus{
		PodNames: existingPodNames,
		Replicas: len(existingPodNames),
	}
	if !reflect.DeepEqual(instance.Status, status) {
		instance.Status = status //把期望状态给运行态
		err := r.Client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.Log.Error(err, "更新pod失败")
			return ctrl.Result{}, err
		}
	}

	//4：len（pod） > 运行中的len（pod.replicas), 期望值小，需要 scale down delete
	if len(existingPodNames) > instance.Spec.Replicas {
		//delete
		r.Log.Info("正在删除pod,当前的podNames和期望的replicas：", existingPodNames, instance.Spec.Replicas)
		pod := existingPods.Items[0]
		err := r.Client.Delete(context.TODO(), &pod)
		if err != nil {
			r.Log.Error(err, "删除pod失败")
			return ctrl.Result{}, err
		}
	}

	//5：len（pod） < 运行中的len（pod.replicas), 期望值大，需要 scale up create
	if len(existingPodNames) > instance.Spec.Replicas {
		//delete
		r.Log.Info("正在创建pod,当前的podNames和期望的replicas：", existingPodNames, instance.Spec.Replicas)
		//Define a new Pod object
		pod := newPodForCR(instance)

		//Set ImoocPod instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			return ctrl.Result{}, nil
		}
		err = r.Create(context.TODO(), pod)
		if err != nil {
			r.Log.Error(err, "创建pod失败")
			return ctrl.Result{}, err
		}
	}

	/*	//Define a new Pod object
		pod := newPodForCR(instance)

		//Set ImoocPod instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			return ctrl.Result{}, nil
		}

		//Check if the Pod already exists
		found := &corev1.Pod{}
		err = r.Get(context.TODO(), types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}, found)
		if err != nil && errors.IsNotFound(err) {
			r.Log.Info("Create a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			err = r.Create(context.TODO(), pod)
			if err != nil {
				return ctrl.Result{}, err
			}
			//Pod创建成功，不用重新排队 -dont requeue
			return ctrl.Result{}, nil
		} else if err != nil {
			return ctrl.Result{}, err
		}

		//pod已经存在，不用重新排队
		r.Log.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)*/
	return ctrl.Result{Requeue: true}, nil
}

func (r *ImoocPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&xxxv1.ImoocPod{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Pod{}).
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
			GenerateName: cr.Name + "-pod",
			Namespace:    cr.Namespace,
			Labels:       lables,
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
