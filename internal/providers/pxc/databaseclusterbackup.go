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

type databaseClusterBackup struct {
	client.Client
}

var _ runtime.DatabaseClusterBackupController = (*databaseClusterBackup)(nil)

func (p *databaseClusterBackup) RegisterSources(b *builder.Builder) error {
	b.Owns(&pxcv1.PerconaXtraDBClusterBackup{})
	return nil
}

func (p *databaseClusterBackup) Ensure(ctx context.Context, db *v1alpha1.DatabaseClusterBackup) error {
	pxc := &pxcv1.PerconaXtraDBClusterBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.GetName(),
			Namespace: db.GetNamespace(),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, p.Client, pxc, func() error {
		// TODO: build a pxc-backup object from the DatabaseClusterBackup object
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (p *databaseClusterBackup) Observe(ctx context.Context, name, namespace string) (v1alpha1.DatabaseClusterBackupStatus, error) {
	pxc := &pxcv1.PerconaXtraDBClusterBackup{}
	if err := p.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, pxc); err != nil {
		return v1alpha1.DatabaseClusterBackupStatus{}, err
	}

	// TODO: build a DatabaseClusterStatus object from the PXC object state.
	return v1alpha1.DatabaseClusterBackupStatus{}, nil
}

func (p *databaseClusterBackup) HandleDelete(ctx context.Context, bkp *v1alpha1.DatabaseClusterBackup) (bool, error) {
	pxc := &pxcv1.PerconaXtraDBClusterBackup{}
	if err := p.Get(ctx, client.ObjectKey{Namespace: bkp.GetNamespace(), Name: bkp.GetName()}, pxc); err != nil {
		return false, client.IgnoreNotFound(err)
	}
	// TODO: handle finalizers in the pxc object.
	return true, nil
}
