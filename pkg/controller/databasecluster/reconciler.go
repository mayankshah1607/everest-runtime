package databasecluster

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/runtime"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	Client     client.Client
	Controller runtime.DatabaseClusterController
}

func (r *Reconciler) Setup(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		Named("DatabaseCluster").
		For(&everestv1alpha1.DatabaseCluster{})

	if err := r.Controller.RegisterSources(b); err != nil {
		return err
	}

	return b.Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	db := &everestv1alpha1.DatabaseCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, db); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Reconciling DatabaseCluster", "req", req)

	if !db.GetDeletionTimestamp().IsZero() {
		ok, err := r.Controller.HandleDelete(ctx, db.Name, db.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}
		result := ctrl.Result{}
		if !ok {
			result.Requeue = true
		}
		return result, nil
	}

	if err := r.Controller.Ensure(ctx, db); err != nil {
		return ctrl.Result{}, err
	}

	status, err := r.Controller.Observe(ctx, db.Name, db.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	db.Status = status
	if err := r.Client.Status().Update(ctx, db); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
