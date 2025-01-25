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

type databaseClusterBackup struct {
	client.Client
}

var _ controller.DatabaseClusterBackupController = (*databaseClusterBackup)(nil)

func (p *databaseClusterBackup) GetSources(ctrl.Manager) ([]source.Source, error) {
	return []source.Source{}, nil
}

func (p *databaseClusterBackup) Reconcile(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseClusterBackup) error {
	return nil
}

func (p *databaseClusterBackup) Observe(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseClusterBackup) (everestv1alpha1.DatabaseClusterBackupStatus, error) {
	return v1alpha1.DatabaseClusterBackupStatus{}, nil
}

func (p *databaseClusterBackup) HandleDelete(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseClusterBackup) (bool, error) {
	return true, nil
}
