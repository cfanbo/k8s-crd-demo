/*
Copyright 2023 sxf.

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

package controller

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	techv1 "github.com/cfanbo/sample/api/v1"
)

const (
	KIND = "SuperService"
)

// SuperServiceReconciler reconciles a SuperService object
type SuperServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tech.example.io,resources=superservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tech.example.io,resources=superservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tech.example.io,resources=superservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;update;patch;get;list;watch
//+kubebuilder:rbac:groups=*,resources=services,verbs=create;update;patch;get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SuperService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *SuperServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	llog := log.FromContext(ctx)

	llog.Info("监听到更新。。。。。。。。")

	// TODO(user): your logic here

	// 读取当前资源实例
	instance := &techv1.SuperService{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 当前实例被删除
	if instance.DeletionTimestamp != nil {
		return ctrl.Result{}, err
	}

	// 如果当前实例对应的 deployment/service  不存在，则创建关联资源
	// 如果存在，判断是否需要更新
	//   如果需要更新，则直接更新
	//   如果不需要更新，则正常返回

	// 查找或创建 deployment
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		if errors.IsNotFound(err) {
			llog.Info("deployment 不存在，重新创建")
			// 创建 deployment
			deployment = constructDeployment(instance)
			if err := r.Create(ctx, deployment); err != nil {
				llog.Error(err, "创建 deployment 出错：")
				return ctrl.Result{}, err
			}

			// Annotations
			if instance.Annotations == nil {
				instance.Annotations = make(map[string]string)
			}
			jsonData, err := json.Marshal(instance.Spec)
			if err != nil {
				llog.Error(err, "toJSON 出错")
				return ctrl.Result{}, err
			}
			instance.Annotations["spec"] = string(jsonData)

			// update instance
			if err := r.Update(ctx, instance); err != nil {
				llog.Error(err, "创建 cr 出错：")
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	// 查找或创建 Service
	currService := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, currService); err != nil {
		if errors.IsNotFound(err) {
			llog.Info("Service 不存在，重新创建")
			currService = constructService(instance)
			if err := r.Create(ctx, currService); err != nil {
				llog.Error(err, "创建 service 出错：")
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	// oldSpec 数据库中存储的记录
	oldSpec := techv1.SuperServiceSpec{}
	if value, ok := instance.Annotations["spec"]; ok && value != "" {
		if err := json.Unmarshal([]byte(value), &oldSpec); err != nil {
			llog.Error(err, "解析JSON失败："+instance.Annotations["spec"])
			return ctrl.Result{}, err
		}
	}

	// 对比实际字段与注解中的信息是否一致
	if !reflect.DeepEqual(oldSpec, instance.Spec) {
		// 同步 instance
		if instance.Annotations == nil {
			instance.Annotations = make(map[string]string)
		}
		jsonData, err := json.Marshal(instance.Spec)
		if err != nil {
			llog.Error(err, "解析 JSON 失败")
			return ctrl.Result{}, err
		}

		// 更新资源变更次数 r.Status().Update()
		instance.Status.Count = instance.Status.Count + 1
		if err := r.Status().Update(ctx, instance); err != nil {
			llog.Error(err, "更新 cr.Status 失败")
			return ctrl.Result{}, err
		}

		// 更新资源实际状态  r.Update()
		instance.Annotations["spec"] = string(jsonData)
		if err := r.Update(ctx, instance); err != nil {
			llog.Error(err, "更新 cr 失败")
			return ctrl.Result{}, err
		}

		// 同步 deployment
		currDeployment := &appsv1.Deployment{}
		if err := r.Get(ctx, req.NamespacedName, currDeployment); err != nil {
			llog.Error(err, "同步 deployment 出错")
			return ctrl.Result{}, err
		}

		newDeploy := constructDeployment(instance)
		currDeployment.Spec = newDeploy.Spec
		if err := r.Update(ctx, currDeployment); err != nil {
			llog.Error(err, "更新 deployment 出错")
			return ctrl.Result{}, err
		}

		// 同步 service
		currService := &corev1.Service{}
		if err := r.Get(ctx, req.NamespacedName, currService); err != nil {
			llog.Error(err, "读取 Service 失败")
			return ctrl.Result{}, err
		}

		newService := constructService(instance)
		currService.Spec = newService.Spec
		if err := r.Update(ctx, currService); err != nil {
			llog.Error(err, "更新 Service 资源失败")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SuperServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&techv1.SuperService{}).
		Complete(r)
}

func constructDeployment(instance *techv1.SuperService) *appsv1.Deployment {
	labels := map[string]string{
		"app": instance.Name,
	}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,

			Labels: labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(instance, schema.GroupVersionKind{
					Group:   techv1.GroupVersion.Group,
					Version: techv1.GroupVersion.Version,
					Kind:    KIND,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: selector,
			Replicas: instance.Spec.Size,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: newContainers(instance),
				},
			},
		},
		//Status: appsv1.DeploymentStatus{},
	}

}

func constructService(instance *techv1.SuperService) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(instance, schema.GroupVersionKind{
					Group:   techv1.GroupVersion.Group,
					Version: techv1.GroupVersion.Version,
					Kind:    KIND,
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: instance.Spec.Ports,
			Selector: map[string]string{
				"app": instance.Name,
			},
		},
		//Status: corev1.ServiceStatus{},
	}
}

func newContainers(instance *techv1.SuperService) []corev1.Container {
	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range instance.Spec.Ports {
		cPort := corev1.ContainerPort{ContainerPort: svcPort.TargetPort.IntVal}
		containerPorts = append(containerPorts, cPort)
	}

	return []corev1.Container{
		{
			Name:            instance.Name,
			Image:           instance.Spec.Image,
			Ports:           containerPorts,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env:             instance.Spec.Envs,
		},
	}

}
