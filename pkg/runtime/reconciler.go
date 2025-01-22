package runtime

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

type Controller[T, Status any] interface {
	RegisterSources(*builder.Builder) error
	Ensure(context.Context, *T) error
	Observe(ctx context.Context, name, namespace string) (Status, error)
	HandleDelete(ctx context.Context, name, namespace string) (bool, error)
}

type (
	DatabaseClusterController        = Controller[everestv1alpha1.DatabaseCluster, everestv1alpha1.DatabaseClusterStatus]
	DatabaseClusterBackupController  = Controller[everestv1alpha1.DatabaseClusterBackup, everestv1alpha1.DatabaseClusterBackupStatus]
	DatabaseClusterRestoreController = Controller[everestv1alpha1.DatabaseClusterRestore, everestv1alpha1.DatabaseClusterRestoreStatus]
)
