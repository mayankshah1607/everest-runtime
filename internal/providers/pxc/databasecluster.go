package pxc

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/runtime"
	"github.com/percona/everest-operator/api/v1alpha1"
	pxcv1 "github.com/percona/percona-xtradb-cluster-operator/pkg/apis/pxc/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type databaseCluster struct {
	client.Client
}

var _ runtime.DatabaseClusterController = (*databaseCluster)(nil)

func (p *databaseCluster) RegisterSources(b *builder.Builder) error {
	b.Owns(&pxcv1.PerconaXtraDBCluster{})
	return nil
}

func (p *databaseCluster) Ensure(ctx context.Context, db *v1alpha1.DatabaseCluster) error {
	pxc := &pxcv1.PerconaXtraDBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.GetName(),
			Namespace: db.GetNamespace(),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, p.Client, pxc, func() error {
		// TODO: build a PXC object from the DatabaseCluster object
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (p *databaseCluster) Observe(ctx context.Context, name, namespace string) (v1alpha1.DatabaseClusterStatus, error) {
	pxc := &pxcv1.PerconaXtraDBCluster{}
	if err := p.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, pxc); err != nil {
		return v1alpha1.DatabaseClusterStatus{}, err
	}

	// TODO: build a DatabaseClusterStatus object from the PXC object state.
	return v1alpha1.DatabaseClusterStatus{}, nil
}

func (p *databaseCluster) HandleDelete(ctx context.Context, name, namespace string) (bool, error) {
	pxc := &pxcv1.PerconaXtraDBCluster{}
	if err := p.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, pxc); err != nil {
		return false, client.IgnoreNotFound(err)
	}
	// TODO: handle finalizers in the pxc object.
	return true, nil
}
