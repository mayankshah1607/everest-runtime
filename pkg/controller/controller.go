package controller

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type DatabaseClusterController interface {
	// GetSources returns the event sources that the reconciler should watch.
	GetSources(ctrl.Manager) ([]source.Source, error)
	// Reconcile reconciles the provided DatabaseCluster object.
	Reconcile(context.Context, client.Client, *everestv1alpha1.DatabaseCluster) error
	// HandleDelete handles the deletion of the provided DatabaseCluster object.
	// This is useful to perform cleanup operations and resolving any finalizers that may be set.
	// Clients that don't require any cleanup can provide a no-op implementation.
	// Returns a boolean indicating if the object was deleted, and an error if any occurred.
	HandleDelete(context.Context, client.Client, *everestv1alpha1.DatabaseCluster) (bool, error)
	// Restart contains the logic for restarting the DatabaseCluster.
	// The implementation is responsible for tracking the state of the restart.
	Restart(context.Context, client.Client, *everestv1alpha1.DatabaseCluster) error
	// Observe the state of the DatabaseCluster and return a DatabaseClusterStatus object.
	Observe(context.Context, client.Client, *everestv1alpha1.DatabaseCluster) (everestv1alpha1.DatabaseClusterStatus, error)
}

type DatabaseClusterBackupController interface {
	// GetSources returns the event sources that the reconciler should watch.
	GetSources(ctrl.Manager) ([]source.Source, error)
	// Reconcile reconciles the provided DatabaseClusterBackup object.
	Reconcile(context.Context, client.Client, *everestv1alpha1.DatabaseClusterBackup) error
	// HandleDelete handles the deletion of the provided DatabaseClusterBackup object.
	// This is useful to perform cleanup operations and resolving any finalizers that may be set.
	// Clients that don't require any cleanup can provide a no-op implementation.
	// Returns a boolean indicating if the object was deleted, and an error if any occurred.
	HandleDelete(context.Context, client.Client, *everestv1alpha1.DatabaseClusterBackup) (bool, error)
	// Observe the state of the DatabaseCluster and return a DatabaseClusterStatus object.
	Observe(context.Context, client.Client, *everestv1alpha1.DatabaseClusterBackup) (everestv1alpha1.DatabaseClusterBackupStatus, error)
}

type DatabaseClusterRestoreController interface {
	// GetSources returns the event sources that the reconciler should watch.
	GetSources(ctrl.Manager) ([]source.Source, error)
	// Reconcile reconciles the provided DatabaseClusterRestore object.
	Reconcile(context.Context, client.Client, *everestv1alpha1.DatabaseClusterRestore) error
	// HandleDelete handles the deletion of the provided DatabaseClusterRestore object.
	// This is useful to perform cleanup operations and resolving any finalizers that may be set.
	// Clients that don't require any cleanup can provide a no-op implementation.
	// Returns a boolean indicating if the object was deleted, and an error if any occurred.
	HandleDelete(context.Context, client.Client, *everestv1alpha1.DatabaseClusterRestore) (bool, error)
	// Observe the state of the DatabaseCluster and return a DatabaseClusterStatus object.
	Observe(context.Context, client.Client, *everestv1alpha1.DatabaseClusterRestore) (everestv1alpha1.DatabaseClusterRestoreStatus, error)
}
