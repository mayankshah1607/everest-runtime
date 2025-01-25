package pxc

import (
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PXC struct {
	controller.DatabaseClusterController
	controller.DatabaseClusterBackupController
	controller.DatabaseClusterRestoreController
}

// New returns a new PXC provider.
func New(c client.Client) *PXC {
	return &PXC{
		DatabaseClusterController:        &databaseCluster{Client: c},
		DatabaseClusterBackupController:  &databaseClusterBackup{Client: c},
		DatabaseClusterRestoreController: &databaseClusterRestore{Client: c},
	}
}
