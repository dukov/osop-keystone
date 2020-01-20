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
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	k8sapps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1alpha1 "github.com/dukov/osop-keystone/api/v1alpha1"

	commonk8s "github.com/dukov/osop-common/pkg/k8s"
)

var (
	ownerKey = ".metadata.controller"
	apiGVStr = openstackv1alpha1.GroupVersion.String()
)

// KeystoneServerReconciler reconciles a KeystoneServer object
type KeystoneServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=openstack.osop.org,resources=keystoneservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.osop.org,resources=keystoneservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openstack.osop.org,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.osop.org,resources=deployments/status,verbs=get

func (r *KeystoneServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var keystoneSrv openstackv1alpha1.KeystoneServer
	ctx := context.Background()
	log := r.Log.WithValues("keystoneserver", req.NamespacedName)
	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("keystone-server")}
	if err := r.Get(ctx, req.NamespacedName, &keystoneSrv); err != nil {
		log.Error(err, "unable to fetch Keystone servers")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var kDepls k8sapps.DeploymentList
	if err := r.List(ctx, &kDepls, client.InNamespace(req.Namespace), client.MatchingField(ownerKey, req.Name)); err != nil {
		log.Error(err, "deployments list error")
		return ctrl.Result{}, err
	}

	cm, err := r.createConfigMap(keystoneSrv)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Creating ConfigMap", "ConfigMap", cm)
	if err = r.Patch(ctx, &cm, client.Apply, applyOpts...); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("ConfigMap Created")

	dep, err := r.createDeployment(keystoneSrv)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Creating Deployment", "Deployment", dep)
	if err = r.Patch(ctx, &dep, client.Apply, applyOpts...); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("Deployment Created")

	return ctrl.Result{}, nil
}

func (r *KeystoneServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&k8sapps.Deployment{}, ownerKey, func(rawObj runtime.Object) []string {
		depl := rawObj.(*k8sapps.Deployment)
		owner := metav1.GetControllerOf(depl)
		if owner == nil {
			return nil
		}

		if owner.APIVersion != apiGVStr || owner.Kind != "KeystoneServer" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&openstackv1alpha1.KeystoneServer{}).
		Owns(&k8sapps.Deployment{}).
		Complete(r)
}

func (r *KeystoneServerReconciler) createDeployment(srv openstackv1alpha1.KeystoneServer) (k8sapps.Deployment, error) {
	vol := commonk8s.NewVolume("etc-keystone", srv.Name)

	container := commonk8s.NewContainer("keystone-api", srv.Spec.Image, []string{"keystone-wsgi-public"})
	volMount := corev1.VolumeMount{
		Name:      "etc-keystone",
		MountPath: "/etc/keystone",
	}

	container.AddVolume(volMount)
	labels := map[string]string{
		"component": "api",
	}

	depl := commonk8s.NewDeployment(srv.Name, srv.Namespace, srv.Spec.Replicas, labels)
	depl.AddContainer(container)
	depl.AddVolume(vol)

	if err := ctrl.SetControllerReference(&srv, depl.Obj, r.Scheme); err != nil {
		return *depl.Obj, err
	}
	return *depl.Obj, nil
}

func (r *KeystoneServerReconciler) createConfigMap(srv openstackv1alpha1.KeystoneServer) (corev1.ConfigMap, error) {
	cfg := make(map[string]string)
	mergeConfig(KeystoneConfigDefaults, srv.Spec.Config)
	var content []string
	for section, opts := range KeystoneConfigDefaults {
		content = append(content, fmt.Sprintf("[%s]", section))
		for key, val := range opts {
			content = append(content, fmt.Sprintf("%s = %s", key, val))
		}
	}

	cfg[KyestoneConfigFilename] = strings.Join(content, "\n")

	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      srv.Name,
			Namespace: srv.Namespace,
		},
		Data: cfg,
	}

	if err := ctrl.SetControllerReference(&srv, &cm, r.Scheme); err != nil {
		return cm, err
	}

	return cm, nil

}
