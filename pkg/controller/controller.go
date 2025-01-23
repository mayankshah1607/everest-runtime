package controller

import (
	"context"
	"fmt"

	"github.com/mayankshah1607/everest-runtime/pkg/controller/databasecluster"
	"github.com/mayankshah1607/everest-runtime/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Controller struct {
	Manager ctrl.Manager

	DatabaseClusterController        runtime.DatabaseClusterController
	DatabaseClusterBackupController  runtime.DatabaseClusterBackupController
	DatabaseClusterRestoreController runtime.DatabaseClusterRestoreController
	DatabaseEngine                   runtime.DatabaseEngine
}

func (c *Controller) Start() error {
	if err := c.ensureDatabaseEngine(); err != nil {
		return err
	}
	if err := c.initControllers(context.Background()); err != nil {
		return err
	}
	return c.Manager.Start(ctrl.SetupSignalHandler())
}

func (c *Controller) initControllers(ctx context.Context) error {
	targetEngine := c.DatabaseEngine.GetName()
	if targetEngine == "" {
		return fmt.Errorf("result of DatabaseEngine.GetName() is empty")
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

func (c *Controller) ensureDatabaseEngine() error {
	// TODO: create a DatabaseEngine object here.
	return nil
}
