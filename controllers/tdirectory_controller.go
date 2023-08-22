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

	testoperatorsv1 "tdirectory-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

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
	log := ctrllog.FromContext(ctx)
	tDirectory := &testoperatorsv1.Tdirectory{}
	existingTdirectoryDeployment := &appsv1.Deployment{}
	existingService := &corev1.Service{}

	log.Info("Event received")
	log.Info("Request: ", "req", req)

	// CR deleted : check if  the Deployment and the Service must be deleted
	err := r.Get(ctx, req.NamespacedName, tDirectory)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("resource not found, check if a deployment must be deleted.")

			// Delete Deployment
			err = r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, existingTdirectoryDeployment)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info("Nothing to do, no deployment found.")
					return ctrl.Result{}, nil
				} else {
					log.Error(err, "Failed to get Deployment")
					return ctrl.Result{}, err
				}
			} else {
				log.Info("Deployment exists: delete it")
				r.Delete(ctx, existingTdirectoryDeployment)
			}

			// Delete Service
			err = r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, existingService)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info("Nothing to do, no service found.")
					return ctrl.Result{}, nil
				} else {
					log.Error(err, "Failed to get Service")
					return ctrl.Result{}, err
				}
			} else {
				log.Info("Service exists: delete it")
				r.Delete(ctx, existingService)
				return ctrl.Result{}, nil
			}
		}
	} else {
		mongoDBSvc := getmongoDBSvc(tDirectory)
		msvc := &corev1.Service{}
		err = r.Get(ctx, types.NamespacedName{Name: "mongodb-service", Namespace: tDirectory.Namespace}, msvc)
		if err != nil {
			if errors.IsNotFound(err) {
				controllerutil.SetControllerReference(tDirectory, mongoDBSvc, r.Scheme)
				err = r.Create(ctx, mongoDBSvc)
				if err != nil {
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{}, err
			}
		} else if !reflect.DeepEqual(mongoDBSvc.Spec, msvc.Spec) {
			mongoDBSvc.ObjectMeta = msvc.ObjectMeta
			controllerutil.SetControllerReference(tDirectory, mongoDBSvc, r.Scheme)
			err = r.Update(ctx, mongoDBSvc)
			if err != nil {
				return ctrl.Result{}, err
			}
			log.Info("mongodb-service updated")
		}
		mongoDBSS := getMongoDBStatefulsets(tDirectory)
		mss := &appsv1.StatefulSet{}
		err = r.Get(ctx, types.NamespacedName{Name: "mongodb", Namespace: tDirectory.Namespace}, mss)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Info("mongodb statefulset not found, will be created")
				controllerutil.SetControllerReference(tDirectory, mongoDBSS, r.Scheme)
				err = r.Create(ctx, mongoDBSS)
				if err != nil {
					return ctrl.Result{}, err
				}
			} else {
				log.Info("failed to get mongodb statefulset")
				return ctrl.Result{}, err
			}
		} else if !reflect.DeepEqual(mongoDBSS.Spec, mss.Spec) {
			r.UpdateVolume(ctx, tDirectory)
			mongoDBSS.ObjectMeta = mss.ObjectMeta
			mongoDBSS.Spec.VolumeClaimTemplates = mss.Spec.VolumeClaimTemplates
			controllerutil.SetControllerReference(tDirectory, mongoDBSS, r.Scheme)
			err = r.Update(ctx, mongoDBSS)
			if err != nil {
				return ctrl.Result{}, err
			}
			log.Info("mongodb statefulset updated")
		}
		// Check if the deployment already exists, if not: create a new one.
		err = r.Get(ctx, types.NamespacedName{Name: tDirectory.Name, Namespace: tDirectory.Namespace}, existingTdirectoryDeployment)
		if err != nil && errors.IsNotFound(err) {
			// Define a new deployment
			tDirectoryDeployment := r.createDeployment(tDirectory)
			log.Info("Creating a new Deployment", "Deployment.Namespace", tDirectoryDeployment.Namespace, "Deployment.Name", tDirectoryDeployment.Name)

			err = r.Create(ctx, tDirectoryDeployment)
			if err != nil {
				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", tDirectoryDeployment.Namespace, "Deployment.Name", tDirectoryDeployment.Name)
				return ctrl.Result{}, err
			}
		} else if err == nil {
			// Deployment exists, check if the Deployment must be updated
			var replicaCount int32 = tDirectory.Spec.TdirectoryApp.Replicas
			if *existingTdirectoryDeployment.Spec.Replicas != replicaCount {
				log.Info("Number of replicas changes, update the deployment!")
				existingTdirectoryDeployment.Spec.Replicas = &replicaCount
				err = r.Update(ctx, existingTdirectoryDeployment)
				if err != nil {
					log.Error(err, "Failed to update Deployment", "Deployment.Namespace", existingTdirectoryDeployment.Namespace, "Deployment.Name", existingTdirectoryDeployment.Name)
					return ctrl.Result{}, err
				}
			}
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		// Check if the service already exists, if not: create a new one
		err = r.Get(ctx, types.NamespacedName{Name: tDirectory.Name, Namespace: tDirectory.Namespace}, existingService)
		if err != nil && errors.IsNotFound(err) {
			// Create the Service
			newService := r.createService(tDirectory)
			log.Info("✨ Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			err = r.Create(ctx, newService)
			if err != nil {
				log.Error(err, "Failed to create new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
				return ctrl.Result{}, err
			}
		} else if err == nil {
			// Service exists, check if the port has to be updated.
			var port int32 = tDirectory.Spec.TdirectoryApp.Port
			if existingService.Spec.Ports[0].Port != port {
				log.Info("Port number changes, update the service!")
				existingService.Spec.Ports[0].Port = port
				err = r.Update(ctx, existingService)
				if err != nil {
					log.Error(err, "Failed to update Service", "Service.Namespace", existingService.Namespace, "Service.Name", existingService.Name)
					return ctrl.Result{}, err
				}
			}
		} else if err != nil {
			log.Error(err, "Failed to get Service")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// Create a Deployment for the Nginx server.
func (r *TdirectoryReconciler) createDeployment(tDirectory *testoperatorsv1.Tdirectory) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tDirectory.Name,
			Namespace: tDirectory.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &tDirectory.Spec.TdirectoryApp.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "tdirectory-server"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "tdirectory-server"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "tdirectory:v1",
						Name:  "tele-directory",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "http",
							Protocol:      "TCP",
						}},
					}},
				},
			},
		},
	}
	return deployment
}

