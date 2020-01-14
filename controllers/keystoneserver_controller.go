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

	if err = r.Create(ctx, &cm); err != nil {
		return ctrl.Result{}, err
	}

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
