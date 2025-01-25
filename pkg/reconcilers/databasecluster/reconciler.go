package databasecluster

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// RestartAnnotationKey is the annotation key that triggers a restart of the database cluster.
	RestartAnnotationKey = "everest.percona.com/restart"
)

// Reconciler reconciles a DatabaseCluster.
type Reconciler struct {
	client.Client
	Controller         controller.DatabaseClusterController
	DatabaseEngineName string
}

func (r *Reconciler) Setup(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		Named("DatabaseCluster").
		For(&everestv1alpha1.DatabaseCluster{}).
		Build(r)
	if err != nil {
		return err
	}

	sources, err := r.Controller.GetSources(mgr)
	if err != nil {
		return err
	}

	for _, src := range sources {
		c.Watch(src)
	}
	return nil
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
		ok, err := r.Controller.HandleDelete(ctx, r, db)
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

	if err := r.Controller.Reconcile(ctx, r, db); err != nil {
		return ctrl.Result{}, err
	}

	status, err := r.Controller.Observe(ctx, r, db)
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
	if _, required := annotations[RestartAnnotationKey]; !required {
		return nil
	}
	return r.Controller.Restart(ctx, r, db)
}
