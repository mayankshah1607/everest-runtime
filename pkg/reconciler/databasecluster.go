package reconciler

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
	controller *controller.DatabaseClusterController
	manager    ctrl.Manager
	client.Client
}

func (r *Reconciler) GetManager() ctrl.Manager {
	return r.manager
}

// New returns a new Reconciler for DatabaseClusterController.
func New(c *controller.DatabaseClusterController) *Reconciler {
	managerOpts := ctrl.Options{}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), managerOpts)
	if err != nil {
		return nil
	}
	r := &Reconciler{
		controller: c,
		manager:    mgr,
		Client:     mgr.GetClient(),
	}
	if err := r.setup(); err != nil {
		return nil
	}
	return r
}

func (r *Reconciler) Start() error {
	return r.manager.Start(ctrl.SetupSignalHandler())
}

func (r *Reconciler) setup() error {
	// use this filter to filter out all DatabaseCluster objects that this reconciler should not handle
	filter := predicate.NewPredicateFuncs(func(object client.Object) bool {
		db, ok := object.(*v2alpha1.DatabaseCluster)
		if !ok {
			return false
		}
		labels := db.GetLabels()
		val, ok := labels["reconciler"]
		return ok && val == r.controller.GetEngineName()
	})

	b := ctrl.NewControllerManagedBy(r.manager).For(&v2alpha1.DatabaseCluster{}, builder.WithPredicates(filter))
	b.Named("DatabaseClusterController")

	// set watch on owned objects.
	for _, obj := range r.controller.GetWatches().GetOwned() {
		b.Owns(obj)
	}
	// set custom watches
	for _, watch := range r.controller.GetWatches().GetCustomWatches() {
		b.Watches(watch.GetObject(), handler.EnqueueRequestsFromMapFunc(watch.GetRequestMapper()))
	}
	return b.Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	db := &v2alpha1.DatabaseCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, db); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	plan := r.controller.GetPlan()

	// Handle cleanup
	if !db.GetDeletionTimestamp().IsZero() {
		for _, step := range plan.GetCleanup().GetSteps() {
			fn := step.Fn
			mustWait := step.Wait
			result, err := fn(ctx, r.Client, db)
			if err != nil {
				return reconcile.Result{}, err
			}
			if !result.Done && mustWait {
				return reconcile.Result{RequeueAfter: result.RequeueAfter}, nil
			}
		}
		return reconcile.Result{}, nil
	}

	// Handle reconciliation
	for _, step := range plan.GetSync().GetSteps() {
		fn := step.Fn
		mustWait := step.Wait
		result, err := fn(ctx, r.Client, db)
		if err != nil {
			return reconcile.Result{}, err
		}
		if !result.Done && mustWait {
			return reconcile.Result{RequeueAfter: result.RequeueAfter}, nil
		}
	}

	// Sync status
	stsFn := r.controller.GetPlan().GetStatusGetter()
	sts, err := stsFn(ctx, r.Client, db)
	if err != nil {
		return reconcile.Result{}, err
	}
	db.Status = sts
	if err := r.Client.Status().Update(ctx, db); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
