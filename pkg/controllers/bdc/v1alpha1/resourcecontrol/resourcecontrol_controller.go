/*
Copyright 2023 KDP(Kubernetes Data Platform).

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

package resourcecontrol

import (
	"bpaas-core-operator/api/bdc/common"
	conditiontype "bpaas-core-operator/api/bdc/condition"
	bdcv1alpha1 "bpaas-core-operator/api/bdc/v1alpha1"
	pkgcommon "bpaas-core-operator/pkg/common"
	bdcctrl "bpaas-core-operator/pkg/controllers/bdc"
	"bpaas-core-operator/pkg/controllers/bdc/constants"
	"bpaas-core-operator/pkg/controllers/bdc/parser"
	"bpaas-core-operator/pkg/controllers/utils/condition"
	"bpaas-core-operator/pkg/controllers/utils/dispatch"
	"bpaas-core-operator/pkg/utils"
	"bpaas-core-operator/version"
	"context"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler reconciles a ResourceControl object
type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	options
}

type options struct {
	defRevLimit          int
	concurrentReconciles int
	ignoreDefNoCtrlReq   bool
	controllerVersion    string
}

//+kubebuilder:rbac:groups=bdc.bdos.io,resources=resourcecontrols,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bdc.bdos.io,resources=resourcecontrols/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bdc.bdos.io,resources=resourcecontrols/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceControl object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.InfoS("Reconcile resource-control policy", "", klog.KRef(req.Namespace, req.Name))

	// Lookup the resourceControl instance for this reconcile request
	var resourceControl bdcv1alpha1.ResourceControl
	if err := r.Get(ctx, req.NamespacedName, &resourceControl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//klog.InfoS("resourceControl", "bdc.cs", resourceControl)

	// Set BigDataCluster as metadata.ownerReferences
	var bigDataCluster bdcv1alpha1.BigDataCluster
	if err := r.Get(ctx, client.ObjectKey{Name: resourceControl.GetAnnotations()[constants.AnnotationBDCName]}, &bigDataCluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	resourceControl.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         bigDataCluster.APIVersion,
			Kind:               bigDataCluster.Kind,
			Name:               bigDataCluster.Name,
			UID:                bigDataCluster.UID,
			Controller:         pointer.Bool(true),
			BlockOwnerDeletion: pointer.Bool(true),
		},
	})
	err := r.patchOwnerReferencer(ctx, &resourceControl)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Replace template.parameter with BigDataCluster Object spec
	bdcParser := parser.NewParser(r.Client)

	// Dispatch manifests
	bdcDispatcher := dispatch.NewManifestsDispatcher(r.Client)
	bdcFile, err := bdcParser.GenerateBigDataClusterFile(ctx, &resourceControl)
	if err != nil {
		klog.Error(err, "[Generate BigDataClusterFile]")
		return ctrl.Result{}, err
	}

	manifests, err := bdcFile.PrepareManifests(ctx, req)
	if err != nil {
		klog.Error(err, "[Handle PrepareManifests]")
		return ctrl.Result{}, err
	}
	// klog.InfoS("ResourceControl", "output manifests", manifests)

	if len(manifests) > 0 {
		if err := bdcDispatcher.Dispatch(ctx, manifests...); err != nil {
			klog.Error(err, "[Handle Apply Manifests]")
		}
		klog.Info("Successfully generated manifests")
	}
	//bdcFile.ReferredObjects = manifests

	var appliedResource []common.BDCObjectReference
	for _, mf := range manifests {
		if mf == nil {
			continue
		}
		referencedObjectStatus := utils.Object2RawExtension(mf.Object["status"])
		appliedResource = append(appliedResource, common.BDCObjectReference{
			ObjectReference: corev1.ObjectReference{
				APIVersion: mf.GetAPIVersion(),
				Kind:       mf.GetKind(),
				Name:       mf.GetName(),
				Namespace:  mf.GetNamespace(),
				UID:        mf.GetUID(),
			},
			Status:  *referencedObjectStatus,
			BDCName: resourceControl.GetAnnotations()[constants.AnnotationBDCName],
		})

	}
	klog.InfoS("Collect", "appliedResources", appliedResource)

	// Check current status.appliedResources in the new manifests or not
	// If not in the new manifests, it means that the resource has been deleted
	lastAppliedResource := resourceControl.Status.AppliedResources
	for _, lar := range lastAppliedResource {
		err := r.deleteAppliedResource(ctx, &lar, manifests)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// UpdateStatus
	resourceControl.Status.AppliedResources = appliedResource
	resourceControl.Status.SchemaConfigMapRef = bdcFile.BDCTemplate.FullTemplate.XDefinitionSchemaName
	err = r.UpdateStatus(ctx, &resourceControl)
	if err != nil {
		err = condition.PatchCondition(ctx, r, &resourceControl,
			conditiontype.ReconcileError(fmt.Errorf(constants.ErrCreateBDCResource, resourceControl.Kind, resourceControl.Name, err)))
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	err = condition.PatchCondition(ctx, r, &resourceControl, conditiontype.ReconcileSuccess())
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) deleteAppliedResource(ctx context.Context, appliedResource *common.BDCObjectReference, manifests []*unstructured.Unstructured) error {

	isDeleted := true
	larObjRef := pkgcommon.ObjectReference{
		ObjectReference: corev1.ObjectReference{
			APIVersion: appliedResource.APIVersion,
			Kind:       appliedResource.Kind,
			Name:       appliedResource.Name,
			Namespace:  appliedResource.Namespace,
			UID:        appliedResource.UID,
		},
	}
	for _, mf := range manifests {
		mfObjRef := pkgcommon.ObjectReference{
			ObjectReference: corev1.ObjectReference{
				APIVersion: mf.GetAPIVersion(),
				Kind:       mf.GetKind(),
				Name:       mf.GetName(),
				Namespace:  mf.GetNamespace(),
				UID:        mf.GetUID(),
			},
		}
		if larObjRef.Equal(mfObjRef) {
			isDeleted = false
			break
		}

	}
	// 创建一个 unstructured.Unstructured 对象
	u := &unstructured.Unstructured{}
	u.SetKind(larObjRef.ObjectReference.Kind)
	u.SetName(larObjRef.ObjectReference.Name)
	u.SetNamespace(larObjRef.ObjectReference.Namespace)
	u.SetAPIVersion(larObjRef.ObjectReference.APIVersion)
	if isDeleted {
		if err := r.Delete(ctx, u); err != nil && !kerrors.IsNotFound(err) {
			return errors.Wrapf(err, "cannot delete manifest, namespace: %s name: %s apiVersion: %s kind: %s", u.GetNamespace(), u.GetName(), u.GetAPIVersion(), u.GetKind())
		}
		return nil
	}
	return nil

}

// UpdateStatus update Status with retry.RetryOnConflict
func (r *Reconciler) UpdateStatus(ctx context.Context, bdc *bdcv1alpha1.ResourceControl, opts ...client.SubResourceUpdateOption) error {
	status := bdc.DeepCopy().Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if err = r.Get(ctx, client.ObjectKey{Name: bdc.Name}, bdc); err != nil {
			return
		}
		bdc.Status = status
		return r.Status().Update(ctx, bdc, opts...)
	})
}

func (r *Reconciler) patchOwnerReferencer(ctx context.Context, bdc *bdcv1alpha1.ResourceControl) error {
	if err := r.Patch(ctx, bdc, client.Merge); err != nil {
		klog.Info(err, "unable to patch annotation")
	}
	klog.InfoS("patch", "Object", bdc.Name, "OwnerReferencer", bdc.OwnerReferences)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.concurrentReconciles,
		}).
		For(&bdcv1alpha1.ResourceControl{}).
		Complete(r)
}

func Setup(mgr ctrl.Manager, args bdcctrl.Args) error {
	r := Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("bdc-resource-control-controller"),
		options:  parseOptions(args),
	}
	return r.SetupWithManager(mgr)
}

func parseOptions(args bdcctrl.Args) options {
	return options{
		defRevLimit:          args.DefRevisionLimit,
		concurrentReconciles: args.ConcurrentReconciles,
		ignoreDefNoCtrlReq:   args.IgnoreDefinitionWithoutControllerRequirement,
		controllerVersion:    version.CoreVersion,
	}
}
