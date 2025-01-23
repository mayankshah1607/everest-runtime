package databasecluster

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/runtime"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	restartAnnotationKey = "everest.percona.com/restart"
)

type Reconciler struct {
	Client             client.Client
	Controller         runtime.DatabaseClusterController
	DatabaseEngineName string
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

	if string(db.Spec.Engine.Type) != r.DatabaseEngineName {
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling DatabaseCluster", "req", req)

	if !db.GetDeletionTimestamp().IsZero() {
		ok, err := r.Controller.HandleDelete(ctx, db)
		if err != nil {
			return ctrl.Result{}, err
		}
		result := ctrl.Result{}
		if !ok {
			result.Requeue = true
		}
		return result, nil
	}

	if err := r.handleRestart(ctx, db); err != nil {
		return ctrl.Result{}, err
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

func (r *Reconciler) handleRestart(ctx context.Context, db *everestv1alpha1.DatabaseCluster) error {
	annotations := db.GetAnnotations()
	if _, required := annotations[restartAnnotationKey]; !required {
		return nil
	}
	return r.Controller.HandleRestart(ctx, db)
}
