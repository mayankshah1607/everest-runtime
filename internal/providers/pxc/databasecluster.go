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

type databaseCluster struct {
	client.Client
}

var _ controller.DatabaseClusterController = (*databaseCluster)(nil)

func (p *databaseCluster) GetSources(ctrl.Manager) ([]source.Source, error) {
	return []source.Source{}, nil
}

func (p *databaseCluster) Reconcile(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseCluster) error {
	return nil
}

func (p *databaseCluster) Observe(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseCluster) (everestv1alpha1.DatabaseClusterStatus, error) {
	return v1alpha1.DatabaseClusterStatus{}, nil
}

func (p *databaseCluster) HandleDelete(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseCluster) (bool, error) {
	return true, nil
}

func (p *databaseCluster) Restart(ctx context.Context, c client.Client, db *everestv1alpha1.DatabaseCluster) error {
	return nil
}
