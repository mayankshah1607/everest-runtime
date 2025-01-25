package pxc

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"github.com/percona/everest-operator/api/v1alpha1"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type databaseClusterRestore struct {
	client.Client
}

var _ controller.DatabaseClusterRestoreController = (*databaseClusterRestore)(nil)

func (p *databaseClusterRestore) GetSources(ctrl.Manager) ([]source.Source, error) {
	return []source.Source{}, nil
}

func (p *databaseClusterRestore) Reconcile(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseClusterRestore) error {
	return nil
}

func (p *databaseClusterRestore) Observe(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseClusterRestore) (everestv1alpha1.DatabaseClusterRestoreStatus, error) {
	return v1alpha1.DatabaseClusterRestoreStatus{}, nil
}

func (p *databaseClusterRestore) HandleDelete(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseClusterRestore) (bool, error) {
	return true, nil
}
