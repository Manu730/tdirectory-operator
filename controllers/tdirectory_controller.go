/*
Copyright 2023.

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
	"reflect"

	tdrTypes "tdirectory-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("controller_tdirectory")

// TdirectoryReconciler reconciles a Tdirectory object
type TdirectoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=testoperators.tdirectory.com,resources=tdirectories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=testoperators.tdirectory.com,resources=tdirectories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=testoperators.tdirectory.com,resources=tdirectories/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tdirectory object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *TdirectoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	rLog := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rLog.Info("Reconciling Tdirectory")
	td := &tdrTypes.Tdirectory{}
	if err = r.Get(ctx, request.NamespacedName, tdr); err != nil && errors.IsNotFound(err) {
		rLog.Error("Request object not found")
		return reconcile.Result{}, nil
	} else if err != nil {
		rLog.Error(err, "Failed to get tdirectory object")
		return reconcile.Result{}, err
	}
	if err = r.UpdateResource(ctx, td); err != nil {
		rLog.Error(err, "Failed to update tdirectory resource")
		return reconcile.Result{}, err
	}
	r.Status().Update(ctx, td)
	return ctrl.Result{}, nil
}

func (r *TdirectoryReconciler) UpdateResource(ctx context.Context, tdr *tdrTypes.Tdirectory) error {
	var err error
	rLog := log.WithValues("Namespace", tdr.Namespace)
	mongoSvc := r.getmongoSvc(tdr)
	msvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: tdr.Spec.TdirectoryDB.Name, Namespace: tdr.Namespace}, msvc)
	if err != nil && errors.IsNotFound(err) {
		controllerutil.SetControllerReference(tdr, mongoDBSvc, r.scheme)
		if e := r.Create(ctx, mongoDBSvc); e != nil {
			rLog.Error(err, "Error creating mongo service")
			return e
		}
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(mongoSvc.Spec, msvc.Spec) {
		mongoDBSvc.ObjectMeta = msvc.ObjectMeta
		controllerutil.SetControllerReference(bookstore, mongoDBSvc, r.scheme)
		if err = r.Update(ctx, mongoDBSvc); err != nil {
			rLog.Error(err, "Error updating mongo service")
			return err
		}
	}
	mongoSS := r.getMongoDBStatefulsets(tdr)
	mss := &appsv1.StatefulSet{}
	if err = r.Get(ctx, types.NamespacedName{Name: tdr.Spec.TdirectoryDB.Name, Namespace: tdr.Namespace}, mss); err != nil && errors.IsNotFound(err) {
		controllerutil.SetControllerReference(tdr, mongoSS, r.scheme)
		if e := r.Create(ctx, mongoSS); e != nil {
			return e
		}
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(mongoSS.Spec, mss.Spec) {
		if err = r.updateMongoVolume(ctx, tdr); err != nil {
			return err
		}
		mongoSS.ObjectMeta = mss.ObjectMeta
		mongoSS.Spec.VolumeClaimTemplates = mss.Spec.VolumeClaimTemplates
		controllerutil.SetControllerReference(tdr, mongoSS, r.scheme)
		if err = r.Update(ctx, mongoSS); err != nil {
			return err
		}
	}
	tdrSvc := t.getTdirectoryAppSvc(tdr)
	tsvc := &corev1.Service{}
	if err = r.Get(ctx, types.NamespacedName{Name: tdr.Spec.TdirectoryApp.Name, Namespace: tdr.Namespace}, tsvc); err != nil && errors.IsNotFound(err) {
		controllerutil.SetControllerReference(tdr, tdrSvc, r.scheme)
		if e := r.Create(ctx, tdrSvc); e != nil {
			return e
		}
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(tdrSvc.Spec, tsvc.Spec) {
		tdrSvc.ObjectMeta = tsvc.ObjectMeta
		tdrSvc.Spec.ClusterIP = tsvc.Spec.ClusterIP
		controllerutil.SetControllerReference(tdr, tdrSvc, r.scheme)
		if err = r.Update(ctx, tdrSvc); err != nil {
			return err
		}
	}

	tdrDep := t.getTdirectoryAppDeploy(tdr)
	tdep := &appsv1.Deployment{}
	if err = r.Get(ctx, types.NamespacedName{Name: tdr.Spec.TdirectoryApp.Name, Namespace: tdr.Namespace}, tdep); err != nil && errors.IsNotFound(err) {
		controllerutil.SetControllerReference(tdr, tdrDep, r.scheme)
		if e := r.Create(ctx, tdrDep); e != nil {
			return e
		}
	}

	if !reflect.DeepEqual(tdrDep.Spec, tdep.Spec) {
		tdrDep.ObjectMeta = tdep.ObjectMeta
		controllerutil.SetControllerReference(tdr, tdrDep, r.scheme)
		if err = r.Update(ctx, tdrDep); err != nil {
			return err
		}
	}
	return nil
}

func (r *TdirectoryReconciler) getmongoSvc(tdr *tdrTypes.Tdirectory) *corev1.Service {
	var p []corev1.ServicePort
	p = append(p, corev1.ServicePort{
		Name: "tcp-port",
		Port: tdr.Spec.TdirectoryDB.Port,
	})
	return &corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: corev1.SchemeGroupVersion.String()}, ObjectMeta: metav1.ObjectMeta{
		Name: tdr.Spec.TdirectoryDB.Name, Namespace: tdr.Namespace, Labels: map[string]string{"app": tdr.Spec.TdirectoryDB.Name}}, Spec: corev1.ServiceSpec{Ports: p, Selector: map[string]string{"app": tdr.Spec.TdirectoryDB.Name, ClusterIP: "None"}}}
}

func (r *TdirectoryReconciler) getMongoDBStatefulsets(tdr *tdrTypes.Tdirectory) *appsv1.StatefulSet {
	var cnts []corev1.Container
	cnt := corev1.Container{Name: tdr.Spec.TdirectoryDB.Name, Image: tdr.Spec.TdirectoryDB.Repository + ":" + tdr.Spec.TdirectoryDB.Tag, ImagePullPolicy: tdr.Spec.TdirectoryDB.ImagePullPolicy}
	cnts = append(cnts, cnt)
	podTempSpec := corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": tdr.Spec.TdirectoryDB.Name}}, Spec: corev1.PodSpec{Containers: cnts}}
	return &appsv1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{Kind: "StatefulSet", APIVersion: appsv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: tdr.Spec.TdirectoryDB.Name, Namespace: tdr.Namespace, Labels: map[string]string{"app": tdr.Spec.TdirectoryDB.Name}},
		Spec:       appsv1.StatefulSetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": tdr.Spec.TdirectoryDB.Name}}, Replicas: &tdr.Spec.TdirectoryDB.Replicas, Template: podTempSpec, ServiceName: tdr.Spec.TdirectoryDB.Name, VolumeClaimTemplates: r.getMongoVolClaimTemplate(tdr.Spec.TdirectoryDB.DBSize)},
	}
}

func (r *TdirectoryReconciler) getMongoVolClaimTemplate(s resource.Quantity) []corev1.PersistentVolumeClaim {
	sClass := "standard"
	rr := corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: s}}
	mongopvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: tdr.Spec.TdirectoryDB.Name + "-pvc"},
		Spec:       PersistentVolumeClaimSpec{AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}, Resources: rr, StorageClassName: &sClass},
	}
	return []corev1.PersistentVolumeClaim{mongopvc}
}

func (r *TdirectoryReconciler) updateMongoVolume(ctx context.Context, tdr *tdrTypes.Tdirectory) error {
	var err error
	rLog := log.WithValues("Namespace", tdr.Namespace)
	mpvc := &corev1.PersistentVolumeClaim{}
	if err = r.Get(ctx, types.NamespacedName{Name: tdr.Spec.TdirectoryDB.Name + "-pvc", Namespace: tdr.Namespace}, mpvc); err != nil {
		return err
	}
	if mpvc.Spec.Resources.Requests[corev1.ResourceStorage] != tdr.Spec.TdirectoryDB.DBSize {
		mpvc.Spec.Resources.Requests[corev1.ResourceStorage] = bookstore.Spec.BookDB.DBSize
		return r.Update(ctx, mpvc)
	}
	return nil
}

func (r *TdirectoryReconciler) getTdirectoryAppSvc(tdr *tdrTypes.Tdirectory) {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tdr.Spec.TdirectoryApp.Name,
			Namespace: tdr.Namespace,
			Labels:    map[string]string{"app": tdr.Spec.TdirectoryApp.Name},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				Name:       "tcp-port",
				Port:       tdr.Spec.TdirectoryApp.Port,
				TargetPort: intstr.FromInt(tdr.Spec.TdirectoryApp.TargetPort),
			},
			Type:     tdr.Spec.TdirectoryApp.ServiceType,
			Selector: map[string]string{"app": tdr.Spec.TdirectoryApp.Name},
		},
	}
}

func (r *TdirectoryReconciler) getTdirectoryAppDeploy(tdr *tdrTypes.Tdirectory) {
	cnt := corev1.Container{
		Name:            tdr.Spec.TdirectoryApp.Name,
		Image:           tdr.Spec.TdirectoryDB.Repository + ":" + tdr.Spec.TdirectoryDB.Tag,
		ImagePullPolicy: tdr.Spec.TdirectoryApp.ImagePullPolicy,
	}
	podTempSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": tdr.Spec.TdirectoryApp.Name},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{cnt},
		},
	}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tdr.Spec.TdirectoryApp.Name,
			Namespace: tdr.Namespace,
			Labels:    map[string]string{"app": tdr.Spec.TdirectoryApp.Name},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": tdr.Spec.TdirectoryApp.Name}},
			Replicas: &tdr.Spec.TdirectoryApp.Replicas,
			Template: podTempSpec,
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TdirectoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tdrTypes.Tdirectory{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
