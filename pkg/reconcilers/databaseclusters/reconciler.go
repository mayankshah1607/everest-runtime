package databaseclusters

import (
	"context"
	"fmt"

	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type Reconciler struct {
	client.Client
	Controller controller.DatabaseClusterController
	pluginName string
}

func newDatabaseClusterPredicates(t string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		db, ok := object.(*v2alpha1.DatabaseCluster)
		if !ok {
			return false
		}
		return db.Spec.Plugin == t
	})
}

func (r *Reconciler) Setup(mgr manager.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		Watches(
			&v2alpha1.DatabaseCluster{},
			&handler.EnqueueRequestForObject{},
			// We will filter out objects based on the Plugin name,
			// but atm we don't have the Plugin CRD, so we will just handle everything.
			// builder.WithPredicates(newDatabaseClusterPredicates(r.pluginName)),
		).
		Named("DatabaseCluster").
		Build(r)
	if err != nil {
		return err
	}

	for _, src := range r.Controller.GetSources(mgr) {
		if err := c.Watch(src); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	db := &v2alpha1.DatabaseCluster{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, db); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Reconciling DatabaseCluster", "namespace", db.Namespace, "name", db.Name)

	if !db.GetDeletionTimestamp().IsZero() {
		done, err := r.Controller.Delete(ctx, r, db)
		if err != nil {
			log.Error(err, "Delete failed")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: !done}, nil
	}

	// aggregate the pod details including defaults from the DatabaseClusterDefinition
	// and set the internal field.
	if err := r.attachPodInfo(ctx, db); err != nil {
		log.Error(err, "attachPodInfo failed")
		return ctrl.Result{}, err
	}

	rr, err := r.Controller.Reconcile(ctx, r, db)
	if err != nil {
		log.Error(err, "Reconcile failed")
		return ctrl.Result{}, err
	}

	st, err := r.Controller.GetStatus(ctx, r, db)
	if err != nil {
		log.Error(err, "GetStatus failed")
		return ctrl.Result{}, err
	}

	db.Status = st
	if err := r.Status().Update(ctx, db); err != nil {
		log.Error(err, "Status update failed")
		return ctrl.Result{}, err
	}
	return rr, nil
}

// We don't have a mechanism to find the DBDefinition for the DatabaseCluster.
// The Plugin CRD will manage this for us, but we don't have it yet, so we shall
// just hard-code references for now.
const dbcDefRef = "clickhouse-definition"

func (r *Reconciler) attachPodInfo(ctx context.Context, db *v2alpha1.DatabaseCluster) error {
	def := &v2alpha1.DatabaseClusterDefinition{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: db.Namespace,
		Name:      dbcDefRef,
	}, def); err != nil {
		return err
	}

	for i, cmp := range db.Spec.Components {
		cmpDef, ok := def.Spec.Definitions.Components[cmp.Type]
		if !ok {
			return fmt.Errorf("component definition not found for %s", cmp.Type)
		}
		db.Spec.Components[i].PodSpec = cmpDef.Defaults
		// TODO: once we have ComponentVersions, we will set the Image from there.
	}
	return nil
}
