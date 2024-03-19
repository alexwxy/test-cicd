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

package orgresourcecontrol

import (
	bdcv1alpha1 "bpaas-core-operator/api/bdc/v1alpha1"
	bdcctrl "bpaas-core-operator/pkg/controllers/bdc"
	"bpaas-core-operator/pkg/controllers/bdc/constants"
	"bpaas-core-operator/pkg/utils"
	"bpaas-core-operator/version"
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler reconciles a OrgResourceControl object
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

//+kubebuilder:rbac:groups=bdc.bdos.io,resources=orgresourcecontrols,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bdc.bdos.io,resources=orgresourcecontrols/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bdc.bdos.io,resources=orgresourcecontrols/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OrgResourceControl object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.InfoS("Reconcile org-resource-quota", "", klog.KRef(req.Namespace, req.Name))

	// Lookup the resourceControl instance for this reconcile request
	var orgResourceControl bdcv1alpha1.OrgResourceControl
	if err := r.Get(ctx, req.NamespacedName, &orgResourceControl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List ResourceControl by label "bdc.bdos.io/org"
	var resourceControlList bdcv1alpha1.ResourceControlList
	var bdcOrgLabels client.MatchingLabels

	if orgResourceControl.GetLabels()[constants.LabelBDCOrgName] == "" {
		klog.InfoS("Not found", "", orgResourceControl.Name, "label", constants.LabelBDCOrgName)
		return ctrl.Result{}, nil
	}
	bdcOrgLabels = map[string]string{
		constants.LabelBDCOrgName: orgResourceControl.GetLabels()[constants.LabelBDCOrgName],
	}
	var listRCOpts = []client.ListOption{
		bdcOrgLabels,
	}

	if err := r.List(ctx, &resourceControlList, listRCOpts...); err != nil {
		klog.ErrorS(err, "", "")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// klog.InfoS("List", constants.LabelBDCOrgName, orgResourceControl.GetLabels()[constants.LabelBDCOrgName], "resourceControls", resourceControlList)
	var details []bdcv1alpha1.OrgResourceControlStatusDetail
	var resourceQuotaStatus corev1.ResourceQuotaStatus
	totalUsed := corev1.ResourceList{}
	totalHard := corev1.ResourceList{}
	for _, rcItem := range resourceControlList.Items {
		var singleBDCResourceControlStatus bdcv1alpha1.OrgResourceControlStatusDetail

		for _, rcar := range rcItem.Status.AppliedResources {
			if rcar.Kind == "ResourceQuota" {
				singleBDCResourceControlStatus.BDCName = rcar.BDCName

				unstructured, err := utils.RawExtension2Unstructured(rcar.Status.DeepCopy())
				if err != nil {
					return ctrl.Result{}, err
				}
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, &resourceQuotaStatus)
				if err != nil {
					return ctrl.Result{}, err
				}
				for resource, quantity := range resourceQuotaStatus.Used {
					r2 := totalUsed[resource]
					r2.Add(quantity)
					totalUsed[resource] = r2
				}

				for resource, quantity := range resourceQuotaStatus.Hard {
					r3 := totalHard[resource]
					r3.Add(quantity)
					totalHard[resource] = r3
				}

				singleBDCResourceControlStatus.Status = *utils.Object2RawExtension(resourceQuotaStatus)
				break
			}
		}
		details = append(details, singleBDCResourceControlStatus)
	}

	orgResourceControl.Status.Details = details
	orgResourceControl.Status.TotalUsed = totalUsed
	orgResourceControl.Status.TotalHard = totalHard

	// Calculate resource quota

	err := r.UpdateStatus(ctx, &orgResourceControl)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// UpdateStatus update Status with retry.RetryOnConflict
func (r *Reconciler) UpdateStatus(ctx context.Context, bdc *bdcv1alpha1.OrgResourceControl, opts ...client.SubResourceUpdateOption) error {
	status := bdc.DeepCopy().Status
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if err = r.Get(ctx, client.ObjectKey{Name: bdc.Name}, bdc); err != nil {
			return
		}
		bdc.Status = status
		return r.Status().Update(ctx, bdc, opts...)
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.concurrentReconciles,
		}).
		For(&bdcv1alpha1.OrgResourceControl{}).
		Complete(r)
}

func Setup(mgr ctrl.Manager, args bdcctrl.Args) error {
	r := Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("bdc-org-resource-control-controller"),
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
