package main

import (
	"github.com/mayankshah1607/everest-runtime/internal/providers/pxc"
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme = runtime.NewScheme()
)

func main() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}
	pxcProv := pxc.New(mgr.GetClient())

	c := controller.Controller{
		Manager:                          mgr,
		DatabaseClusterController:        pxcProv.DatabaseClusterController,
		DatabaseClusterBackupController:  pxcProv.DatabaseClusterBackupController,
		DatabaseClusterRestoreController: pxcProv.DatabaseClusterRestoreController,
		DatabaseEngine:                   pxcProv.DatabaseEngine,
	}

	if err := c.Start(); err != nil {
		panic(err)
	}
}