// Create a Service for the Nginx server.
func (r *TdirectoryReconciler) createService(tDirectory *testoperatorsv1.Tdirectory) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tDirectory.Name,
			Namespace: tDirectory.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "tdirectory-server",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       tDirectory.Spec.TdirectoryApp.Port,
					TargetPort: intstr.FromInt(80),
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

	return service
}

func getmongoDBSvc(tdr *testoperatorsv1.Tdirectory) *corev1.Service {

	p := make([]corev1.ServicePort, 0)
	servicePort := corev1.ServicePort{
		Name: "tcp-port",
		Port: tdr.Spec.TdirectoryDB.Port,
	}
	p = append(p, servicePort)
	mongoDBSvc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb-service",
			Namespace: tdr.Namespace,
			Labels:    map[string]string{"app": "tdirectory-mongodb"},
		},
		Spec: corev1.ServiceSpec{
			Ports:     p,
			Selector:  map[string]string{"app": "tdirectory-mongodb"},
			ClusterIP: "None",
		},
	}
	return mongoDBSvc
}

func getMongoDBStatefulsets(tdr *testoperatorsv1.Tdirectory) *appsv1.StatefulSet {

	cnts := make([]corev1.Container, 0)
	cnt := corev1.Container{
		Name:            "mongodb",
		Image:           "mongo:latest",
		ImagePullPolicy: tdr.Spec.TdirectoryDB.ImagePullPolicy,
	}
	cnts = append(cnts, cnt)
	podTempSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "tdirectory-mongodb"},
		},
		Spec: corev1.PodSpec{
			Containers: cnts,
		},
	}
	mongoss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb",
			Namespace: tdr.Namespace,
			Labels:    map[string]string{"app": "tdirectory-mongodb"},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "tdirectory-mongodb"},
			},
			Replicas:    &tdr.Spec.TdirectoryDB.Replicas,
			Template:    podTempSpec,
			ServiceName: "mongodb-service",
			//VolumeClaimTemplates: volClaimTemplate(),
			VolumeClaimTemplates: volClaimTemplate(tdr.Spec.TdirectoryDB.DBSize),
		},
	}
	return mongoss
}

func volClaimTemplate(DBSize resource.Quantity) []corev1.PersistentVolumeClaim {

	storageClass := "standard"
	mongorr := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			//corev1.ResourceStorage: resource.MustParse(DBSize),
			corev1.ResourceStorage: DBSize,
		},
	}
	accessModeList := make([]corev1.PersistentVolumeAccessMode, 0)
	accessModeList = append(accessModeList, corev1.ReadWriteOnce)
	mongopvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mongodb-pvc",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      accessModeList,
			Resources:        mongorr,
			StorageClassName: &storageClass,
		},
	}
	pvcList := make([]corev1.PersistentVolumeClaim, 0)
	pvcList = append(pvcList, mongopvc)
	return pvcList
}

func (r *TdirectoryReconciler) UpdateVolume(ctx context.Context, tdr *testoperatorsv1.Tdirectory) error {

	log := ctrllog.FromContext(ctx)
	mpvc := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: "mongodb-pvc-mongodb-0", Namespace: tdr.Namespace}, mpvc)
	if err != nil {
		return nil
	}
	if mpvc.Spec.Resources.Requests[corev1.ResourceStorage] != tdr.Spec.TdirectoryDB.DBSize {
		log.Info("Need to expand the mongodb volume")
		mpvc.Spec.Resources.Requests[corev1.ResourceStorage] = tdr.Spec.TdirectoryDB.DBSize
		err := r.Update(ctx, mpvc)
		if err != nil {
			log.Info("Error in expanding the mongodb volume")
			return err
		}
		log.Info("mongodb volume updated successfully")
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TdirectoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testoperatorsv1.Tdirectory{}).
		Watches(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}, builder.WithPredicates(predicate.Funcs{
			// Check only delete events for a service
			UpdateFunc: func(e event.UpdateEvent) bool {
				return false
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
			},
		})).
		Complete(r)
}
