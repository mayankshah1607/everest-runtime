package pxc

import (
	"github.com/mayankshah1607/everest-runtime/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PXC struct {
	runtime.DatabaseClusterController
	runtime.DatabaseClusterBackupController
	runtime.DatabaseClusterRestoreController
	runtime.DatabaseEngine
}

// New returns a new PXC provider.
func New(c client.Client) *PXC {
	return &PXC{
		DatabaseClusterController:        &databaseCluster{Client: c},
		DatabaseClusterBackupController:  &databaseClusterBackup{Client: c},
		DatabaseClusterRestoreController: &databaseClusterRestore{Client: c},
		DatabaseEngine:                   &engine{Client: c, name: "pxc"},
	}
}
