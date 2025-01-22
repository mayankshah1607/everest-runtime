package controller

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/controller/databasecluster"
	"github.com/mayankshah1607/everest-runtime/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Controller struct {
	Manager ctrl.Manager

	DatabaseClusterController        runtime.DatabaseClusterController
	DatabaseClusterBackupController  runtime.DatabaseClusterBackupController
	DatabaseClusterRestoreController runtime.DatabaseClusterRestoreController
}

func (c *Controller) Start() error {
	if err := c.initControllers(context.Background()); err != nil {
		return err
	}
	return c.Manager.Start(ctrl.SetupSignalHandler())
}

func (c *Controller) initControllers(ctx context.Context) error {
	if c.DatabaseClusterController != nil {
		if err := (&databasecluster.Reconciler{
			Client:     c.Manager.GetClient(),
			Controller: c.DatabaseClusterController,
		}).Setup(c.Manager); err != nil {
			return err
		}
	}

	// TODO: DBC and DBB controllers
	return nil
}
