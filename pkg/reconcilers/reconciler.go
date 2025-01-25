package reconcilers

import (
	"context"
	"fmt"

	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"github.com/mayankshah1607/everest-runtime/pkg/extension"
	"github.com/mayankshah1607/everest-runtime/pkg/reconcilers/databasecluster"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
	Manager ctrl.Manager

	DatabaseClusterController        controller.DatabaseClusterController
	DatabaseClusterBackupController  controller.DatabaseClusterBackupController
	DatabaseClusterRestoreController controller.DatabaseClusterRestoreController

	Extension *extension.Extension
}

func (r *Reconciler) Start(ctx context.Context) error {
	return r.Manager.Start(ctx)
}

func (c *Reconciler) initControllers(ctx context.Context) error {
	targetEngine := c.Extension.Engine
	if targetEngine == "" {
		return fmt.Errorf("extension.Name cannot be empty")
	}
	if c.DatabaseClusterController != nil {
		if err := (&databasecluster.Reconciler{
			Client:             c.Manager.GetClient(),
			Controller:         c.DatabaseClusterController,
			DatabaseEngineName: targetEngine,
		}).Setup(c.Manager); err != nil {
			return err
		}
	}

	// TODO: DBC and DBB controllers
	return nil
}
